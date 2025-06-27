package server

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"crypto/internal/adapters/exchange"
	v1 "crypto/internal/adapters/handler/http/v1"
	redis "crypto/internal/adapters/redis"
	postgres "crypto/internal/adapters/repository"

	"crypto/internal/config"
	"crypto/internal/core/port"
	"crypto/internal/core/service/marketsrv"
	"crypto/pkg/concurrency"

	_ "github.com/lib/pq"
)

type App struct {
	cfg    *config.Config
	router *http.ServeMux
	server *http.Server
	db     *sql.DB

	// MarketFlow specific components
	priceService     port.PriceService
	priceRepository  port.PricesRepository
	cacheRepository  port.CacheRepository
	exchangeAdapters map[string]port.ExchangeAdapter
	dataProcessor    *concurrency.DataProcessor
	batchWriter      *concurrency.BatchWriter

	// Concurrency control
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Mode management
	currentMode port.AppMode
	modeMutex   sync.RWMutex

	// Graceful shutdown
	shutdownCh chan os.Signal
}

func NewApp(cfg *config.Config) *App {
	ctx, cancel := context.WithCancel(context.Background())

	return &App{
		cfg:              cfg,
		ctx:              ctx,
		cancel:           cancel,
		exchangeAdapters: make(map[string]port.ExchangeAdapter),
		currentMode:      port.AppMode(cfg.App.Mode),
		shutdownCh:       make(chan os.Signal, 1),
	}
}

func (app *App) Initialize() error {
	slog.Info("Initializing MarketFlow application...")
	app.router = http.NewServeMux()

	// 1. Database connection
	if err := app.initDatabase(); err != nil {
		slog.Error("Connection to database failed", "error", err)
		return err
	}
	slog.Info("Database connected successfully")

	// 2. Redis connection
	if err := app.initRedis(); err != nil {
		slog.Error("Connection to Redis failed", "error", err)
		// Redis is optional - continue without it
		slog.Warn("Continuing without Redis cache")
	} else {
		slog.Info("Redis connected successfully")
	}

	// 3. Initialize repositories
	app.priceRepository = postgres.NewPricesRepository(app.db)

	// 4. Initialize services
	app.priceService = marketsrv.NewPriceService(app.priceRepository, app.cacheRepository)

	// 5. Initialize concurrency components
	if err := app.initConcurrencyComponents(); err != nil {
		slog.Error("Failed to initialize concurrency components", "error", err)
		return err
	}
	slog.Info("Concurrency components initialized successfully")

	// 6. Initialize exchange adapters
	if err := app.initExchangeAdapters(); err != nil {
		slog.Error("Failed to initialize exchange adapters", "error", err)
		return err
	}
	slog.Info("Exchange adapters initialized successfully")

	// 7. Initialize HTTP server
	if err := app.initHTTPServer(); err != nil {
		slog.Error("Failed to initialize HTTP server", "error", err)
		return err
	}
	slog.Info("HTTP server initialized successfully")

	// 8. Start background tasks
	go app.startBackgroundTasks()

	return nil
}

func (app *App) initDatabase() error {
	dbConn, err := postgres.NewDbConnInstance(&app.cfg.Database)
	if err != nil {
		return err
	}
	app.db = dbConn
	return nil
}

func (app *App) initRedis() error {
	cacheRepo, err := redis.NewCacheRepository(&app.cfg.Redis)
	if err != nil {
		return err
	}
	app.cacheRepository = cacheRepo
	return nil
}

func (app *App) initConcurrencyComponents() error {
	slog.Info("Initializing BatchWriter and DataProcessor...")

	// 1. Create BatchWriter for PostgreSQL
	app.batchWriter = concurrency.NewBatchWriter(
		app.db,
		app.cfg.Processing.BatchSize,
		app.cfg.GetBatchTimeout(),
	)

	// 2. Create DataProcessor for aggregation
	app.dataProcessor = concurrency.NewDataProcessor(
		app.batchWriter,
		app.cacheRepository,
		app.cfg.GetAggregationInterval(),
		app.cfg.Processing.WorkersPerExchange,
	)

	return nil
}

