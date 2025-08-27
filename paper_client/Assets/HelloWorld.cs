using UnityEngine;
using UnityEngine.InputSystem;

public class HelloWorld : MonoBehaviour
{
    void Start()
    {
        Debug.Log("Hello World from Unity!");
        Debug.Log("This feels familiar coming from ActionScript!");
    }

    void Update()
    {
        // New Input System syntax
        if (Keyboard.current.spaceKey.wasPressedThisFrame)
        {
            Debug.Log("Space pressed! Frame rate: " + (1.0f / Time.deltaTime));
        }
    }
}