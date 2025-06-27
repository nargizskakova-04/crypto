package domain

import "time"

// Application constants
const (
	AppName    = "MarketFlow"
	AppVersion = "1.0.0"
)

// Default configuration values
const (
	DefaultPort            = 8080
	DefaultMode            = "test"
	DefaultReadTimeout     = 30 * time.Second
	DefaultWriteTimeout    = 30 * time.Second
	DefaultShutdownTimeout = 30 * time.Second
)

// Database constants
const (
	DefaultMaxConnections     = 25
	DefaultMaxIdleConnections = 10
	DefaultConnMaxLifetime    = 30 * time.Minute
	DefaultConnectionTimeout  = 10 * time.Second
)

// Redis constants
const (
	DefaultRedisTTL     = 60 * time.Second
	DefaultRedisTimeout = 5 * time.Second
	DefaultRedisDB      = 0
)

// Processing constants
const (
	DefaultBatchSize           = 50
	DefaultBatchTimeout        = 30 * time.Second
	DefaultWorkersPerExchange  = 5
	DefaultAggregationInterval = 1 * time.Minute
	DefaultCleanupInterval     = 5 * time.Minute
	DefaultHealthCheckInterval = 30 * time.Second
)

// Exchange connection constants
const (
	DefaultExchangeHost      = "localhost"
	DefaultReconnectInterval = 5 * time.Second
	DefaultConnectionRetries = 3
	DefaultStreamBufferSize  = 100
	DefaultErrorBufferSize   = 10
)

// Data retention constants
const (
	MaxPriceHistoryAge = 2 * time.Minute  // Храним в памяти данные за последние 2 минуты
	MaxCacheAge        = 60 * time.Second // TTL для Redis кеша
	DefaultQueryPeriod = 24 * time.Hour   // Период по умолчанию для API запросов
)

// API constants
const (
	MaxPeriodDuration = 7 * 24 * time.Hour // Максимальный период для API запросов
	MinPeriodDuration = 1 * time.Second    // Минимальный период для API запросов
	DefaultAPITimeout = 30 * time.Second   // Таймаут для API запросов
)

// Buffer and queue sizes
const (
	DefaultChannelBufferSize    = 1000
	DefaultWorkerQueueSize      = 1000
	DefaultBatchWriterQueueSize = 100
	MaxConcurrentConnections    = 1000
)

// Validation limits
const (
	MinPrice              = 0.00000001 // Минимальная цена (1 сатоши в BTC)
	MaxPrice              = 1000000000 // Максимальная цена (1 млрд)
	MaxSymbolLength       = 20
	MaxExchangeNameLength = 50
	MaxBatchSize          = 1000
	MaxWorkersPerExchange = 100
)

// HTTP response constants
const (
	HTTPTimeoutDuration = 30 * time.Second
	MaxRequestBodySize  = 1024 * 1024 // 1MB
)

// Logging constants
const (
	LogFieldSymbol    = "symbol"
	LogFieldExchange  = "exchange"
	LogFieldPrice     = "price"
	LogFieldTimestamp = "timestamp"
	LogFieldWorkerID  = "worker_id"
	LogFieldBatchSize = "batch_size"
	LogFieldError     = "error"
	LogFieldDuration  = "duration"
	LogFieldMode      = "mode"
	LogFieldStatus    = "status"
)

// Key prefixes for Redis
const (
	RedisKeyPrefix          = "marketflow:"
	RedisLatestPricePrefix  = RedisKeyPrefix + "latest_price:"
	RedisPriceHistoryPrefix = RedisKeyPrefix + "price_history:"
	RedisHealthPrefix       = RedisKeyPrefix + "health:"
	RedisStatsPrefix        = RedisKeyPrefix + "stats:"
)

// Business rules constants
const (
	// Минимальное количество обновлений для валидной агрегации
	MinUpdatesForAggregation = 1

	// Максимальное отклонение цены от предыдущей (в процентах)
	MaxPriceDeviationPercent = 50.0

	// Время жизни подключения к бирже до переподключения
	MaxConnectionLifetime = 1 * time.Hour

	// Максимальное время без данных от биржи
	MaxDataSilence = 5 * time.Minute
)

// Test mode constants
const (
	TestDataGenerationInterval = 200 * time.Millisecond
	TestPriceVolatility        = 0.05 // 5% максимальное изменение цены
	TestMaxPriceDeviation      = 0.20 // 20% максимальное отклонение от базовой цены
)

// Exchange port constants (for live mode)
const (
	Exchange1Port = 40101
	Exchange2Port = 40102
	Exchange3Port = 40103
)

// Base prices for test mode (in USDT) - ИСПРАВЛЕНО
var BasePrices = map[Symbol]float64{
	BTCUSDT:  45000.0,
	ETHUSDT:  3000.0,
	DOGEUSDT: 0.07,
	TONUSDT:  2.5,
	SOLUSDT:  100.0,
}

