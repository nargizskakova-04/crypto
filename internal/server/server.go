package server

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"crypto/internal/adapters/repository/postgres"
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

	// Database connection
	// dbConn, err := postgres.NewDbConnInstance(&app.cfg.Repository)
	// if err != nil {
	// 	app.logger.Error("Connection to db failed", err)
	// 	return err
	// }
	// app.db = dbConn
	// app.logger.Info("Database connected successfully")

	// Initialize repositories
	// userRepository := postgres.NewUserRepository(dbConn)

	// Initialize Rick and Morty API client
	// rickAPIImpl := rickAPI.NewRickAPI(app.logger)

	// Initialize S3 storage with config values
	// s3Config := app.cfg.S3
	// app.logger.Info("Initializing S3 storage",
	// 	"endpoint", s3Config.Endpoint,
	// 	"post_bucket", s3Config.PostBucket,
	// 	"comment_bucket", s3Config.CommentBucket)

	// s3Storage := storage.NewS3Storage(
	// 	s3Config.Endpoint,
	// 	s3Config.AccessKey,
	// 	s3Config.SecretKey,
	// 	s3Config.PostBucket,
	// 	s3Config.CommentBucket,
	// 	s3Config.UseSSL,
	// 	app.logger,
	// )

	// Initialize services
	// userService := usersrv.NewUserService(userRepository, app.logger)

	// Initialize middleware
	// sessionMiddleware := middleware.NewSessionMiddleware(userService, sessionService, characterService, app.logger)

	// Initialize handlers
	// postHandler := v1.NewPostHandler(postService, commentService, s3Storage, app.logger)

	// Set up routes
	// v1.SetRoutes(app.router, postHandler, sessionMiddleware)

	// // Add static file server for uploaded files (fallback for local development)
	// fileServer := http.FileServer(http.Dir("./uploads"))
	// app.router.Handle("/uploads/", http.StripPrefix("/uploads/", fileServer))

	// Start background tasks
	// go app.startBackgroundTasks(postService, sessionService)

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

// func (app *App) startBackgroundTasks(postService port.PostService, sessionService port.SessionService) {
// 	postTicker := time.NewTicker(1 * time.Minute)
// 	defer postTicker.Stop()

// 	sessionTicker := time.NewTicker(1 * time.Hour)
// 	defer sessionTicker.Stop()

// 	for {
// 		select {
// 		case <-postTicker.C:
// 			if err := postService.CheckPostsForDeletion(); err != nil {
// 				app.logger.Error("Failed to check posts for deletion", err)
// 			}
// 		case <-sessionTicker.C:
// 			if err := sessionService.CleanupExpiredSessions(); err != nil {
// 				app.logger.Error("Failed to cleanup expired sessions", err)
// 			}
// 		}
// 	}
// }

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
