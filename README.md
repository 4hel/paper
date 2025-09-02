## Client Package Structure

| Package | Imports | Description |
|---------|---------|-------------|
| GameInitializer | Paper.UI.Core | Entry point component that creates and initializes the main GameUI system |
| Paper.Network | NativeWebSocket | WebSocket message protocol and GameServerClient for real-time server communication |
| Paper.UI.Core | Paper.Network, Paper.UI.Panels | Main UI coordinator managing canvas, event system, and panel transitions |
| Paper.UI.Panels | None | UI panels for login (LoginPanel) and game interaction (GamePanel) with button and text management |

## TODO

### * Re-Architect:
    in @gameserver/internal/gateway/handler.go - when the handler receives make_choice from the player this is passed to the lobby like lobby.MakeChoice() make a code review and evalutate if this architecture makes sense, because conceptually a lobby is for matching players and player moves should be handled by the game room, I would assume.
    
    Current Code: Works but violates clean architecture principles. The lobby acting as a message forwarder creates unnecessary coupling and doesn't align with its conceptual purpose.
    Better Approach: Gateway should route game messages directly to game rooms based on client state, keeping lobby focused on matchmaking only.


## Server Package Structure

| Package | Imports | Description |
|---------|---------|-------------|
| main (cmd/paperserver) | internal/gateway | HTTP server wrapper with WebSocket handler and graceful shutdown mechanism |
| main (cmd/client) | gorilla/websocket, internal/types | Command-line client for testing the game server with text-based interface |
| internal/types | gorilla/websocket | Message structures, client connection management, and WebSocket communication types |
| internal/gateway | gorilla/websocket, internal/lobby, internal/types | WebSocket connection handler with pump-based architecture for bidirectional communication |
| internal/lobby | internal/gameroom, internal/types | Player matchmaking, game room management, and client state transitions |
| internal/gameroom | internal/types | Rock Paper Scissors game logic and player interaction management |

## Server Structs Reference

| Package  | Name          | Methods                                                                                                                                                          | Source File                   | Purpose |
|----------|---------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------------|---------|
| main     | Server        | Start, Shutdown                                                                                                                                                  | cmd/paperserver/main.go       | HTTP server wrapper with WebSocket handler for testing |
| types    | BaseGameEvent | _(no methods)_                                                                                                                                                   | internal/types/message.go     | Base structure for all WebSocket game events |
| types    | ErrorMessage  | _(no methods)_                                                                                                                                                   | internal/types/message.go     | Server message for error responses |
| types    | Client        | SetName, GetName, Close, IsClosed                                                                                                                                | internal/types/client.go      | WebSocket client connection with state management |
| gateway  | Handler       | HandleWebSocket, addClient, removeClient, readPump, writePump, handleMessage, Close                                                                              | internal/gateway/handler.go   | WebSocket connection manager and message router |
| lobby    | Lobby         | AddClient, RemoveClient, JoinLobby, startGame, sendPlayerWaiting, sendGameStarting, sendError, MakeChoice, PlayAgain, joinLobbyInternal, onGameEnd, Close        | internal/lobby/lobby.go       | Player matchmaking and game room management |
| gameroom | GameRoom      | StartFirstRound, MakeChoice, processRound, determineWinner, startRound, endGame, getClientByID, sendRoundResult, sendRoundStart, sendGameEnded, sendError, Close | internal/gameroom/gameroom.go | Rock Paper Scissors game logic and state |

## Server Package Structure Diagram

