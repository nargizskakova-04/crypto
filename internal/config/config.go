package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	App        App                 `json:"app"`
	Database   Database            `json:"database"`
	Redis      Redis               `json:"redis"`
	Exchanges  map[string]Exchange `json:"exchanges"`
	Processing Processing          `json:"processing"`
}

type App struct {
	Port         int    `json:"port"`
	Mode         string `json:"mode"` // "live" или "test"
	ReadTimeout  string `json:"read_timeout"`
	WriteTimeout string `json:"write_timeout"`
}

type Database struct {
	Host               string `json:"host"`
	Port               int    `json:"port"`
	Name               string `json:"name"`
	Username           string `json:"username"`
	Password           string `json:"password"`
	SSLMode            string `json:"ssl_mode"`
	MaxConnections     int    `json:"max_connections"`
	MaxIdleConnections int    `json:"max_idle_connections"`
}

type Redis struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
	TTL      string `json:"ttl"`
}

type Exchange struct {
	Host              string `json:"host"`
	Port              int    `json:"port"`
	ReconnectInterval string `json:"reconnect_interval"`
}

type Processing struct {
	BatchSize           int    `json:"batch_size"`
	BatchTimeout        string `json:"batch_timeout"`
	WorkersPerExchange  int    `json:"workers_per_exchange"`
	AggregationInterval string `json:"aggregation_interval"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	// Override with environment variables
	overrideFromEnv(&cfg)

	return &cfg, nil
}

func overrideFromEnv(cfg *Config) {
	// App
	if port := os.Getenv("APP_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.App.Port = p
		}
	}
	if mode := os.Getenv("APP_MODE"); mode != "" {
		cfg.App.Mode = mode
	}

	// Database
	if host := os.Getenv("DB_HOST"); host != "" {
		cfg.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Database.Port = p
		}
	}
	if user := os.Getenv("DB_USER"); user != "" {
		cfg.Database.Username = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		cfg.Database.Password = password
	}
	if name := os.Getenv("DB_NAME"); name != "" {
		cfg.Database.Name = name
	}

	// Redis
	if host := os.Getenv("REDIS_HOST"); host != "" {
		cfg.Redis.Host = host
	}
	if port := os.Getenv("REDIS_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Redis.Port = p
		}
	}
	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		cfg.Redis.Password = password
	}
}

// Helper methods

func (cfg *Config) GetDatabaseDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)
}

func (cfg *Config) GetRedisAddress() string {
	return fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
}

func (cfg *Config) GetReadTimeout() time.Duration {
	if d, err := time.ParseDuration(cfg.App.ReadTimeout); err == nil {
		return d
	}
	return 30 * time.Second // default
}

func (cfg *Config) GetWriteTimeout() time.Duration {
	if d, err := time.ParseDuration(cfg.App.WriteTimeout); err == nil {
		return d
	}
	return 30 * time.Second // default
}

func (cfg *Config) GetRedisTTL() time.Duration {
	if d, err := time.ParseDuration(cfg.Redis.TTL); err == nil {
		return d
	}
	return 60 * time.Second // default
}

func (cfg *Config) GetBatchTimeout() time.Duration {
	if d, err := time.ParseDuration(cfg.Processing.BatchTimeout); err == nil {
		return d
	}
	return 30 * time.Second // default
}

func (cfg *Config) GetAggregationInterval() time.Duration {
	if d, err := time.ParseDuration(cfg.Processing.AggregationInterval); err == nil {
		return d
	}
	return 1 * time.Minute // default
}

func (cfg *Config) GetExchangeReconnectInterval(exchangeName string) time.Duration {
	if exchange, ok := cfg.Exchanges[exchangeName]; ok {
		if d, err := time.ParseDuration(exchange.ReconnectInterval); err == nil {
			return d
		}
	}
	return 5 * time.Second // default
}
