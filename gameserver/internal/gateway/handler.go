package gateway

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/4hel/paper/gameserver/internal/lobby"
	"github.com/4hel/paper/gameserver/internal/types"
	"github.com/gorilla/websocket"
)

/*
Handler manages WebSocket connections and message routing.

Main entry point: HandleWebSocket() upgrades HTTP connections to WebSocket.

The Handler uses a pump-based architecture for bidirectional communication:
  - readPump:  Connection → Application (pulls data from WebSocket, pushes to message handler)
  - writePump: Application → Connection (pulls data from Send channel, pushes to WebSocket)
*/
type Handler struct {
	upgrader websocket.Upgrader
	lobby    *lobby.Lobby
	clients  map[string]*types.Client
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewHandler creates a new WebSocket handler
func NewHandler() *Handler {
	ctx, cancel := context.WithCancel(context.Background())

	return &Handler{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins for development
				return true
			},
		},
		lobby:   lobby.NewLobby(),
		clients: make(map[string]*types.Client),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// HandleWebSocket upgrades HTTP connection to WebSocket
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Create client
	clientID := generateClientID()
	client := types.NewClient(clientID, conn)

	// Add client to handler and lobby
	h.addClient(client)
	h.lobby.AddClient(client)

	// Start client goroutines
	go h.writePump(client)
	go h.readPump(client)

	log.Printf("New WebSocket connection established: %s", clientID)
}

// addClient adds a client to the handler's client map
func (h *Handler) addClient(client *types.Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client.ID] = client
}

// removeClient removes a client from handler and lobby
func (h *Handler) removeClient(client *types.Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.clients, client.ID)
	h.lobby.RemoveClient(client.ID)
	client.Close()
}

// readPump handles incoming messages from client
func (h *Handler) readPump(client *types.Client) {
	defer h.removeClient(client)

	// Set read deadline and pong handler
	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		select {
		case <-client.Ctx.Done():
			return
		default:
			var event types.BaseGameEvent
			err := client.Conn.ReadJSON(&event)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error for client %s: %v", client.ID, err)
				}
				return
			}

			h.handleMessage(client, event)
		}
	}
}

// writePump handles outgoing messages to client
func (h *Handler) writePump(client *types.Client) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case <-client.Ctx.Done():
			return
		case event, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.Conn.WriteJSON(event); err != nil {
				log.Printf("Write error for client %s: %v", client.ID, err)
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming messages from clients
func (h *Handler) handleMessage(client *types.Client, event types.BaseGameEvent) {
	switch event.Type {
	case "join_lobby":
		var joinMsg types.JoinLobbyMessage
		if err := json.Unmarshal(event.Data, &joinMsg); err != nil {
			log.Printf("Failed to unmarshal join_lobby message from client %s: %v", client.ID, err)
			return
		}

		if err := h.lobby.JoinLobby(client.ID, joinMsg); err != nil {
			log.Printf("Failed to join lobby for client %s: %v", client.ID, err)
		}

	case "make_choice":
		var choiceMsg types.MakeChoiceMessage
		if err := json.Unmarshal(event.Data, &choiceMsg); err != nil {
			log.Printf("Failed to unmarshal make_choice message from client %s: %v", client.ID, err)
			return
		}
		
		if err := h.lobby.MakeChoice(client.ID, choiceMsg.Choice); err != nil {
			log.Printf("Failed to make choice for client %s: %v", client.ID, err)
		}

	case "play_again":
		if err := h.lobby.PlayAgain(client.ID); err != nil {
			log.Printf("Failed to play again for client %s: %v", client.ID, err)
		}

	case "disconnect":
		log.Printf("Client %s requested disconnect", client.ID)
		client.Close()

	default:
		log.Printf("Unknown message type '%s' from client %s", event.Type, client.ID)
	}
}

// generateClientID generates a unique client ID
func generateClientID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	randBytes := make([]byte, length)
	
	_, err := rand.Read(randBytes)
	if err != nil {
		// Fallback to timestamp-based generation if crypto/rand fails
		seed := time.Now().UnixNano()
		for i := range b {
			seed = (seed*1103515245 + 12345) & 0x7FFFFFFF // Keep positive
			b[i] = charset[seed%int64(len(charset))]
		}
		return string(b)
	}
	
	for i, randByte := range randBytes {
		b[i] = charset[randByte%byte(len(charset))]
	}
	return string(b)
}

// Close shuts down the handler
func (h *Handler) Close() {
	h.cancel()
	h.lobby.Close()

	h.mu.Lock()
	defer h.mu.Unlock()

	for _, client := range h.clients {
		client.Close()
	}
}
