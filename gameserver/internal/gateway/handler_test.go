package gateway

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/4hel/paper/gameserver/internal/types"
)

func TestHandler_ConcurrentConnections(t *testing.T) {
	handler := NewHandler()
	defer handler.Close()

	server := httptest.NewServer(http.HandlerFunc(handler.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect multiple clients concurrently
	const numClients = 10
	var wg sync.WaitGroup
	connections := make([]*websocket.Conn, numClients)

	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				t.Errorf("Client %d failed to connect: %v", idx, err)
				return
			}
			connections[idx] = conn
		}(i)
	}

	wg.Wait()

	// Give connections time to be registered
	time.Sleep(50 * time.Millisecond)

	// Verify all connections are tracked
	handler.mu.RLock()
	clientCount := len(handler.clients)
	handler.mu.RUnlock()

	if clientCount != numClients {
		t.Errorf("Expected %d clients, got %d", numClients, clientCount)
	}

	// Close all connections
	for i, conn := range connections {
		if conn != nil {
			conn.Close()
		} else {
			t.Errorf("Connection %d is nil", i)
		}
	}

	// Wait for cleanup
	time.Sleep(100 * time.Millisecond)

	// Verify clients are cleaned up
	handler.mu.RLock()
	finalCount := len(handler.clients)
	handler.mu.RUnlock()

	if finalCount != 0 {
		t.Errorf("Expected 0 clients after cleanup, got %d", finalCount)
	}
}

func TestHandler_ConcurrentJoinLobby(t *testing.T) {
	handler := NewHandler()
	defer handler.Close()

	server := httptest.NewServer(http.HandlerFunc(handler.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect clients and join lobby concurrently
	const numClients = 5
	var wg sync.WaitGroup
	connections := make([]*websocket.Conn, numClients)

	// Connect all clients first
	for i := 0; i < numClients; i++ {
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("Client %d failed to connect: %v", i, err)
		}
		connections[i] = conn
	}

	// Send join_lobby messages concurrently
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			
			joinData, _ := json.Marshal(types.JoinLobbyMessage{
				Name: "Player" + string(rune('0'+idx)),
			})
			joinEvent := types.BaseGameEvent{
				Type: "join_lobby",
				Data: joinData,
			}

			err := connections[idx].WriteJSON(joinEvent)
			if err != nil {
				t.Errorf("Client %d failed to send join_lobby: %v", idx, err)
			}
		}(i)
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond) // Allow processing

	// Clean up
	for _, conn := range connections {
		conn.Close()
	}
}

func TestHandler_ClientDisconnectDuringGame(t *testing.T) {
	handler := NewHandler()
	defer handler.Close()

	server := httptest.NewServer(http.HandlerFunc(handler.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect two clients
	conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal("Failed to connect client 1:", err)
	}

	conn2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal("Failed to connect client 2:", err)
	}

	// Both clients join lobby
	joinData1, _ := json.Marshal(types.JoinLobbyMessage{Name: "Alice"})
	joinEvent1 := types.BaseGameEvent{Type: "join_lobby", Data: joinData1}
	conn1.WriteJSON(joinEvent1)

	joinData2, _ := json.Marshal(types.JoinLobbyMessage{Name: "Bob"})
	joinEvent2 := types.BaseGameEvent{Type: "join_lobby", Data: joinData2}
	conn2.WriteJSON(joinEvent2)

	time.Sleep(50 * time.Millisecond) // Allow game to start

	// Abruptly disconnect one client (simulates network failure)
	conn1.Close()

	// Give time for cleanup
	time.Sleep(100 * time.Millisecond)

	// Verify no panic occurred and second client still tracked
	handler.mu.RLock()
	clientCount := len(handler.clients)
	handler.mu.RUnlock()

	if clientCount != 1 {
		t.Errorf("Expected 1 client remaining, got %d", clientCount)
	}

	// Clean up
	conn2.Close()
}

func TestHandler_StressTestConnectDisconnect(t *testing.T) {
	handler := NewHandler()
	defer handler.Close()

	server := httptest.NewServer(http.HandlerFunc(handler.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Rapid connect/disconnect cycles
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			
			// Connect
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				t.Errorf("Client %d failed to connect: %v", idx, err)
				return
			}

			// Join lobby
			joinData, _ := json.Marshal(types.JoinLobbyMessage{
				Name: "Player" + string(rune('A'+idx%26)),
			})
			joinEvent := types.BaseGameEvent{Type: "join_lobby", Data: joinData}
			conn.WriteJSON(joinEvent)

			// Disconnect quickly
			time.Sleep(time.Millisecond * time.Duration(idx%10+1))
			conn.Close()
		}(i)
	}

	wg.Wait()
	time.Sleep(200 * time.Millisecond) // Allow cleanup

	// Verify all clients cleaned up
	handler.mu.RLock()
	clientCount := len(handler.clients)
	handler.mu.RUnlock()

	if clientCount != 0 {
		t.Errorf("Expected 0 clients after stress test, got %d", clientCount)
	}
}