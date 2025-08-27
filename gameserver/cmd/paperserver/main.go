package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

	// Start server
	port := ":8080"
	log.Printf("Paper game server starting on port %s", port)
	log.Printf("WebSocket endpoint: ws://localhost%s/ws", port)

	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Shutting down server...")
		wsHandler.Close()
		os.Exit(0)
	}()

	// Start HTTP server
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
