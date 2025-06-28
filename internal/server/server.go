package server

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"

	"crypto/internal/adapters/repository/postgres"
	"crypto/internal/config"

	"github.com/redis/go-redis/v9"
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
