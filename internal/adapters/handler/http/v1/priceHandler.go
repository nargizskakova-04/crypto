package v1

import (
	"net/http"

	"crypto/internal/core/port"
)

type PriceHandler struct {
	priceService port.PriceService
}

func NewPriceHandler(
	priceService port.PriceService,
) *PriceHandler {
	return &PriceHandler{
		priceService: priceService,
	}
}

func (h *PriceHandler) GetLatestPrice(w http.ResponseWriter, r *http.Request) {
}

func (h *PriceHandler) GetLatestPriceByExchange(w http.ResponseWriter, r *http.Request) {
}

func (h *PriceHandler) GetHighestPrice(w http.ResponseWriter, r *http.Request) {
}

func (h *PriceHandler) GetHighestPriceByExchange(w http.ResponseWriter, r *http.Request) {
}

func (h *PriceHandler) GetLowestPrice(w http.ResponseWriter, r *http.Request) {
}

func (h *PriceHandler) GetLowestPriceByExchange(w http.ResponseWriter, r *http.Request) {
}

func (h *PriceHandler) GetAveragePrice(w http.ResponseWriter, r *http.Request) {
}

func (h *PriceHandler) GetAveragePriceByExchange(w http.ResponseWriter, r *http.Request) {
}
