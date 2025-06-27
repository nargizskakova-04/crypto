package v1

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"crypto/internal/core/domain"
	"crypto/internal/core/port"
)

type PriceHandler struct {
	priceService port.PriceService
}

func NewPriceHandler(priceService port.PriceService) *PriceHandler {
	return &PriceHandler{
		priceService: priceService,
	}
}

// GET /prices/latest/{symbol}
func (h *PriceHandler) GetLatestPrice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract symbol from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	symbol := strings.ToUpper(parts[2])

	slog.Info("Getting latest price", "symbol", symbol)

	price, err := h.priceService.GetLatestPrice(r.Context(), symbol)
	if err != nil {
		slog.Error("Failed to get latest price", "symbol", symbol, "error", err)
		http.Error(w, "Price not found", http.StatusNotFound)
		return
	}

	respondJSON(w, price)
}

// GET /prices/latest/{exchange}/{symbol}
func (h *PriceHandler) GetLatestExchangePrice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract exchange and symbol from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	exchange := parts[2]
	symbol := strings.ToUpper(parts[3])

	slog.Info("Getting latest exchange price", "exchange", exchange, "symbol", symbol)

	price, err := h.priceService.GetLatestExchangePrice(r.Context(), symbol, exchange)
	if err != nil {
		slog.Error("Failed to get latest exchange price", "exchange", exchange, "symbol", symbol, "error", err)
		http.Error(w, "Price not found", http.StatusNotFound)
		return
	}

	respondJSON(w, price)
}

// GET /prices/highest/{symbol}[?period={duration}]
func (h *PriceHandler) GetHighestPrice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract symbol from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	symbol := strings.ToUpper(parts[2])

	// Parse period parameter
	period := parsePeriodParameter(r)

	slog.Info("Getting highest price", "symbol", symbol, "period", period)

	price, err := h.priceService.GetHighestPrice(r.Context(), symbol, period)
	if err != nil {
		slog.Error("Failed to get highest price", "symbol", symbol, "error", err)
		http.Error(w, "Price not found", http.StatusNotFound)
		return
	}

	respondJSON(w, price)
}

// GET /prices/highest/{exchange}/{symbol}[?period={duration}]
func (h *PriceHandler) GetHighestExchangePrice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract exchange and symbol from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	exchange := parts[2]
	symbol := strings.ToUpper(parts[3])

	// Parse period parameter
	period := parsePeriodParameter(r)

	slog.Info("Getting highest exchange price", "exchange", exchange, "symbol", symbol, "period", period)

	price, err := h.priceService.GetHighestExchangePrice(r.Context(), symbol, exchange, period)
	if err != nil {
		slog.Error("Failed to get highest exchange price", "exchange", exchange, "symbol", symbol, "error", err)
		http.Error(w, "Price not found", http.StatusNotFound)
		return
	}

	respondJSON(w, price)
}

// GET /prices/lowest/{symbol}[?period={duration}]
func (h *PriceHandler) GetLowestPrice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract symbol from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	symbol := strings.ToUpper(parts[2])

	// Parse period parameter
	period := parsePeriodParameter(r)

	slog.Info("Getting lowest price", "symbol", symbol, "period", period)

	price, err := h.priceService.GetLowestPrice(r.Context(), symbol, period)
	if err != nil {
		slog.Error("Failed to get lowest price", "symbol", symbol, "error", err)
		http.Error(w, "Price not found", http.StatusNotFound)
		return
	}

	respondJSON(w, price)
}

// GET /prices/lowest/{exchange}/{symbol}[?period={duration}]
func (h *PriceHandler) GetLowestExchangePrice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract exchange and symbol from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	exchange := parts[2]
	symbol := strings.ToUpper(parts[3])

	// Parse period parameter
	period := parsePeriodParameter(r)

	slog.Info("Getting lowest exchange price", "exchange", exchange, "symbol", symbol, "period", period)

	price, err := h.priceService.GetLowestExchangePrice(r.Context(), symbol, exchange, period)
	if err != nil {
		slog.Error("Failed to get lowest exchange price", "exchange", exchange, "symbol", symbol, "error", err)
		http.Error(w, "Price not found", http.StatusNotFound)
		return
	}

	respondJSON(w, price)
}

