using UnityEngine;
using Paper.UI.Core;

public class GameInitializer : MonoBehaviour
{
    void Start()
    {
        Debug.Log("Game initializing...");
        Debug.Log("Paper Rock Scissors client starting up!");
        
        // Create GameUI which will handle everything
        GameObject uiObject = new GameObject("GameUI");
        uiObject.AddComponent<GameUI>();
    }
}