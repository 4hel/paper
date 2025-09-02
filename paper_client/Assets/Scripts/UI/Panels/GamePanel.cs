using UnityEngine;
using UnityEngine.UI;
using System;

namespace Paper.UI.Panels
{
    public class GamePanel : MonoBehaviour
    {
        private GameObject panel;
        private Text opponentNameText;
        private Text gameStatusText;
        private Text resultText;
        private Button rockButton;
        private Button paperButton;
        private Button scissorsButton;
        private Button playAgainButton;
        private Button disconnectButton;
        
        public event Action<string> OnChoiceMade;
        public event Action OnPlayAgainRequested;
        public event Action OnDisconnectRequested;
        
        public void Initialize(Canvas parentCanvas)
        {
            CreatePanel(parentCanvas);
        }
        
        void CreatePanel(Canvas parentCanvas)
        {
            panel = new GameObject("GamePanel");
            panel.transform.SetParent(parentCanvas.transform, false);
            
            Image panelImage = panel.AddComponent<Image>();
            panelImage.color = new Color(0.1f, 0.3f, 0.1f, 0.8f); // Different color for game panel
            
            RectTransform panelRect = panel.GetComponent<RectTransform>();
            panelRect.sizeDelta = new Vector2(500, 300);
            panelRect.anchoredPosition = Vector2.zero;
            
            // Title
            CreateText("Rock Paper Scissors", panel, new Vector2(0, 120), 24);
            
            // Opponent name
            opponentNameText = CreateText("vs Opponent", panel, new Vector2(0, 80), 18, "OpponentName");
            
            // Game status
            gameStatusText = CreateText("Make your choice!", panel, new Vector2(0, 40), 16, "GameStatus");
            
            // Choice buttons
            rockButton = CreateButton("✊ Rock", panel, new Vector2(-120, -20));
            rockButton.onClick.AddListener(() => OnChoiceClick("rock"));
            
            paperButton = CreateButton("📄 Paper", panel, new Vector2(0, -20));
            paperButton.onClick.AddListener(() => OnChoiceClick("paper"));
            
            scissorsButton = CreateButton("✂️ Scissors", panel, new Vector2(120, -20));
            scissorsButton.onClick.AddListener(() => OnChoiceClick("scissors"));
            
            // Result text area
            resultText = CreateText("", panel, new Vector2(0, -80), 14, "ResultText");
            
            // End game buttons (initially hidden)
            playAgainButton = CreateButton("Play Again", panel, new Vector2(-80, -120));
            playAgainButton.onClick.AddListener(() => OnPlayAgainRequested?.Invoke());
            playAgainButton.gameObject.SetActive(false);
            
            disconnectButton = CreateButton("Disconnect", panel, new Vector2(80, -120));
            disconnectButton.onClick.AddListener(() => OnDisconnectRequested?.Invoke());
            disconnectButton.gameObject.SetActive(false);
            
            // Initially hidden
            panel.SetActive(false);
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
        
        Button CreateButton(string text, GameObject parent, Vector2 position)
        {
            GameObject buttonObject = new GameObject("GameButton");
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
        
        void OnChoiceClick(string choice)
        {
            OnChoiceMade?.Invoke(choice);
            gameStatusText.text = $"You chose {choice}. Waiting for opponent...";
            SetChoiceButtonsEnabled(false);
        }
        
        public void SetOpponentName(string opponentName)
        {
            opponentNameText.text = $"vs {opponentName}";
        }
        
        public void UpdateGameStatus(string status)
        {
            gameStatusText.text = status;
        }
        
        public void UpdateResultText(string result)
        {
            resultText.text = result;
        }
        
        public void SetChoiceButtonsEnabled(bool enabled)
        {
            rockButton.interactable = enabled;
            paperButton.interactable = enabled;
            scissorsButton.interactable = enabled;
        }
        
        public void ShowChoiceButtons()
        {
            rockButton.gameObject.SetActive(true);
            paperButton.gameObject.SetActive(true);
            scissorsButton.gameObject.SetActive(true);
            
            playAgainButton.gameObject.SetActive(false);
            disconnectButton.gameObject.SetActive(false);
            
            SetChoiceButtonsEnabled(true);
        }
        
        public void ShowEndGameButtons()
        {
            rockButton.gameObject.SetActive(false);
            paperButton.gameObject.SetActive(false);
            scissorsButton.gameObject.SetActive(false);
            
            playAgainButton.gameObject.SetActive(true);
            disconnectButton.gameObject.SetActive(true);
        }
        
        public void ShowWaitingState()
        {
            rockButton.gameObject.SetActive(false);
            paperButton.gameObject.SetActive(false);
            scissorsButton.gameObject.SetActive(false);
            playAgainButton.gameObject.SetActive(false);
            disconnectButton.gameObject.SetActive(false);
            
            UpdateGameStatus("Returning to lobby...");
            UpdateResultText("Waiting for opponent to join...");
        }
        
        public void Show()
        {
            panel.SetActive(true);
        }
        
        public void Hide()
        {
            panel.SetActive(false);
        }
    }
}