package port

import (
	"context"
	"time"

	"crypto/internal/core/domain"
)

// AppMode тип для режима приложения
type AppMode string

const (
	ModeLive AppMode = "live"
	ModeTest AppMode = "test"
)

// PricesRepository интерфейс для работы с PostgreSQL (хранение агрегированных данных)
type PricesRepository interface {
	// Последние цены из агрегированных данных
	GetLatestPrice(ctx context.Context, symbol string) (*domain.PriceResponse, error)
	GetLatestExchangePrice(ctx context.Context, symbol, exchange string) (*domain.PriceResponse, error)

	// Максимальные цены за период
	GetHighestPrice(ctx context.Context, symbol string) (*domain.PriceResponse, error)
	GetHighestPriceExchange(ctx context.Context, symbol, exchange string) (*domain.PriceResponse, error)
	GetHighestPriceInDuration(ctx context.Context, symbol string, from, to time.Time) (*domain.PriceResponse, error)
	GetHighestPriceInDurationExchange(ctx context.Context, symbol, exchange string, from, to time.Time) (*domain.PriceResponse, error)

	// Минимальные цены за период
	GetLowestPrice(ctx context.Context, symbol string) (*domain.PriceResponse, error)
	GetLowestPriceExchange(ctx context.Context, symbol, exchange string) (*domain.PriceResponse, error)
	GetLowestPriceInDuration(ctx context.Context, symbol string, from, to time.Time) (*domain.PriceResponse, error)
	GetLowestPriceInDurationExchange(ctx context.Context, symbol, exchange string, from, to time.Time) (*domain.PriceResponse, error)

	// Средние цены за период
	GetAveragePrice(ctx context.Context, symbol string) (*domain.PriceResponse, error)
	GetAveragePriceExchange(ctx context.Context, symbol, exchange string) (*domain.PriceResponse, error)
	GetAveragePriceInDuration(ctx context.Context, symbol string, from, to time.Time) (*domain.PriceResponse, error)
	GetAveragePriceInDurationExchange(ctx context.Context, symbol, exchange string, from, to time.Time) (*domain.PriceResponse, error)

	// Сохранение агрегированных данных (батчинг)
	SaveAggregatedData(ctx context.Context, data []domain.AggregatedData) error

	// Получение статистики
	GetPriceStatistics(ctx context.Context, symbol, exchange string, from, to time.Time) (*domain.PriceStatistics, error)

	// Проверка состояния
	HealthCheck(ctx context.Context) error
}

// CacheRepository интерфейс для работы с Redis (кеширование последних цен)
type CacheRepository interface {
	// Сохранение последних цен в реальном времени
	SetLatestPrice(ctx context.Context, update domain.PriceUpdate) error

	// Получение последних цен из кеша
	GetLatestPrice(ctx context.Context, symbol string) (*domain.PriceResponse, error)
	GetLatestExchangePrice(ctx context.Context, symbol, exchange string) (*domain.PriceResponse, error)

	// Получение исторических данных для агрегации (последние N минут)
	GetPricesInTimeRange(ctx context.Context, symbol, exchange string, from, to time.Time) ([]domain.PriceUpdate, error)

	// Управление TTL и очистка
	CleanupOldPrices(ctx context.Context, olderThan time.Time) error
	SetTTL(ctx context.Context, key string, ttl time.Duration) error

	// Проверка состояния
	HealthCheck(ctx context.Context) error
}

// ExchangeAdapter интерфейс для подключения к биржам (Live и Test режимы)
type ExchangeAdapter interface {
	// Подключение к бирже
	Connect(ctx context.Context) error

	// Получение потока данных о ценах
	GetPriceStream(ctx context.Context) (<-chan domain.PriceUpdate, <-chan error)

	// Отключение от биржи
	Disconnect() error

	// Информация о бирже
	GetExchangeInfo() domain.ExchangeInfo

	// Проверка состояния подключения
	IsConnected() bool
}

// DataProcessor интерфейс для обработки данных (конкурентность и агрегация)
type DataProcessor interface {
	// Запуск обработки данных
	Start(ctx context.Context) error

	// Обработка одного обновления цены
	ProcessPriceUpdate(update domain.PriceUpdate) error

	// Остановка обработки
	Stop() error

	// Получение канала для результатов агрегации
	GetAggregationResults() <-chan domain.AggregatedData

	// Статистика обработки
	GetProcessingStats() map[string]interface{}
}

// PriceService интерфейс для бизнес-логики работы с ценами
type PriceService interface {
	// API методы для получения цен
	GetLatestPrice(ctx context.Context, symbol string) (*domain.PriceResponse, error)
	GetLatestExchangePrice(ctx context.Context, symbol, exchange string) (*domain.PriceResponse, error)

	// Получение экстремальных цен за период
	GetHighestPrice(ctx context.Context, symbol string, period time.Duration) (*domain.PriceResponse, error)
	GetHighestExchangePrice(ctx context.Context, symbol, exchange string, period time.Duration) (*domain.PriceResponse, error)

	GetLowestPrice(ctx context.Context, symbol string, period time.Duration) (*domain.PriceResponse, error)
	GetLowestExchangePrice(ctx context.Context, symbol, exchange string, period time.Duration) (*domain.PriceResponse, error)

	// Получение средних цен за период
	GetAveragePrice(ctx context.Context, symbol string, period time.Duration) (*domain.PriceResponse, error)
	GetAverageExchangePrice(ctx context.Context, symbol, exchange string, period time.Duration) (*domain.PriceResponse, error)

	// Получение статистики
	GetPriceStatistics(ctx context.Context, symbol, exchange string, period time.Duration) (*domain.PriceStatistics, error)
}

