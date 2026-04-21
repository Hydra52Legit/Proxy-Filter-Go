package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("🌐 HTTP Proxy Server with Filtering")
	fmt.Println(strings.Repeat("=", 60))

	// Load configuration and blacklist
	config, err := LoadConfig("blacklist.txt")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Create filter
	filter := NewFilter(config)

	// Create server
	server := NewServer(config, filter)

	// Start server in a goroutine
	go func() {
		log.Printf("🚀 Server starting on %s", config.ListenAddr)
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server error:", err)
		}
	}()

	// Wait for interrupt signal (Ctrl+C)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("\n🛑 Shutting down server...")

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Gracefully stop the server
	if err := server.Stop(ctx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	log.Println("✅ Server stopped")
}