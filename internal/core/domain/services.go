package domain

import (
	"fmt"
	"math"
	"time"
)

// PriceValidator предоставляет методы валидации цен
type PriceValidator struct{}

// NewPriceValidator создает новый валидатор цен
func NewPriceValidator() *PriceValidator {
	return &PriceValidator{}
}

// ValidatePriceUpdate проверяет валидность обновления цены
func (v *PriceValidator) ValidatePriceUpdate(update PriceUpdate) error {
	if err := update.Validate(); err != nil {
		return err
	}

	// Дополнительные бизнес-правила валидации
	if !v.isPriceReasonable(update.Symbol, update.Price) {
		return NewValidationError("price", update.Price, "price is unreasonable for symbol")
	}

	if !v.isTimestampRecent(update.Timestamp) {
		return NewValidationError("timestamp", update.Timestamp, "timestamp is too old")
	}

	return nil
}

// ValidatePriceDeviation проверяет отклонение цены от предыдущей
func (v *PriceValidator) ValidatePriceDeviation(currentPrice, previousPrice float64) error {
	if previousPrice <= 0 {
		return nil // Нет предыдущей цены для сравнения
	}

	deviation := math.Abs(currentPrice-previousPrice) / previousPrice * 100
	maxDeviation := 50.0 // Максимальное отклонение 50%

	if deviation > maxDeviation {
		return NewValidationError("price", currentPrice,
			fmt.Sprintf("price deviation %.2f%% exceeds maximum allowed %.2f%%",
				deviation, maxDeviation))
	}

	return nil
}

// isPriceReasonable проверяет разумность цены для символа
func (v *PriceValidator) isPriceReasonable(symbol string, price float64) bool {
	// Получаем базовую цену для символа из констант
	var basePrice float64
	switch symbol {
	case "BTCUSDT":
		basePrice = 45000.0
	case "ETHUSDT":
		basePrice = 3000.0
	case "DOGEUSDT":
		basePrice = 0.07
	case "TONUSDT":
		basePrice = 2.5
	case "SOLUSDT":
		basePrice = 100.0
	default:
		basePrice = 1000.0 // default
	}

	// Цена не должна отклоняться более чем в 10 раз от базовой
	minReasonablePrice := basePrice / 10
	maxReasonablePrice := basePrice * 10

	return price >= minReasonablePrice && price <= maxReasonablePrice
}

// isTimestampRecent проверяет актуальность timestamp
func (v *PriceValidator) isTimestampRecent(timestamp time.Time) bool {
	// Timestamp не должен быть старше 5 минут
	return time.Since(timestamp) <= 5*time.Minute
}

// PriceAggregator предоставляет методы агрегации цен
type PriceAggregator struct{}

// NewPriceAggregator создает новый агрегатор цен
func NewPriceAggregator() *PriceAggregator {
	return &PriceAggregator{}
}

// AggregatePrices агрегирует список обновлений цен
func (a *PriceAggregator) AggregatePrices(updates []PriceUpdate, exchange string, symbol string) (*AggregatedData, error) {
	if len(updates) == 0 {
		return nil, fmt.Errorf("no price updates to aggregate")
	}

	minUpdatesForAggregation := 1
	if len(updates) < minUpdatesForAggregation {
		return nil, fmt.Errorf("insufficient updates for aggregation: got %d, need at least %d",
			len(updates), minUpdatesForAggregation)
	}

	// Валидируем, что все обновления для одной биржи и символа
	for _, update := range updates {
		if update.Exchange != exchange || update.Symbol != symbol {
			return nil, fmt.Errorf("inconsistent exchange/symbol in updates")
		}
	}

	// Вычисляем статистику
	sum := 0.0
	minPrice := updates[0].Price
	maxPrice := updates[0].Price

	for _, update := range updates {
		sum += update.Price
		if update.Price < minPrice {
			minPrice = update.Price
		}
		if update.Price > maxPrice {
			maxPrice = update.Price
		}
	}

	avgPrice := sum / float64(len(updates))

	aggregated := &AggregatedData{
		PairName:     symbol,
		Exchange:     exchange,
		Timestamp:    time.Now(),
		AveragePrice: avgPrice,
		MinPrice:     minPrice,
		MaxPrice:     maxPrice,
	}

	// Валидируем результат агрегации
	if err := aggregated.Validate(); err != nil {
		return nil, fmt.Errorf("aggregation validation failed: %w", err)
	}

	return aggregated, nil
}

