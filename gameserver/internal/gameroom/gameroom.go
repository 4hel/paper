package gameroom

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/4hel/paper/gameserver/internal/types"
)

// Choice represents a player's choice
type Choice string

const (
	Rock     Choice = "rock"
	Paper    Choice = "paper"
	Scissors Choice = "scissors"
)

// GameRoom manages a Rock Paper Scissors game between two players
type GameRoom struct {
	ID             string
	Player1        *types.Client
	Player2        *types.Client
	Player1Wins    int
	Player2Wins    int
	CurrentRound   int
	Player1Choice  Choice
	Player2Choice  Choice
	Player1Ready   bool
	Player2Ready   bool
	GameEnded      bool
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	onGameEnd      func(gameRoomID string) // Callback to notify when game ends
}

// NewGameRoom creates a new game room for two players
func NewGameRoom(id string, player1, player2 *types.Client, onGameEnd func(string)) *GameRoom {
	ctx, cancel := context.WithCancel(context.Background())
	
	room := &GameRoom{
		ID:           id,
		Player1:      player1,
		Player2:      player2,
		CurrentRound: 1,
		ctx:          ctx,
		cancel:       cancel,
		onGameEnd:    onGameEnd,
	}

	// Set players' game room ID and status
	player1.GameRoomID = id
	player1.InGame = true
	player1.InLobby = false

	player2.GameRoomID = id
	player2.InGame = true
	player2.InLobby = false

	// Don't start the round immediately - let the lobby send game_starting first
	log.Printf("GameRoom %s created for players %s and %s", id, player1.GetName(), player2.GetName())
	return room
}

// StartFirstRound begins the first round of the game
func (gr *GameRoom) StartFirstRound() {
	gr.startRound()
}

// MakeChoice processes a player's choice
func (gr *GameRoom) MakeChoice(clientID string, choice Choice) error {
	gr.mu.Lock()
	defer gr.mu.Unlock()

	if gr.GameEnded {
		return nil // Game already ended
	}

	// Validate choice
	if choice != Rock && choice != Paper && choice != Scissors {
		gr.sendError(gr.getClientByID(clientID), "Invalid choice. Use rock, paper, or scissors")
		return nil
	}

	// Record the choice
	if clientID == gr.Player1.ID {
		gr.Player1Choice = choice
		gr.Player1Ready = true
		log.Printf("GameRoom %s: Player1 %s chose %s", gr.ID, gr.Player1.GetName(), choice)
	} else if clientID == gr.Player2.ID {
		gr.Player2Choice = choice
		gr.Player2Ready = true
		log.Printf("GameRoom %s: Player2 %s chose %s", gr.ID, gr.Player2.GetName(), choice)
	} else {
		return nil // Player not in this game
	}

	// Check if both players have made their choices
	if gr.Player1Ready && gr.Player2Ready {
		gr.processRound()
	}

	return nil
}

// processRound determines the winner and sends results
func (gr *GameRoom) processRound() {
	var shouldStartNextRound bool
	// Determine round winner
	result1, result2 := gr.determineWinner(gr.Player1Choice, gr.Player2Choice)

	// Update scores
	if result1 == "win" {
		gr.Player1Wins++
	} else if result2 == "win" {
		gr.Player2Wins++
	}

	log.Printf("GameRoom %s Round %d: %s vs %s - Score: %d-%d", 
		gr.ID, gr.CurrentRound, gr.Player1Choice, gr.Player2Choice, gr.Player1Wins, gr.Player2Wins)

	// Send round results
	gr.sendRoundResult(gr.Player1, result1, string(gr.Player1Choice), string(gr.Player2Choice))
	gr.sendRoundResult(gr.Player2, result2, string(gr.Player2Choice), string(gr.Player1Choice))

	// Reset choices for next round
	gr.Player1Choice = ""
	gr.Player2Choice = ""
	gr.Player1Ready = false
	gr.Player2Ready = false

	// Check if game is over (best of 3)
	if gr.Player1Wins >= 2 || gr.Player2Wins >= 2 || gr.CurrentRound >= 3 {
		gr.endGame()
	} else {
		// Prepare for next round
		gr.CurrentRound++
		shouldStartNextRound = true
	}
	
	// Start next round in a goroutine to avoid deadlock
	if shouldStartNextRound {
		go gr.startRound()
	}
}

// determineWinner returns the result for player1 and player2
func (gr *GameRoom) determineWinner(choice1, choice2 Choice) (string, string) {
	if choice1 == choice2 {
		return "draw", "draw"
	}

	// Rock Paper Scissors logic
	switch {
	case (choice1 == Rock && choice2 == Scissors) ||
		 (choice1 == Paper && choice2 == Rock) ||
		 (choice1 == Scissors && choice2 == Paper):
		return "win", "lose"
	default:
		return "lose", "win"
	}
}