// Exchange names for test mode - ИСПРАВЛЕНО
var TestExchangeNames = []ExchangeName{
	"TestExchange1",
	"TestExchange2",
	"TestExchange3",
}

// Live exchange names - ИСПРАВЛЕНО
var LiveExchangeNames = []ExchangeName{
	"exchange1",
	"exchange2",
	"exchange3",
}

// Supported periods for API queries
var SupportedPeriods = map[string]time.Duration{
	"1s":  1 * time.Second,
	"3s":  3 * time.Second,
	"5s":  5 * time.Second,
	"10s": 10 * time.Second,
	"30s": 30 * time.Second,
	"1m":  1 * time.Minute,
	"3m":  3 * time.Minute,
	"5m":  5 * time.Minute,
	"15m": 15 * time.Minute,
	"30m": 30 * time.Minute,
	"1h":  1 * time.Hour,
	"4h":  4 * time.Hour,
	"12h": 12 * time.Hour,
	"24h": 24 * time.Hour,
}

// HTTP status messages
var StatusMessages = map[int]string{
	200: "OK",
	400: "Bad Request",
	404: "Not Found",
	405: "Method Not Allowed",
	500: "Internal Server Error",
	503: "Service Unavailable",
}

// Health check thresholds
const (
	HealthCheckCriticalErrorThreshold = 10 // Критичное количество ошибок за период
	HealthCheckWarningErrorThreshold  = 5  // Предупреждение при количестве ошибок
	HealthCheckPeriod                 = 1 * time.Minute
)

// Rate limiting (for future use)
const (
	DefaultRateLimit        = 100 // requests per minute
	DefaultBurstLimit       = 20  // burst requests
	RateLimitWindowDuration = 1 * time.Minute
)

// Metrics constants (for future monitoring)
const (
	MetricPriceUpdatesTotal = "price_updates_total"
	MetricProcessingLatency = "processing_latency_seconds"
	MetricActiveConnections = "active_connections"
	MetricBatchWrites       = "batch_writes_total"
	MetricCacheHits         = "cache_hits_total"
	MetricCacheMisses       = "cache_misses_total"
	MetricErrors            = "errors_total"
)

// Helper functions - ОБНОВЛЕНЫ

// GetBasePriceForSymbol возвращает базовую цену для символа (для test mode)
func GetBasePriceForSymbol(symbol Symbol) float64 {
	if price, exists := BasePrices[symbol]; exists {
		return price
	}
	return 1000.0 // default price
}

// GetBasePriceForSymbolString возвращает базовую цену для символа (строковая версия)
func GetBasePriceForSymbolString(symbolStr string) float64 {
	symbol, err := NewSymbol(symbolStr)
	if err != nil {
		return 1000.0 // default price
	}
	return GetBasePriceForSymbol(symbol)
}

// IsSupportedPeriod проверяет, поддерживается ли период
func IsSupportedPeriod(periodStr string) bool {
	_, exists := SupportedPeriods[periodStr]
	return exists
}

// ParsePeriod парсит строку периода в Duration
func ParsePeriod(periodStr string) (time.Duration, bool) {
	duration, exists := SupportedPeriods[periodStr]
	return duration, exists
}

// GetAllSupportedPeriods возвращает все поддерживаемые периоды
func GetAllSupportedPeriods() []string {
	periods := make([]string, 0, len(SupportedPeriods))
	for period := range SupportedPeriods {
		periods = append(periods, period)
	}
	return periods
}

// GetAllTestExchangeNames возвращает имена тестовых бирж как строки
func GetAllTestExchangeNames() []string {
	names := make([]string, len(TestExchangeNames))
	for i, name := range TestExchangeNames {
		names[i] = name.String()
	}
	return names
}

// GetAllLiveExchangeNames возвращает имена live бирж как строки
func GetAllLiveExchangeNames() []string {
	names := make([]string, len(LiveExchangeNames))
	for i, name := range LiveExchangeNames {
		names[i] = name.String()
	}
	return names
}

// IsValidBatchSize проверяет валидность размера батча
func IsValidBatchSize(size int) bool {
	return size > 0 && size <= MaxBatchSize
}

// IsValidWorkerCount проверяет валидность количества воркеров
func IsValidWorkerCount(count int) bool {
	return count > 0 && count <= MaxWorkersPerExchange
}

// IsValidPrice проверяет валидность цены (float64 версия)
func IsValidPrice(price float64) bool {
	return price >= MinPrice && price <= MaxPrice
}

// IsValidPriceValue проверяет валидность цены (Price версия)
func IsValidPriceValue(price Price) bool {
	return price.IsPositive() && price.Float64() <= MaxPrice
}

// IsRecentTimestamp проверяет, недавний ли timestamp
func IsRecentTimestamp(timestamp time.Time, maxAge time.Duration) bool {
	return time.Since(timestamp) <= maxAge
}