// GET /prices/average/{symbol}
func (h *PriceHandler) GetAveragePrice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract symbol from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	symbol := strings.ToUpper(parts[2])

	// Parse period parameter
	period := parsePeriodParameter(r)

	slog.Info("Getting average price", "symbol", symbol, "period", period)

	price, err := h.priceService.GetAveragePrice(r.Context(), symbol, period)
	if err != nil {
		slog.Error("Failed to get average price", "symbol", symbol, "error", err)
		http.Error(w, "Price not found", http.StatusNotFound)
		return
	}

	respondJSON(w, price)
}

// GET /prices/average/{exchange}/{symbol}[?period={duration}]
func (h *PriceHandler) GetAverageExchangePrice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract exchange and symbol from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	exchange := parts[2]
	symbol := strings.ToUpper(parts[3])

	// Parse period parameter
	period := parsePeriodParameter(r)

	slog.Info("Getting average exchange price", "exchange", exchange, "symbol", symbol, "period", period)

	price, err := h.priceService.GetAverageExchangePrice(r.Context(), symbol, exchange, period)
	if err != nil {
		slog.Error("Failed to get average exchange price", "exchange", exchange, "symbol", symbol, "error", err)
		http.Error(w, "Price not found", http.StatusNotFound)
		return
	}

	respondJSON(w, price)
}

// Helper functions

func parsePeriodParameter(r *http.Request) time.Duration {
	periodStr := r.URL.Query().Get("period")
	if periodStr == "" {
		return 24 * time.Hour // default period
	}

	// Parse duration string (e.g., "1s", "3m", "5m")
	if duration, err := time.ParseDuration(periodStr); err == nil {
		return duration
	}

	// Try to parse as seconds (e.g., "30" -> 30 seconds)
	if seconds, err := strconv.Atoi(periodStr); err == nil {
		return time.Duration(seconds) * time.Second
	}

	// Default fallback
	return 24 * time.Hour
}

