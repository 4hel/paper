# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview
Paper is a real-time multiplayer Rock Paper Scissors game with a Unity client and Go WebSocket server. The architecture follows a clean separation with the gameserver handling all multiplayer logic and the Unity client providing the game interface.

## Development Commands

### Go Server (gameserver/)
```bash
# Navigate to server directory
cd gameserver

# Run comprehensive tests (race detection + verbose output)
make test

# Run end-to-end tests (automated player vs player game)
make end2end

# Build server and client
make build

# Run the server (starts on :8080)
make server

# Run client (replace PlayerName with actual name)
make client name=PlayerName

# Clean build artifacts  
make clean

# Manual commands (if needed)
go mod tidy                    # Update dependencies
go run cmd/paperserver/main.go # Run server directly

# Test server startup (requires coreutils: brew install coreutils)
timeout 3s go run cmd/paperserver/main.go

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

1. **Gateway** (`internal/gateway/`): Uses pump-based architecture
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

### Gateway Connection Management
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

## Testing Strategy

The codebase includes comprehensive tests with focus on concurrency and race condition detection:

### Primary Test Command
```bash
make test  # Runs all tests with race detection (-race) and verbose output (-v)
```

**This single command provides maximum bug detection including:**
- Race condition detection (catches concurrency bugs)
- Verbose test output with detailed logging
- Tests all internal packages (./internal/...)

### Test Coverage Areas
- **Client lifecycle**: Connection management, double-close scenarios, concurrent operations
- **Lobby system**: Player matching, name validation, concurrent join/remove operations  
- **Gateway integration**: WebSocket handling, message routing, connection cleanup
- **Stress testing**: High-load scenarios with rapid connect/disconnect cycles

**Critical**: Always run `make test` before committing changes. The race detector has caught critical bugs including double-close panics and array index errors that would cause production crashes.