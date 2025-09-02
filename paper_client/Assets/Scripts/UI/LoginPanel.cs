using UnityEngine;
using UnityEngine.UI;
using System;

namespace Scripts.UI
{
    public class LoginPanel : MonoBehaviour
    {
        private GameObject panel;
        private InputField nameInput;
        private Button joinButton;
        private Text statusText;
        
        public event Action<string> OnJoinRequested;
        
        public void Initialize(Canvas parentCanvas)
        {
            CreatePanel(parentCanvas);
        }
        
        void CreatePanel(Canvas parentCanvas)
        {
            panel = new GameObject("LoginPanel");
            panel.transform.SetParent(parentCanvas.transform, false);
            
            Image panelImage = panel.AddComponent<Image>();
            panelImage.color = new Color(0.2f, 0.2f, 0.2f, 0.8f);
            
            RectTransform panelRect = panel.GetComponent<RectTransform>();
            panelRect.sizeDelta = new Vector2(400, 200);
            panelRect.anchoredPosition = Vector2.zero;
            
            // Create title text
            CreateText("Rock Paper Scissors", panel, new Vector2(0, 60), 24);
            
            // Create name input field
            nameInput = CreateInputField("Enter your name...", panel, new Vector2(0, 10));
            
            // Create join button
            joinButton = CreateButton("Join Game", panel, new Vector2(0, -40));
            joinButton.onClick.AddListener(OnJoinButtonClick);
            
            // Create status text (initially empty)
            statusText = CreateText("", panel, new Vector2(0, -80), 16, "StatusText");
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
        
        void OnJoinButtonClick()
        {
            string playerName = nameInput.text.Trim();
            
            if (string.IsNullOrEmpty(playerName))
            {
                UpdateStatus("Please enter a name!");
                return;
            }
            
            OnJoinRequested?.Invoke(playerName);
        }
        
        public void SetJoinButtonEnabled(bool enabled)
        {
            joinButton.interactable = enabled;
        }
        
        public void UpdateStatus(string status)
        {
            if (statusText != null)
            {
                statusText.text = status;
            }
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