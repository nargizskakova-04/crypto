package repository

import (
	"context"
	"crypto/internal/config"
	"crypto/internal/core/domain"
	"crypto/internal/core/port"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/lib/pq"
)

// PricesRepository implements the PostgreSQL repository for market data
type PricesRepository struct {
	db *sql.DB
}

// NewPricesRepository creates a new PostgreSQL repository instance
func NewPricesRepository(db *sql.DB) port.PricesRepository {
	return &PricesRepository{
		db: db,
	}
}

// NewDbConnInstance creates a new database connection
func NewDbConnInstance(cfg *config.Database) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.Username,
		cfg.Password,
		cfg.Name,
		cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxConnections)
	db.SetMaxIdleConns(cfg.MaxIdleConnections)
	db.SetConnMaxLifetime(30 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Info("PostgreSQL connection established successfully")
	return db, nil
}

// GetLatestPrice returns the most recent price for a symbol across all exchanges
func (r *PricesRepository) GetLatestPrice(ctx context.Context, symbol string) (*domain.PriceResponse, error) {
	query := `
		SELECT pair_name, exchange, timestamp, average_price 
		FROM prices 
		WHERE pair_name = $1 
		ORDER BY timestamp DESC 
		LIMIT 1
	`

	var response domain.PriceResponse
	err := r.db.QueryRowContext(ctx, query, symbol).Scan(
		&response.Symbol,
		&response.Exchange,
		&response.Timestamp,
		&response.Price,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrPriceNotFound
		}
		return nil, fmt.Errorf("failed to get latest price: %w", err)
	}

	return &response, nil
}

// GetLatestExchangePrice returns the most recent price for a symbol from a specific exchange
func (r *PricesRepository) GetLatestExchangePrice(ctx context.Context, symbol, exchange string) (*domain.PriceResponse, error) {
	query := `
		SELECT pair_name, exchange, timestamp, average_price 
		FROM prices 
		WHERE pair_name = $1 AND exchange = $2 
		ORDER BY timestamp DESC 
		LIMIT 1
	`

	var response domain.PriceResponse
	err := r.db.QueryRowContext(ctx, query, symbol, exchange).Scan(
		&response.Symbol,
		&response.Exchange,
		&response.Timestamp,
		&response.Price,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrPriceNotFound
		}
		return nil, fmt.Errorf("failed to get latest exchange price: %w", err)
	}

	return &response, nil
}

// GetHighestPrice returns the highest price for a symbol across all exchanges
func (r *PricesRepository) GetHighestPrice(ctx context.Context, symbol string) (*domain.PriceResponse, error) {
	query := `
		SELECT pair_name, exchange, timestamp, max_price 
		FROM prices 
		WHERE pair_name = $1 
		ORDER BY max_price DESC, timestamp DESC 
		LIMIT 1
	`

	var response domain.PriceResponse
	err := r.db.QueryRowContext(ctx, query, symbol).Scan(
		&response.Symbol,
		&response.Exchange,
		&response.Timestamp,
		&response.Price,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrPriceNotFound
		}
		return nil, fmt.Errorf("failed to get highest price: %w", err)
	}

	return &response, nil
}

// GetHighestPriceExchange returns the highest price for a symbol from a specific exchange
func (r *PricesRepository) GetHighestPriceExchange(ctx context.Context, symbol, exchange string) (*domain.PriceResponse, error) {
	query := `
		SELECT pair_name, exchange, timestamp, max_price 
		FROM prices 
		WHERE pair_name = $1 AND exchange = $2 
		ORDER BY max_price DESC, timestamp DESC 
		LIMIT 1
	`

	var response domain.PriceResponse
	err := r.db.QueryRowContext(ctx, query, symbol, exchange).Scan(
		&response.Symbol,
		&response.Exchange,
		&response.Timestamp,
		&response.Price,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrPriceNotFound
		}
		return nil, fmt.Errorf("failed to get highest price for exchange: %w", err)
	}

	return &response, nil
}

// GetHighestPriceInDuration returns the highest price for a symbol within a time range
func (r *PricesRepository) GetHighestPriceInDuration(ctx context.Context, symbol string, from, to time.Time) (*domain.PriceResponse, error) {
	query := `
		SELECT pair_name, exchange, timestamp, max_price 
		FROM prices 
		WHERE pair_name = $1 AND timestamp BETWEEN $2 AND $3 
		ORDER BY max_price DESC, timestamp DESC 
		LIMIT 1
	`

	var response domain.PriceResponse
	err := r.db.QueryRowContext(ctx, query, symbol, from, to).Scan(
		&response.Symbol,
		&response.Exchange,
		&response.Timestamp,
		&response.Price,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrPriceNotFound
		}
		return nil, fmt.Errorf("failed to get highest price in duration: %w", err)
	}

	return &response, nil
}

