package exchange

import (
	"bufio"
	"context"
	"crypto/internal/core/domain"
	"crypto/internal/core/port"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"time"
)

// LiveExchangeAdapter подключается к реальным exchange simulators
type LiveExchangeAdapter struct {
	name              string
	host              string
	port              int
	reconnectInterval time.Duration
	conn              net.Conn
	connected         bool
}

func NewLiveExchangeAdapter(name, host string, port int, reconnectInterval time.Duration) port.ExchangeAdapter {
	return &LiveExchangeAdapter{
		name:              name,
		host:              host,
		port:              port,
		reconnectInterval: reconnectInterval,
		connected:         false,
	}
}

func (a *LiveExchangeAdapter) Connect(ctx context.Context) error {
	address := fmt.Sprintf("%s:%d", a.host, a.port)
	slog.Info("Connecting to exchange", "name", a.name, "address", address)

	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", address, err)
	}

	a.conn = conn
	a.connected = true

	slog.Info("Successfully connected to exchange", "name", a.name, "address", address)
	return nil
}

func (a *LiveExchangeAdapter) GetPriceStream(ctx context.Context) (<-chan domain.PriceUpdate, <-chan error) {
	priceStream := make(chan domain.PriceUpdate, 100)
	errorStream := make(chan error, 10)

	go func() {
		defer close(priceStream)
		defer close(errorStream)

		if a.conn == nil {
			errorStream <- fmt.Errorf("not connected to exchange")
			return
		}

		scanner := bufio.NewScanner(a.conn)
		scanner.Buffer(make([]byte, 64*1024), 64*1024) // Увеличиваем буфер

		for {
			select {
			case <-ctx.Done():
				slog.Info("Price stream context cancelled", "exchange", a.name)
				return
			default:
				if !scanner.Scan() {
					if err := scanner.Err(); err != nil {
						errorStream <- fmt.Errorf("scanner error: %w", err)
					} else {
						errorStream <- fmt.Errorf("connection closed by server")
					}
					return
				}

				line := scanner.Text()
				if line == "" {
					continue
				}

				// Парсим JSON данные от exchange simulator
				var marketData domain.MarketData
				if err := json.Unmarshal([]byte(line), &marketData); err != nil {
					slog.Warn("Failed to parse market data", "exchange", a.name, "data", line, "error", err)
					continue
				}

				// Устанавливаем exchange name если не указан
				if marketData.Exchange == "" {
					marketData.Exchange = a.name
				}

				// Конвертируем в PriceUpdate
				priceUpdate := marketData.ToPriceUpdate()

				// Валидируем данные
				if err := priceUpdate.Validate(); err != nil {
					slog.Warn("Invalid price update", "exchange", a.name, "error", err)
					continue
				}

				select {
				case priceStream <- priceUpdate:
					slog.Debug("Price update sent",
						"exchange", a.name,
						"symbol", priceUpdate.Symbol,
						"price", priceUpdate.Price)
				case <-ctx.Done():
					return
				default:
					slog.Warn("Price stream buffer full, dropping update", "exchange", a.name)
				}
			}
		}
	}()

	return priceStream, errorStream
}

func (a *LiveExchangeAdapter) Disconnect() error {
	if a.conn != nil {
		slog.Info("Disconnecting from exchange", "name", a.name)
		err := a.conn.Close()
		a.conn = nil
		a.connected = false
		return err
	}
	return nil
}

func (a *LiveExchangeAdapter) GetExchangeInfo() domain.ExchangeInfo {
	status := "disconnected"
	if a.connected {
		status = "connected"
	}

	return domain.ExchangeInfo{
		Name:   a.name,
		Host:   a.host,
		Port:   a.port,
		Status: status,
	}
}

func (a *LiveExchangeAdapter) IsConnected() bool {
	return a.connected
}

// internal/adapters/exchange/test_adapter.go
