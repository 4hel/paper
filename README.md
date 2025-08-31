## Go Server Structs Reference

| Package  | Name          | Methods                                                                                                                                                          | Source File                   | Purpose |
|----------|---------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------------|---------|
| main     | Server        | Start, Shutdown                                                                                                                                                  | cmd/paperserver/main.go       | HTTP server wrapper with WebSocket handler for testing |
| types    | BaseGameEvent | _(no methods)_                                                                                                                                                   | internal/types/message.go     | Base structure for all WebSocket game events |
| types    | ErrorMessage  | _(no methods)_                                                                                                                                                   | internal/types/message.go     | Server message for error responses |
| types    | Client        | SetName, GetName, Close, IsClosed                                                                                                                                | internal/types/client.go      | WebSocket client connection with state management |
| gateway  | Handler       | HandleWebSocket, addClient, removeClient, readPump, writePump, handleMessage, Close                                                                              | internal/gateway/handler.go   | WebSocket connection manager and message router |
| lobby    | Lobby         | AddClient, RemoveClient, JoinLobby, startGame, sendPlayerWaiting, sendGameStarting, sendError, MakeChoice, PlayAgain, joinLobbyInternal, onGameEnd, Close        | internal/lobby/lobby.go       | Player matchmaking and game room management |
| gameroom | GameRoom      | StartFirstRound, MakeChoice, processRound, determineWinner, startRound, endGame, getClientByID, sendRoundResult, sendRoundStart, sendGameEnded, sendError, Close | internal/gameroom/gameroom.go | Rock Paper Scissors game logic and state |

## Complete Game Flow Sequence

```mermaid
sequenceDiagram
    participant C1 as Client 1
    participant C2 as Client 2
    participant H as Handler
    participant L as Lobby
    participant GR as GameRoom

    Note over C1,GR: Connection & Lobby Phase
    C1->>H: WebSocket Connect
    H->>H: HandleWebSocket()
    H->>H: Create Client, addClient()
    H->>L: AddClient(client1)
    H->>H: Start readPump() & writePump()
    
    C1->>H: join_lobby {name: "Alice"}
    H->>H: handleMessage()
    H->>L: JoinLobby(client1, "Alice")
    L->>L: Check waiting players (none)
    L->>C1: player_waiting
    Note over L: Client1 added to waitingPlayers
    
    C2->>H: WebSocket Connect
    H->>H: HandleWebSocket()
    H->>H: Create Client, addClient()
    H->>L: AddClient(client2)
    H->>H: Start readPump() & writePump()
    
    C2->>H: join_lobby {name: "Bob"}
    H->>H: handleMessage()
    H->>L: JoinLobby(client2, "Bob")
    L->>L: Found waiting player (Alice)
    L->>L: startGame(Alice, Bob)
    
    Note over L,GR: Game Room Creation
    L->>GR: NewGameRoom(id, player1, player2)
    GR->>GR: Set players InGame=true, InLobby=false
    L->>C1: game_starting {opponent_name: "Bob"}
    L->>C2: game_starting {opponent_name: "Alice"}
    L->>GR: StartFirstRound()
    
    Note over C1,GR: Game Play Phase
    GR->>GR: startRound()
    GR->>C1: round_start {round_number: 1}
    GR->>C2: round_start {round_number: 1}
    
    C1->>H: make_choice {choice: "rock"}
    H->>H: handleMessage()
    H->>L: MakeChoice(client1, "rock")
    L->>GR: MakeChoice(client1, "rock")
    GR->>GR: Record Player1Choice, set Player1Ready=true
    
    C2->>H: make_choice {choice: "scissors"}
    H->>H: handleMessage()
    H->>L: MakeChoice(client2, "scissors")
    L->>GR: MakeChoice(client2, "scissors")
    GR->>GR: Record Player2Choice, set Player2Ready=true
    GR->>GR: processRound() - Both ready
    GR->>GR: determineWinner() - Player1 wins
    GR->>C1: round_result {result: "win", your_choice: "rock", opponent_choice: "scissors"}
    GR->>C2: round_result {result: "lose", your_choice: "scissors", opponent_choice: "rock"}
    
    Note over GR: Round 2 (if game continues)
    GR->>GR: startRound() (async)
    GR->>C1: round_start {round_number: 2}
    GR->>C2: round_start {round_number: 2}
    
    Note over C1,GR: ... More rounds until best of 3 ...
    
    Note over GR: Game End (Player1 wins 2-1)
    GR->>GR: endGame()
    GR->>C1: game_ended {result: "win"}
    GR->>C2: game_ended {result: "lose"}
    GR->>GR: Reset players: InGame=false, InLobby=true
    GR->>L: onGameEnd(gameRoomID)
    L->>GR: Close() & delete from gameRooms
    
    Note over C1,GR: Post-Game Options
    alt Play Again
        C1->>H: play_again
        H->>H: handleMessage()
        H->>L: PlayAgain(client1)
        L->>L: Reset client state, joinLobbyInternal()
        L->>C1: player_waiting (or match with another player)
    else Disconnect
        C1->>H: disconnect
        H->>H: handleMessage() 
        H->>H: client.Close()
        H->>H: removeClient()
        H->>L: RemoveClient(client1)
    end
    
    Note over C2,GR: Client2 similar options...
```

# Paper - Multiplayer Rock Paper Scissors

A real-time multiplayer Rock Paper Scissors game with Unity client and Go WebSocket server.

## Prerequisites

### Unity Client Dependencies
**REQUIRED**: Install NativeWebSocket package in Unity:
1. Open Unity Package Manager (Window → Package Manager)
2. Click "+" → Add package from git URL
3. Enter: `https://github.com/endel/NativeWebSocket.git#upm`
4. Click "Add"

Without this package, the Unity client will not compile.

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