// CalculateStatistics вычисляет расширенную статистику по ценам
func (a *PriceAggregator) CalculateStatistics(updates []PriceUpdate, period string) (*PriceStatistics, error) {
	if len(updates) == 0 {
		return nil, fmt.Errorf("no price updates for statistics")
	}

	// Сортируем по времени (простая сортировка для небольших массивов)
	sortedUpdates := make([]PriceUpdate, len(updates))
	copy(sortedUpdates, updates)

	for i := 0; i < len(sortedUpdates)-1; i++ {
		for j := i + 1; j < len(sortedUpdates); j++ {
			if sortedUpdates[i].Timestamp.After(sortedUpdates[j].Timestamp) {
				sortedUpdates[i], sortedUpdates[j] = sortedUpdates[j], sortedUpdates[i]
			}
		}
	}

	// Вычисляем статистику
	sum := 0.0
	minPrice := sortedUpdates[0].Price
	maxPrice := sortedUpdates[0].Price

	for _, update := range sortedUpdates {
		sum += update.Price
		if update.Price < minPrice {
			minPrice = update.Price
		}
		if update.Price > maxPrice {
			maxPrice = update.Price
		}
	}

	avgPrice := sum / float64(len(sortedUpdates))
	firstUpdate := sortedUpdates[0]
	lastUpdate := sortedUpdates[len(sortedUpdates)-1]

	stats := &PriceStatistics{
		Symbol:       firstUpdate.Symbol,
		Exchange:     firstUpdate.Exchange,
		Period:       period,
		Count:        len(sortedUpdates),
		AveragePrice: avgPrice,
		MinPrice:     minPrice,
		MaxPrice:     maxPrice,
		FirstPrice:   firstUpdate.Price,
		LastPrice:    lastUpdate.Price,
		StartTime:    firstUpdate.Timestamp,
		EndTime:      lastUpdate.Timestamp,
	}

	return stats, nil
}

// ExchangeConnectionManager управляет состоянием подключений к биржам
type ExchangeConnectionManager struct {
	connections map[string]*ConnectionState
}

// ConnectionState представляет состояние подключения к бирже
type ConnectionState struct {
	Exchange           string
	Status             string // "connected", "disconnected", "reconnecting", "failed"
	LastConnectedAt    time.Time
	LastDisconnectedAt time.Time
	ConnectionAttempts int
	LastError          error
}

// NewExchangeConnectionManager создает новый менеджер подключений
func NewExchangeConnectionManager() *ExchangeConnectionManager {
	return &ExchangeConnectionManager{
		connections: make(map[string]*ConnectionState),
	}
}

// UpdateConnectionStatus обновляет статус подключения к бирже
func (m *ExchangeConnectionManager) UpdateConnectionStatus(exchange string, status string, err error) {
	state, exists := m.connections[exchange]
	if !exists {
		state = &ConnectionState{
			Exchange: exchange,
		}
		m.connections[exchange] = state
	}

	state.Status = status
	state.LastError = err

	switch status {
	case "connected":
		state.LastConnectedAt = time.Now()
		state.ConnectionAttempts = 0
	case "disconnected", "failed":
		state.LastDisconnectedAt = time.Now()
	case "reconnecting":
		state.ConnectionAttempts++
	}
}

// GetConnectionState возвращает состояние подключения к бирже
func (m *ExchangeConnectionManager) GetConnectionState(exchange string) (*ConnectionState, bool) {
	state, exists := m.connections[exchange]
	return state, exists
}

// GetAllConnectionStates возвращает состояния всех подключений
func (m *ExchangeConnectionManager) GetAllConnectionStates() map[string]*ConnectionState {
	result := make(map[string]*ConnectionState)
	for exchange, state := range m.connections {
		// Создаем копию для безопасности
		stateCopy := *state
		result[exchange] = &stateCopy
	}
	return result
}

// ShouldReconnect определяет, нужно ли переподключаться к бирже
func (m *ExchangeConnectionManager) ShouldReconnect(exchange string) bool {
	state, exists := m.connections[exchange]
	if !exists {
		return true // Первое подключение
	}

	// Не переподключаемся, если уже подключены или переподключаемся
	if state.Status == "connected" || state.Status == "reconnecting" {
		return false
	}

	// Не переподключаемся, если слишком много попыток
	maxRetries := 3
	if state.ConnectionAttempts >= maxRetries {
		return false
	}

	// Переподключаемся, если прошло достаточно времени с последней попытки
	reconnectInterval := 5 * time.Second
	timeSinceLastAttempt := time.Since(state.LastDisconnectedAt)
	return timeSinceLastAttempt >= reconnectInterval
}

// HealthChecker проверяет здоровье компонентов системы
type HealthChecker struct{}

// NewHealthChecker создает новый проверщик здоровья
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{}
}

