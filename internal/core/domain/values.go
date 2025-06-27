package domain

import (
	"fmt"
	"strings"
	"time"
)

// AppMode представляет режим работы приложения
type AppMode string

const (
	ModeLive AppMode = "live"
	ModeTest AppMode = "test"
)

// IsValid проверяет валидность режима
func (m AppMode) IsValid() bool {
	return m == ModeLive || m == ModeTest
}

// String возвращает строковое представление режима
func (m AppMode) String() string {
	return string(m)
}

// Symbol представляет торговую пару
type Symbol string

// Поддерживаемые торговые пары
const (
	BTCUSDT  Symbol = "BTCUSDT"
	ETHUSDT  Symbol = "ETHUSDT"
	DOGEUSDT Symbol = "DOGEUSDT"
	TONUSDT  Symbol = "TONUSDT"
	SOLUSDT  Symbol = "SOLUSDT"
)

// ValidSymbols содержит все поддерживаемые торговые пары
var ValidSymbols = map[Symbol]bool{
	BTCUSDT:  true,
	ETHUSDT:  true,
	DOGEUSDT: true,
	TONUSDT:  true,
	SOLUSDT:  true,
}

// NewSymbol создает новый Symbol с валидацией
func NewSymbol(s string) (Symbol, error) {
	symbol := Symbol(strings.ToUpper(s))
	if !symbol.IsValid() {
		return "", fmt.Errorf("%w: %s", ErrInvalidSymbol, s)
	}
	return symbol, nil
}

// IsValid проверяет, поддерживается ли торговая пара
func (s Symbol) IsValid() bool {
	return ValidSymbols[s]
}

// String возвращает строковое представление символа
func (s Symbol) String() string {
	return string(s)
}

// GetBaseCurrency возвращает базовую валюту (например, BTC для BTCUSDT)
func (s Symbol) GetBaseCurrency() string {
	str := string(s)
	if strings.HasSuffix(str, "USDT") {
		return str[:len(str)-4]
	}
	return str
}

// GetQuoteCurrency возвращает котируемую валюту (например, USDT для BTCUSDT)
func (s Symbol) GetQuoteCurrency() string {
	str := string(s)
	if strings.HasSuffix(str, "USDT") {
		return "USDT"
	}
	return ""
}

// IsValidSymbol проверяет валидность строки символа
func IsValidSymbol(s string) bool {
	symbol := Symbol(strings.ToUpper(s))
	return symbol.IsValid()
}

// GetValidSymbols возвращает список всех поддерживаемых торговых пар
func GetValidSymbols() []string {
	symbols := make([]string, 0, len(ValidSymbols))
	for symbol := range ValidSymbols {
		symbols = append(symbols, string(symbol))
	}
	return symbols
}

// GetAllSymbols возвращает все Symbol как slice
func GetAllSymbols() []Symbol {
	symbols := make([]Symbol, 0, len(ValidSymbols))
	for symbol := range ValidSymbols {
		symbols = append(symbols, symbol)
	}
	return symbols
}

// ExchangeName представляет имя биржи
type ExchangeName string

// Поддерживаемые биржи
const (
	Exchange1     ExchangeName = "exchange1"
	Exchange2     ExchangeName = "exchange2"
	Exchange3     ExchangeName = "exchange3"
	TestExchange1 ExchangeName = "TestExchange1"
	TestExchange2 ExchangeName = "TestExchange2"
	TestExchange3 ExchangeName = "TestExchange3"
	TestGenerator ExchangeName = "test-generator"
)

// NewExchangeName создает новое имя биржи
func NewExchangeName(name string) ExchangeName {
	return ExchangeName(strings.ToLower(name))
}

// String возвращает строковое представление имени биржи
func (e ExchangeName) String() string {
	return string(e)
}

// IsTestExchange проверяет, является ли биржа тестовой
func (e ExchangeName) IsTestExchange() bool {
	return strings.HasPrefix(string(e), "test") || strings.Contains(string(e), "Test")
}

// Period представляет временной период
type Period time.Duration

