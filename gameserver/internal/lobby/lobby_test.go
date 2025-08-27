package lobby

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/4hel/paper/gameserver/internal/types"
)

func createMockClient(t *testing.T, id string) *types.Client {
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
	t.Cleanup(client.Close)
	return client
}

func TestLobby_DuplicateNames(t *testing.T) {
	lobby := NewLobby()
	defer lobby.Close()

	client1 := createMockClient(t, "client1")
	client2 := createMockClient(t, "client2")

	lobby.AddClient(client1)
	lobby.AddClient(client2)

	// First client joins with name "Alice"
	err1 := lobby.JoinLobby("client1", types.JoinLobbyMessage{Name: "Alice"})
	if err1 != nil {
		t.Errorf("First client should join successfully: %v", err1)
	}

	// Second client tries same name - should fail
	err2 := lobby.JoinLobby("client2", types.JoinLobbyMessage{Name: "Alice"})
	if err2 == nil {
		t.Error("Second client with duplicate name should fail")
	}
}

func TestLobby_RemoveClientMultipleTimes(t *testing.T) {
	lobby := NewLobby()
	defer lobby.Close()

	client := createMockClient(t, "client123")
	lobby.AddClient(client)

	// Remove client multiple times - should not panic
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lobby.RemoveClient("client123")
		}()
	}

	wg.Wait() // Should not panic
}

func TestLobby_ConcurrentJoinAndRemove(t *testing.T) {
	lobby := NewLobby()
	defer lobby.Close()

	// Create multiple clients
	clients := make([]*types.Client, 10)
	for i := 0; i < 10; i++ {
		clients[i] = createMockClient(t, "client"+string(rune('0'+i)))
		lobby.AddClient(clients[i])
	}

	var wg sync.WaitGroup

	// Concurrent joins
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			clientID := "client" + string(rune('0'+idx))
			playerName := "Player" + string(rune('0'+idx))
			lobby.JoinLobby(clientID, types.JoinLobbyMessage{Name: playerName})
		}(i)
	}

	// Concurrent removes
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			clientID := "client" + string(rune('0'+idx))
			lobby.RemoveClient(clientID)
		}(i)
	}

	wg.Wait() // Should not panic or deadlock
}

func TestLobby_PlayerMatching(t *testing.T) {
	lobby := NewLobby()
	defer lobby.Close()

	client1 := createMockClient(t, "client1")
	client2 := createMockClient(t, "client2")

	lobby.AddClient(client1)
	lobby.AddClient(client2)

	// First player joins - should be waiting
	err1 := lobby.JoinLobby("client1", types.JoinLobbyMessage{Name: "Alice"})
	if err1 != nil {
		t.Errorf("First player should join successfully: %v", err1)
	}

	if !client1.InLobby {
		t.Error("First client should be in lobby")
	}

	// Second player joins - should start game
	err2 := lobby.JoinLobby("client2", types.JoinLobbyMessage{Name: "Bob"})
	if err2 != nil {
		t.Errorf("Second player should join successfully: %v", err2)
	}

	// Both should be in game now
	if client1.InLobby || !client1.InGame {
		t.Error("Client1 should be in game, not in lobby")
	}
	if client2.InLobby || !client2.InGame {
		t.Error("Client2 should be in game, not in lobby")
	}
}

func TestLobby_EmptyName(t *testing.T) {
	lobby := NewLobby()
	defer lobby.Close()

	client := createMockClient(t, "client1")
	lobby.AddClient(client)

	// Empty name should fail
	err := lobby.JoinLobby("client1", types.JoinLobbyMessage{Name: ""})
	if err == nil {
		t.Error("Empty name should fail")
	}
}

func TestLobby_NonExistentClient(t *testing.T) {
	lobby := NewLobby()
	defer lobby.Close()

	// Try to join with non-existent client
	err := lobby.JoinLobby("nonexistent", types.JoinLobbyMessage{Name: "Alice"})
	if err == nil {
		t.Error("Non-existent client should fail")
	}
}