package marketsrv

import (
	"context"
	"crypto/internal/core/domain"
	"crypto/internal/core/port"
	"log/slog"
	"time"
)

// PriceService implements the business logic for price operations
type PriceService struct {
	priceRepository port.PricesRepository
	cacheRepository port.CacheRepository
	validator       *domain.PriceValidator
}

// NewPriceService creates a new price service instance
func NewPriceService(priceRepo port.PricesRepository, cacheRepo port.CacheRepository) port.PriceService {
	return &PriceService{
		priceRepository: priceRepo,
		cacheRepository: cacheRepo,
		validator:       domain.NewPriceValidator(),
	}
}

// GetLatestPrice returns the latest price for a symbol from any exchange
func (s *PriceService) GetLatestPrice(ctx context.Context, symbol string) (*domain.PriceResponse, error) {
	// Validate symbol
	if !domain.IsValidSymbol(symbol) {
		return nil, domain.ErrSymbolNotSupported(symbol)
	}

	// Try cache first (Redis)
	if s.cacheRepository != nil {
		if price, err := s.cacheRepository.GetLatestPrice(ctx, symbol); err == nil {
			slog.Debug("Latest price retrieved from cache", "symbol", symbol)
			return price, nil
		} else if err != domain.ErrPriceNotFound {
			slog.Warn("Cache lookup failed, falling back to database", "symbol", symbol, "error", err)
		}
	}

	// Fallback to database
	price, err := s.priceRepository.GetLatestPrice(ctx, symbol)
	if err != nil {
		slog.Error("Failed to get latest price", "symbol", symbol, "error", err)
		return nil, err
	}

	slog.Debug("Latest price retrieved from database", "symbol", symbol)
	return price, nil
}

// GetLatestExchangePrice returns the latest price for a symbol from a specific exchange
func (s *PriceService) GetLatestExchangePrice(ctx context.Context, symbol, exchange string) (*domain.PriceResponse, error) {
	// Validate inputs
	if !domain.IsValidSymbol(symbol) {
		return nil, domain.ErrSymbolNotSupported(symbol)
	}
	if exchange == "" {
		return nil, domain.ErrEmptyExchange
	}

	// Try cache first (Redis)
	if s.cacheRepository != nil {
		if price, err := s.cacheRepository.GetLatestExchangePrice(ctx, symbol, exchange); err == nil {
			slog.Debug("Latest exchange price retrieved from cache", "symbol", symbol, "exchange", exchange)
			return price, nil
		} else if err != domain.ErrPriceNotFound {
			slog.Warn("Cache lookup failed, falling back to database", "symbol", symbol, "exchange", exchange, "error", err)
		}
	}

	// Fallback to database
	price, err := s.priceRepository.GetLatestExchangePrice(ctx, symbol, exchange)
	if err != nil {
		slog.Error("Failed to get latest exchange price", "symbol", symbol, "exchange", exchange, "error", err)
		return nil, err
	}

	slog.Debug("Latest exchange price retrieved from database", "symbol", symbol, "exchange", exchange)
	return price, nil
}

// GetHighestPrice returns the highest price for a symbol within a period
func (s *PriceService) GetHighestPrice(ctx context.Context, symbol string, period time.Duration) (*domain.PriceResponse, error) {
	// Validate inputs
	if !domain.IsValidSymbol(symbol) {
		return nil, domain.ErrSymbolNotSupported(symbol)
	}
	if period <= 0 {
		return nil, domain.ErrInvalidPeriod
	}

	// Calculate time range
	to := time.Now()
	from := to.Add(-period)

	// For very recent data (< 5 minutes), try cache first
	if period <= 5*time.Minute && s.cacheRepository != nil {
		if price, err := s.getHighestPriceFromCache(ctx, symbol, "", from, to); err == nil {
			return price, nil
		}
	}

	// Use database for longer periods or if cache fails
	if period <= time.Hour {
		// For short periods, use time range query
		return s.priceRepository.GetHighestPriceInDuration(ctx, symbol, from, to)
	}

	// For longer periods, get overall highest
	return s.priceRepository.GetHighestPrice(ctx, symbol)
}

