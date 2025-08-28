package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/4hel/paper/gameserver/internal/types"
)

// Player represents an automated test player
type Player struct {
	name   string
	conn   *websocket.Conn
	t      *testing.T
	wg     *sync.WaitGroup
	gameResults chan string
}

// NewPlayer creates a new automated player
func NewPlayer(name string, serverURL string, t *testing.T, wg *sync.WaitGroup) (*Player, error) {
	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		return nil, err
	}

	return &Player{
		name:        name,
		conn:        conn,
		t:           t,
		wg:          wg,
		gameResults: make(chan string, 1),
	}, nil
}

// Close closes the player connection
func (p *Player) Close() {
	p.conn.Close()
	close(p.gameResults)
}

// Play starts the automated gameplay for this player
func (p *Player) Play() {
	defer p.wg.Done()
	defer p.Close()

	p.t.Logf("[%s] Starting gameplay", p.name)

	// Send join_lobby message
	joinData, _ := json.Marshal(types.JoinLobbyMessage{Name: p.name})
	joinEvent := types.BaseGameEvent{
		Type: "join_lobby",
		Data: joinData,
	}

	if err := p.conn.WriteJSON(joinEvent); err != nil {
		p.t.Errorf("[%s] Failed to send join_lobby: %v", p.name, err)
		return
	}

	p.t.Logf("[%s] Sent join_lobby", p.name)

	// Game state tracking
	choices := []string{"rock", "paper", "scissors"}

	// Message handling loop
	for {
		var event types.BaseGameEvent
		err := p.conn.ReadJSON(&event)
		if err != nil {
			p.t.Logf("[%s] Connection closed: %v", p.name, err)
			return
		}

		p.t.Logf("[%s] Received: %s", p.name, event.Type)

		switch event.Type {
		case "player_waiting":
			p.t.Logf("[%s] Waiting for opponent...", p.name)

		case "game_starting":
			p.t.Logf("[%s] Game starting!", p.name)

		case "round_start":
			p.t.Logf("[%s] Round starting, making choice...", p.name)
			
			// Make random choice after small delay
			go func() {
				time.Sleep(time.Duration(rand.Intn(100)+50) * time.Millisecond)
				
				choice := choices[rand.Intn(len(choices))]
				choiceData, _ := json.Marshal(types.MakeChoiceMessage{Choice: choice})
				choiceEvent := types.BaseGameEvent{
					Type: "make_choice",
					Data: choiceData,
				}

				if err := p.conn.WriteJSON(choiceEvent); err != nil {
					p.t.Errorf("[%s] Failed to send choice: %v", p.name, err)
					return
				}

				p.t.Logf("[%s] Chose: %s", p.name, choice)
			}()

		case "round_result":
			
			var result types.RoundResultMessage
			if err := json.Unmarshal(event.Data, &result); err != nil {
				p.t.Errorf("[%s] Failed to parse round result: %v", p.name, err)
				continue
			}
			
			p.t.Logf("[%s] Round result: %s (you: %s, opponent: %s)", 
				p.name, result.Result, result.YourChoice, result.OpponentChoice)

		case "game_ended":
			
			var gameEnd types.GameEndedMessage
			if err := json.Unmarshal(event.Data, &gameEnd); err != nil {
				p.t.Errorf("[%s] Failed to parse game end: %v", p.name, err)
				continue
			}
			
			p.t.Logf("[%s] Game ended: %s", p.name, gameEnd.Result)
			
			// Send result to test and exit
			p.gameResults <- gameEnd.Result
			return

		case "error":
			var errorMsg types.ErrorMessage
			if err := json.Unmarshal(event.Data, &errorMsg); err != nil {
				p.t.Errorf("[%s] Failed to parse error: %v", p.name, err)
			} else {
				p.t.Errorf("[%s] Received error: %s", p.name, errorMsg.Message)
			}
			return

		default:
			p.t.Logf("[%s] Unknown message type: %s", p.name, event.Type)
		}
	}
}

// GetGameResult waits for the game to complete and returns the result
func (p *Player) GetGameResult() string {
	select {
	case result := <-p.gameResults:
		return result
	case <-time.After(30 * time.Second):
		p.t.Errorf("[%s] Timeout waiting for game result", p.name)
		return "timeout"
	}
}

// TestEndToEndGame tests a complete game between two automated players
func TestEndToEndGame(t *testing.T) {
	// Find an available port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal("Failed to find available port:", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	serverAddr := fmt.Sprintf(":%d", port)
	serverURL := fmt.Sprintf("ws://localhost:%d/ws", port)

	// Create and start server
	server := NewServer(serverAddr)
	
	// Start server in goroutine
	serverDone := make(chan error, 1)
	go func() {
		serverDone <- server.Start()
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	t.Logf("Starting end-to-end test with server on port %d", port)

	// Create two automated players
	var wg sync.WaitGroup
	wg.Add(2)

	player1, err := NewPlayer("TestPlayer1", serverURL, t, &wg)
	if err != nil {
		t.Fatal("Failed to create player1:", err)
	}

	player2, err := NewPlayer("TestPlayer2", serverURL, t, &wg)
	if err != nil {
		player1.Close()
		t.Fatal("Failed to create player2:", err)
	}

	// Start both players
	go player1.Play()
	go player2.Play()

	// Wait for both games to complete or timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		t.Log("Both players completed their games")
	case <-time.After(30 * time.Second):
		t.Fatal("Test timeout - games did not complete in time")
	}

	// Get results from both players
	result1 := player1.GetGameResult()
	result2 := player2.GetGameResult()

	t.Logf("Player1 result: %s", result1)
	t.Logf("Player2 result: %s", result2)

	// Verify results are complementary (one wins, one loses, or both draw)
	if result1 == "win" && result2 != "lose" {
		t.Errorf("Inconsistent results: Player1=%s, Player2=%s", result1, result2)
	} else if result1 == "lose" && result2 != "win" {
		t.Errorf("Inconsistent results: Player1=%s, Player2=%s", result1, result2)
	} else if result1 == "draw" && result2 != "draw" {
		t.Errorf("Inconsistent results: Player1=%s, Player2=%s", result1, result2)
	} else if result1 != "win" && result1 != "lose" && result1 != "draw" && result1 != "timeout" {
		t.Errorf("Invalid result for Player1: %s", result1)
	} else if result2 != "win" && result2 != "lose" && result2 != "draw" && result2 != "timeout" {
		t.Errorf("Invalid result for Player2: %s", result2)
	}

	// Shutdown server
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		t.Errorf("Failed to shutdown server: %v", err)
	}

	t.Log("End-to-end test completed successfully")
}