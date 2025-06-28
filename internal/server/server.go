package server

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"crypto/internal/adapters/repository/postgres"
	// "your-project/internal/cache"       // TODO: раскомментировать когда готов
	// "your-project/internal/handlers/v1" // TODO: раскомментировать когда готов
	"crypto/internal/config"

	"github.com/redis/go-redis/v9"

	_ "github.com/lib/pq"
)

type App struct {
	cfg         *config.Config
	router      *http.ServeMux
	db          *sql.DB
	redisClient *redis.Client
}

func NewApp(cfg *config.Config) *App {
	return &App{
		cfg: cfg,
	}
}

func (app *App) Initialize() error {
	slog.Info("Initializing application...")
	app.router = http.NewServeMux()

	// Database connection
	dbConn, err := postgres.NewDbConnInstance(&app.cfg.Repository)
	if err != nil {
		slog.Error("Connection to database failed", "error", err)
		return err
	}
	app.db = dbConn
	slog.Info("Database connected successfully")

	// Redis connection
	redisClient := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", app.cfg.Cache.RedisHost, app.cfg.Cache.RedisPort),
		Password:     app.cfg.Cache.RedisPassword,
		DB:           app.cfg.Cache.RedisDB,
		PoolSize:     app.cfg.Cache.PoolSize,
		MinIdleConns: app.cfg.Cache.MinIdleConns,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		slog.Warn("Redis connection failed, continuing without cache", "error", err)
		// Fallback mechanism - продолжаем без Redis
		app.redisClient = nil
	} else {
		app.redisClient = redisClient
		slog.Info("Redis connected successfully")
	}

	// Initialize cache adapter
	// var cacheAdapter *cache.RedisAdapter
	// if app.redisClient != nil {
	// 	cacheAdapter = cache.NewRedisAdapter(
	// 		fmt.Sprintf("%s:%d", app.cfg.Cache.RedisHost, app.cfg.Cache.RedisPort),
	// 		app.cfg.Cache.RedisPassword,
	// 		app.cfg.Cache.RedisDB,
	// 	)
	// }

	// ВРЕМЕННО: простые роуты для проверки
	app.router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := `{"status":"ok","database":"connected","redis":"`
		if app.redisClient != nil {
			response += `connected"`
		} else {
			response += `disconnected"`
		}
		response += `}`
		w.Write([]byte(response))
	})

	app.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message":"Market Data API is running"}`))
	})
	// TODO: Добавить настоящие handlers когда будут готовы
	// marketDataRepository := postgres.NewMarketDataRepository(dbConn)
	// marketDataService := service.NewMarketDataService(marketDataRepository, nil)
	// marketDataHandler := v1.NewMarketDataHandler(marketDataService)
	// v1.SetMarketRoutes(app.router, marketDataHandler, healthHandler)

	// Start background tasks (ВРЕМЕННО УПРОЩЕНО)

	go app.startMarketDataProcessor()

	slog.Info("Application initialized successfully")
	return nil
}

func (app *App) Run() {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", app.cfg.App.Port),
		Handler: app.router,
	}

	slog.Info("Starting server", "port", app.cfg.App.Port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Server error", err)
		panic(err)
	}
}

// Background task для обработки market data (ВРЕМЕННО ОТКЛЮЧЕН)
func (app *App) startMarketDataProcessor() {
	// TODO: Добавить когда сервис будет готов
	slog.Info("Market data processor placeholder - not implemented yet")

	// Пока просто логируем каждую минуту что система работает
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			slog.Info("Market data system is running...")
		}
	}
}