```mermaid
graph TB
    subgraph main_paperserver ["main (cmd/paperserver)"]
        direction LR
        Server["Server<br/>httpServer: *http.Server<br/>wsHandler: *gateway.Handler"]
        http_Server["http.Server<br/>Addr: string<br/>Handler: http.Handler"]
        http_ServeMux["http.ServeMux"]
        context_Context["context.Context"]
        os_Signal["os.Signal"]
        
        Server ~~~ http_Server ~~~ http_ServeMux ~~~ context_Context ~~~ os_Signal
        
        class Server green
        class http_Server,http_ServeMux,context_Context,os_Signal blue
    end

    subgraph main_client ["main (cmd/client)"]
        direction LR
        websocket_Conn["websocket.Conn"]
        tls_Config["tls.Config<br/>ServerName: string"]
        websocket_Dialer["websocket.Dialer<br/>TLSClientConfig: *tls.Config"]
        
        websocket_Conn ~~~ tls_Config ~~~ websocket_Dialer
        
        class websocket_Conn,tls_Config,websocket_Dialer grey
    end

    subgraph types_pkg ["types"]
        direction LR
        Client["Client<br/>ID: string<br/>Name: string<br/>Conn: *websocket.Conn<br/>Send: chan BaseGameEvent<br/>InLobby: bool<br/>InGame: bool<br/>GameRoomID: string<br/>mu: sync.RWMutex<br/>Ctx: context.Context<br/>cancel: context.CancelFunc<br/>closed: bool"]
        BaseGameEvent["BaseGameEvent<br/>Type: string<br/>Data: json.RawMessage"]
        MessageStructs["Message Structs<br/>JoinLobbyMessage<br/>MakeChoiceMessage<br/>PlayAgainMessage<br/>DisconnectMessage<br/>PlayerWaitingMessage<br/>GameStartingMessage<br/>RoundResultMessage<br/>RoundStartMessage<br/>GameEndedMessage<br/>ErrorMessage"]
        websocket_Conn2["websocket.Conn"]
        context_Context2["context.Context"]
        sync_RWMutex["sync.RWMutex"]
        json_RawMessage["json.RawMessage"]
        
        Client ~~~ BaseGameEvent ~~~ MessageStructs ~~~ websocket_Conn2 ~~~ context_Context2 ~~~ sync_RWMutex ~~~ json_RawMessage
        
        class Client,BaseGameEvent,MessageStructs green
        class websocket_Conn2,context_Context2,sync_RWMutex,json_RawMessage blue
    end

    subgraph gateway_pkg ["gateway"]
        direction LR
        Handler["Handler<br/>upgrader: websocket.Upgrader<br/>lobby: *lobby.Lobby<br/>clients: map[string]*types.Client<br/>mu: sync.RWMutex<br/>ctx: context.Context<br/>cancel: context.CancelFunc"]
        websocket_Upgrader["websocket.Upgrader<br/>ReadBufferSize: int<br/>WriteBufferSize: int<br/>CheckOrigin: func"]
        http_Request["http.Request"]
        http_ResponseWriter["http.ResponseWriter"]
        websocket_Conn3["websocket.Conn"]
        context_Context3["context.Context"]
        sync_RWMutex2["sync.RWMutex"]
        time_Ticker["time.Ticker"]
        
        Handler ~~~ websocket_Upgrader ~~~ http_Request ~~~ http_ResponseWriter ~~~ websocket_Conn3 ~~~ context_Context3 ~~~ sync_RWMutex2 ~~~ time_Ticker
        
        class Handler green
        class websocket_Upgrader,http_Request,http_ResponseWriter,websocket_Conn3,context_Context3,sync_RWMutex2,time_Ticker blue
    end

    subgraph lobby_pkg ["lobby"]
        direction LR
        Lobby["Lobby<br/>clients: map[string]*types.Client<br/>waitingPlayers: map[string]*types.Client<br/>gameRooms: map[string]*gameroom.GameRoom<br/>gameRoomCounter: int<br/>mu: sync.RWMutex<br/>ctx: context.Context<br/>cancel: context.CancelFunc"]
        context_Context4["context.Context"]
        sync_RWMutex3["sync.RWMutex"]
        
        Lobby ~~~ context_Context4 ~~~ sync_RWMutex3
        
        class Lobby green
        class context_Context4,sync_RWMutex3 blue
    end

    subgraph gameroom_pkg ["gameroom"]
        direction LR
        GameRoom["GameRoom<br/>ID: string<br/>Player1: *types.Client<br/>Player2: *types.Client<br/>Player1Wins: int<br/>Player2Wins: int<br/>CurrentRound: int<br/>Player1Choice: Choice<br/>Player2Choice: Choice<br/>Player1Ready: bool<br/>Player2Ready: bool<br/>GameEnded: bool<br/>mu: sync.RWMutex<br/>ctx: context.Context<br/>cancel: context.CancelFunc<br/>onGameEnd: func(string)"]
        Choice["Choice (type string)<br/>Rock, Paper, Scissors"]
        context_Context5["context.Context"]
        sync_RWMutex4["sync.RWMutex"]
        
        GameRoom ~~~ Choice ~~~ context_Context5 ~~~ sync_RWMutex4
        
        class GameRoom,Choice green
        class context_Context5,sync_RWMutex4 blue
    end
    
    %% Package imports as directed edges
    main_paperserver --> gateway_pkg
    main_client --> types_pkg
    gateway_pkg --> lobby_pkg
    gateway_pkg --> types_pkg
    lobby_pkg --> gameroom_pkg
    lobby_pkg --> types_pkg
    gameroom_pkg --> types_pkg
    
    %% Style definitions
    classDef green fill:#90EE90,stroke:#006400,stroke-width:2px,color:#000000
    classDef blue fill:#ADD8E6,stroke:#000080,stroke-width:2px,color:#000000
    classDef grey fill:#D3D3D3,stroke:#696969,stroke-width:2px,color:#000000
    
    %% Subgraph styles
    style main_paperserver fill:#FFFFE0,stroke:#000000,stroke-width:2px,color:#000000
    style main_client fill:#FFFFE0,stroke:#000000,stroke-width:2px,color:#000000
    style types_pkg fill:#FFFFE0,stroke:#000000,stroke-width:2px,color:#000000
    style gateway_pkg fill:#FFFFE0,stroke:#000000,stroke-width:2px,color:#000000
    style lobby_pkg fill:#FFFFE0,stroke:#000000,stroke-width:2px,color:#000000
    style gameroom_pkg fill:#FFFFE0,stroke:#000000,stroke-width:2px,color:#000000
```

## Server Package Structure Diagram

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