// GetHighestPriceInDurationExchange returns the highest price for a symbol from a specific exchange within a time range
func (r *PricesRepository) GetHighestPriceInDurationExchange(ctx context.Context, symbol, exchange string, from, to time.Time) (*domain.PriceResponse, error) {
	query := `
		SELECT pair_name, exchange, timestamp, max_price 
		FROM prices 
		WHERE pair_name = $1 AND exchange = $2 AND timestamp BETWEEN $3 AND $4 
		ORDER BY max_price DESC, timestamp DESC 
		LIMIT 1
	`

	var response domain.PriceResponse
	err := r.db.QueryRowContext(ctx, query, symbol, exchange, from, to).Scan(
		&response.Symbol,
		&response.Exchange,
		&response.Timestamp,
		&response.Price,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrPriceNotFound
		}
		return nil, fmt.Errorf("failed to get highest price in duration for exchange: %w", err)
	}

	return &response, nil
}

// GetLowestPrice returns the lowest price for a symbol across all exchanges
func (r *PricesRepository) GetLowestPrice(ctx context.Context, symbol string) (*domain.PriceResponse, error) {
	query := `
		SELECT pair_name, exchange, timestamp, min_price 
		FROM prices 
		WHERE pair_name = $1 
		ORDER BY min_price ASC, timestamp DESC 
		LIMIT 1
	`

	var response domain.PriceResponse
	err := r.db.QueryRowContext(ctx, query, symbol).Scan(
		&response.Symbol,
		&response.Exchange,
		&response.Timestamp,
		&response.Price,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrPriceNotFound
		}
		return nil, fmt.Errorf("failed to get lowest price: %w", err)
	}

	return &response, nil
}

// GetLowestPriceExchange returns the lowest price for a symbol from a specific exchange
func (r *PricesRepository) GetLowestPriceExchange(ctx context.Context, symbol, exchange string) (*domain.PriceResponse, error) {
	query := `
		SELECT pair_name, exchange, timestamp, min_price 
		FROM prices 
		WHERE pair_name = $1 AND exchange = $2 
		ORDER BY min_price ASC, timestamp DESC 
		LIMIT 1
	`

	var response domain.PriceResponse
	err := r.db.QueryRowContext(ctx, query, symbol, exchange).Scan(
		&response.Symbol,
		&response.Exchange,
		&response.Timestamp,
		&response.Price,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrPriceNotFound
		}
		return nil, fmt.Errorf("failed to get lowest price for exchange: %w", err)
	}

	return &response, nil
}

// GetLowestPriceInDuration returns the lowest price for a symbol within a time range
func (r *PricesRepository) GetLowestPriceInDuration(ctx context.Context, symbol string, from, to time.Time) (*domain.PriceResponse, error) {
	query := `
		SELECT pair_name, exchange, timestamp, min_price 
		FROM prices 
		WHERE pair_name = $1 AND timestamp BETWEEN $2 AND $3 
		ORDER BY min_price ASC, timestamp DESC 
		LIMIT 1
	`

	var response domain.PriceResponse
	err := r.db.QueryRowContext(ctx, query, symbol, from, to).Scan(
		&response.Symbol,
		&response.Exchange,
		&response.Timestamp,
		&response.Price,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrPriceNotFound
		}
		return nil, fmt.Errorf("failed to get lowest price in duration: %w", err)
	}

	return &response, nil
}

// GetLowestPriceInDurationExchange returns the lowest price for a symbol from a specific exchange within a time range
func (r *PricesRepository) GetLowestPriceInDurationExchange(ctx context.Context, symbol, exchange string, from, to time.Time) (*domain.PriceResponse, error) {
	query := `
		SELECT pair_name, exchange, timestamp, min_price 
		FROM prices 
		WHERE pair_name = $1 AND exchange = $2 AND timestamp BETWEEN $3 AND $4 
		ORDER BY min_price ASC, timestamp DESC 
		LIMIT 1
	`

	var response domain.PriceResponse
	err := r.db.QueryRowContext(ctx, query, symbol, exchange, from, to).Scan(
		&response.Symbol,
		&response.Exchange,
		&response.Timestamp,
		&response.Price,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrPriceNotFound
		}
		return nil, fmt.Errorf("failed to get lowest price in duration for exchange: %w", err)
	}

	return &response, nil
}

// GetAveragePrice returns the average price for a symbol across all exchanges
func (r *PricesRepository) GetAveragePrice(ctx context.Context, symbol string) (*domain.PriceResponse, error) {
	query := `
		SELECT pair_name, exchange, timestamp, average_price 
		FROM prices 
		WHERE pair_name = $1 
		ORDER BY timestamp DESC 
		LIMIT 1
	`

	var response domain.PriceResponse
	err := r.db.QueryRowContext(ctx, query, symbol).Scan(
		&response.Symbol,
		&response.Exchange,
		&response.Timestamp,
		&response.Price,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrPriceNotFound
		}
		return nil, fmt.Errorf("failed to get average price: %w", err)
	}

	return &response, nil
}

