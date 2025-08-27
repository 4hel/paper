package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/4hel/paper/gameserver/internal/gateway"
)

func main() {
	// Create WebSocket handler
	wsHandler := gateway.NewHandler()
	defer wsHandler.Close()

	// Setup HTTP routes
	http.HandleFunc("/ws", wsHandler.HandleWebSocket)
	
	// Add a simple health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Create HTTP server
	port := ":8080"
	server := &http.Server{
		Addr: port,
	}

	log.Printf("Paper game server starting on port %s", port)
	log.Printf("WebSocket endpoint: ws://localhost%s/ws", port)

	// Setup graceful shutdown with timeout
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Shutting down server...")
		
		// Create shutdown context with timeout
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		
		// Close WebSocket handler first with timeout
		log.Println("Closing WebSocket handler...")
		done := make(chan struct{})
		go func() {
			wsHandler.Close()
			close(done)
		}()
		
		// Wait for handler to close or timeout
		select {
		case <-done:
			log.Println("WebSocket handler closed")
		case <-time.After(2 * time.Second):
			log.Println("WebSocket handler close timed out, forcing shutdown")
		}
		
		// Small delay to let connections close
		time.Sleep(100 * time.Millisecond)
		
		// Shutdown HTTP server
		log.Println("Shutting down HTTP server...")
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server forced to shutdown after timeout: %v", err)
		} else {
			log.Println("HTTP server shutdown gracefully")
		}
		
		log.Println("Server shutdown complete")
		os.Exit(0) // Force exit if graceful shutdown completed
	}()

	// Start HTTP server
	log.Println("Server is ready to handle requests")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Server failed to start:", err)
	}
}
