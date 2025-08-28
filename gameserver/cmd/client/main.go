package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/4hel/paper/gameserver/internal/types"
)

func main() {
	// Parse command line arguments
	var name = flag.String("name", "", "Player name (required)")
	var server = flag.String("server", "localhost:8080", "Server address")
	var forceHTTP = flag.Bool("http", false, "Force HTTP instead of HTTPS for production servers")
	flag.Parse()

	if *name == "" {
		fmt.Println("Usage: go run cmd/client/main.go -name <player_name> [-server localhost:8080]")
		fmt.Println("\nDeveloper Client - prints raw JSON protocol messages")
		fmt.Println("Commands during gameplay:")
		fmt.Println("  1, 2, 3     - Rock, Paper, Scissors choices")
		fmt.Println("  play        - Play again after game ends")
		fmt.Println("  quit        - Disconnect from server")
		os.Exit(1)
	}

	// Connect to WebSocket server
	protocol := "ws"
	if !*forceHTTP && strings.Contains(*server, ".") && !strings.HasPrefix(*server, "localhost") && !strings.HasPrefix(*server, "127.0.0.1") {
		protocol = "wss"
	}
	url := fmt.Sprintf("%s://%s/ws", protocol, *server)
	fmt.Printf("[DEV CLIENT] Connecting to %s as '%s'\n", url, *name)

	// Configure dialer for production servers
	dialer := websocket.DefaultDialer
	if protocol == "wss" {
		dialer.TLSClientConfig = &tls.Config{
			ServerName: strings.Split(*server, ":")[0], // Extract hostname for SNI
		}
	}

	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer conn.Close()

	fmt.Printf("[DEV CLIENT] Connected! WebSocket established\n")
	fmt.Printf("[DEV CLIENT] Commands: 1=rock, 2=paper, 3=scissors, play, quit\n")
	fmt.Printf("[DEV CLIENT] ------- PROTOCOL MESSAGES -------\n")

	// Send join_lobby message
	joinData, _ := json.Marshal(types.JoinLobbyMessage{Name: *name})
	joinEvent := types.BaseGameEvent{
		Type: "join_lobby",
		Data: joinData,
	}

	jsonOut, _ := json.MarshalIndent(joinEvent, "", "  ")
	fmt.Printf("[SEND] %s\n", string(jsonOut))

	if err := conn.WriteJSON(joinEvent); err != nil {
		log.Fatal("Failed to send join_lobby:", err)
	}

	// Create channels for communication
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	done := make(chan struct{})
	inputChan := make(chan string)

	// Start input reader goroutine
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			inputChan <- strings.TrimSpace(scanner.Text())
		}
	}()

	// Track game state for input validation
	var inGame bool = false
	var waitingForChoice bool = false

	// Read messages from server
	go func() {
		defer close(done)
		for {
			var event types.BaseGameEvent
			err := conn.ReadJSON(&event)
			if err != nil {
				fmt.Printf("[ERROR] Connection closed: %v\n", err)
				return
			}

			// Print raw JSON received
			jsonIn, _ := json.MarshalIndent(event, "", "  ")
			fmt.Printf("[RECV] %s\n", string(jsonIn))

			// Basic state tracking for input validation
			switch event.Type {
			case "player_waiting":
				inGame = false
				waitingForChoice = false
				fmt.Printf("[DEV CLIENT] Waiting for opponent...\n")
			case "game_starting":
				inGame = true
				// Don't change waitingForChoice here - round_start will set it
			case "round_start":
				inGame = true // Ensure we're in game when round starts
				waitingForChoice = true
				fmt.Printf("[DEV CLIENT] Enter your choice: 1=rock, 2=paper, 3=scissors\n")
			case "round_result":
				waitingForChoice = false
			case "game_ended":
				inGame = false
				waitingForChoice = false
				fmt.Printf("[DEV CLIENT] Game ended. Enter: play (to play again) or quit (to disconnect)\n")
			}
		}
	}()

	// Main game loop
	for {
		select {
		case <-done:
			fmt.Printf("[DEV CLIENT] Connection closed\n")
			return

		case <-interrupt:
			fmt.Printf("\n[DEV CLIENT] Interrupt received, disconnecting...\n")
			
			// Send disconnect message
			disconnectData, _ := json.Marshal(types.DisconnectMessage{})
			disconnectEvent := types.BaseGameEvent{
				Type: "disconnect",
				Data: disconnectData,
			}
			
			jsonOut, _ := json.MarshalIndent(disconnectEvent, "", "  ")
			fmt.Printf("[SEND] %s\n", string(jsonOut))
			
			conn.WriteJSON(disconnectEvent)
			
			// Close connection gracefully
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return

		case input := <-inputChan:
			var eventToSend *types.BaseGameEvent
			
			if waitingForChoice {
				// Handle game choice input
				var choice string
				switch input {
				case "1":
					choice = "rock"
				case "2":
					choice = "paper"
				case "3":
					choice = "scissors"
				default:
					fmt.Printf("[DEV CLIENT] Invalid choice '%s'. Use: 1=rock, 2=paper, 3=scissors\n", input)
					continue
				}

				// Send make_choice message
				choiceData, _ := json.Marshal(types.MakeChoiceMessage{Choice: choice})
				eventToSend = &types.BaseGameEvent{
					Type: "make_choice",
					Data: choiceData,
				}

			} else if !inGame {
				// Handle post-game or lobby input
				switch strings.ToLower(input) {
				case "play":
					// Send play_again message
					playAgainData, _ := json.Marshal(types.PlayAgainMessage{})
					eventToSend = &types.BaseGameEvent{
						Type: "play_again",
						Data: playAgainData,
					}

				case "quit", "exit", "q":
					// Send disconnect message
					disconnectData, _ := json.Marshal(types.DisconnectMessage{})
					eventToSend = &types.BaseGameEvent{
						Type: "disconnect",
						Data: disconnectData,
					}

				default:
					fmt.Printf("[DEV CLIENT] Unknown command '%s'. Available: play, quit\n", input)
					continue
				}
			} else {
				fmt.Printf("[DEV CLIENT] In game but not waiting for choice. Current state: inGame=%v, waitingForChoice=%v\n", inGame, waitingForChoice)
				continue
			}

			// Send the event and print JSON
			if eventToSend != nil {
				jsonOut, _ := json.MarshalIndent(*eventToSend, "", "  ")
				fmt.Printf("[SEND] %s\n", string(jsonOut))

				if err := conn.WriteJSON(*eventToSend); err != nil {
					fmt.Printf("[ERROR] Failed to send message: %v\n", err)
				}

				// If disconnect, exit
				if eventToSend.Type == "disconnect" {
					time.Sleep(100 * time.Millisecond) // Brief delay for message to send
					return
				}
			}
		}
	}
}