// GetHighestExchangePrice returns the highest price for a symbol from a specific exchange within a period
func (s *PriceService) GetHighestExchangePrice(ctx context.Context, symbol, exchange string, period time.Duration) (*domain.PriceResponse, error) {
	// Validate inputs
	if !domain.IsValidSymbol(symbol) {
		return nil, domain.ErrSymbolNotSupported(symbol)
	}
	if exchange == "" {
		return nil, domain.ErrEmptyExchange
	}
	if period <= 0 {
		return nil, domain.ErrInvalidPeriod
	}

	// Calculate time range
	to := time.Now()
	from := to.Add(-period)

	// For very recent data (< 5 minutes), try cache first
	if period <= 5*time.Minute && s.cacheRepository != nil {
		if price, err := s.getHighestPriceFromCache(ctx, symbol, exchange, from, to); err == nil {
			return price, nil
		}
	}

	// Use database for longer periods or if cache fails
	if period <= time.Hour {
		// For short periods, use time range query
		return s.priceRepository.GetHighestPriceInDurationExchange(ctx, symbol, exchange, from, to)
	}

	// For longer periods, get overall highest
	return s.priceRepository.GetHighestPriceExchange(ctx, symbol, exchange)
}

// GetLowestPrice returns the lowest price for a symbol within a period
func (s *PriceService) GetLowestPrice(ctx context.Context, symbol string, period time.Duration) (*domain.PriceResponse, error) {
	// Validate inputs
	if !domain.IsValidSymbol(symbol) {
		return nil, domain.ErrSymbolNotSupported(symbol)
	}
	if period <= 0 {
		return nil, domain.ErrInvalidPeriod
	}

	// Calculate time range
	to := time.Now()
	from := to.Add(-period)

	// For very recent data (< 5 minutes), try cache first
	if period <= 5*time.Minute && s.cacheRepository != nil {
		if price, err := s.getLowestPriceFromCache(ctx, symbol, "", from, to); err == nil {
			return price, nil
		}
	}

	// Use database for longer periods or if cache fails
	if period <= time.Hour {
		// For short periods, use time range query
		return s.priceRepository.GetLowestPriceInDuration(ctx, symbol, from, to)
	}

	// For longer periods, get overall lowest
	return s.priceRepository.GetLowestPrice(ctx, symbol)
}

// GetLowestExchangePrice returns the lowest price for a symbol from a specific exchange within a period
func (s *PriceService) GetLowestExchangePrice(ctx context.Context, symbol, exchange string, period time.Duration) (*domain.PriceResponse, error) {
	// Validate inputs
	if !domain.IsValidSymbol(symbol) {
		return nil, domain.ErrSymbolNotSupported(symbol)
	}
	if exchange == "" {
		return nil, domain.ErrEmptyExchange
	}
	if period <= 0 {
		return nil, domain.ErrInvalidPeriod
	}

	// Calculate time range
	to := time.Now()
	from := to.Add(-period)

	// For very recent data (< 5 minutes), try cache first
	if period <= 5*time.Minute && s.cacheRepository != nil {
		if price, err := s.getLowestPriceFromCache(ctx, symbol, exchange, from, to); err == nil {
			return price, nil
		}
	}

	// Use database for longer periods or if cache fails
	if period <= time.Hour {
		// For short periods, use time range query
		return s.priceRepository.GetLowestPriceInDurationExchange(ctx, symbol, exchange, from, to)
	}

	// For longer periods, get overall lowest
	return s.priceRepository.GetLowestPriceExchange(ctx, symbol, exchange)
}

// GetAveragePrice returns the average price for a symbol within a period
func (s *PriceService) GetAveragePrice(ctx context.Context, symbol string, period time.Duration) (*domain.PriceResponse, error) {
	// Validate inputs
	if !domain.IsValidSymbol(symbol) {
		return nil, domain.ErrSymbolNotSupported(symbol)
	}
	if period <= 0 {
		return nil, domain.ErrInvalidPeriod
	}

	// Calculate time range
	to := time.Now()
	from := to.Add(-period)

	// For very recent data (< 5 minutes), try cache first
	if period <= 5*time.Minute && s.cacheRepository != nil {
		if price, err := s.getAveragePriceFromCache(ctx, symbol, "", from, to); err == nil {
			return price, nil
		}
	}

	// Use database for longer periods or if cache fails
	if period <= time.Hour {
		// For short periods, use time range query
		return s.priceRepository.GetAveragePriceInDuration(ctx, symbol, from, to)
	}

	// For longer periods, get recent average
	return s.priceRepository.GetAveragePrice(ctx, symbol)
}

