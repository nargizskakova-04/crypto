package domain

import (
	"fmt"
	"time"
)

// PriceUpdate представляет сырое обновление цены от биржи
type PriceUpdate struct {
	Exchange  string    `json:"exchange"`
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"timestamp"`
}

// MarketData представляет данные от exchange simulators (совместимость)
type MarketData struct {
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Timestamp int64   `json:"timestamp"`
	Exchange  string  `json:"exchange"`
}

// AggregatedData представляет агрегированные данные за период (для PostgreSQL)
type AggregatedData struct {
	PairName     string    `json:"pair_name" db:"pair_name"`
	Exchange     string    `json:"exchange" db:"exchange"`
	Timestamp    time.Time `json:"timestamp" db:"timestamp"`
	AveragePrice float64   `json:"average_price" db:"average_price"`
	MinPrice     float64   `json:"min_price" db:"min_price"`
	MaxPrice     float64   `json:"max_price" db:"max_price"`
}

// PriceResponse представляет ответ API для клиентов
type PriceResponse struct {
	Symbol    string    `json:"symbol"`
	Exchange  string    `json:"exchange,omitempty"`
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"timestamp"`
}

// HealthStatus представляет статус системы
type HealthStatus struct {
	Status       string                 `json:"status"`
	Timestamp    time.Time              `json:"timestamp"`
	Version      string                 `json:"version"`
	Mode         string                 `json:"mode"`
	Dependencies map[string]interface{} `json:"dependencies"`
}

// ExchangeInfo представляет информацию о бирже
type ExchangeInfo struct {
	Name   string `json:"name"`
	Host   string `json:"host"`
	Port   int    `json:"port"`
	Status string `json:"status"` // "connected", "disconnected", "reconnecting"
}

