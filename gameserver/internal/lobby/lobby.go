package lobby

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/4hel/paper/gameserver/internal/gameroom"
	"github.com/4hel/paper/gameserver/internal/types"
)

// Lobby manages player matchmaking and game rooms
type Lobby struct {
	clients        map[string]*types.Client
	waitingPlayers map[string]*types.Client
	gameRooms      map[string]*gameroom.GameRoom
	gameRoomCounter int
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
		gameRooms:      make(map[string]*gameroom.GameRoom),
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

	// Generate game room ID
	l.gameRoomCounter++
	gameRoomID := fmt.Sprintf("room-%d", l.gameRoomCounter)

	// Create game room
	gameRoom := gameroom.NewGameRoom(gameRoomID, player1, player2, l.onGameEnd)
	l.gameRooms[gameRoomID] = gameRoom

	// Send game starting messages first (before round_start)
	l.sendGameStarting(player1, player2.GetName())
	l.sendGameStarting(player2, player1.GetName())

	// Now start the first round after game_starting messages are sent
	gameRoom.StartFirstRound()

	log.Printf("Game starting between %s (%s) and %s (%s) in room %s", 
		player1.ID, player1.GetName(), 
		player2.ID, player2.GetName(),
		gameRoomID)
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

// MakeChoice forwards a player's choice to their game room
func (l *Lobby) MakeChoice(clientID string, choice string) error {
	l.mu.RLock()
	defer l.mu.RUnlock()

	client, exists := l.clients[clientID]
	if !exists {
		return fmt.Errorf("client %s not found", clientID)
	}

	if client.GameRoomID == "" {
		return fmt.Errorf("client %s not in a game room", clientID)
	}

	gameRoom, exists := l.gameRooms[client.GameRoomID]
	if !exists {
		return fmt.Errorf("game room %s not found", client.GameRoomID)
	}

	return gameRoom.MakeChoice(clientID, gameroom.Choice(choice))
}

// PlayAgain handles when a player wants to play another game
func (l *Lobby) PlayAgain(clientID string) error {
	log.Printf("PlayAgain: ENTRY - called for client %s", clientID)
	
	l.mu.Lock()
	defer l.mu.Unlock()

	log.Printf("PlayAgain: MUTEX ACQUIRED for client %s", clientID)

	client, exists := l.clients[clientID]
	if !exists {
		log.Printf("PlayAgain: ERROR - Client %s not found in clients map", clientID)
		return fmt.Errorf("client %s not found", clientID)
	}

	log.Printf("PlayAgain: Client %s (%s) wants to play again", clientID, client.GetName())
	log.Printf("PlayAgain: Current client state - InGame: %v, InLobby: %v, GameRoomID: %s", 
		client.InGame, client.InLobby, client.GameRoomID)

	// Reset client state
	client.InGame = false
	client.InLobby = true
	client.GameRoomID = ""

	log.Printf("PlayAgain: Reset client state - InGame: %v, InLobby: %v, GameRoomID: %s", 
		client.InGame, client.InLobby, client.GameRoomID)
	log.Printf("PlayAgain: Current waiting players count: %d", len(l.waitingPlayers))

	// Re-join the lobby for matchmaking
	err := l.joinLobbyInternal(clientID, types.JoinLobbyMessage{Name: client.GetName()})
	if err != nil {
		log.Printf("PlayAgain: ERROR - joinLobbyInternal failed for client %s: %v", clientID, err)
	} else {
		log.Printf("PlayAgain: SUCCESS - joinLobbyInternal succeeded for client %s", clientID)
	}
	
	log.Printf("PlayAgain: EXIT - returning for client %s", clientID)
	return err
}

// joinLobbyInternal is the internal version without locking (already locked)
func (l *Lobby) joinLobbyInternal(clientID string, joinMsg types.JoinLobbyMessage) error {
	log.Printf("joinLobbyInternal: Processing client %s with name '%s'", clientID, joinMsg.Name)
	
	client, exists := l.clients[clientID]
	if !exists {
		log.Printf("joinLobbyInternal: Client %s not found in clients map", clientID)
		return fmt.Errorf("client %s not found", clientID)
	}

	// Check if there's another player waiting
	log.Printf("joinLobbyInternal: Current waiting players count: %d", len(l.waitingPlayers))
	if len(l.waitingPlayers) > 0 {
		log.Printf("joinLobbyInternal: Found waiting players, attempting to match...")
		// Match with first waiting player
		for _, waitingClient := range l.waitingPlayers {
			log.Printf("joinLobbyInternal: Matching %s (%s) with waiting client %s (%s)", 
				clientID, joinMsg.Name, waitingClient.ID, waitingClient.GetName())
			// Start game between client and waitingClient
			l.startGame(client, waitingClient)
			return nil
		}
	} else {
		// No one waiting, add to waiting list
		log.Printf("joinLobbyInternal: No waiting players, adding %s (%s) to waiting list", clientID, joinMsg.Name)
		l.waitingPlayers[clientID] = client
		l.sendPlayerWaiting(client)
		log.Printf("Client %s (%s) is waiting for opponent", clientID, joinMsg.Name)
	}

	return nil
}

// onGameEnd is called when a game room finishes
func (l *Lobby) onGameEnd(gameRoomID string) {
	log.Printf("onGameEnd: ENTRY - called for game room %s", gameRoomID)
	
	// Run asynchronously to avoid deadlock with the lobby mutex
	go func() {
		log.Printf("onGameEnd: ASYNC GOROUTINE - starting cleanup for game room %s", gameRoomID)
		
		l.mu.Lock()
		log.Printf("onGameEnd: MUTEX ACQUIRED for game room %s", gameRoomID)
		defer l.mu.Unlock()

		if gameRoom, exists := l.gameRooms[gameRoomID]; exists {
			gameRoom.Close()
			delete(l.gameRooms, gameRoomID)
			log.Printf("Game room %s destroyed", gameRoomID)
		}
		
		log.Printf("onGameEnd: EXIT - completed for game room %s", gameRoomID)
	}()
	
	log.Printf("onGameEnd: SCHEDULED ASYNC CLEANUP for game room %s", gameRoomID)
}

// Close shuts down the lobby
func (l *Lobby) Close() {
	l.cancel()
	
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Close all game rooms
	for _, gameRoom := range l.gameRooms {
		gameRoom.Close()
	}
	
	for _, client := range l.clients {
		client.Close()
	}
}