// CheckSystemHealth проверяет общее здоровье системы
func (h *HealthChecker) CheckSystemHealth(dependencies map[string]interface{}) string {
	if len(dependencies) == 0 {
		return "down"
	}

	hasHealthyComponents := false
	hasDegradedComponents := false
	hasDownComponents := false

	for _, dep := range dependencies {
		if depMap, ok := dep.(map[string]interface{}); ok {
			if status, ok := depMap["status"].(string); ok {
				switch status {
				case "up":
					hasHealthyComponents = true
				case "degraded":
					hasDegradedComponents = true
				case "down":
					hasDownComponents = true
				}
			}
		}
	}

	// Если PostgreSQL недоступен, система не работает
	if pgDep, exists := dependencies["postgresql"]; exists {
		if pgMap, ok := pgDep.(map[string]interface{}); ok {
			if status, ok := pgMap["status"].(string); ok && status == "down" {
				return "down"
			}
		}
	}

	if hasDownComponents {
		return "degraded"
	}

	if hasDegradedComponents {
		return "degraded"
	}

	if hasHealthyComponents {
		return "ok"
	}

	return "down"
}

// CheckDependencyHealth проверяет здоровье конкретной зависимости
func (h *HealthChecker) CheckDependencyHealth(name string, checkFunc func() error) map[string]interface{} {
	result := map[string]interface{}{
		"name":       name,
		"checked_at": time.Now(),
	}

	err := checkFunc()
	if err != nil {
		result["status"] = "down"
		result["error"] = err.Error()
	} else {
		result["status"] = "up"
	}

	return result
}

// ModeManager управляет режимами работы приложения
type ModeManager struct {
	currentMode string // "live" или "test"
}

// NewModeManager создает новый менеджер режимов
func NewModeManager(initialMode string) *ModeManager {
	return &ModeManager{
		currentMode: initialMode,
	}
}

// GetCurrentMode возвращает текущий режим
func (m *ModeManager) GetCurrentMode() string {
	return m.currentMode
}

// SwitchMode переключает режим работы
func (m *ModeManager) SwitchMode(newMode string) error {
	if newMode != "live" && newMode != "test" {
		return fmt.Errorf("invalid mode: %s", newMode)
	}

	if m.currentMode == newMode {
		return nil // Уже в нужном режиме
	}

	m.currentMode = newMode
	return nil
}

// IsLiveMode проверяет, работает ли система в live режиме
func (m *ModeManager) IsLiveMode() bool {
	return m.currentMode == "live"
}

// IsTestMode проверяет, работает ли система в test режиме
func (m *ModeManager) IsTestMode() bool {
	return m.currentMode == "test"
}

// DataQualityChecker проверяет качество данных
type DataQualityChecker struct{}

// NewDataQualityChecker создает новый проверщик качества данных
func NewDataQualityChecker() *DataQualityChecker {
	return &DataQualityChecker{}
}

// CheckDataQuality проверяет качество массива обновлений цен
func (c *DataQualityChecker) CheckDataQuality(updates []PriceUpdate) error {
	if len(updates) == 0 {
		return fmt.Errorf("no data to check")
	}

	// Проверяем каждое обновление
	for i, update := range updates {
		if err := update.Validate(); err != nil {
			return fmt.Errorf("invalid update at index %d: %w", i, err)
		}
	}

	// Проверяем временную последовательность
	if err := c.checkTimeSequence(updates); err != nil {
		return fmt.Errorf("time sequence validation failed: %w", err)
	}

	// Проверяем разумность цен
	if err := c.checkPriceReasonableness(updates); err != nil {
		return fmt.Errorf("price reasonableness check failed: %w", err)
	}

	return nil
}

// checkTimeSequence проверяет временную последовательность данных
func (c *DataQualityChecker) checkTimeSequence(updates []PriceUpdate) error {
	if len(updates) < 2 {
		return nil
	}

	for i := 1; i < len(updates); i++ {
		if updates[i].Timestamp.Before(updates[i-1].Timestamp) {
			return fmt.Errorf("timestamps are not in chronological order at index %d", i)
		}

		// Проверяем, что временные промежутки разумные
		timeDiff := updates[i].Timestamp.Sub(updates[i-1].Timestamp)
		if timeDiff > 1*time.Hour {
			return fmt.Errorf("time gap too large between updates: %v", timeDiff)
		}
	}

	return nil
}

// checkPriceReasonableness проверяет разумность цен
func (c *DataQualityChecker) checkPriceReasonableness(updates []PriceUpdate) error {
	validator := NewPriceValidator()

	for i := 1; i < len(updates); i++ {
		if updates[i].Symbol == updates[i-1].Symbol {
			if err := validator.ValidatePriceDeviation(updates[i].Price, updates[i-1].Price); err != nil {
				return fmt.Errorf("price deviation check failed at index %d: %w", i, err)
			}
		}
	}

	return nil
}
