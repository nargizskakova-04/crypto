package v1

import (
	"crypto/internal/core/port"
)

type MarketDataHandler struct {
	marketInterface marketInterface
}

func NewMarketDataHandler(
	marketInterface marketInterface,
) *MarketDataHandler {
	return &MarketDataHandler{
		marketInterface: marketInterface,
	}
}

func SetMarketDataHandler(
	marketService port.MarketService,
	cashe port.Cashe,
)
