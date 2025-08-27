package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

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

	// Track game state
	var inGame bool = false
	var waitingForChoice bool = false

	// Read messages from server
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
				inGame = false

			case "game_starting":
				var gameStart types.GameStartingMessage
				json.Unmarshal(event.Data, &gameStart)
				fmt.Printf("🎮 Game starting! Opponent: %s\n", gameStart.OpponentName)
				fmt.Println("Get ready for Rock Paper Scissors!")
				inGame = true

			case "round_start":
				var roundStart types.RoundStartMessage
				json.Unmarshal(event.Data, &roundStart)
				fmt.Printf("\n🥊 Round %d\n", roundStart.RoundNumber)
				fmt.Println("Make your choice:")
				fmt.Println("1 = Rock, 2 = Paper, 3 = Scissors")
				fmt.Print("Enter choice (1-3): ")
				waitingForChoice = true

			case "round_result":
				var result types.RoundResultMessage
				json.Unmarshal(event.Data, &result)
				
				fmt.Printf("\n📊 Round Result: %s\n", strings.ToUpper(result.Result))
				fmt.Printf("Your choice: %s\n", result.YourChoice)
				fmt.Printf("Opponent: %s\n", result.OpponentChoice)
				
				if result.Result == "win" {
					fmt.Println("🎉 You won this round!")
				} else if result.Result == "lose" {
					fmt.Println("😞 You lost this round!")
				} else {
					fmt.Println("🤝 Draw! Same choice!")
				}
				waitingForChoice = false

			case "game_ended":
				var gameEnd types.GameEndedMessage
				json.Unmarshal(event.Data, &gameEnd)
				
				fmt.Printf("\n🏁 Game Over! Result: %s\n", strings.ToUpper(gameEnd.Result))
				if gameEnd.Result == "win" {
					fmt.Println("🏆 Congratulations! You won the game!")
				} else if gameEnd.Result == "lose" {
					fmt.Println("💔 You lost the game. Better luck next time!")
				} else {
					fmt.Println("🤝 Game ended in a draw!")
				}
				
				fmt.Println("\nOptions:")
				fmt.Println("Type 'play' to play again")
				fmt.Println("Type 'quit' to disconnect")
				fmt.Print("Your choice: ")
				inGame = false

			case "error":
				var errorMsg types.ErrorMessage
				json.Unmarshal(event.Data, &errorMsg)
				fmt.Printf("❌ Error: %s\n", errorMsg.Message)

			default:
				fmt.Printf("📨 Received: %s\n", event.Type)
			}
		}
	}()

	// Main game loop
	for {
		select {
		case <-done:
			fmt.Println("Connection closed")
			return

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
			return

		case input := <-inputChan:
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
					fmt.Printf("Invalid choice '%s'. Please enter 1, 2, or 3: ", input)
					continue
				}

				// Send make_choice message
				choiceData, _ := json.Marshal(types.MakeChoiceMessage{Choice: choice})
				choiceEvent := types.BaseGameEvent{
					Type: "make_choice",
					Data: choiceData,
				}

				if err := conn.WriteJSON(choiceEvent); err != nil {
					log.Printf("Failed to send choice: %v", err)
				} else {
					fmt.Printf("You chose: %s\n", choice)
					fmt.Println("⏳ Waiting for opponent's choice...")
					waitingForChoice = false
				}

			} else if !inGame {
				// Handle post-game input
				switch strings.ToLower(input) {
				case "play":
					// Send play_again message
					playAgainData, _ := json.Marshal(types.PlayAgainMessage{})
					playAgainEvent := types.BaseGameEvent{
						Type: "play_again",
						Data: playAgainData,
					}

					if err := conn.WriteJSON(playAgainEvent); err != nil {
						log.Printf("Failed to send play_again: %v", err)
					} else {
						fmt.Println("Searching for new opponent...")
					}

				case "quit", "exit", "q":
					fmt.Println("Goodbye!")
					
					// Send disconnect message
					disconnectData, _ := json.Marshal(types.DisconnectMessage{})
					disconnectEvent := types.BaseGameEvent{
						Type: "disconnect",
						Data: disconnectData,
					}
					conn.WriteJSON(disconnectEvent)
					return

				default:
					fmt.Printf("Unknown command '%s'. Type 'play' or 'quit': ", input)
				}
			}
		}
	}
}