package redis

import (
	"context"
	"crypto/internal/config"
	"crypto/internal/core/domain"
	"crypto/internal/core/port"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheRepository implements Redis-based caching for market data
type CacheRepository struct {
	client *redis.Client
	ttl    time.Duration
}

// NewCacheRepository creates a new Redis cache repository
func NewCacheRepository(cfg *config.Redis) (port.CacheRepository, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Parse TTL
	ttl, err := time.ParseDuration(cfg.TTL)
	if err != nil {
		ttl = 60 * time.Second // default TTL
	}

	slog.Info("Redis connection established successfully", "addr", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))

	return &CacheRepository{
		client: rdb,
		ttl:    ttl,
	}, nil
}

// SetLatestPrice stores the latest price update in Redis
func (r *CacheRepository) SetLatestPrice(ctx context.Context, update domain.PriceUpdate) error {
	// Store latest price with key format: "latest_price:SYMBOL:EXCHANGE"
	latestKey := fmt.Sprintf("%s%s:%s", domain.RedisLatestPricePrefix, update.Symbol, update.Exchange)

	// Store general latest price with key format: "latest_price:SYMBOL"
	generalKey := fmt.Sprintf("%s%s", domain.RedisLatestPricePrefix, update.Symbol)

	priceData := domain.PriceResponse{
		Symbol:    update.Symbol,
		Exchange:  update.Exchange,
		Price:     update.Price,
		Timestamp: update.Timestamp,
	}

	data, err := json.Marshal(priceData)
	if err != nil {
		return fmt.Errorf("failed to marshal price data: %w", err)
	}

	pipe := r.client.Pipeline()

	// Set latest price for specific exchange
	pipe.Set(ctx, latestKey, data, r.ttl)

	// Set general latest price (could be from any exchange)
	pipe.Set(ctx, generalKey, data, r.ttl)

	// Store in price history for aggregation (sorted set with timestamp as score)
	historyKey := fmt.Sprintf("%s%s:%s", domain.RedisPriceHistoryPrefix, update.Symbol, update.Exchange)
	pipe.ZAdd(ctx, historyKey, redis.Z{
		Score:  float64(update.Timestamp.Unix()),
		Member: data,
	})

	// Set TTL for history key
	pipe.Expire(ctx, historyKey, r.ttl)

	// Clean old entries (keep only last 2 minutes)
	cutoff := update.Timestamp.Add(-2 * time.Minute).Unix()
	pipe.ZRemRangeByScore(ctx, historyKey, "0", strconv.FormatInt(cutoff, 10))

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute Redis pipeline: %w", err)
	}

	slog.Debug("Stored price in Redis",
		"symbol", update.Symbol,
		"exchange", update.Exchange,
		"price", update.Price)

	return nil
}

// GetLatestPrice retrieves the latest price for a symbol (from any exchange)
func (r *CacheRepository) GetLatestPrice(ctx context.Context, symbol string) (*domain.PriceResponse, error) {
	key := fmt.Sprintf("%s%s", domain.RedisLatestPricePrefix, symbol)

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, domain.ErrPriceNotFound
		}
		return nil, fmt.Errorf("failed to get latest price from Redis: %w", err)
	}

	var priceResponse domain.PriceResponse
	if err := json.Unmarshal([]byte(data), &priceResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal price data: %w", err)
	}

	return &priceResponse, nil
}

// GetLatestExchangePrice retrieves the latest price for a symbol from a specific exchange
func (r *CacheRepository) GetLatestExchangePrice(ctx context.Context, symbol, exchange string) (*domain.PriceResponse, error) {
	key := fmt.Sprintf("%s%s:%s", domain.RedisLatestPricePrefix, symbol, exchange)

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, domain.ErrPriceNotFound
		}
		return nil, fmt.Errorf("failed to get latest exchange price from Redis: %w", err)
	}

	var priceResponse domain.PriceResponse
	if err := json.Unmarshal([]byte(data), &priceResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal price data: %w", err)
	}

	return &priceResponse, nil
}

// GetPricesInTimeRange retrieves price updates for a symbol and exchange within a time range
func (r *CacheRepository) GetPricesInTimeRange(ctx context.Context, symbol, exchange string, from, to time.Time) ([]domain.PriceUpdate, error) {
	key := fmt.Sprintf("%s%s:%s", domain.RedisPriceHistoryPrefix, symbol, exchange)

	// Use ZRANGEBYSCORE to get prices within time range
	fromScore := strconv.FormatInt(from.Unix(), 10)
	toScore := strconv.FormatInt(to.Unix(), 10)

	results, err := r.client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: fromScore,
		Max: toScore,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get prices in time range from Redis: %w", err)
	}

	var updates []domain.PriceUpdate
	for _, result := range results {
		var priceResponse domain.PriceResponse
		if err := json.Unmarshal([]byte(result), &priceResponse); err != nil {
			slog.Warn("Failed to unmarshal price data from Redis", "error", err)
			continue
		}

		update := domain.PriceUpdate{
			Exchange:  priceResponse.Exchange,
			Symbol:    priceResponse.Symbol,
			Price:     priceResponse.Price,
			Timestamp: priceResponse.Timestamp,
		}
		updates = append(updates, update)
	}

	return updates, nil
}

