# Paper - Multiplayer Rock Paper Scissors

A real-time multiplayer Rock Paper Scissors game with Unity client and Go WebSocket server.

## Architecture

### Server Components (Go)
- **Gateway**: Manages client connections and message routing using pump-based architecture
- **Lobby**: Matches players and creates game rooms
- **Game Room**: Handles game logic and state management

#### Gateway Architecture
The Gateway uses a pump-based architecture for efficient bidirectional communication:

- **readPump**: Connection → Application
  - Continuously reads incoming messages from WebSocket connections
  - Handles message parsing and routing to appropriate handlers
  - Manages connection timeouts and ping/pong for connection health

- **writePump**: Application → Connection  
  - Continuously writes outgoing messages from internal channels to WebSocket
  - Handles message serialization and delivery
  - Manages write timeouts and connection cleanup

Each client connection runs two separate goroutines (readPump + writePump) for non-blocking, concurrent message processing. This pattern ensures that slow reads don't block writes and vice versa.

#### Server Architecture Flow

```mermaid
graph TD
    %% Client connections
    C1[Client 1] -.-> |WebSocket| GW[Gateway Handler]
    C2[Client 2] -.-> |WebSocket| GW
    C3[Client N] -.-> |WebSocket| GW
    
    %% Gateway components
    GW --> |readPump| RP1[Read Pump 1]
    GW --> |readPump| RP2[Read Pump 2] 
    GW --> |readPump| RPN[Read Pump N]
    
    GW --> |writePump| WP1[Write Pump 1]
    GW --> |writePump| WP2[Write Pump 2]
    GW --> |writePump| WPN[Write Pump N]
    
    %% Message routing
    RP1 --> |handleMessage| MH[Message Handler]
    RP2 --> |handleMessage| MH
    RPN --> |handleMessage| MH
    
    %% Message types and routing
    MH --> |join_lobby| LB[Lobby Manager]
    MH --> |make_choice| LB
    MH --> |play_again| LB
    MH --> |disconnect| LB
    
    %% Lobby management
    LB --> |matchmaking| WP[Waiting Players Queue]
    LB --> |create game| GR1[Game Room 1]
    LB --> |create game| GR2[Game Room 2]
    LB --> |create game| GRN[Game Room N]
    
    %% Game room operations
    GR1 --> |round_start| WP1
    GR1 --> |round_result| WP1
    GR1 --> |game_ended| WP1
    
    GR2 --> |round_start| WP2
    GR2 --> |round_result| WP2
    GR2 --> |game_ended| WP2
    
    %% Game room lifecycle
    GR1 --> |onGameEnd| LB
    GR2 --> |onGameEnd| LB
    GRN --> |onGameEnd| LB
    
    %% Error handling
    MH --> |error messages| WP1
    MH --> |error messages| WP2
    MH --> |error messages| WPN
    
    %% State management
    subgraph "Client State"
        CS1[Client 1 State<br/>InGame: true/false<br/>InLobby: true/false<br/>GameRoomID: string]
        CS2[Client 2 State<br/>InGame: true/false<br/>InLobby: true/false<br/>GameRoomID: string]
    end
    
    LB -.-> CS1
    LB -.-> CS2
    GR1 -.-> CS1
    GR1 -.-> CS2
    
    %% Message flow legend with high contrast colors
    classDef clientStyle fill:#2196f3,stroke:#1976d2,stroke-width:2px,color:#ffffff
    classDef gatewayStyle fill:#9c27b0,stroke:#7b1fa2,stroke-width:2px,color:#ffffff
    classDef lobbyStyle fill:#4caf50,stroke:#388e3c,stroke-width:2px,color:#ffffff
    classDef gameStyle fill:#ff9800,stroke:#f57c00,stroke-width:2px,color:#ffffff
    
    class C1,C2,C3 clientStyle
    class GW,RP1,RP2,RPN,WP1,WP2,WPN,MH gatewayStyle
    class LB,WP lobbyStyle
    class GR1,GR2,GRN gameStyle
```

### Client (Unity)
- **Game UI**: Rock Paper Scissors interface
- **WebSocket Client**: Communication with game server
- **Game State Manager**: Client-side game state

## Game Flow

### Client State Machine
```mermaid
stateDiagram-v2
    [*] --> Disconnected
    Disconnected --> ChoosingName : connect to gameserver
    ChoosingName --> InLobby : send join_lobby
    ChoosingName --> Disconnected : connection error
    
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
- `join_lobby` - Join lobby with player name
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
│   │   └── gateway/     # Client gateway and WebSocket handling
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