func (app *App) initExchangeAdapters() error {
	slog.Info("Initializing exchange adapters", "mode", app.currentMode)

	if app.currentMode == "live" {
		// Live Mode: connect to provided exchange simulators
		for exchangeName, exchangeConfig := range app.cfg.Exchanges {
			adapter := exchange.NewLiveExchangeAdapter(
				exchangeName,
				exchangeConfig.Host,
				exchangeConfig.Port,
				app.cfg.GetExchangeReconnectInterval(exchangeName),
			)
			app.exchangeAdapters[exchangeName] = adapter
		}
	} else {
		// Test Mode: create synthetic data generator
		testAdapter := exchange.NewTestExchangeAdapter("test-generator")
		app.exchangeAdapters["test-generator"] = testAdapter
	}

	return nil
}

func (app *App) initHTTPServer() error {
	slog.Info("Setting up HTTP routes...")

	// Initialize handlers
	priceHandler := v1.NewPriceHandler(app.priceService)
	modeHandler := v1.NewModeHandler(app)
	healthHandler := v1.NewHealthHandler(
		app.priceRepository,
		app.cacheRepository,
		app.exchangeAdapters,
		app, // Pass app as ModeGetter since it implements GetCurrentMode()
	)

	// Set up routes
	v1.SetRoutes(app.router, priceHandler, modeHandler, healthHandler)

	app.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", app.cfg.App.Port),
		Handler:      app.router,
		ReadTimeout:  app.cfg.GetReadTimeout(),
		WriteTimeout: app.cfg.GetWriteTimeout(),
	}

	slog.Info("HTTP server configured successfully", "port", app.cfg.App.Port)
	return nil
}

func (app *App) SetupGracefulShutdown() {
	signal.Notify(app.shutdownCh, syscall.SIGINT, syscall.SIGTERM)
}

func (app *App) Run() error {
	slog.Info("Starting MarketFlow application...")

	// 1. Start concurrency components
	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		if err := app.batchWriter.Start(app.ctx); err != nil {
			slog.Error("BatchWriter error", "error", err)
		}
	}()

	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		if err := app.dataProcessor.Start(app.ctx); err != nil {
			slog.Error("DataProcessor error", "error", err)
		}
	}()

	// 2. Start exchange adapters
	for name, adapter := range app.exchangeAdapters {
		app.wg.Add(1)
		go app.runExchangeAdapter(name, adapter)
	}

	// 3. Start HTTP server
	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		slog.Info("Starting HTTP server", "port", app.cfg.App.Port)

		if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server error", "error", err)
		}
	}()

	// 4. Wait for shutdown signal
	go func() {
		<-app.shutdownCh
		slog.Info("Received shutdown signal, initiating graceful shutdown...")
		app.Shutdown()
	}()

	// 5. Wait for all components to finish
	app.wg.Wait()
	slog.Info("All components stopped")
	return nil
}

func (app *App) runExchangeAdapter(name string, adapter port.ExchangeAdapter) {
	defer app.wg.Done()

	slog.Info("Starting exchange adapter", "exchange", name)

	for {
		select {
		case <-app.ctx.Done():
			slog.Info("Stopping exchange adapter", "exchange", name)
			adapter.Disconnect()
			return
		default:
			// Connect to exchange
			if err := adapter.Connect(app.ctx); err != nil {
				slog.Error("Failed to connect to exchange", "exchange", name, "error", err)

				// Wait before reconnecting
				select {
				case <-app.ctx.Done():
					return
				case <-time.After(5 * time.Second):
					continue
				}
			}

			// Get price stream
			priceStream, errorStream := adapter.GetPriceStream(app.ctx)

			// Process data
			for {
				select {
				case <-app.ctx.Done():
					adapter.Disconnect()
					return
				case update, ok := <-priceStream:
					if !ok {
						slog.Warn("Price stream closed", "exchange", name)
						goto reconnect
					}

					// Send to DataProcessor
					if err := app.dataProcessor.ProcessPriceUpdate(update); err != nil {
						slog.Error("Failed to process price update", "exchange", name, "error", err)
					}

				case err, ok := <-errorStream:
					if !ok {
						slog.Warn("Error stream closed", "exchange", name)
						goto reconnect
					}

					slog.Error("Exchange adapter error", "exchange", name, "error", err)
					goto reconnect
				}
			}

		reconnect:
			adapter.Disconnect()
			slog.Info("Reconnecting to exchange", "exchange", name)

			// Wait before reconnecting
			select {
			case <-app.ctx.Done():
				return
			case <-time.After(5 * time.Second):
				// Continue to reconnect
			}
		}
	}
}