// GetAverageExchangePrice returns the average price for a symbol from a specific exchange within a period
func (s *PriceService) GetAverageExchangePrice(ctx context.Context, symbol, exchange string, period time.Duration) (*domain.PriceResponse, error) {
	// Validate inputs
	if !domain.IsValidSymbol(symbol) {
		return nil, domain.ErrSymbolNotSupported(symbol)
	}
	if exchange == "" {
		return nil, domain.ErrEmptyExchange
	}
	if period <= 0 {
		return nil, domain.ErrInvalidPeriod
	}

	// Calculate time range
	to := time.Now()
	from := to.Add(-period)

	// For very recent data (< 5 minutes), try cache first
	if period <= 5*time.Minute && s.cacheRepository != nil {
		if price, err := s.getAveragePriceFromCache(ctx, symbol, exchange, from, to); err == nil {
			return price, nil
		}
	}

	// Use database for longer periods or if cache fails
	if period <= time.Hour {
		// For short periods, use time range query
		return s.priceRepository.GetAveragePriceInDurationExchange(ctx, symbol, exchange, from, to)
	}

	// For longer periods, get recent average
	return s.priceRepository.GetAveragePriceExchange(ctx, symbol, exchange)
}

// GetPriceStatistics returns detailed price statistics for a symbol and exchange within a period
func (s *PriceService) GetPriceStatistics(ctx context.Context, symbol, exchange string, period time.Duration) (*domain.PriceStatistics, error) {
	// Validate inputs
	if !domain.IsValidSymbol(symbol) {
		return nil, domain.ErrSymbolNotSupported(symbol)
	}
	if exchange == "" {
		return nil, domain.ErrEmptyExchange
	}
	if period <= 0 {
		return nil, domain.ErrInvalidPeriod
	}

	// Calculate time range
	to := time.Now()
	from := to.Add(-period)

	// For recent data, try to get from cache and calculate statistics
	if period <= 5*time.Minute && s.cacheRepository != nil {
		if stats, err := s.getStatisticsFromCache(ctx, symbol, exchange, from, to, period); err == nil {
			return stats, nil
		}
	}

	// Fallback to database
	return s.priceRepository.GetPriceStatistics(ctx, symbol, exchange, from, to)
}

// Helper methods for cache operations

// getHighestPriceFromCache gets the highest price from cache data
func (s *PriceService) getHighestPriceFromCache(ctx context.Context, symbol, exchange string, from, to time.Time) (*domain.PriceResponse, error) {
	if exchange == "" {
		// Get from all exchanges and find highest
		return s.getHighestPriceFromAllExchanges(ctx, symbol, from, to)
	}

	updates, err := s.cacheRepository.GetPricesInTimeRange(ctx, symbol, exchange, from, to)
	if err != nil || len(updates) == 0 {
		return nil, domain.ErrPriceNotFound
	}

	highest := updates[0]
	for _, update := range updates {
		if update.Price > highest.Price {
			highest = update
		}
	}

	return &domain.PriceResponse{
		Symbol:    highest.Symbol,
		Exchange:  highest.Exchange,
		Price:     highest.Price,
		Timestamp: highest.Timestamp,
	}, nil
}

// getLowestPriceFromCache gets the lowest price from cache data
func (s *PriceService) getLowestPriceFromCache(ctx context.Context, symbol, exchange string, from, to time.Time) (*domain.PriceResponse, error) {
	if exchange == "" {
		// Get from all exchanges and find lowest
		return s.getLowestPriceFromAllExchanges(ctx, symbol, from, to)
	}

	updates, err := s.cacheRepository.GetPricesInTimeRange(ctx, symbol, exchange, from, to)
	if err != nil || len(updates) == 0 {
		return nil, domain.ErrPriceNotFound
	}

	lowest := updates[0]
	for _, update := range updates {
		if update.Price < lowest.Price {
			lowest = update
		}
	}

	return &domain.PriceResponse{
		Symbol:    lowest.Symbol,
		Exchange:  lowest.Exchange,
		Price:     lowest.Price,
		Timestamp: lowest.Timestamp,
	}, nil
}

