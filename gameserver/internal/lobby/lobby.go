package lobby

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/4hel/paper/gameserver/internal/types"
)

// Lobby manages player matchmaking
type Lobby struct {
	clients        map[string]*types.Client
	waitingPlayers map[string]*types.Client
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
}

// NewLobby creates a new lobby instance
func NewLobby() *Lobby {
	ctx, cancel := context.WithCancel(context.Background())
	return &Lobby{
		clients:        make(map[string]*types.Client),
		waitingPlayers: make(map[string]*types.Client),
		ctx:            ctx,
		cancel:         cancel,
	}
}

// AddClient adds a client to the lobby
func (l *Lobby) AddClient(client *types.Client) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.clients[client.ID] = client
	log.Printf("Client %s added to lobby", client.ID)
}

// RemoveClient removes a client from the lobby
func (l *Lobby) RemoveClient(clientID string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if _, exists := l.clients[clientID]; exists {
		delete(l.clients, clientID)
		delete(l.waitingPlayers, clientID)
		log.Printf("Client %s removed from lobby", clientID)
	}
}

// JoinLobby processes a client joining the lobby with their name
func (l *Lobby) JoinLobby(clientID string, joinMsg types.JoinLobbyMessage) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	client, exists := l.clients[clientID]
	if !exists {
		return fmt.Errorf("client %s not found", clientID)
	}

	// Validate name
	if joinMsg.Name == "" {
		l.sendError(client, "Name cannot be empty")
		return fmt.Errorf("empty name for client %s", clientID)
	}

	// Check if name is already taken
	for _, waitingClient := range l.waitingPlayers {
		if waitingClient.GetName() == joinMsg.Name {
			l.sendError(client, "Name already taken")
			return fmt.Errorf("name %s already taken", joinMsg.Name)
		}
	}

	// Set client name and add to lobby
	client.SetName(joinMsg.Name)
	client.InLobby = true
	
	// Check if there's another player waiting
	if len(l.waitingPlayers) > 0 {
		// Match with first waiting player
		for _, waitingClient := range l.waitingPlayers {
			// Start game between client and waitingClient
			l.startGame(client, waitingClient)
			return nil
		}
	} else {
		// No one waiting, add to waiting list
		l.waitingPlayers[clientID] = client
		l.sendPlayerWaiting(client)
		log.Printf("Client %s (%s) is waiting for opponent", clientID, joinMsg.Name)
	}

	return nil
}

// startGame initiates a game between two players
func (l *Lobby) startGame(player1, player2 *types.Client) {
	// Remove both players from waiting list
	delete(l.waitingPlayers, player1.ID)
	delete(l.waitingPlayers, player2.ID)

	// Mark players as in game
	player1.InLobby = false
	player1.InGame = true
	player2.InLobby = false
	player2.InGame = true

	// Send game starting messages
	l.sendGameStarting(player1, player2.GetName())
	l.sendGameStarting(player2, player1.GetName())

	log.Printf("Game starting between %s (%s) and %s (%s)", 
		player1.ID, player1.GetName(), 
		player2.ID, player2.GetName())
}

// sendPlayerWaiting sends player_waiting message to client
func (l *Lobby) sendPlayerWaiting(client *types.Client) {
	data, _ := json.Marshal(types.PlayerWaitingMessage{})
	event := types.BaseGameEvent{
		Type: "player_waiting",
		Data: data,
	}
	
	select {
	case client.Send <- event:
	default:
		log.Printf("Failed to send player_waiting to client %s", client.ID)
	}
}

// sendGameStarting sends game_starting message to client
func (l *Lobby) sendGameStarting(client *types.Client, opponentName string) {
	data, _ := json.Marshal(types.GameStartingMessage{
		OpponentName: opponentName,
	})
	event := types.BaseGameEvent{
		Type: "game_starting",
		Data: data,
	}
	
	select {
	case client.Send <- event:
	default:
		log.Printf("Failed to send game_starting to client %s", client.ID)
	}
}

// sendError sends error message to client
func (l *Lobby) sendError(client *types.Client, message string) {
	data, _ := json.Marshal(types.ErrorMessage{
		Message: message,
	})
	event := types.BaseGameEvent{
		Type: "error",
		Data: data,
	}
	
	select {
	case client.Send <- event:
	default:
		log.Printf("Failed to send error to client %s", client.ID)
	}
}

// Close shuts down the lobby
func (l *Lobby) Close() {
	l.cancel()
	
	l.mu.Lock()
	defer l.mu.Unlock()
	
	for _, client := range l.clients {
		client.Close()
	}
}