func (app *App) startBackgroundTasks() {
	// Cleanup ticker for Redis (удаляем старые данные каждые 5 минут)
	cleanupTicker := time.NewTicker(5 * time.Minute)
	defer cleanupTicker.Stop()

	// Health check ticker (проверяем состояние каждые 30 секунд)
	healthTicker := time.NewTicker(30 * time.Second)
	defer healthTicker.Stop()

	for {
		select {
		case <-app.ctx.Done():
			slog.Info("Background tasks stopping")
			return
		case <-cleanupTicker.C:
			if app.cacheRepository != nil {
				cutoff := time.Now().Add(-2 * time.Minute) // Удаляем данные старше 2 минут
				if err := app.cacheRepository.CleanupOldPrices(app.ctx, cutoff); err != nil {
					slog.Error("Failed to cleanup old prices from Redis", "error", err)
				}
			}
		case <-healthTicker.C:
			// Периодическая проверка состояния компонентов
			app.performHealthCheck()
		}
	}
}

func (app *App) performHealthCheck() {
	// Check database
	if err := app.priceRepository.HealthCheck(app.ctx); err != nil {
		slog.Error("Database health check failed", "error", err)
	}

	// Check Redis
	if app.cacheRepository != nil {
		if err := app.cacheRepository.HealthCheck(app.ctx); err != nil {
			slog.Error("Redis health check failed", "error", err)
		}
	}

	// Check exchanges
	for name, adapter := range app.exchangeAdapters {
		if !adapter.IsConnected() {
			slog.Warn("Exchange adapter disconnected", "exchange", name)
		}
	}
}

// Mode management methods
func (app *App) SwitchToLiveMode() error {
	app.modeMutex.Lock()
	defer app.modeMutex.Unlock()

	if app.currentMode == "live" {
		return nil // Already in live mode
	}

	slog.Info("Switching to live mode...")

	// Stop current adapters
	for _, adapter := range app.exchangeAdapters {
		adapter.Disconnect()
	}

	// Wait a bit for disconnection
	time.Sleep(1 * time.Second)

	// Clear adapters
	app.exchangeAdapters = make(map[string]port.ExchangeAdapter)

	// Create live adapters
	for exchangeName, exchangeConfig := range app.cfg.Exchanges {
		adapter := exchange.NewLiveExchangeAdapter(
			exchangeName,
			exchangeConfig.Host,
			exchangeConfig.Port,
			app.cfg.GetExchangeReconnectInterval(exchangeName),
		)
		app.exchangeAdapters[exchangeName] = adapter
	}

	// Start new adapters
	for name, adapter := range app.exchangeAdapters {
		app.wg.Add(1)
		go app.runExchangeAdapter(name, adapter)
	}

	app.currentMode = "live"
	slog.Info("Successfully switched to live mode")
	return nil
}

func (app *App) SwitchToTestMode() error {
	app.modeMutex.Lock()
	defer app.modeMutex.Unlock()

	if app.currentMode == "test" {
		return nil // Already in test mode
	}

	slog.Info("Switching to test mode...")

	// Stop current adapters
	for _, adapter := range app.exchangeAdapters {
		adapter.Disconnect()
	}

	// Wait a bit for disconnection
	time.Sleep(1 * time.Second)

	// Clear adapters
	app.exchangeAdapters = make(map[string]port.ExchangeAdapter)

	// Create test adapter
	testAdapter := exchange.NewTestExchangeAdapter("test-generator")
	app.exchangeAdapters["test-generator"] = testAdapter

	// Start test adapter
	app.wg.Add(1)
	go app.runExchangeAdapter("test-generator", testAdapter)

	app.currentMode = "test"
	slog.Info("Successfully switched to test mode")
	return nil
}

func (app *App) GetCurrentMode() port.AppMode {
	app.modeMutex.RLock()
	defer app.modeMutex.RUnlock()
	return app.currentMode
}

// Graceful shutdown
func (app *App) Shutdown() {
	slog.Info("Shutting down MarketFlow application...")

	// Stop HTTP server
	if app.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := app.server.Shutdown(ctx); err != nil {
			slog.Error("HTTP server shutdown error", "error", err)
		}
	}

	// Cancel context to stop all goroutines
	app.cancel()

	// Stop exchange adapters
	for _, adapter := range app.exchangeAdapters {
		adapter.Disconnect()
	}

	// Stop concurrency components
	if app.dataProcessor != nil {
		app.dataProcessor.Stop()
	}
	if app.batchWriter != nil {
		app.batchWriter.Stop()
	}

	// Close database
	if app.db != nil {
		app.db.Close()
	}

	slog.Info("MarketFlow application shutdown completed")
}
