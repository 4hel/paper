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

// Server wraps the HTTP server and WebSocket handler for easier testing
type Server struct {
	httpServer *http.Server
	wsHandler  *gateway.Handler
}

// NewServer creates a new server instance
func NewServer(port string) *Server {
	wsHandler := gateway.NewHandler()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", wsHandler.HandleWebSocket)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return &Server{
		httpServer: &http.Server{
			Addr:    port,
			Handler: mux,
		},
		wsHandler: wsHandler,
	}
}

// Start starts the server (blocking)
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	// Close WebSocket handler first
	s.wsHandler.Close()
	
	// Then shutdown HTTP server
	return s.httpServer.Shutdown(ctx)
}

func main() {
	port := ":8080"
	server := NewServer(port)

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
		
		// Shutdown server gracefully
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server forced to shutdown after timeout: %v", err)
		} else {
			log.Println("Server shutdown gracefully")
		}
		
		log.Println("Server shutdown complete")
		os.Exit(0)
	}()

	// Start HTTP server
	log.Println("Server is ready to handle requests")
	if err := server.Start(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Server failed to start:", err)
	}
}
