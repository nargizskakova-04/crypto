package main

import (
	"crypto/internal/app"
	"flag"
	"fmt"
	"os"
)

// INFO:
// MarketFlow - Real-Time Market Data Processing System
// domain layer - the fundamental operations and rules for market data processing, price aggregation, and concurrency patterns
// adapter layer - data access layer for PostgreSQL, Redis, and exchange connections
//
// TODO: Implement real-time data fetching from cryptocurrency exchanges (Live Mode)
// TODO: Implement test data generation (Test Mode)
// TODO: Implement concurrency patterns: Fan-in, Fan-out, Worker Pool, Generator
// TODO: Store aggregated data in PostgreSQL with batching
// TODO: Cache recent prices in Redis for quick access
// TODO: Provide REST API for querying market data and statistics
// TODO: Handle graceful shutdown and failover scenarios

// General functionality:
//	Process real-time cryptocurrency price updates from multiple sources
//	Support Live Mode (connect to provided exchange simulators on ports 40101-40103)
//	Support Test Mode (generate synthetic market data locally)
//	Use hexagonal architecture with clean separation of concerns
//	Implement concurrency patterns for efficient data processing
//	Store minute-aggregated data (average, min, max prices) in PostgreSQL
//	Cache latest prices in Redis for fast API responses
//	Provide REST API endpoints for price queries and system health
//	Handle failover and reconnection scenarios gracefully

func main() {
	// Define command line flags
	var (
		port     = flag.Int("port", 0, "Port number")
		helpFlag = flag.Bool("help", false, "Show help message")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  marketflow [--port <N>]\n")
		fmt.Fprintf(os.Stderr, "  marketflow --help\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  --port N     Port number\n")
	}

	flag.Parse()

	// Show help if requested
	if *helpFlag {
		flag.Usage()
		return
	}

	fmt.Println("Starting MarketFlow application...")

	// Start the application
	if err := app.Start(*port); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start application: %v\n", err)
		os.Exit(1)
	}

	// Keep application running
	select {}
}
