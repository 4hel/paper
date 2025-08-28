using UnityEngine;
using UnityEngine.UI;
using UnityEngine.EventSystems;
using UnityEngine.InputSystem.UI;

public class GameUI : MonoBehaviour
{
    private Canvas canvas;
    private InputField nameInput;
    private Button joinButton;
    private GameServerClient gameClient;
    
    // UI Panels
    private GameObject loginPanel;
    private GameObject gamePanel;
    
    // Game UI elements
    private Text opponentNameText;
    private Text gameStatusText;
    private Button rockButton;
    private Button paperButton;
    private Button scissorsButton;
    
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
        
        // Create login panel
        CreateLoginPanel();
        
        // Create game panel (initially hidden)
        CreateGamePanel();
    }
    
    void CreateLoginPanel()
    {
        loginPanel = new GameObject("LoginPanel");
        loginPanel.transform.SetParent(canvas.transform, false);
        
        Image panelImage = loginPanel.AddComponent<Image>();
        panelImage.color = new Color(0.2f, 0.2f, 0.2f, 0.8f);
        
        RectTransform panelRect = loginPanel.GetComponent<RectTransform>();
        panelRect.sizeDelta = new Vector2(400, 200);
        panelRect.anchoredPosition = Vector2.zero;
        
        // Create title text
        CreateText("Rock Paper Scissors", loginPanel, new Vector2(0, 60), 24);
        
        // Create name input field
        nameInput = CreateInputField("Enter your name...", loginPanel, new Vector2(0, 10));
        
        // Create join button
        joinButton = CreateButton("Join Game", loginPanel, new Vector2(0, -40));
        joinButton.onClick.AddListener(OnJoinButtonClick);
        
        // Create status text (initially empty)
        CreateText("", loginPanel, new Vector2(0, -80), 16, "StatusText");
    }
    
    void CreateGamePanel()
    {
        gamePanel = new GameObject("GamePanel");
        gamePanel.transform.SetParent(canvas.transform, false);
        
        Image panelImage = gamePanel.AddComponent<Image>();
        panelImage.color = new Color(0.1f, 0.3f, 0.1f, 0.8f); // Different color for game panel
        
        RectTransform panelRect = gamePanel.GetComponent<RectTransform>();
        panelRect.sizeDelta = new Vector2(500, 300);
        panelRect.anchoredPosition = Vector2.zero;
        
        // Title
        CreateText("Rock Paper Scissors", gamePanel, new Vector2(0, 120), 24);
        
        // Opponent name
        opponentNameText = CreateText("vs Opponent", gamePanel, new Vector2(0, 80), 18, "OpponentName");
        
        // Game status
        gameStatusText = CreateText("Make your choice!", gamePanel, new Vector2(0, 40), 16, "GameStatus");
        
        // Choice buttons
        rockButton = CreateButton("✊ Rock", gamePanel, new Vector2(-120, -20));
        rockButton.onClick.AddListener(() => OnChoiceClick("rock"));
        
        paperButton = CreateButton("📄 Paper", gamePanel, new Vector2(0, -20));
        paperButton.onClick.AddListener(() => OnChoiceClick("paper"));
        
        scissorsButton = CreateButton("✂️ Scissors", gamePanel, new Vector2(120, -20));
        scissorsButton.onClick.AddListener(() => OnChoiceClick("scissors"));
        
        // Result text area
        CreateText("", gamePanel, new Vector2(0, -80), 14, "ResultText");
        
        // Initially hidden
        gamePanel.SetActive(false);
    }
    
    Text CreateText(string content, GameObject parent, Vector2 position, int fontSize, string name = "Text")
    {
        GameObject textObject = new GameObject(name);
        textObject.transform.SetParent(parent.transform, false);
        
        Text text = textObject.AddComponent<Text>();
        text.text = content;
        text.fontSize = fontSize;
        text.color = Color.white;
        text.alignment = TextAnchor.MiddleCenter;
        text.font = Resources.GetBuiltinResource<Font>("LegacyRuntime.ttf");
        
        RectTransform textRect = textObject.GetComponent<RectTransform>();
        textRect.sizeDelta = new Vector2(350, 30);
        textRect.anchoredPosition = position;
        
        return text;
    }
    
    InputField CreateInputField(string placeholder, GameObject parent, Vector2 position)
    {
        GameObject inputObject = new GameObject("NameInput");
        inputObject.transform.SetParent(parent.transform, false);
        
        Image inputImage = inputObject.AddComponent<Image>();
        inputImage.color = Color.white;
        
        InputField inputField = inputObject.AddComponent<InputField>();
        
        // Create child text object for the input
        GameObject textObject = new GameObject("Text");
        textObject.transform.SetParent(inputObject.transform, false);
        
        Text inputText = textObject.AddComponent<Text>();
        inputText.text = "";
        inputText.fontSize = 16;
        inputText.color = Color.black;
        inputText.font = Resources.GetBuiltinResource<Font>("LegacyRuntime.ttf");
        
        RectTransform textRect = textObject.GetComponent<RectTransform>();
        textRect.sizeDelta = new Vector2(-20, -10);
        textRect.anchoredPosition = Vector2.zero;
        textRect.anchorMin = Vector2.zero;
        textRect.anchorMax = Vector2.one;
        textRect.offsetMin = new Vector2(10, 5);
        textRect.offsetMax = new Vector2(-10, -5);
        
        inputField.textComponent = inputText;
        
        // Create placeholder
        GameObject placeholderObject = new GameObject("Placeholder");
        placeholderObject.transform.SetParent(inputObject.transform, false);
        
        Text placeholderText = placeholderObject.AddComponent<Text>();
        placeholderText.text = placeholder;
        placeholderText.fontSize = 16;
        placeholderText.color = new Color(0.5f, 0.5f, 0.5f, 1f);
        placeholderText.font = Resources.GetBuiltinResource<Font>("LegacyRuntime.ttf");
        
        RectTransform placeholderRect = placeholderObject.GetComponent<RectTransform>();
        placeholderRect.sizeDelta = new Vector2(-20, -10);
        placeholderRect.anchoredPosition = Vector2.zero;
        placeholderRect.anchorMin = Vector2.zero;
        placeholderRect.anchorMax = Vector2.one;
        placeholderRect.offsetMin = new Vector2(10, 5);
        placeholderRect.offsetMax = new Vector2(-10, -5);
        
        inputField.placeholder = placeholderText;
        
        RectTransform inputRect = inputObject.GetComponent<RectTransform>();
        inputRect.sizeDelta = new Vector2(300, 35);
        inputRect.anchoredPosition = position;
        
        return inputField;
    }
    
    Button CreateButton(string text, GameObject parent, Vector2 position)
    {
        GameObject buttonObject = new GameObject("JoinButton");
        buttonObject.transform.SetParent(parent.transform, false);
        
        Image buttonImage = buttonObject.AddComponent<Image>();
        buttonImage.color = new Color(0.2f, 0.6f, 1f, 1f);
        
        Button button = buttonObject.AddComponent<Button>();
        
        GameObject buttonTextObject = new GameObject("Text");
        buttonTextObject.transform.SetParent(buttonObject.transform, false);
        
        Text buttonText = buttonTextObject.AddComponent<Text>();
        buttonText.text = text;
        buttonText.fontSize = 18;
        buttonText.color = Color.white;
        buttonText.alignment = TextAnchor.MiddleCenter;
        buttonText.font = Resources.GetBuiltinResource<Font>("LegacyRuntime.ttf");
        
        RectTransform buttonTextRect = buttonTextObject.GetComponent<RectTransform>();
        buttonTextRect.anchorMin = Vector2.zero;
        buttonTextRect.anchorMax = Vector2.one;
        buttonTextRect.sizeDelta = Vector2.zero;
        buttonTextRect.anchoredPosition = Vector2.zero;
        
        RectTransform buttonRect = buttonObject.GetComponent<RectTransform>();
        buttonRect.sizeDelta = new Vector2(150, 40);
        buttonRect.anchoredPosition = position;
        
        return button;
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
    
    void OnJoinButtonClick()
    {
        string playerName = nameInput.text.Trim();
        
        if (string.IsNullOrEmpty(playerName))
        {
            UpdateStatus("Please enter a name!");
            return;
        }
        
        if (!gameClient.IsConnected)
        {
            UpdateStatus("Not connected to server!");
            return;
        }
        
        gameClient.JoinLobby(playerName);
        
        joinButton.interactable = false;
        UpdateStatus($"Joining as {playerName}...");
    }
    
    void OnServerConnected()
    {
        UpdateStatus("Connected to server!");
        joinButton.interactable = true;
    }
    
    void OnMessageReceived(string messageType, string dataJson)
    {
        Debug.Log($"Received {messageType}: {dataJson}");
        
        switch (messageType)
        {
            case "player_waiting":
                var waitingMsg = GameMessageHelper.ParsePlayerWaiting(dataJson);
                UpdateStatus("Waiting for opponent...");
                break;
                
            case "game_starting":
                var startingMsg = GameMessageHelper.ParseGameStarting(dataJson);
                SwitchToGameView(startingMsg.opponent_name);
                break;
                
            case "round_start":
                var roundStartMsg = GameMessageHelper.ParseRoundStart(dataJson);
                UpdateGameStatus($"Round {roundStartMsg.round_number} - Make your choice!");
                SetChoiceButtonsEnabled(true);
                break;
                
            case "round_result":
                var resultMsg = GameMessageHelper.ParseRoundResult(dataJson);
                UpdateGameStatus($"You {resultMsg.result}!");
                UpdateResultText($"You: {resultMsg.your_choice} | Opponent: {resultMsg.opponent_choice}");
                break;
                
            case "game_ended":
                var endMsg = GameMessageHelper.ParseGameEnded(dataJson);
                UpdateGameStatus($"Game Over! You {endMsg.result}!");
                UpdateResultText("Click 'Play Again' or disconnect");
                // Could add a "Play Again" button here
                break;
                
            case "error":
                var errorMsg = GameMessageHelper.ParseError(dataJson);
                UpdateStatus($"Error: {errorMsg.message}");
                joinButton.interactable = true;
                break;
                
            default:
                UpdateStatus($"Unknown message: {messageType}");
                break;
        }
    }
    
    void OnError(string error)
    {
        UpdateStatus($"Error: {error}");
        joinButton.interactable = true;
    }
    
    void SwitchToGameView(string opponentName)
    {
        loginPanel.SetActive(false);
        gamePanel.SetActive(true);
        
        opponentNameText.text = $"vs {opponentName}";
        gameStatusText.text = "Game starting...";
        
        // Enable choice buttons
        SetChoiceButtonsEnabled(true);
    }
    
    void SwitchToLoginView()
    {
        gamePanel.SetActive(false);
        loginPanel.SetActive(true);
        joinButton.interactable = true;
    }
    
    void OnChoiceClick(string choice)
    {
        gameClient.MakeChoice(choice);
        gameStatusText.text = $"You chose {choice}. Waiting for opponent...";
        
        // Disable buttons until next round
        SetChoiceButtonsEnabled(false);
    }
    
    void SetChoiceButtonsEnabled(bool enabled)
    {
        rockButton.interactable = enabled;
        paperButton.interactable = enabled;
        scissorsButton.interactable = enabled;
    }
    
    void UpdateGameStatus(string status)
    {
        if (gameStatusText != null)
        {
            gameStatusText.text = status;
        }
    }
    
    void UpdateResultText(string result)
    {
        Text resultText = GameObject.Find("ResultText")?.GetComponent<Text>();
        if (resultText != null)
        {
            resultText.text = result;
        }
    }
    
    void UpdateStatus(string status)
    {
        Text statusText = GameObject.Find("StatusText")?.GetComponent<Text>();
        if (statusText != null)
        {
            statusText.text = status;
        }
    }
}