// PriceStatistics представляет расширенную статистику по ценам
type PriceStatistics struct {
	Symbol        string    `json:"symbol"`
	Exchange      string    `json:"exchange,omitempty"`
	Period        string    `json:"period"`
	Count         int       `json:"count"`
	AveragePrice  float64   `json:"average_price"`
	MinPrice      float64   `json:"min_price"`
	MaxPrice      float64   `json:"max_price"`
	FirstPrice    float64   `json:"first_price"`
	LastPrice     float64   `json:"last_price"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Volatility    float64   `json:"volatility,omitempty"`     // Волатильность (стандартное отклонение)
	PriceChange   float64   `json:"price_change,omitempty"`   // Изменение цены (последняя - первая)
	PercentChange float64   `json:"percent_change,omitempty"` // Процентное изменение
}

// Methods for PriceUpdate

// GetKey возвращает уникальный ключ для комбинации биржа:символ
func (p PriceUpdate) GetKey() string {
	return fmt.Sprintf("%s:%s", p.Exchange, p.Symbol)
}

// ToMarketData конвертирует PriceUpdate в MarketData
func (p PriceUpdate) ToMarketData() MarketData {
	return MarketData{
		Symbol:    p.Symbol,
		Price:     p.Price,
		Timestamp: p.Timestamp.Unix(),
		Exchange:  p.Exchange,
	}
}

// ToPriceResponse конвертирует PriceUpdate в PriceResponse
func (p PriceUpdate) ToPriceResponse() PriceResponse {
	return PriceResponse{
		Symbol:    p.Symbol,
		Exchange:  p.Exchange,
		Price:     p.Price,
		Timestamp: p.Timestamp,
	}
}

// Validate проверяет валидность PriceUpdate
func (p PriceUpdate) Validate() error {
	if p.Exchange == "" {
		return ErrEmptyExchange
	}
	if p.Symbol == "" {
		return ErrEmptySymbol
	}
	if !IsValidSymbol(p.Symbol) {
		return fmt.Errorf("%w: %s", ErrInvalidSymbol, p.Symbol)
	}
	if p.Price <= 0 {
		return fmt.Errorf("%w: got %f", ErrInvalidPrice, p.Price)
	}
	if p.Timestamp.IsZero() {
		return ErrEmptyTimestamp
	}
	return nil
}

// Methods for MarketData

// ToPriceUpdate конвертирует MarketData в PriceUpdate
func (m MarketData) ToPriceUpdate() PriceUpdate {
	return PriceUpdate{
		Exchange:  m.Exchange,
		Symbol:    m.Symbol,
		Price:     m.Price,
		Timestamp: time.Unix(m.Timestamp, 0),
	}
}

// Methods for AggregatedData

// GetKey возвращает уникальный ключ для комбинации биржа:символ
func (a AggregatedData) GetKey() string {
	return fmt.Sprintf("%s:%s", a.Exchange, a.PairName)
}

// ToPriceResponse конвертирует AggregatedData в PriceResponse (используя среднюю цену)
func (a AggregatedData) ToPriceResponse() PriceResponse {
	return PriceResponse{
		Symbol:    a.PairName,
		Exchange:  a.Exchange,
		Price:     a.AveragePrice,
		Timestamp: a.Timestamp,
	}
}

// ToPriceStatistics конвертирует AggregatedData в PriceStatistics
func (a AggregatedData) ToPriceStatistics(period string, count int) PriceStatistics {
	return PriceStatistics{
		Symbol:       a.PairName,
		Exchange:     a.Exchange,
		Period:       period,
		Count:        count,
		AveragePrice: a.AveragePrice,
		MinPrice:     a.MinPrice,
		MaxPrice:     a.MaxPrice,
		FirstPrice:   a.MinPrice,                    // Приближение
		LastPrice:    a.MaxPrice,                    // Приближение
		StartTime:    a.Timestamp.Add(-time.Minute), // Период агрегации
		EndTime:      a.Timestamp,
	}
}

// Validate проверяет валидность AggregatedData
func (a AggregatedData) Validate() error {
	if a.Exchange == "" {
		return ErrEmptyExchange
	}
	if a.PairName == "" {
		return ErrEmptySymbol
	}
	if !IsValidSymbol(a.PairName) {
		return fmt.Errorf("%w: %s", ErrInvalidSymbol, a.PairName)
	}
	if a.AveragePrice <= 0 || a.MinPrice <= 0 || a.MaxPrice <= 0 {
		return ErrInvalidPrice
	}
	if a.MinPrice > a.AveragePrice || a.AveragePrice > a.MaxPrice {
		return ErrInvalidPriceRange
	}
	if a.Timestamp.IsZero() {
		return ErrEmptyTimestamp
	}
	return nil
}

// Methods for PriceStatistics

// GetPriceChange возвращает абсолютное изменение цены
func (p PriceStatistics) GetPriceChange() float64 {
	return p.LastPrice - p.FirstPrice
}

// GetPercentChange возвращает процентное изменение цены
func (p PriceStatistics) GetPercentChange() float64 {
	if p.FirstPrice == 0 {
		return 0
	}
	return ((p.LastPrice - p.FirstPrice) / p.FirstPrice) * 100
}

// IsPositiveChange проверяет, выросла ли цена
func (p PriceStatistics) IsPositiveChange() bool {
	return p.LastPrice > p.FirstPrice
}

// GetDuration возвращает продолжительность периода
func (p PriceStatistics) GetDuration() time.Duration {
	return p.EndTime.Sub(p.StartTime)
}

// GetAveragePriceRange возвращает диапазон цен относительно средней
func (p PriceStatistics) GetAveragePriceRange() (float64, float64) {
	avgDeviationUp := p.MaxPrice - p.AveragePrice
	avgDeviationDown := p.AveragePrice - p.MinPrice
	return avgDeviationDown, avgDeviationUp
}

// Validate проверяет валидность статистики
func (p PriceStatistics) Validate() error {
	if p.Symbol == "" {
		return ErrEmptySymbol
	}
	if p.Count <= 0 {
		return fmt.Errorf("count must be positive, got: %d", p.Count)
	}
	if p.MinPrice <= 0 || p.MaxPrice <= 0 || p.AveragePrice <= 0 {
		return ErrInvalidPrice
	}
	if p.MinPrice > p.AveragePrice || p.AveragePrice > p.MaxPrice {
		return ErrInvalidPriceRange
	}
	if p.StartTime.After(p.EndTime) {
		return fmt.Errorf("start time cannot be after end time")
	}
	return nil
}

// ToSummary возвращает краткое описание статистики
func (p PriceStatistics) ToSummary() string {
	change := p.GetPercentChange()
	direction := "📈"
	if change < 0 {
		direction = "📉"
	} else if change == 0 {
		direction = "➡️"
	}

	return fmt.Sprintf("%s %s: %d updates, %.2f avg, %.2f%% change %s",
		p.Exchange, p.Symbol, p.Count, p.AveragePrice, change, direction)
}

// IsOlderThan проверяет, старше ли цена указанного времени
func (p PriceResponse) IsOlderThan(duration time.Duration) bool {
	return time.Since(p.Timestamp) > duration
}

// Methods for HealthStatus

// IsHealthy возвращает true если система работает нормально
func (h HealthStatus) IsHealthy() bool {
	return h.Status == "ok"
}

// AddDependency добавляет информацию о зависимости
func (h *HealthStatus) AddDependency(name string, status string, details map[string]interface{}) {
	if h.Dependencies == nil {
		h.Dependencies = make(map[string]interface{})
	}

	dep := map[string]interface{}{
		"status": status,
	}

	for k, v := range details {
		dep[k] = v
	}

	h.Dependencies[name] = dep
}

// Methods for ExchangeInfo

// IsConnected возвращает true если биржа подключена
func (e ExchangeInfo) IsConnected() bool {
	return e.Status == "connected"
}

// GetAddress возвращает адрес биржи
func (e ExchangeInfo) GetAddress() string {
	if e.Port == 0 {
		return e.Host
	}
	return fmt.Sprintf("%s:%d", e.Host, e.Port)
}
