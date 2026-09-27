package main

import (
	"log"
	"os"

	"github.com/example/healthcheckmonitor/config"
	"github.com/example/healthcheckmonitor/internal/logger"
	"github.com/example/healthcheckmonitor/internal/monitor"
	"github.com/example/healthcheckmonitor/internal/server"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env if present
	_ = godotenv.Load()

	// Initialise logger early
	logger.Init()

	// Load configuration
	cfg, err := config.Init()
	if err != nil {
		logger.Log.Fatalf("failed to load configuration: %v", err)
	}

	// Create and start monitor
	mon := monitor.New(cfg)
	mon.Start()
	defer mon.Stop()

	// Determine listening address
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	// Initialise HTTP server
	srv := server.New(addr, mon)

	// Run server (blocks until interrupt)
	srv.Run()
}