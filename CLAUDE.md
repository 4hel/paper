# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview
Paper is a real-time multiplayer Rock Paper Scissors game with a Unity client and Go WebSocket server. The architecture follows a clean separation with the gameserver handling all multiplayer logic and the Unity client providing the game interface.

## Development Commands

### Go Server (gameserver/)
```bash
# Navigate to server directory
cd gameserver

# Install/update dependencies
go mod tidy

# Build the server
go build ./cmd/paperserver

# Run the server (starts on :8080)
go run cmd/paperserver/main.go

# Test WebSocket endpoint
# Connect to: ws://localhost:8080/ws
# Health check: http://localhost:8080/health
```

### Unity Client (paper_client/)
- Open `paper_client` directory in Unity Editor
- Connect to `ws://localhost:8080/ws` for local development
- Build and run or play in editor for testing

## Architecture

### Server Architecture (Go)
The server uses a message-driven architecture with three main components:

1. **WebSocket Handler** (`internal/websocket/`): Uses pump-based architecture
   - `readPump`: Connection → Application (reads from WebSocket)
   - `writePump`: Application → Connection (writes to WebSocket)
   - Each client runs two separate goroutines for non-blocking I/O

2. **Lobby System** (`internal/lobby/`): Manages player matchmaking
   - Handles player name validation and uniqueness
   - Matches waiting players automatically
   - Transitions players from lobby to game state

3. **Message System** (`internal/types/`): Type-safe WebSocket communication
   - `BaseGameEvent`: Base structure for all messages with type hierarchy
   - Client and Server message types for Unity C# compatibility
   - Thread-safe client connection management

### Message Protocol
All WebSocket messages follow the BaseGameEvent structure:
```json
{
  "type": "message_type",
  "data": { ...message_specific_data... }
}
```

**Client → Server**: `join_lobby`, `make_choice`, `play_again`, `disconnect`  
**Server → Client**: `player_waiting`, `game_starting`, `round_result`, `round_start`, `game_ended`, `error`

### Client State Flow
1. **ChoosingName**: Client connects and sends `join_lobby` with player name
2. **Lobby States**: Receives `player_waiting` or `game_starting`
3. **Game States**: Best-of-3 Rock Paper Scissors rounds
4. **End Game**: Option to `play_again` or `disconnect`

## Key Implementation Details

### WebSocket Connection Management
- Each client has dedicated read/write goroutines with independent error handling
- Automatic ping/pong for connection health (54s intervals)
- Graceful shutdown with context cancellation
- Thread-safe client state management with RWMutex

### Lobby Matching Logic
- First-come-first-served matching (no skill-based matchmaking)
- Name uniqueness enforced across waiting players
- Automatic game initiation when two players are matched
- Players removed from waiting queue when matched

### Message Inheritance for Unity
The BaseGameEvent structure is designed for Unity C# client compatibility, allowing for type-safe message handling with inheritance patterns commonly used in C# game development.