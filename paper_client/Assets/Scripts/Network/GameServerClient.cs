using UnityEngine;
using System;
using System.Collections;
using NativeWebSocket;

namespace Paper.Network
{
    public class GameServerClient : MonoBehaviour
    {
        [SerializeField] private string serverUrl = "wss://paperserver-prd.dingodream.org/ws";
        
        private WebSocket webSocket;
        
        public event Action<string, string> OnMessageReceived; // (messageType, dataJson)
        public event Action OnConnected;
        public event Action<WebSocketCloseCode> OnDisconnected;
        public event Action<string> OnError;
        
        public bool IsConnected => webSocket?.State == WebSocketState.Open;
        
        void Start()
        {
            DontDestroyOnLoad(gameObject);
            Debug.Log("GameServerClient initialized");
        }
        
        void Update()
        {
            webSocket?.DispatchMessageQueue();
        }
        
        public async void Connect()
        {
            if (webSocket?.State == WebSocketState.Open)
                return;
                
            try
            {
                webSocket = new WebSocket(serverUrl);
                
                webSocket.OnOpen += () =>
                {
                    Debug.Log("Connected to game server");
                    OnConnected?.Invoke();
                };
                
                webSocket.OnMessage += (bytes) =>
                {
                    string jsonMessage = System.Text.Encoding.UTF8.GetString(bytes);
                    try
                    {
                        GameMessageHelper.IncomingGameEvent gameEvent = GameMessageHelper.ParseBaseEvent(jsonMessage);
                        if (gameEvent != null)
                        {
                            OnMessageReceived?.Invoke(gameEvent.type, gameEvent.data);
                        }
                    }
                    catch (Exception e)
                    {
                        Debug.LogError($"Failed to parse message: {jsonMessage}. Error: {e.Message}");
                    }
                };
                
                webSocket.OnError += (error) =>
                {
                    Debug.LogError($"WebSocket Error: {error}");
                    OnError?.Invoke(error);
                };
                
                webSocket.OnClose += (closeCode) =>
                {
                    Debug.Log($"Connection closed: {closeCode}");
                    OnDisconnected?.Invoke(closeCode);
                };
                
                await webSocket.Connect();
            }
            catch (Exception e)
            {
                Debug.LogError($"Connection failed: {e.Message}");
                OnError?.Invoke(e.Message);
            }
        }
        
        public new async void SendMessage(string message)
        {
            if (IsConnected)
            {
                await webSocket.SendText(message);
                Debug.Log($"Sent: {message}");
            }
            else
            {
                Debug.LogWarning("Cannot send message: not connected");
            }
        }
        
        // Typed message methods
        public void JoinLobby(string playerName)
        {
            string message = GameMessageHelper.CreateJoinLobby(playerName);
            SendMessage(message);
        }
        
        public void MakeChoice(string choice)
        {
            string message = GameMessageHelper.CreateMakeChoice(choice);
            SendMessage(message);
        }
        
        public void PlayAgain()
        {
            string message = GameMessageHelper.CreatePlayAgain();
            SendMessage(message);
        }
        
        public void DisconnectFromGame()
        {
            string message = GameMessageHelper.CreateDisconnect();
            SendMessage(message);
        }
        
        public async void Disconnect()
        {
            if (webSocket != null)
            {
                await webSocket.Close();
            }
        }
        
        void OnDestroy()
        {
            Disconnect();
        }
        
        void OnApplicationPause(bool pauseStatus)
        {
            // Only disconnect on mobile platforms when pausing
            // Don't disconnect in editor when losing focus
            #if UNITY_ANDROID || UNITY_IOS
            if (pauseStatus)
                Disconnect();
            #endif
        }
        
        void OnApplicationQuit()
        {
            Disconnect();
        }
    }
}