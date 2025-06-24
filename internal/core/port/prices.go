package port

import (
	"time"

	"crypto/internal/core/domain"
)

type PricesRepository interface {
	GetLatestPrice(symbol string) (domain.GetPrice, error)
	GetLatestExchangePrice(symbol string, exchange string) (domain.GetPrice, error)

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
