package types

import (
	"context"
	"sync"

	"github.com/gorilla/websocket"
)

// Client represents a connected WebSocket client
type Client struct {
	ID         string
	Name       string
	Conn       *websocket.Conn
	Send       chan BaseGameEvent
	InLobby    bool
	InGame     bool
	GameRoomID string
	mu         sync.RWMutex
	Ctx        context.Context
	cancel     context.CancelFunc
	closed     bool
}

// NewClient creates a new client instance
func NewClient(id string, conn *websocket.Conn) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		ID:      id,
		Conn:    conn,
		Send:    make(chan BaseGameEvent, 256),
		InLobby: false,
		InGame:  false,
		Ctx:     ctx,
		cancel:  cancel,
	}
}

// SetName sets the client's name safely
func (c *Client) SetName(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Name = name
}

// GetName gets the client's name safely
func (c *Client) GetName() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Name
}

// Close closes the client connection and cancels context
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.closed {
		return // Already closed
	}
	
	c.closed = true
	c.cancel()
	close(c.Send)
	c.Conn.Close()
}

// IsClosed returns true if the client has been closed
func (c *Client) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
}