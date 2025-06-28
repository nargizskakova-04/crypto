package app

// import (
// 	"fmt"
// 	"forum/internal/config"
// 	"forum/internal/server"
// 	"log"
// )

// const cfgPath = "./config/config.json"

// func Start() error {
// 	log.Println("Loading configuration...")
// 	config, err := config.GetConfig(cfgPath)
// 	if err != nil {
// 		return fmt.Errorf("failed to get config: %w", err)
// 	}

// 	log.Printf("Configuration loaded: port=%d\n", config.App.Port)

// 	log.Println("Creating application instance...")
// 	app := server.NewApp(config)

// 	log.Println("Initializing application...")
// 	if err := app.Initialize(); err != nil {
// 		return fmt.Errorf("failed to initialize app: %w", err)
// 	}

// 	log.Println("Starting server...")
// 	app.Run()

// 	log.Println("Server stopped")
// 	return nil
// }
// d
