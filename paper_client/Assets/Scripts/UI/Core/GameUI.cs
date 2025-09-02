using UnityEngine;
using UnityEngine.UI;
using UnityEngine.EventSystems;
using UnityEngine.InputSystem.UI;
using Paper.Network;
using Paper.UI.Panels;

namespace Paper.UI.Core
{
    public class GameUI : MonoBehaviour
    {
        private Canvas canvas;
        private GameServerClient gameClient;
        private LoginPanel loginPanel;
        private GamePanel gamePanel;
        
        void Start()
        {
            CreateUI();
            SetupGameClient();
        }
        
        void CreateUI()
        {
            // Create EventSystem (required for UI input)
            if (FindObjectOfType<EventSystem>() == null)
            {
                GameObject eventSystemObject = new GameObject("EventSystem");
                eventSystemObject.AddComponent<EventSystem>();
                eventSystemObject.AddComponent<InputSystemUIInputModule>();
            }
            
            // Create Canvas
            GameObject canvasObject = new GameObject("Canvas");
            canvas = canvasObject.AddComponent<Canvas>();
            canvas.renderMode = RenderMode.ScreenSpaceOverlay;
            canvasObject.AddComponent<CanvasScaler>();
            canvasObject.AddComponent<GraphicRaycaster>();
            
            // Create panels
            CreatePanels();
        }
        
        void CreatePanels()
        {
            // Create login panel
            GameObject loginObject = new GameObject("LoginPanel");
            loginPanel = loginObject.AddComponent<LoginPanel>();
            loginPanel.Initialize(canvas);
            loginPanel.OnJoinRequested += OnJoinRequested;
            
            // Create game panel (initially hidden)
            GameObject gameObject = new GameObject("GamePanel");
            gamePanel = gameObject.AddComponent<GamePanel>();
            gamePanel.Initialize(canvas);
            gamePanel.OnChoiceMade += OnChoiceMade;
            gamePanel.OnPlayAgainRequested += OnPlayAgainRequested;
            gamePanel.OnDisconnectRequested += OnDisconnectRequested;
        }
        
        void SetupGameClient()
        {
            GameObject serverObject = new GameObject("GameServerClient");
            gameClient = serverObject.AddComponent<GameServerClient>();
            
            gameClient.OnConnected += OnServerConnected;
            gameClient.OnMessageReceived += OnMessageReceived;
            gameClient.OnError += OnError;
            
            gameClient.Connect();
        }
        
        void OnJoinRequested(string playerName)
        {
            if (!gameClient.IsConnected)
            {
                loginPanel.UpdateStatus("Not connected to server!");
                return;
            }
            
            gameClient.JoinLobby(playerName);
            loginPanel.SetJoinButtonEnabled(false);
            loginPanel.UpdateStatus($"Joining as {playerName}...");
        }
        
        void OnChoiceMade(string choice)
        {
            gameClient.MakeChoice(choice);
        }
        
        void OnPlayAgainRequested()
        {
            gameClient.PlayAgain();
            gamePanel.ShowWaitingState();
        }
        
        void OnDisconnectRequested()
        {
            gameClient.DisconnectFromGame();
            SwitchToLoginView();
        }
        
        void OnServerConnected()
        {
            loginPanel.UpdateStatus("Connected to server!");
            loginPanel.SetJoinButtonEnabled(true);
        }
        
        void OnMessageReceived(string messageType, string dataJson)
        {
            Debug.Log($"Received {messageType}: {dataJson}");
            
            switch (messageType)
            {
                case "player_waiting":
                    var waitingMsg = GameMessageHelper.ParsePlayerWaiting(dataJson);
                    // If we're already in game view, show waiting state
                    if (gamePanel.gameObject.activeInHierarchy)
                    {
                        gamePanel.ShowWaitingState();
                    }
                    else
                    {
                        loginPanel.UpdateStatus("Waiting for opponent...");
                    }
                    break;
                    
                case "game_starting":
                    var startingMsg = GameMessageHelper.ParseGameStarting(dataJson);
                    SwitchToGameView(startingMsg.opponent_name);
                    gamePanel.ShowChoiceButtons(); // Make sure choice buttons are visible
                    break;
                    
                case "round_start":
                    var roundStartMsg = GameMessageHelper.ParseRoundStart(dataJson);
                    gamePanel.UpdateGameStatus($"Round {roundStartMsg.round_number} - Make your choice!");
                    gamePanel.ShowChoiceButtons(); // Show choice buttons for new round
                    break;
                    
                case "round_result":
                    var resultMsg = GameMessageHelper.ParseRoundResult(dataJson);
                    gamePanel.UpdateGameStatus($"You {resultMsg.result}!");
                    gamePanel.UpdateResultText($"You: {resultMsg.your_choice} | Opponent: {resultMsg.opponent_choice}");
                    break;
                    
                case "game_ended":
                    var endMsg = GameMessageHelper.ParseGameEnded(dataJson);
                    gamePanel.UpdateGameStatus($"Game Over! You {endMsg.result}!");
                    gamePanel.UpdateResultText("Choose your next action:");
                    gamePanel.ShowEndGameButtons(); // Show Play Again and Disconnect buttons
                    break;
                    
                case "error":
                    var errorMsg = GameMessageHelper.ParseError(dataJson);
                    loginPanel.UpdateStatus($"Error: {errorMsg.message}");
                    loginPanel.SetJoinButtonEnabled(true);
                    break;
                    
                default:
                    loginPanel.UpdateStatus($"Unknown message: {messageType}");
                    break;
            }
        }
        
        void OnError(string error)
        {
            loginPanel.UpdateStatus($"Error: {error}");
            loginPanel.SetJoinButtonEnabled(true);
        }
        
        void SwitchToGameView(string opponentName)
        {
            loginPanel.Hide();
            gamePanel.Show();
            gamePanel.SetOpponentName(opponentName);
            gamePanel.UpdateGameStatus("Game starting...");
        }
        
        void SwitchToLoginView()
        {
            gamePanel.Hide();
            loginPanel.Show();
            loginPanel.SetJoinButtonEnabled(true);
        }
    }
}