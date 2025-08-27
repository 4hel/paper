package gameroom

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/4hel/paper/gameserver/internal/types"
)

func createMockClient(t *testing.T, id, name string) *types.Client {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		select {}
	}))
	t.Cleanup(server.Close)

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal("Failed to connect:", err)
	}

	client := types.NewClient(id, conn)
	client.SetName(name)
	t.Cleanup(client.Close)
	return client
}

func TestGameRoom_RockPaperScissorsLogic(t *testing.T) {
	tests := []struct {
		choice1  Choice
		choice2  Choice
		result1  string
		result2  string
	}{
		// Rock vs others
		{Rock, Rock, "draw", "draw"},
		{Rock, Paper, "lose", "win"},
		{Rock, Scissors, "win", "lose"},
		
		// Paper vs others
		{Paper, Rock, "win", "lose"},
		{Paper, Paper, "draw", "draw"},
		{Paper, Scissors, "lose", "win"},
		
		// Scissors vs others
		{Scissors, Rock, "lose", "win"},
		{Scissors, Paper, "win", "lose"},
		{Scissors, Scissors, "draw", "draw"},
	}

	for _, tt := range tests {
		t.Run(string(tt.choice1)+"_vs_"+string(tt.choice2), func(t *testing.T) {
			player1 := createMockClient(t, "player1", "Alice")
			player2 := createMockClient(t, "player2", "Bob")

			gameRoom := NewGameRoom("test-room", player1, player2, func(roomID string) {
				// Game end callback for cleanup
			})
			defer gameRoom.Close()

			// Make choices
			err1 := gameRoom.MakeChoice(player1.ID, tt.choice1)
			err2 := gameRoom.MakeChoice(player2.ID, tt.choice2)

			if err1 != nil {
				t.Errorf("Player1 choice failed: %v", err1)
			}
			if err2 != nil {
				t.Errorf("Player2 choice failed: %v", err2)
			}

			// Wait for message processing
			time.Sleep(10 * time.Millisecond)

			// First message should be round_start
			select {
			case event := <-player1.Send:
				if event.Type != "round_start" {
					t.Errorf("Expected round_start first, got %s", event.Type)
				}
			case <-time.After(100 * time.Millisecond):
				t.Error("No round_start message received for player1")
			}

			select {
			case event := <-player2.Send:
				if event.Type != "round_start" {
					t.Errorf("Expected round_start first, got %s", event.Type)
				}
			case <-time.After(100 * time.Millisecond):
				t.Error("No round_start message received for player2")
			}

			// Second message should be round_result
			select {
			case event := <-player1.Send:
				if event.Type != "round_result" {
					t.Errorf("Expected round_result, got %s", event.Type)
				}
			case <-time.After(100 * time.Millisecond):
				t.Error("No round_result message received for player1")
			}

			select {
			case event := <-player2.Send:
				if event.Type != "round_result" {
					t.Errorf("Expected round_result, got %s", event.Type)
				}
			case <-time.After(100 * time.Millisecond):
				t.Error("No round_result message received for player2")
			}
		})
	}
}

func TestGameRoom_BestOfThreeFlow(t *testing.T) {
	player1 := createMockClient(t, "player1", "Alice")
	player2 := createMockClient(t, "player2", "Bob")

	var gameEndCalled bool
	var gameEndedRoomID string
	gameRoom := NewGameRoom("test-room", player1, player2, func(roomID string) {
		gameEndCalled = true
		gameEndedRoomID = roomID
	})
	defer gameRoom.Close()

	// Wait for initial round_start message
	time.Sleep(10 * time.Millisecond)
	
	// Clear initial round_start messages
	select {
	case <-player1.Send:
	default:
	}
	select {
	case <-player2.Send:
	default:
	}

	// Round 1: Player1 wins (Rock vs Scissors)
	gameRoom.MakeChoice(player1.ID, Rock)
	gameRoom.MakeChoice(player2.ID, Scissors)
	
	time.Sleep(10 * time.Millisecond)
	
	// Check round 1 results
	if gameRoom.Player1Wins != 1 {
		t.Errorf("Expected Player1Wins = 1, got %d", gameRoom.Player1Wins)
	}
	if gameRoom.Player2Wins != 0 {
		t.Errorf("Expected Player2Wins = 0, got %d", gameRoom.Player2Wins)
	}
	if gameRoom.CurrentRound != 2 {
		t.Errorf("Expected CurrentRound = 2, got %d", gameRoom.CurrentRound)
	}
	if gameRoom.GameEnded {
		t.Error("Game should not be ended after round 1")
	}

	// Clear round_result messages
	select {
	case <-player1.Send:
	default:
	}
	select {
	case <-player2.Send:
	default:
	}

	// Wait for round_start for round 2
	time.Sleep(2100 * time.Millisecond) // Wait for delay + processing
	
	// Clear round_start messages
	select {
	case <-player1.Send:
	default:
	}
	select {
	case <-player2.Send:
	default:
	}

	// Round 2: Player1 wins again (Paper vs Rock)
	gameRoom.MakeChoice(player1.ID, Paper)
	gameRoom.MakeChoice(player2.ID, Rock)
	
	time.Sleep(10 * time.Millisecond)

	// Check final state
	if gameRoom.Player1Wins != 2 {
		t.Errorf("Expected Player1Wins = 2, got %d", gameRoom.Player1Wins)
	}
	if !gameRoom.GameEnded {
		t.Error("Game should be ended after player1 reaches 2 wins")
	}
	
	// Wait for callback
	time.Sleep(10 * time.Millisecond)
	
	if !gameEndCalled {
		t.Error("Game end callback should have been called")
	}
	if gameEndedRoomID != "test-room" {
		t.Errorf("Expected room ID 'test-room', got '%s'", gameEndedRoomID)
	}

	// Check that players are no longer in game
	if player1.InGame {
		t.Error("Player1 should no longer be in game")
	}
	if player2.InGame {
		t.Error("Player2 should no longer be in game")
	}
}