// GetAveragePriceExchange returns the average price for a symbol from a specific exchange
func (r *PricesRepository) GetAveragePriceExchange(ctx context.Context, symbol, exchange string) (*domain.PriceResponse, error) {
	query := `
		SELECT pair_name, exchange, timestamp, average_price 
		FROM prices 
		WHERE pair_name = $1 AND exchange = $2 
		ORDER BY timestamp DESC 
		LIMIT 1
	`

	var response domain.PriceResponse
	err := r.db.QueryRowContext(ctx, query, symbol, exchange).Scan(
		&response.Symbol,
		&response.Exchange,
		&response.Timestamp,
		&response.Price,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrPriceNotFound
		}
		return nil, fmt.Errorf("failed to get average price for exchange: %w", err)
	}

	return &response, nil
}

// GetAveragePriceInDuration returns the average price for a symbol within a time range
func (r *PricesRepository) GetAveragePriceInDuration(ctx context.Context, symbol string, from, to time.Time) (*domain.PriceResponse, error) {
	query := `
		SELECT $1 as pair_name, '' as exchange, $3 as timestamp, AVG(average_price) as average_price
		FROM prices 
		WHERE pair_name = $1 AND timestamp BETWEEN $2 AND $3
	`

	var response domain.PriceResponse
	err := r.db.QueryRowContext(ctx, query, symbol, from, to).Scan(
		&response.Symbol,
		&response.Exchange,
		&response.Timestamp,
		&response.Price,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrPriceNotFound
		}
		return nil, fmt.Errorf("failed to get average price in duration: %w", err)
	}

	return &response, nil
}

// GetAveragePriceInDurationExchange returns the average price for a symbol from a specific exchange within a time range
func (r *PricesRepository) GetAveragePriceInDurationExchange(ctx context.Context, symbol, exchange string, from, to time.Time) (*domain.PriceResponse, error) {
	query := `
		SELECT $1 as pair_name, $2 as exchange, $4 as timestamp, AVG(average_price) as average_price
		FROM prices 
		WHERE pair_name = $1 AND exchange = $2 AND timestamp BETWEEN $3 AND $4
	`

	var response domain.PriceResponse
	err := r.db.QueryRowContext(ctx, query, symbol, exchange, from, to).Scan(
		&response.Symbol,
		&response.Exchange,
		&response.Timestamp,
		&response.Price,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrPriceNotFound
		}
		return nil, fmt.Errorf("failed to get average price in duration for exchange: %w", err)
	}

	return &response, nil
}

// SaveAggregatedData saves aggregated data in batch
func (r *PricesRepository) SaveAggregatedData(ctx context.Context, data []domain.AggregatedData) error {
	if len(data) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO prices (pair_name, exchange, timestamp, average_price, min_price, max_price) 
		VALUES ($1, $2, $3, $4, $5, $6)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, item := range data {
		if err := item.Validate(); err != nil {
			slog.Warn("Skipping invalid aggregated data", "error", err)
			continue
		}

		_, err = stmt.ExecContext(ctx,
			item.PairName,
			item.Exchange,
			item.Timestamp,
			item.AveragePrice,
			item.MinPrice,
			item.MaxPrice,
		)
		if err != nil {
			return fmt.Errorf("failed to insert aggregated data: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	slog.Info("Successfully saved aggregated data", "count", len(data))
	return nil
}

// GetPriceStatistics returns price statistics for a symbol and exchange within a time range
func (r *PricesRepository) GetPriceStatistics(ctx context.Context, symbol, exchange string, from, to time.Time) (*domain.PriceStatistics, error) {
	query := `
		SELECT 
			COUNT(*) as count,
			AVG(average_price) as avg_price,
			MIN(min_price) as min_price,
			MAX(max_price) as max_price,
			MIN(timestamp) as start_time,
			MAX(timestamp) as end_time
		FROM prices 
		WHERE pair_name = $1 AND exchange = $2 AND timestamp BETWEEN $3 AND $4
	`

	var stats domain.PriceStatistics
	var count int
	var startTime, endTime time.Time

	err := r.db.QueryRowContext(ctx, query, symbol, exchange, from, to).Scan(
		&count,
		&stats.AveragePrice,
		&stats.MinPrice,
		&stats.MaxPrice,
		&startTime,
		&endTime,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrDataNotAvailable
		}
		return nil, fmt.Errorf("failed to get price statistics: %w", err)
	}

	stats.Symbol = symbol
	stats.Exchange = exchange
	stats.Count = count
	stats.StartTime = startTime
	stats.EndTime = endTime
	stats.Period = to.Sub(from).String()

	return &stats, nil
}

// HealthCheck performs a health check on the database connection
func (r *PricesRepository) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := r.db.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	// Test a simple query
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM prices LIMIT 1").Scan(&count)
	if err != nil {
		return fmt.Errorf("database query test failed: %w", err)
	}

	return nil
}
