package port

import (
	"context"
	"crypto/internal/core/domain"
	"time"
)

type PriceRepository interface {
	GetLatestPrice(symbol string) (domain.GetPrice, error)
	GetLatestPriceByExchange(symbol string, exchange string) (domain.GetPrice, error)

	GetHighestPrice(symbol string) (domain.GetPrice, error)
	GetHighestPriceExchange(symbol string, exchange string) (domain.GetPrice, error)
	GetHighestPriceInDuration(symbol string, from time.Time, to time.Time) (domain.GetPrice, error)
	GetHighestPriceInDurationExchange(symbol string, exchange string, from time.Time, to time.Time) (domain.GetPrice, error)

	GetLowestPrice(symbol string) (domain.GetPrice, error)
	GetLowestPriceExchange(symbol string, exchange string) (domain.GetPrice, error)
	GetLowestPriceInDuration(symbol string, from time.Time, to time.Time) (domain.GetPrice, error)
	GetLowestPriceInDurationExchange(symbol string, exchange string, from time.Time, to time.Time) (domain.GetPrice, error)

	GetAveragePrice(symbol string) (domain.GetPrice, error)
	GetAveragePriceExchange(symbol string, exchange string) (domain.GetPrice, error)
	GetAveragePriceInDurationExchange(symbol string, exchange string, from time.Time, to time.Time) (domain.GetPrice, error)
}

type PriceService interface {
	GetLatestPrice(ctx context.Context, symbol string) (*domain.MarketData, error)

	GetLatestPriceByExchange(ctx context.Context, symbol, exchange string) (*domain.MarketData, error)

	GetHighestPrice(ctx context.Context, symbol string, period time.Duration) (*domain.MarketData, error)

	GetHighestPriceByExchange(ctx context.Context, symbol, exchange string, period time.Duration) (*domain.MarketData, error)

	GetLowestPrice(ctx context.Context, symbol string, period time.Duration) (*domain.MarketData, error)

	GetLowestPriceByExchange(ctx context.Context, symbol, exchange string, period time.Duration) (*domain.MarketData, error)

	GetAveragePrice(ctx context.Context, symbol string, period time.Duration) (*domain.MarketData, error)

	GetAveragePriceByExchange(ctx context.Context, symbol, exchange string, period time.Duration) (*domain.MarketData, error)
}
