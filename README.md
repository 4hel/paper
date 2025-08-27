# Paper - Multiplayer Rock Paper Scissors

A real-time multiplayer Rock Paper Scissors game with Unity client and Go WebSocket server.

## Architecture

### Server Components (Go)
- **WebSocket Handler**: Manages client connections and message routing
- **Lobby**: Matches players and creates game rooms
- **Game Room**: Handles game logic and state management

### Client (Unity)
- **Game UI**: Rock Paper Scissors interface
- **WebSocket Client**: Communication with game server
- **Game State Manager**: Client-side game state

## Game Flow

### Client State Machine
```mermaid
stateDiagram-v2
    [*] --> Disconnected
    Disconnected --> InLobby : connect to gameserver
    
    %% Lobby states
    InLobby --> WaitingInLobby : receive player_waiting
    InLobby --> WaitingForChoice : receive game_starting
    WaitingInLobby --> WaitingForChoice : receive game_starting
    
    %% Game room states
    WaitingForChoice --> WaitingForOpponentChoice : send make_choice
    WaitingForChoice --> ShowResult : receive round_result
    WaitingForOpponentChoice --> ShowResult : receive round_result
    
    ShowResult --> WaitingForChoice : receive round_start
    ShowResult --> GameEnded : receive game_ended
    
    GameEnded --> InLobby : send play_again
    GameEnded --> Disconnected : send disconnect
    InLobby --> Disconnected : send disconnect
```

## WebSocket Message Protocol

### Client → Server Messages
- `make_choice` - Submit Rock/Paper/Scissors choice
- `play_again` - Return to lobby after game ends
- `disconnect` - Leave server

### Server → Client Messages  
- `player_waiting` - Waiting for opponent in lobby
- `game_starting` - Opponent found, entering game
- `round_result` - Round outcome (win/lose/draw) 
- `round_start` - Next round beginning
- `game_ended` - Final game result
- `error` - Error message

## Project Structure

```
/
├── gameserver/          # Go WebSocket server
│   ├── cmd/paperserver/ # Server executable
│   ├── internal/        # Server components
│   │   ├── lobby/       # Lobby management
│   │   ├── game/        # Game logic
│   │   └── websocket/   # WebSocket handling
│   └── go.mod
├── paper_client/        # Unity client
│   ├── Assets/
│   │   ├── Scripts/     # Game scripts
│   │   ├── Scenes/      # Unity scenes
│   │   └── UI/          # UI components
│   └── ...
└── README.md
```

## Development Setup

### Server
```bash
cd gameserver
go mod tidy
go run cmd/paperserver/main.go
```

### Client  
1. Open `paper_client` in Unity
2. Build and run or play in editor
3. Connect to `ws://localhost:8080/ws`

## Game Rules
- Best of 3 rounds wins
- Rock beats Scissors
- Scissors beats Paper  
- Paper beats Rock
- Same choice = Draw (replay round)