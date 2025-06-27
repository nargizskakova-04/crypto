package exchange

import (
	"context"
	"log/slog"
	"math/rand"
	"time"

	"crypto/internal/core/domain"
	"crypto/internal/core/port"
)

// TestExchangeAdapter генерирует синтетические данные для Test Mode
type TestExchangeAdapter struct {
	name      string
	connected bool
}

func NewTestExchangeAdapter(name string) port.ExchangeAdapter {
	return &TestExchangeAdapter{
		name:      name,
		connected: false,
	}
}

func (a *TestExchangeAdapter) Connect(ctx context.Context) error {
	slog.Info("Connecting test exchange adapter", "name", a.name)
	a.connected = true
	return nil
}

func (a *TestExchangeAdapter) GetPriceStream(ctx context.Context) (<-chan domain.PriceUpdate, <-chan error) {
	priceStream := make(chan domain.PriceUpdate, 100)
	errorStream := make(chan error, 10)

	// Стартовые цены для каждой пары
	basePrices := map[string]float64{
		"BTCUSDT":  45000.0,
		"ETHUSDT":  3000.0,
		"DOGEUSDT": 0.07,
		"TONUSDT":  2.5,
		"SOLUSDT":  100.0,
	}

	// Текущие цены (будут колебаться)
	currentPrices := make(map[string]float64)
	for symbol, price := range basePrices {
		currentPrices[symbol] = price
	}

	// Различные "биржи" для Test Mode
	exchanges := []string{"TestExchange1", "TestExchange2", "TestExchange3"}

	go func() {
		defer close(priceStream)
		defer close(errorStream)

		ticker := time.NewTicker(200 * time.Millisecond) // Генерируем данные каждые 200мс
		defer ticker.Stop()

		slog.Info("Test exchange adapter started generating data", "name", a.name)

		for {
			select {
			case <-ctx.Done():
				slog.Info("Test exchange adapter stopped", "name", a.name)
				return
			case <-ticker.C:
				// Генерируем случайное обновление
				symbols := []string{"BTCUSDT", "ETHUSDT", "DOGEUSDT", "TONUSDT", "SOLUSDT"}
				symbol := symbols[rand.Intn(len(symbols))]
				exchange := exchanges[rand.Intn(len(exchanges))]

				// Получаем текущую цену и добавляем случайное изменение
				currentPrice := currentPrices[symbol]
				basePrice := basePrices[symbol]

				// Случайное изменение от -5% до +5%
				change := (rand.Float64() - 0.5) * 0.1 // -0.05 to +0.05
				newPrice := currentPrice * (1 + change)

				// Ограничиваем отклонение от базовой цены (макс ±20%)
				minPrice := basePrice * 0.8
				maxPrice := basePrice * 1.2
				if newPrice < minPrice {
					newPrice = minPrice + rand.Float64()*(basePrice-minPrice)
				} else if newPrice > maxPrice {
					newPrice = basePrice + rand.Float64()*(maxPrice-basePrice)
				}

				currentPrices[symbol] = newPrice

				priceUpdate := domain.PriceUpdate{
					Exchange:  exchange,
					Symbol:    symbol,
					Price:     newPrice,
					Timestamp: time.Now(),
				}

				select {
				case priceStream <- priceUpdate:
					slog.Debug("Generated test price update",
						"exchange", exchange,
						"symbol", symbol,
						"price", newPrice)
				case <-ctx.Done():
					return
				default:
					slog.Warn("Test price stream buffer full")
				}
			}
		}
	}()

	return priceStream, errorStream
}

func (a *TestExchangeAdapter) Disconnect() error {
	slog.Info("Disconnecting test exchange adapter", "name", a.name)
	a.connected = false
	return nil
}

func (a *TestExchangeAdapter) GetExchangeInfo() domain.ExchangeInfo {
	status := "disconnected"
	if a.connected {
		status = "connected"
	}

	return domain.ExchangeInfo{
		Name:   a.name,
		Host:   "localhost",
		Port:   0, // N/A for test mode
		Status: status,
	}
}

func (a *TestExchangeAdapter) IsConnected() bool {
	return a.connected
}