// startRound begins a new round
func (gr *GameRoom) startRound() {
	gr.mu.Lock()
	if gr.GameEnded {
		gr.mu.Unlock()
		return
	}
	
	roundNumber := gr.CurrentRound
	player1 := gr.Player1
	player2 := gr.Player2
	gameID := gr.ID
	gr.mu.Unlock()

	log.Printf("GameRoom %s: Starting round %d", gameID, roundNumber)

	// Send messages without holding the mutex to avoid deadlock
	gr.sendRoundStart(player1, roundNumber)
	gr.sendRoundStart(player2, roundNumber)
}


// endGame finishes the game and determines the winner
func (gr *GameRoom) endGame() {
	gr.GameEnded = true

	var result1, result2 string

	// Determine final game result
	if gr.Player1Wins > gr.Player2Wins {
		result1 = "win"
		result2 = "lose"
	} else if gr.Player2Wins > gr.Player1Wins {
		result1 = "lose"
		result2 = "win"
	} else {
		result1 = "draw"
		result2 = "draw"
	}

	log.Printf("GameRoom %s ended: %s (%d) vs %s (%d) - Winner: %s", 
		gr.ID, gr.Player1.GetName(), gr.Player1Wins, gr.Player2.GetName(), gr.Player2Wins,
		func() string {
			if result1 == "win" { return gr.Player1.GetName() }
			if result2 == "win" { return gr.Player2.GetName() }
			return "Draw"
		}())

	// Send game ended messages
	gr.sendGameEnded(gr.Player1, result1)
	gr.sendGameEnded(gr.Player2, result2)

	// Reset player states
	gr.Player1.InGame = false
	gr.Player1.InLobby = true
	gr.Player1.GameRoomID = ""
	gr.Player2.InGame = false
	gr.Player2.InLobby = true
	gr.Player2.GameRoomID = ""

	// Notify that game has ended
	if gr.onGameEnd != nil {
		gr.onGameEnd(gr.ID)
	}
}

// getClientByID returns the client with the given ID
func (gr *GameRoom) getClientByID(clientID string) *types.Client {
	if gr.Player1.ID == clientID {
		return gr.Player1
	} else if gr.Player2.ID == clientID {
		return gr.Player2
	}
	return nil
}

// Message sending functions
func (gr *GameRoom) sendRoundResult(client *types.Client, result, yourChoice, opponentChoice string) {
	if client.IsClosed() {
		return
	}
	
	data, _ := json.Marshal(types.RoundResultMessage{
		Result:         result,
		YourChoice:     yourChoice,
		OpponentChoice: opponentChoice,
	})
	event := types.BaseGameEvent{
		Type: "round_result",
		Data: data,
	}
	
	select {
	case client.Send <- event:
	case <-gr.ctx.Done():
		return
	default:
		log.Printf("Failed to send round_result to client %s", client.ID)
	}
}

func (gr *GameRoom) sendRoundStart(client *types.Client, roundNumber int) {
	if client.IsClosed() {
		return
	}
	
	data, _ := json.Marshal(types.RoundStartMessage{
		RoundNumber: roundNumber,
	})
	event := types.BaseGameEvent{
		Type: "round_start",
		Data: data,
	}
	
	select {
	case client.Send <- event:
	case <-gr.ctx.Done():
		return
	default:
		log.Printf("Failed to send round_start to client %s", client.ID)
	}
}

func (gr *GameRoom) sendGameEnded(client *types.Client, result string) {
	if client.IsClosed() {
		return
	}
	
	data, _ := json.Marshal(types.GameEndedMessage{
		Result: result,
	})
	event := types.BaseGameEvent{
		Type: "game_ended",
		Data: data,
	}
	
	select {
	case client.Send <- event:
	case <-gr.ctx.Done():
		return
	default:
		log.Printf("Failed to send game_ended to client %s", client.ID)
	}
}

func (gr *GameRoom) sendError(client *types.Client, message string) {
	if client.IsClosed() {
		return
	}
	
	data, _ := json.Marshal(types.ErrorMessage{
		Message: message,
	})
	event := types.BaseGameEvent{
		Type: "error",
		Data: data,
	}
	
	select {
	case client.Send <- event:
	case <-gr.ctx.Done():
		return
	default:
		log.Printf("Failed to send error to client %s", client.ID)
	}
}

// Close cleans up the game room
func (gr *GameRoom) Close() {
	gr.cancel()
}