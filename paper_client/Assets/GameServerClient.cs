using UnityEngine;
using System;
using System.Collections;
using NativeWebSocket;

public class GameServerClient : MonoBehaviour
{
    [SerializeField] private string serverUrl = "ws://localhost:8080/ws";
    
    private WebSocket webSocket;
    
    public event Action<string> OnMessageReceived;
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
                string message = System.Text.Encoding.UTF8.GetString(bytes);
                OnMessageReceived?.Invoke(message);
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