func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("Failed to encode JSON response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// internal/adapters/handler/http/v1/mode_handler.go

type ModeHandler struct {
	app ModeManager // Interface для управления режимами
}

type ModeManager interface {
	SwitchToLiveMode() error
	SwitchToTestMode() error
	GetCurrentMode() port.AppMode
}

func NewModeHandler(app ModeManager) *ModeHandler {
	return &ModeHandler{
		app: app,
	}
}

// POST /mode/test
func (h *ModeHandler) SwitchToTestMode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	slog.Info("Switching to test mode via API")

	if err := h.app.SwitchToTestMode(); err != nil {
		slog.Error("Failed to switch to test mode", "error", err)
		http.Error(w, "Failed to switch mode", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status":  "success",
		"message": "Switched to test mode",
		"mode":    "test",
	}

	respondJSON(w, response)
}

// POST /mode/live
func (h *ModeHandler) SwitchToLiveMode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	slog.Info("Switching to live mode via API")

	if err := h.app.SwitchToLiveMode(); err != nil {
		slog.Error("Failed to switch to live mode", "error", err)
		http.Error(w, "Failed to switch mode", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status":  "success",
		"message": "Switched to live mode",
		"mode":    "live",
	}

	respondJSON(w, response)
}

// internal/adapters/handler/http/v1/health_handler.go

type HealthHandler struct {
	priceRepository  port.PricesRepository
	cacheRepository  port.CacheRepository
	exchangeAdapters map[string]port.ExchangeAdapter
}

func NewHealthHandler(
	priceRepository port.PricesRepository,
	cacheRepository port.CacheRepository,
	exchangeAdapters map[string]port.ExchangeAdapter,
) *HealthHandler {
	return &HealthHandler{
		priceRepository:  priceRepository,
		cacheRepository:  cacheRepository,
		exchangeAdapters: exchangeAdapters,
	}
}

// GET /health
func (h *HealthHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	health := domain.HealthStatus{
		Status:       "ok",
		Timestamp:    time.Now(),
		Version:      "1.0.0",
		Dependencies: make(map[string]interface{}),
	}

	// Check PostgreSQL
	if err := h.priceRepository.HealthCheck(r.Context()); err != nil {
		health.Status = "degraded"
		health.Dependencies["postgresql"] = map[string]interface{}{
			"status": "down",
			"error":  err.Error(),
		}
	} else {
		health.Dependencies["postgresql"] = map[string]interface{}{
			"status": "up",
		}
	}

	// Check Redis
	if h.cacheRepository != nil {
		if err := h.cacheRepository.HealthCheck(r.Context()); err != nil {
			health.Dependencies["redis"] = map[string]interface{}{
				"status": "down",
				"error":  err.Error(),
			}
		} else {
			health.Dependencies["redis"] = map[string]interface{}{
				"status": "up",
			}
		}
	} else {
		health.Dependencies["redis"] = map[string]interface{}{
			"status": "not_configured",
		}
	}

	// Check exchanges
	exchanges := make(map[string]interface{})
	for name, adapter := range h.exchangeAdapters {
		info := adapter.GetExchangeInfo()
		exchanges[name] = map[string]interface{}{
			"status": info.Status,
			"host":   info.Host,
			"port":   info.Port,
		}
	}
	health.Dependencies["exchanges"] = exchanges

	// Set overall status based on critical components
	if health.Status == "ok" {
		// PostgreSQL is critical
		if pgStatus, ok := health.Dependencies["postgresql"].(map[string]interface{}); ok {
			if pgStatus["status"] != "up" {
				health.Status = "down"
			}
		}
	}

	statusCode := http.StatusOK
	if health.Status == "down" {
		statusCode = http.StatusServiceUnavailable
	} else if health.Status == "degraded" {
		statusCode = http.StatusPartialContent
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(health)
}

// internal/adapters/handler/http/v1/routes.go

func SetRoutes(
	mux *http.ServeMux,
	priceHandler *PriceHandler,
	modeHandler *ModeHandler,
	healthHandler *HealthHandler,
) {
	// Price API routes
	mux.HandleFunc("/prices/latest/", handlePriceRoutes(priceHandler))
	mux.HandleFunc("/prices/highest/", handlePriceRoutes(priceHandler))
	mux.HandleFunc("/prices/lowest/", handlePriceRoutes(priceHandler))
	mux.HandleFunc("/prices/average/", handlePriceRoutes(priceHandler))

	// Mode API routes
	mux.HandleFunc("/mode/test", modeHandler.SwitchToTestMode)
	mux.HandleFunc("/mode/live", modeHandler.SwitchToLiveMode)

	// Health API
	mux.HandleFunc("/health", healthHandler.GetHealth)

	slog.Info("HTTP routes configured")
}

func handlePriceRoutes(handler *PriceHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.Trim(r.URL.Path, "/")
		parts := strings.Split(path, "/")

		if len(parts) < 3 {
			http.Error(w, "Invalid URL format", http.StatusBadRequest)
			return
		}

		priceType := parts[1] // "latest", "highest", "lowest", "average"

		switch priceType {
		case "latest":
			if len(parts) == 3 { // /prices/latest/{symbol}
				handler.GetLatestPrice(w, r)
			} else if len(parts) == 4 { // /prices/latest/{exchange}/{symbol}
				handler.GetLatestExchangePrice(w, r)
			} else {
				http.Error(w, "Invalid URL format", http.StatusBadRequest)
			}

		case "highest":
			if len(parts) == 3 { // /prices/highest/{symbol}
				handler.GetHighestPrice(w, r)
			} else if len(parts) == 4 { // /prices/highest/{exchange}/{symbol}
				handler.GetHighestExchangePrice(w, r)
			} else {
				http.Error(w, "Invalid URL format", http.StatusBadRequest)
			}

		case "lowest":
			if len(parts) == 3 { // /prices/lowest/{symbol}
				handler.GetLowestPrice(w, r)
			} else if len(parts) == 4 { // /prices/lowest/{exchange}/{symbol}
				handler.GetLowestExchangePrice(w, r)
			} else {
				http.Error(w, "Invalid URL format", http.StatusBadRequest)
			}

		case "average":
			if len(parts) == 3 { // /prices/average/{symbol}
				handler.GetAveragePrice(w, r)
			} else if len(parts) == 4 { // /prices/average/{exchange}/{symbol}
				handler.GetAverageExchangePrice(w, r)
			} else {
				http.Error(w, "Invalid URL format", http.StatusBadRequest)
			}

		default:
			http.Error(w, "Unknown price type", http.StatusNotFound)
		}
	}
}
