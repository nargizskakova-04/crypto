package app

import (
	"fmt"
	"log"

	"crypto/internal/config"
	"crypto/internal/server"
)

const cfgPath = "./config/config.json"

func Start(portFlag int) error {
	log.Println("Loading configuration...")
	config, err := config.LoadConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	// Override port from command line if provided
	if portFlag > 0 {
		config.App.Port = portFlag
		log.Printf("Port overridden from command line: %d\n", portFlag)
	}

	log.Printf("Configuration loaded: port=%d, mode=%s\n", config.App.Port, config.App.Mode)

	log.Println("Creating MarketFlow application instance...")
	app := server.NewApp(config)

	log.Println("Initializing MarketFlow application...")
	if err := app.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize app: %w", err)
	}

	log.Println("Starting MarketFlow server...")

	// Setup graceful shutdown
	app.SetupGracefulShutdown()

	// Run the server
	if err := app.Run(); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	log.Println("MarketFlow server stopped")
	return nil
}