// getAveragePriceFromCache gets the average price from cache data
func (s *PriceService) getAveragePriceFromCache(ctx context.Context, symbol, exchange string, from, to time.Time) (*domain.PriceResponse, error) {
	if exchange == "" {
		// Get from all exchanges and calculate average
		return s.getAveragePriceFromAllExchanges(ctx, symbol, from, to)
	}

	updates, err := s.cacheRepository.GetPricesInTimeRange(ctx, symbol, exchange, from, to)
	if err != nil || len(updates) == 0 {
		return nil, domain.ErrPriceNotFound
	}

	var sum float64
	for _, update := range updates {
		sum += update.Price
	}

	average := sum / float64(len(updates))
	latestUpdate := updates[len(updates)-1]

	return &domain.PriceResponse{
		Symbol:    symbol,
		Exchange:  exchange,
		Price:     average,
		Timestamp: latestUpdate.Timestamp,
	}, nil
}

// getStatisticsFromCache calculates statistics from cache data
func (s *PriceService) getStatisticsFromCache(ctx context.Context, symbol, exchange string, from, to time.Time, period time.Duration) (*domain.PriceStatistics, error) {
	updates, err := s.cacheRepository.GetPricesInTimeRange(ctx, symbol, exchange, from, to)
	if err != nil || len(updates) == 0 {
		return nil, domain.ErrDataNotAvailable
	}

	aggregator := domain.NewPriceAggregator()
	return aggregator.CalculateStatistics(updates, period.String())
}

// Helper methods for multi-exchange operations

// getHighestPriceFromAllExchanges gets highest price across all exchanges
func (s *PriceService) getHighestPriceFromAllExchanges(ctx context.Context, symbol string, from, to time.Time) (*domain.PriceResponse, error) {
	exchanges := domain.GetAllLiveExchangeNames()
	var highest *domain.PriceResponse

	for _, exchange := range exchanges {
		updates, err := s.cacheRepository.GetPricesInTimeRange(ctx, symbol, exchange, from, to)
		if err != nil {
			continue
		}

		for _, update := range updates {
			response := &domain.PriceResponse{
				Symbol:    update.Symbol,
				Exchange:  update.Exchange,
				Price:     update.Price,
				Timestamp: update.Timestamp,
			}

			if highest == nil || response.Price > highest.Price {
				highest = response
			}
		}
	}

	if highest == nil {
		return nil, domain.ErrPriceNotFound
	}

	return highest, nil
}

// getLowestPriceFromAllExchanges gets lowest price across all exchanges
func (s *PriceService) getLowestPriceFromAllExchanges(ctx context.Context, symbol string, from, to time.Time) (*domain.PriceResponse, error) {
	exchanges := domain.GetAllLiveExchangeNames()
	var lowest *domain.PriceResponse

	for _, exchange := range exchanges {
		updates, err := s.cacheRepository.GetPricesInTimeRange(ctx, symbol, exchange, from, to)
		if err != nil {
			continue
		}

		for _, update := range updates {
			response := &domain.PriceResponse{
				Symbol:    update.Symbol,
				Exchange:  update.Exchange,
				Price:     update.Price,
				Timestamp: update.Timestamp,
			}

			if lowest == nil || response.Price < lowest.Price {
				lowest = response
			}
		}
	}

	if lowest == nil {
		return nil, domain.ErrPriceNotFound
	}

	return lowest, nil
}

// getAveragePriceFromAllExchanges gets average price across all exchanges
func (s *PriceService) getAveragePriceFromAllExchanges(ctx context.Context, symbol string, from, to time.Time) (*domain.PriceResponse, error) {
	exchanges := domain.GetAllLiveExchangeNames()
	var allUpdates []domain.PriceUpdate
	var latestTimestamp time.Time

	for _, exchange := range exchanges {
		updates, err := s.cacheRepository.GetPricesInTimeRange(ctx, symbol, exchange, from, to)
		if err != nil {
			continue
		}

		allUpdates = append(allUpdates, updates...)

		// Track latest timestamp
		for _, update := range updates {
			if update.Timestamp.After(latestTimestamp) {
				latestTimestamp = update.Timestamp
			}
		}
	}

	if len(allUpdates) == 0 {
		return nil, domain.ErrPriceNotFound
	}

	var sum float64
	for _, update := range allUpdates {
		sum += update.Price
	}

	average := sum / float64(len(allUpdates))

	return &domain.PriceResponse{
		Symbol:    symbol,
		Exchange:  "", // Multi-exchange average
		Price:     average,
		Timestamp: latestTimestamp,
	}, nil
}
