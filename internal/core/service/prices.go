package service

import (
	"errors"
	"log/slog"

	"crypto/internal/core/port"
)

type PriceService struct {
	priceRepository port.PricesRepository
}

func (s *PriceService) GetLatestPrice(symbol string) (int, error) {
	if symbol != "BTCUSDT" || symbol != "DOGEUSDT" || symbol != "TONUSDT" || symbol != "SOLUSDT" || symbol != "ETHUSDT" {
		slog.Error("There is not sumbol with a name: ", symbol)
		return 0, errors.New("There is not sumbol with that name")
	}
	price, err := s.GetLatestPrice(symbol)
	if err != nil {
		slog.Error(err.Error())
		return 0, err
	}
	return price, nil
}
