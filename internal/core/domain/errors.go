package domain

import (
	"errors"
	"fmt"
	"time"
)

// Базовые доменные ошибки
var (
	// Валидация данных
	ErrInvalidSymbol     = errors.New("invalid symbol")
	ErrInvalidPrice      = errors.New("invalid price")
	ErrInvalidMode       = errors.New("invalid mode")
	ErrInvalidPeriod     = errors.New("invalid period")
	ErrInvalidPriceRange = errors.New("invalid price range")

	// Пустые значения
	ErrEmptySymbol    = errors.New("symbol cannot be empty")
	ErrEmptyExchange  = errors.New("exchange cannot be empty")
	ErrEmptyTimestamp = errors.New("timestamp cannot be empty")

	// Отсутствие данных
	ErrPriceNotFound    = errors.New("price not found")
	ErrExchangeNotFound = errors.New("exchange not found")
	ErrDataNotAvailable = errors.New("data not available")

	// Подключения и сеть
	ErrConnectionFailed  = errors.New("connection failed")
	ErrConnectionTimeout = errors.New("connection timeout")
	ErrDisconnected      = errors.New("disconnected")
	ErrReconnectFailed   = errors.New("reconnect failed")

	// Обработка данных
	ErrProcessingFailed  = errors.New("processing failed")
	ErrBufferFull        = errors.New("buffer is full")
	ErrQueueFull         = errors.New("queue is full")
	ErrWorkerPoolStopped = errors.New("worker pool stopped")

	// Кеш и хранилище
	ErrCacheUnavailable    = errors.New("cache unavailable")
	ErrDatabaseUnavailable = errors.New("database unavailable")
	ErrStorageFailed       = errors.New("storage operation failed")

	// Конфигурация
	ErrInvalidConfig  = errors.New("invalid configuration")
	ErrConfigNotFound = errors.New("configuration not found")
	ErrMissingConfig  = errors.New("missing required configuration")

	// Сервис недоступен
	ErrServiceUnavailable = errors.New("service unavailable")
	ErrServiceTimeout     = errors.New("service timeout")
	ErrServiceOverloaded  = errors.New("service overloaded")

	// Режимы работы
	ErrModeNotSupported = errors.New("mode not supported")
	ErrModeSwitchFailed = errors.New("mode switch failed")

	// Аутентификация и авторизация (на будущее)
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")

	// Лимиты и ограничения
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	ErrQuotaExceeded     = errors.New("quota exceeded")
)

// DomainError представляет доменную ошибку с контекстом
type DomainError struct {
	Code    string
	Message string
	Cause   error
	Context map[string]interface{}
}

// NewDomainError создает новую доменную ошибку
func NewDomainError(code, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Context: make(map[string]interface{}),
	}
}

// NewDomainErrorWithCause создает доменную ошибку с причиной
func NewDomainErrorWithCause(code, message string, cause error) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Cause:   cause,
		Context: make(map[string]interface{}),
	}
}

// Error реализует интерфейс error
func (e *DomainError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap возвращает причину ошибки
func (e *DomainError) Unwrap() error {
	return e.Cause
}

// WithContext добавляет контекст к ошибке
func (e *DomainError) WithContext(key string, value interface{}) *DomainError {
	e.Context[key] = value
	return e
}

// GetContext возвращает значение из контекста
func (e *DomainError) GetContext(key string) interface{} {
	return e.Context[key]
}

// ValidationError представляет ошибку валидации
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

// NewValidationError создает новую ошибку валидации
func NewValidationError(field string, value interface{}, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}

// Error реализует интерфейс error
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s' with value '%v': %s",
		e.Field, e.Value, e.Message)
}

// ConnectionError представляет ошибку подключения
type ConnectionError struct {
	Exchange string
	Host     string
	Port     int
	Cause    error
}

// NewConnectionError создает новую ошибку подключения
func NewConnectionError(exchange, host string, port int, cause error) *ConnectionError {
	return &ConnectionError{
		Exchange: exchange,
		Host:     host,
		Port:     port,
		Cause:    cause,
	}
}

// Error реализует интерфейс error
func (e *ConnectionError) Error() string {
	address := fmt.Sprintf("%s:%d", e.Host, e.Port)
	if e.Cause != nil {
		return fmt.Sprintf("connection error for exchange '%s' at %s: %v",
			e.Exchange, address, e.Cause)
	}
	return fmt.Sprintf("connection error for exchange '%s' at %s", e.Exchange, address)
}

// Unwrap возвращает причину ошибки
func (e *ConnectionError) Unwrap() error {
	return e.Cause
}

// ProcessingError представляет ошибку обработки данных
type ProcessingError struct {
	Stage     string
	Data      interface{}
	Cause     error
	Timestamp time.Time
}

// NewProcessingError создает новую ошибку обработки
func NewProcessingError(stage string, data interface{}, cause error) *ProcessingError {
	return &ProcessingError{
		Stage:     stage,
		Data:      data,
		Cause:     cause,
		Timestamp: time.Now(),
	}
}

// Error реализует интерфейс error
func (e *ProcessingError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("processing error at stage '%s': %v", e.Stage, e.Cause)
	}
	return fmt.Sprintf("processing error at stage '%s'", e.Stage)
}

// Unwrap возвращает причину ошибки
func (e *ProcessingError) Unwrap() error {
	return e.Cause
}

// Вспомогательные функции для создания типичных ошибок

// ErrSymbolNotSupported создает ошибку неподдерживаемого символа
func ErrSymbolNotSupported(symbol string) error {
	return NewValidationError("symbol", symbol, "symbol not supported")
}

// ErrExchangeConnectionFailed создает ошибку подключения к бирже
func ErrExchangeConnectionFailed(exchange, host string, port int, cause error) error {
	return NewConnectionError(exchange, host, port, cause)
}

// ErrDataProcessingFailed создает ошибку обработки данных
func ErrDataProcessingFailed(stage string, data interface{}, cause error) error {
	return NewProcessingError(stage, data, cause)
}

// ErrPriceValidationFailed создает ошибку валидации цены
func ErrPriceValidationFailed(price float64, reason string) error {
	return NewValidationError("price", price, reason)
}

// ErrPeriodValidationFailed создает ошибку валидации периода
func ErrPeriodValidationFailed(period string, reason string) error {
	return NewValidationError("period", period, reason)
}

// IsValidationError проверяет, является ли ошибка ошибкой валидации
func IsValidationError(err error) bool {
	var validationErr *ValidationError
	return errors.As(err, &validationErr)
}

// IsConnectionError проверяет, является ли ошибка ошибкой подключения
func IsConnectionError(err error) bool {
	var connErr *ConnectionError
	return errors.As(err, &connErr)
}

// IsProcessingError проверяет, является ли ошибка ошибкой обработки
func IsProcessingError(err error) bool {
	var procErr *ProcessingError
	return errors.As(err, &procErr)
}

// IsDomainError проверяет, является ли ошибка доменной ошибкой
func IsDomainError(err error) bool {
	var domainErr *DomainError
	return errors.As(err, &domainErr)
}

// WrapError оборачивает ошибку в доменную ошибку
func WrapError(code, message string, cause error) error {
	return NewDomainErrorWithCause(code, message, cause)
}
