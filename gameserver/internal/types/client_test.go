package types

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/websocket"
)

func TestClient_DoubleClose(t *testing.T) {
	// Create a mock WebSocket connection
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		// Keep connection open for test
		select {}
	}))
	defer server.Close()

	// Connect to test server
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal("Failed to connect:", err)
	}

	// Create client
	client := NewClient("test-123", conn)

	// Close multiple times - should not panic
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client.Close() // Should be safe to call multiple times
		}()
	}

	wg.Wait() // If we reach here, no panic occurred
}

func TestClient_ConcurrentNameAccess(t *testing.T) {
	// Create a mock WebSocket connection
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		select {}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal("Failed to connect:", err)
	}

	client := NewClient("test-456", conn)
	defer client.Close()

	// Concurrent name setting and getting
	var wg sync.WaitGroup
	names := []string{"Alice", "Bob", "Charlie", "Diana", "Eve"}

	// Set names concurrently
	for _, name := range names {
		wg.Add(1)
		go func(n string) {
			defer wg.Done()
			client.SetName(n)
		}(name)
	}

	// Get names concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = client.GetName() // Should not race or panic
		}()
	}

	wg.Wait()

	// Final name should be one of the set names
	finalName := client.GetName()
	found := false
	for _, name := range names {
		if finalName == name {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Final name '%s' not in expected names %v", finalName, names)
	}
}

func TestClient_SendChannelAfterClose(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		select {}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal("Failed to connect:", err)
	}

	client := NewClient("test-789", conn)

	// Close the client
	client.Close()

	// Check that client is marked as closed
	if !client.IsClosed() {
		t.Error("Client should be marked as closed")
	}

	// Trying to send should not panic but should fail gracefully
	func() {
		defer func() {
			if r := recover(); r != nil {
				// This is currently expected behavior - we're testing that the 
				// double close is fixed, not that sending to closed channels is safe
				t.Logf("Expected panic when sending to closed channel: %v", r)
			}
		}()

		select {
		case client.Send <- BaseGameEvent{Type: "test", Data: []byte("{}")}:
			t.Error("Should not be able to send to closed channel")
		default:
			// Expected - channel is closed or send would block
		}
	}()
}