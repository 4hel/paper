package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
	"github.com/4hel/paper/gameserver/internal/types"
)

func main() {
	// Parse command line arguments
	var name = flag.String("name", "", "Player name (required)")
	var server = flag.String("server", "localhost:8080", "Server address")
	flag.Parse()

	if *name == "" {
		fmt.Println("Usage: go run cmd/client/main.go -name <player_name> [-server localhost:8080]")
		os.Exit(1)
	}

	// Connect to WebSocket server
	url := fmt.Sprintf("ws://%s/ws", *server)
	fmt.Printf("Connecting to %s as '%s'...\n", url, *name)

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer conn.Close()

	fmt.Printf("Connected! Joining lobby as '%s'\n", *name)

	// Send join_lobby message
	joinData, _ := json.Marshal(types.JoinLobbyMessage{Name: *name})
	joinEvent := types.BaseGameEvent{
		Type: "join_lobby",
		Data: joinData,
	}

	if err := conn.WriteJSON(joinEvent); err != nil {
		log.Fatal("Failed to send join_lobby:", err)
	}

	// Handle graceful shutdown
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Read messages from server
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			var event types.BaseGameEvent
			err := conn.ReadJSON(&event)
			if err != nil {
				fmt.Println("Connection closed:", err)
				return
			}

			switch event.Type {
			case "player_waiting":
				fmt.Println("⏳ Waiting for opponent...")

			case "game_starting":
				var gameStart types.GameStartingMessage
				json.Unmarshal(event.Data, &gameStart)
				fmt.Printf("🎮 Game starting! Opponent: %s\n", gameStart.OpponentName)

			case "error":
				var errorMsg types.ErrorMessage
				json.Unmarshal(event.Data, &errorMsg)
				fmt.Printf("❌ Error: %s\n", errorMsg.Message)

			default:
				fmt.Printf("📨 Received: %s\n", event.Type)
			}
		}
	}()

	// Wait for interrupt or connection close
	select {
	case <-done:
		fmt.Println("Connection closed")
	case <-interrupt:
		fmt.Println("\nDisconnecting...")
		
		// Send disconnect message
		disconnectData, _ := json.Marshal(types.DisconnectMessage{})
		disconnectEvent := types.BaseGameEvent{
			Type: "disconnect",
			Data: disconnectData,
		}
		conn.WriteJSON(disconnectEvent)
		
		// Close connection gracefully
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	}
}