// Предопределенные периоды
const (
	Period1Second   Period = Period(1 * time.Second)
	Period3Seconds  Period = Period(3 * time.Second)
	Period5Seconds  Period = Period(5 * time.Second)
	Period10Seconds Period = Period(10 * time.Second)
	Period30Seconds Period = Period(30 * time.Second)
	Period1Minute   Period = Period(1 * time.Minute)
	Period3Minutes  Period = Period(3 * time.Minute)
	Period5Minutes  Period = Period(5 * time.Minute)
	Period1Hour     Period = Period(1 * time.Hour)
	Period24Hours   Period = Period(24 * time.Hour)
)

// NewPeriod создает Period из time.Duration
func NewPeriod(d time.Duration) Period {
	return Period(d)
}

// Duration возвращает time.Duration
func (p Period) Duration() time.Duration {
	return time.Duration(p)
}

// String возвращает строковое представление периода
func (p Period) String() string {
	return time.Duration(p).String()
}

// IsValid проверяет валидность периода
func (p Period) IsValid() bool {
	return time.Duration(p) > 0 && time.Duration(p) <= Period24Hours.Duration()
}

// Price представляет цену (value object для безопасности типов)
type Price float64

// NewPrice создает новую цену с валидацией
func NewPrice(value float64) (Price, error) {
	if value <= 0 {
		return 0, fmt.Errorf("%w: %f", ErrInvalidPrice, value)
	}
	return Price(value), nil
}

// Float64 возвращает значение цены как float64
func (p Price) Float64() float64 {
	return float64(p)
}

// IsPositive проверяет, положительная ли цена
func (p Price) IsPositive() bool {
	return p > 0
}

// IsZero проверяет, равна ли цена нулю
func (p Price) IsZero() bool {
	return p == 0
}

// String возвращает строковое представление цены
func (p Price) String() string {
	return fmt.Sprintf("%.8f", float64(p))
}

// Compare сравнивает две цены (-1, 0, 1)
func (p Price) Compare(other Price) int {
	if p < other {
		return -1
	}
	if p > other {
		return 1
	}
	return 0
}

// IsGreaterThan проверяет, больше ли цена другой
func (p Price) IsGreaterThan(other Price) bool {
	return p > other
}

// IsLessThan проверяет, меньше ли цена другой
func (p Price) IsLessThan(other Price) bool {
	return p < other
}

// ConnectionStatus представляет статус подключения
type ConnectionStatus string

const (
	StatusConnected    ConnectionStatus = "connected"
	StatusDisconnected ConnectionStatus = "disconnected"
	StatusReconnecting ConnectionStatus = "reconnecting"
	StatusFailed       ConnectionStatus = "failed"
)

// NewConnectionStatus создает новый статус подключения
func NewConnectionStatus(status string) ConnectionStatus {
	return ConnectionStatus(strings.ToLower(status))
}

// String возвращает строковое представление статуса
func (s ConnectionStatus) String() string {
	return string(s)
}

// IsConnected проверяет, подключен ли компонент
func (s ConnectionStatus) IsConnected() bool {
	return s == StatusConnected
}

// IsDisconnected проверяет, отключен ли компонент
func (s ConnectionStatus) IsDisconnected() bool {
	return s == StatusDisconnected
}

// IsReconnecting проверяет, переподключается ли компонент
func (s ConnectionStatus) IsReconnecting() bool {
	return s == StatusReconnecting
}

// HealthStatusValue представляет статус здоровья системы
type HealthStatusValue string

const (
	HealthOK       HealthStatusValue = "ok"
	HealthDegraded HealthStatusValue = "degraded"
	HealthDown     HealthStatusValue = "down"
)

// NewHealthStatus создает новый статус здоровья
func NewHealthStatus(status string) HealthStatusValue {
	return HealthStatusValue(strings.ToLower(status))
}

// String возвращает строковое представление статуса здоровья
func (h HealthStatusValue) String() string {
	return string(h)
}

// IsHealthy проверяет, здорова ли система
func (h HealthStatusValue) IsHealthy() bool {
	return h == HealthOK
}

// IsDegraded проверяет, работает ли система с ограничениями
func (h HealthStatusValue) IsDegraded() bool {
	return h == HealthDegraded
}

// IsDown проверяет, не работает ли система
func (h HealthStatusValue) IsDown() bool {
	return h == HealthDown
}
