package types

import "encoding/json"

// BaseGameEvent represents the base structure for all game events
type BaseGameEvent struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// Client to Server Messages
type JoinLobbyMessage struct {
	Name string `json:"name"`
}

type MakeChoiceMessage struct {
	Choice string `json:"choice"` // "rock", "paper", "scissors"
}

type PlayAgainMessage struct{}

type DisconnectMessage struct{}

// Server to Client Messages
type PlayerWaitingMessage struct{}

type GameStartingMessage struct {
	OpponentName string `json:"opponent_name"`
}

type RoundResultMessage struct {
	Result       string `json:"result"`        // "win", "lose", "draw"
	YourChoice   string `json:"your_choice"`   // "rock", "paper", "scissors"
	OpponentChoice string `json:"opponent_choice"` // "rock", "paper", "scissors"
}

type RoundStartMessage struct {
	RoundNumber int `json:"round_number"`
}

type GameEndedMessage struct {
	Result string `json:"result"` // "win", "lose"
	Score  string `json:"score"`  // "2-1", "2-0", etc.
}

type ErrorMessage struct {
	Message string `json:"message"`
}