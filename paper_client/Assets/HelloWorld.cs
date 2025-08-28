using UnityEngine;

public class HelloWorld : MonoBehaviour
{
    void Start()
    {
        Debug.Log("Hello World from Unity!");
        Debug.Log("This feels familiar coming from ActionScript!");
        
        // Create GameUI which will handle everything
        GameObject uiObject = new GameObject("GameUI");
        uiObject.AddComponent<GameUI>();
    }
}