func TestGameRoom_InvalidChoices(t *testing.T) {
	player1 := createMockClient(t, "player1", "Alice")
	player2 := createMockClient(t, "player2", "Bob")

	gameRoom := NewGameRoom("test-room", player1, player2, nil)
	defer gameRoom.Close()

	// Wait and clear initial round_start messages
	time.Sleep(10 * time.Millisecond)
	select {
	case <-player1.Send:
	default:
	}
	select {
	case <-player2.Send:
	default:
	}

	// Test invalid choice
	err := gameRoom.MakeChoice(player1.ID, Choice("invalid"))
	if err != nil {
		t.Errorf("Expected no error for invalid choice, got %v", err)
	}

	// Should receive error message
	select {
	case event := <-player1.Send:
		if event.Type != "error" {
			t.Errorf("Expected error message, got %s", event.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected error message for invalid choice")
	}
}

func TestGameRoom_NonExistentPlayer(t *testing.T) {
	player1 := createMockClient(t, "player1", "Alice")
	player2 := createMockClient(t, "player2", "Bob")

	gameRoom := NewGameRoom("test-room", player1, player2, nil)
	defer gameRoom.Close()

	// Test choice from non-existent player
	err := gameRoom.MakeChoice("nonexistent", Rock)
	if err != nil {
		t.Errorf("Expected no error for non-existent player, got %v", err)
	}

	// Should not affect game state
	if gameRoom.Player1Ready || gameRoom.Player2Ready {
		t.Error("Players should not be marked as ready")
	}
}

func TestGameRoom_ConcurrentChoices(t *testing.T) {
	player1 := createMockClient(t, "player1", "Alice")
	player2 := createMockClient(t, "player2", "Bob")

	gameRoom := NewGameRoom("test-room", player1, player2, nil)
	defer gameRoom.Close()

	// Make concurrent choices
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		gameRoom.MakeChoice(player1.ID, Rock)
	}()

	go func() {
		defer wg.Done()
		gameRoom.MakeChoice(player2.ID, Scissors)
	}()

	wg.Wait()
	time.Sleep(10 * time.Millisecond)

	// Both players should have made choices and round should be processed
	if gameRoom.Player1Wins != 1 {
		t.Errorf("Expected Player1Wins = 1, got %d", gameRoom.Player1Wins)
	}
}

func TestGameRoom_ChoiceAfterGameEnd(t *testing.T) {
	player1 := createMockClient(t, "player1", "Alice")
	player2 := createMockClient(t, "player2", "Bob")

	gameRoom := NewGameRoom("test-room", player1, player2, nil)
	defer gameRoom.Close()

	// Manually end the game
	gameRoom.mu.Lock()
	gameRoom.GameEnded = true
	gameRoom.mu.Unlock()

	// Try to make a choice after game ended
	err := gameRoom.MakeChoice(player1.ID, Rock)
	if err != nil {
		t.Errorf("Expected no error when making choice after game end, got %v", err)
	}

	// Choice should be ignored
	if gameRoom.Player1Ready {
		t.Error("Player1 should not be marked as ready after game ended")
	}
}

func TestGameRoom_DrawScenarios(t *testing.T) {
	player1 := createMockClient(t, "player1", "Alice")
	player2 := createMockClient(t, "player2", "Bob")

	var gameEndCalled bool
	gameRoom := NewGameRoom("test-room", player1, player2, func(roomID string) {
		gameEndCalled = true
	})
	defer gameRoom.Close()

	// Clear initial messages
	time.Sleep(10 * time.Millisecond)
	select {
	case <-player1.Send:
	default:
	}
	select {
	case <-player2.Send:
	default:
	}

	// Round 1: Draw
	gameRoom.MakeChoice(player1.ID, Rock)
	gameRoom.MakeChoice(player2.ID, Rock)
	
	time.Sleep(10 * time.Millisecond)

	// Check that nobody won
	if gameRoom.Player1Wins != 0 || gameRoom.Player2Wins != 0 {
		t.Errorf("Expected no wins after draw, got %d-%d", gameRoom.Player1Wins, gameRoom.Player2Wins)
	}

	// Round should still advance
	if gameRoom.CurrentRound != 2 {
		t.Errorf("Expected CurrentRound = 2 after draw, got %d", gameRoom.CurrentRound)
	}

	// Clear messages and continue
	select {
	case <-player1.Send:
	default:
	}
	select {
	case <-player2.Send:
	default:
	}

	// Wait for next round
	time.Sleep(2100 * time.Millisecond)
	select {
	case <-player1.Send:
	default:
	}
	select {
	case <-player2.Send:
	default:
	}

	// Round 2: Draw again
	gameRoom.MakeChoice(player1.ID, Paper)
	gameRoom.MakeChoice(player2.ID, Paper)
	
	time.Sleep(10 * time.Millisecond)
	
	// Still no wins
	if gameRoom.Player1Wins != 0 || gameRoom.Player2Wins != 0 {
		t.Errorf("Expected no wins after second draw, got %d-%d", gameRoom.Player1Wins, gameRoom.Player2Wins)
	}

	// Wait for next round
	time.Sleep(2100 * time.Millisecond)
	select {
	case <-player1.Send:
	default:
	}
	select {
	case <-player2.Send:
	default:
	}

	// Round 3: Final draw
	gameRoom.MakeChoice(player1.ID, Scissors)
	gameRoom.MakeChoice(player2.ID, Scissors)
	
	time.Sleep(10 * time.Millisecond)

	// Game should end after 3 rounds even with all draws
	if !gameRoom.GameEnded {
		t.Error("Game should end after 3 rounds even with draws")
	}

	// Wait for callback
	time.Sleep(10 * time.Millisecond)
	
	if !gameEndCalled {
		t.Error("Game end callback should have been called after 3 draws")
	}
}

func TestGameRoom_MaxRoundsReached(t *testing.T) {
	player1 := createMockClient(t, "player1", "Alice")
	player2 := createMockClient(t, "player2", "Bob")

	var gameEndCalled bool
	gameRoom := NewGameRoom("test-room", player1, player2, func(roomID string) {
		gameEndCalled = true
	})
	defer gameRoom.Close()

	// Simulate 3 rounds with 1-1 score going into round 3
	gameRoom.mu.Lock()
	gameRoom.Player1Wins = 1
	gameRoom.Player2Wins = 1
	gameRoom.CurrentRound = 3
	gameRoom.mu.Unlock()

	// Clear any pending messages
	select {
	case <-player1.Send:
	default:
	}
	select {
	case <-player2.Send:
	default:
	}

	// Final round
	gameRoom.MakeChoice(player1.ID, Rock)
	gameRoom.MakeChoice(player2.ID, Scissors)
	
	time.Sleep(10 * time.Millisecond)

	// Game should end
	if !gameRoom.GameEnded {
		t.Error("Game should end after round 3")
	}

	// Player1 should win 2-1
	if gameRoom.Player1Wins != 2 {
		t.Errorf("Expected Player1Wins = 2, got %d", gameRoom.Player1Wins)
	}

	if !gameEndCalled {
		t.Error("Game end callback should have been called")
	}
}

func TestGameRoom_StressTestMultipleGames(t *testing.T) {
	const numGames = 3 // Reduced to avoid overwhelming logs
	var wg sync.WaitGroup
	var gameEndCount int
	var mu sync.Mutex

	for i := 0; i < numGames; i++ {
		wg.Add(1)
		go func(gameID int) {
			defer wg.Done()
			
			player1 := createMockClient(t, fmt.Sprintf("player1-%d", gameID), "Alice")
			player2 := createMockClient(t, fmt.Sprintf("player2-%d", gameID), "Bob")

			gameRoom := NewGameRoom(fmt.Sprintf("test-room-%d", gameID), player1, player2, func(roomID string) {
				mu.Lock()
				gameEndCount++
				mu.Unlock()
			})
			defer gameRoom.Close()

			// Play a quick game (Player1 wins 2-0)
			time.Sleep(10 * time.Millisecond) // Wait for initial setup
			
			// Clear initial messages
			select {
			case <-player1.Send:
			default:
			}
			select {
			case <-player2.Send:
			default:
			}

			// Round 1
			gameRoom.MakeChoice(player1.ID, Rock)
			gameRoom.MakeChoice(player2.ID, Scissors)
			time.Sleep(10 * time.Millisecond)

			// Round 2 - wait for round start
			time.Sleep(2100 * time.Millisecond)
			select {
			case <-player1.Send:
			default:
			}
			select {
			case <-player2.Send:
			default:
			}

			gameRoom.MakeChoice(player1.ID, Paper)
			gameRoom.MakeChoice(player2.ID, Rock)
			time.Sleep(10 * time.Millisecond)
		}(i)
	}

	wg.Wait()

	// All games should have ended
	mu.Lock()
	finalCount := gameEndCount
	mu.Unlock()
	
	if finalCount != numGames {
		t.Errorf("Expected %d games to end, got %d", numGames, finalCount)
	}
}