// CleanupOldPrices removes price data older than the specified time
func (r *CacheRepository) CleanupOldPrices(ctx context.Context, olderThan time.Time) error {
	// Get all price history keys
	pattern := fmt.Sprintf("%s*", domain.RedisPriceHistoryPrefix)
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get price history keys: %w", err)
	}

	if len(keys) == 0 {
		return nil
	}

	pipe := r.client.Pipeline()
	cutoffScore := strconv.FormatInt(olderThan.Unix(), 10)

	// Remove old entries from each price history
	for _, key := range keys {
		pipe.ZRemRangeByScore(ctx, key, "0", cutoffScore)
	}

	// Also clean up latest price keys that might be expired
	latestPattern := fmt.Sprintf("%s*", domain.RedisLatestPricePrefix)
	latestKeys, err := r.client.Keys(ctx, latestPattern).Result()
	if err == nil {
		for _, key := range latestKeys {
			// Check if key exists and is expired
			ttl := r.client.TTL(ctx, key)
			if ttl.Val() < 0 {
				pipe.Del(ctx, key)
			}
		}
	}

	results, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to cleanup old prices: %w", err)
	}

	cleanedCount := 0
	for _, result := range results {
		if cmd, ok := result.(*redis.IntCmd); ok {
			cleanedCount += int(cmd.Val())
		}
	}

	if cleanedCount > 0 {
		slog.Info("Cleaned up old prices from Redis", "entries_removed", cleanedCount)
	}

	return nil
}

// SetTTL sets the TTL for a specific key
func (r *CacheRepository) SetTTL(ctx context.Context, key string, ttl time.Duration) error {
	err := r.client.Expire(ctx, key, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set TTL for key %s: %w", key, err)
	}
	return nil
}

// HealthCheck performs a health check on the Redis connection
func (r *CacheRepository) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Test basic connectivity
	if err := r.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis ping failed: %w", err)
	}

	// Test write and read
	testKey := fmt.Sprintf("%shealth_check", domain.RedisKeyPrefix)
	testValue := fmt.Sprintf("health_check_%d", time.Now().Unix())

	if err := r.client.Set(ctx, testKey, testValue, 10*time.Second).Err(); err != nil {
		return fmt.Errorf("Redis write test failed: %w", err)
	}

	result, err := r.client.Get(ctx, testKey).Result()
	if err != nil {
		return fmt.Errorf("Redis read test failed: %w", err)
	}

	if result != testValue {
		return fmt.Errorf("Redis read/write test failed: expected %s, got %s", testValue, result)
	}

	// Clean up test key
	r.client.Del(ctx, testKey)

	return nil
}

// GetCacheStats returns statistics about the cache
func (r *CacheRepository) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	info, err := r.client.Info(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis info: %w", err)
	}

	stats := make(map[string]interface{})

	// Parse basic stats from info string
	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if strings.Contains(line, ":") && !strings.HasPrefix(line, "#") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// Try to convert to number
				if intVal, err := strconv.Atoi(value); err == nil {
					stats[key] = intVal
				} else if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
					stats[key] = floatVal
				} else {
					stats[key] = value
				}
			}
		}
	}

	// Add custom stats
	latestPricePattern := fmt.Sprintf("%s*", domain.RedisLatestPricePrefix)
	latestPriceKeys, _ := r.client.Keys(ctx, latestPricePattern).Result()
	stats["latest_price_keys"] = len(latestPriceKeys)

	historyPattern := fmt.Sprintf("%s*", domain.RedisPriceHistoryPrefix)
	historyKeys, _ := r.client.Keys(ctx, historyPattern).Result()
	stats["price_history_keys"] = len(historyKeys)

	return stats, nil
}

// FlushPriceData removes all price-related data (for testing/maintenance)
func (r *CacheRepository) FlushPriceData(ctx context.Context) error {
	// Get all marketflow keys
	pattern := fmt.Sprintf("%s*", domain.RedisKeyPrefix)
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get marketflow keys: %w", err)
	}

	if len(keys) == 0 {
		return nil
	}

	// Delete all keys in batches
	batchSize := 100
	for i := 0; i < len(keys); i += batchSize {
		end := i + batchSize
		if end > len(keys) {
			end = len(keys)
		}

		batch := keys[i:end]
		if err := r.client.Del(ctx, batch...).Err(); err != nil {
			return fmt.Errorf("failed to delete keys batch: %w", err)
		}
	}

	slog.Info("Flushed all price data from Redis", "keys_deleted", len(keys))
	return nil
}

// GetPricesCount returns the number of price updates for a symbol and exchange
func (r *CacheRepository) GetPricesCount(ctx context.Context, symbol, exchange string) (int64, error) {
	key := fmt.Sprintf("%s%s:%s", domain.RedisPriceHistoryPrefix, symbol, exchange)

	count, err := r.client.ZCard(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get prices count: %w", err)
	}

	return count, nil
}

// GetLatestPricesForSymbol returns latest prices from all exchanges for a symbol
func (r *CacheRepository) GetLatestPricesForSymbol(ctx context.Context, symbol string) (map[string]*domain.PriceResponse, error) {
	// Search for all keys matching the pattern
	pattern := fmt.Sprintf("%s%s:*", domain.RedisLatestPricePrefix, symbol)
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys for symbol: %w", err)
	}

	if len(keys) == 0 {
		return nil, domain.ErrPriceNotFound
	}

	// Get all values in one operation
	values, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get multiple prices: %w", err)
	}

	result := make(map[string]*domain.PriceResponse)

	for i, value := range values {
		if value == nil {
			continue
		}

		var priceResponse domain.PriceResponse
		if err := json.Unmarshal([]byte(value.(string)), &priceResponse); err != nil {
			slog.Warn("Failed to unmarshal price data", "key", keys[i], "error", err)
			continue
		}

		result[priceResponse.Exchange] = &priceResponse
	}

	if len(result) == 0 {
		return nil, domain.ErrPriceNotFound
	}

	return result, nil
}