// ModeManager интерфейс для управления режимами работы (Live/Test)
type ModeManager interface {
	// Получение текущего режима
	GetCurrentMode() AppMode

	// Переключение режимов
	SwitchToLiveMode(ctx context.Context) error
	SwitchToTestMode(ctx context.Context) error

	// Проверка режима
	IsInLiveMode() bool
	IsInTestMode() bool

	// Получение информации о режиме
	GetModeInfo() map[string]interface{}
}

// HealthChecker интерфейс для проверки состояния системы
type HealthChecker interface {
	// Проверка состояния всей системы
	CheckHealth(ctx context.Context) (*domain.HealthStatus, error)

	// Проверка состояния отдельных компонентов
	CheckDatabase(ctx context.Context) error
	CheckRedis(ctx context.Context) error
	CheckExchanges(ctx context.Context) map[string]error

	// Получение детальной информации о состоянии
	GetDetailedHealth(ctx context.Context) map[string]interface{}
}

// ConfigManager интерфейс для управления конфигурацией
type ConfigManager interface {
	// Получение конфигурации
	GetDatabaseConfig() map[string]interface{}
	GetRedisConfig() map[string]interface{}
	GetExchangeConfig(exchange string) map[string]interface{}
	GetProcessingConfig() map[string]interface{}

	// Обновление конфигурации в runtime (опционально)
	UpdateConfig(key string, value interface{}) error

	// Валидация конфигурации
	ValidateConfig() error
}

// BatchWriter интерфейс для батчевой записи в PostgreSQL
type BatchWriter interface {
	// Добавление данных в батч
	AddData(data domain.AggregatedData) error

	// Принудительная запись батча
	FlushBatch() error

	// Запуск/остановка
	Start(ctx context.Context) error
	Stop() error

	// Статистика батчинга
	GetBatchStats() map[string]interface{}
}

// WorkerPool интерфейс для пула воркеров
type WorkerPool interface {
	// Управление воркерами
	Start() error
	Stop() error

	// Отправка задач
	SubmitTask(task interface{}) error

	// Получение статистики
	GetWorkerStats() map[string]interface{}

	// Масштабирование
	ScaleWorkers(count int) error
}

// EventPublisher интерфейс для публикации событий (опционально, для будущего)
type EventPublisher interface {
	// Публикация событий
	PublishPriceUpdate(update domain.PriceUpdate) error
	PublishModeChange(oldMode, newMode AppMode) error
	PublishHealthChange(status domain.HealthStatus) error

	// Подписка на события
	Subscribe(eventType string, handler func(interface{}) error) error
	Unsubscribe(eventType string) error
}

// MetricsCollector интерфейс для сбора метрик (опционально, для будущего)
type MetricsCollector interface {
	// Счетчики
	IncrementCounter(name string, tags map[string]string)

	// Гистограммы
	RecordDuration(name string, duration time.Duration, tags map[string]string)

	// Gauge метрики
	SetGauge(name string, value float64, tags map[string]string)

	// Получение метрик
	GetMetrics() map[string]interface{}
}

// Logger интерфейс для логирования (опционально, если нужно абстрагироваться от slog)
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, err error, args ...interface{})
	Warn(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

// Validator интерфейс для валидации данных
type Validator interface {
	// Валидация входных данных
	ValidatePriceUpdate(update domain.PriceUpdate) error
	ValidateSymbol(symbol string) error
	ValidateExchange(exchange string) error
	ValidatePeriod(period time.Duration) error

	// Валидация бизнес-правил
	ValidatePriceDeviation(current, previous float64) error
	ValidateDataQuality(updates []domain.PriceUpdate) error
}

// RateLimiter интерфейс для ограничения скорости запросов (опционально)
type RateLimiter interface {
	// Проверка лимитов
	Allow(ctx context.Context, key string) (bool, error)

	// Получение информации о лимитах
	GetLimitInfo(key string) map[string]interface{}

	// Настройка лимитов
	SetLimit(key string, limit int, window time.Duration) error
}

// Cache интерфейс для общего кеширования (опционально, более общий чем CacheRepository)
type Cache interface {
	// Базовые операции
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error

	// Bulk операции
	MGet(ctx context.Context, keys []string) (map[string]interface{}, error)
	MSet(ctx context.Context, data map[string]interface{}, ttl time.Duration) error

	// Управление TTL
	Expire(ctx context.Context, key string, ttl time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Очистка
	FlushAll(ctx context.Context) error
}
