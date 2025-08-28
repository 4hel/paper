using UnityEngine;
using UnityEngine.InputSystem;

public class HelloWorld : MonoBehaviour
{
    private GameServerClient gameClient;
    
    void Start()
    {
        Debug.Log("Hello World from Unity!");
        Debug.Log("This feels familiar coming from ActionScript!");
        
        // Create GameServerClient programmatically
        GameObject serverObject = new GameObject("GameServerClient");
        gameClient = serverObject.AddComponent<GameServerClient>();
        
        // Subscribe to events
        gameClient.OnConnected += OnServerConnected;
        gameClient.OnMessageReceived += OnMessageReceived;
        
        // Connect to server
        gameClient.Connect();
    }

    void Update()
    {
        // New Input System syntax
        if (Keyboard.current.spaceKey.wasPressedThisFrame)
        {

            
            // Test sending a message
            if (gameClient.IsConnected)
            {
                Debug.Log("connected and sending: join_lobby");
                gameClient.SendMessage("{\"type\":\"join_lobby\",\"data\":{\"name\":\"TestPlayer\"}}");
            } else {
                Debug.Log("not connected");
            }
        }
    }
    
    private void OnServerConnected()
    {
        Debug.Log("Successfully connected to game server!");
    }
    
    private void OnMessageReceived(string message)
    {
        Debug.Log($"received: {message}");
    }
}