using System;

namespace Paper.Network
{
    [Serializable]
    public class BaseGameEvent<T>
    {
        public string type;
        public T data;
    }

    // For sending messages with typed data
    [Serializable]
    public class JoinLobbyEvent
    {
        public string type = "join_lobby";
        public JoinLobbyMessage data;
    }

    [Serializable]
    public class MakeChoiceEvent
    {
        public string type = "make_choice";
        public MakeChoiceMessage data;
    }

    [Serializable]
    public class PlayAgainEvent
    {
        public string type = "play_again";
        public PlayAgainMessage data;
    }

    [Serializable]
    public class DisconnectEvent
    {
        public string type = "disconnect";
        public DisconnectMessage data;
    }

    // Client to Server Messages
    [Serializable]
    public class JoinLobbyMessage
    {
        public string name;
    }

    [Serializable]
    public class MakeChoiceMessage
    {
        public string choice; // "rock", "paper", "scissors"
    }

    [Serializable]
    public class PlayAgainMessage
    {
        // Empty message
    }

    [Serializable]
    public class DisconnectMessage
    {
        // Empty message
    }

    // Server to Client Messages
    [Serializable]
    public class PlayerWaitingMessage
    {
        // Empty message
    }

    [Serializable]
    public class GameStartingMessage
    {
        public string opponent_name;
    }

    [Serializable]
    public class RoundResultMessage
    {
        public string result;        // "win", "lose", "draw"
        public string your_choice;   // "rock", "paper", "scissors"
        public string opponent_choice; // "rock", "paper", "scissors"
    }

    [Serializable]
    public class RoundStartMessage
    {
        public int round_number;
    }

    [Serializable]
    public class GameEndedMessage
    {
        public string result; // "win", "lose", "draw"
    }

    [Serializable]
    public class ErrorMessage
    {
        public string message;
    }

    // Utility class for message handling
    public static class GameMessageHelper
    {
        // Send message helpers - convert C# objects to JSON
        public static string CreateJoinLobby(string playerName)
        {
            var envelope = new JoinLobbyEvent 
            { 
                data = new JoinLobbyMessage { name = playerName }
            };
            return UnityEngine.JsonUtility.ToJson(envelope);
        }
        
        public static string CreateMakeChoice(string choice)
        {
            var envelope = new MakeChoiceEvent
            {
                data = new MakeChoiceMessage { choice = choice }
            };
            return UnityEngine.JsonUtility.ToJson(envelope);
        }
        
        public static string CreatePlayAgain()
        {
            var envelope = new PlayAgainEvent
            {
                data = new PlayAgainMessage()
            };
            return UnityEngine.JsonUtility.ToJson(envelope);
        }
        
        public static string CreateDisconnect()
        {
            var envelope = new DisconnectEvent
            {
                data = new DisconnectMessage()
            };
            return UnityEngine.JsonUtility.ToJson(envelope);
        }
        
        // Receive message helpers - parse JSON to C# objects
        public static T ParseMessage<T>(string jsonData)
        {
            return UnityEngine.JsonUtility.FromJson<T>(jsonData);
        }
        
        // Simple BaseGameEvent for parsing incoming messages
        [System.Serializable]
        public class IncomingGameEvent
        {
            public string type;
            public string data;
        }
        
        public static IncomingGameEvent ParseBaseEvent(string json)
        {
            try
            {
                // Extract type field
                int typeStart = json.IndexOf("\"type\":\"") + 8;
                int typeEnd = json.IndexOf("\"", typeStart);
                string messageType = json.Substring(typeStart, typeEnd - typeStart);
                
                // Extract data part
                int dataKeyIndex = json.IndexOf("\"data\":");
                if (dataKeyIndex == -1)
                {
                    return new IncomingGameEvent { type = messageType, data = "{}" };
                }
                
                // Find the start of data value (after the colon)
                int colonIndex = json.IndexOf(":", dataKeyIndex);
                int dataStart = colonIndex + 1;
                
                // Skip whitespace
                while (dataStart < json.Length && char.IsWhiteSpace(json[dataStart]))
                    dataStart++;
                
                // Find the end of the data object
                int dataEnd;
                if (json[dataStart] == '{')
                {
                    // Find matching closing brace
                    int braceCount = 1;
                    dataEnd = dataStart + 1;
                    while (dataEnd < json.Length && braceCount > 0)
                    {
                        if (json[dataEnd] == '{') braceCount++;
                        else if (json[dataEnd] == '}') braceCount--;
                        dataEnd++;
                    }
                    dataEnd--; // Point to the closing brace
                }
                else
                {
                    // Handle other data types (shouldn't happen in our protocol)
                    dataEnd = json.LastIndexOf('}') - 1;
                }
                
                string dataJson = json.Substring(dataStart, dataEnd - dataStart + 1);
                
                return new IncomingGameEvent { type = messageType, data = dataJson };
            }
            catch (System.Exception e)
            {
                UnityEngine.Debug.LogError($"Failed to parse JSON: {json}, Error: {e.Message}");
                return new IncomingGameEvent { type = "error", data = "{}" };
            }
        }
        
        // Type-safe message parsing
        public static PlayerWaitingMessage ParsePlayerWaiting(string dataJson)
        {
            return ParseMessage<PlayerWaitingMessage>(dataJson);
        }
        
        public static GameStartingMessage ParseGameStarting(string dataJson)
        {
            return ParseMessage<GameStartingMessage>(dataJson);
        }
        
        public static RoundResultMessage ParseRoundResult(string dataJson)
        {
            return ParseMessage<RoundResultMessage>(dataJson);
        }
        
        public static RoundStartMessage ParseRoundStart(string dataJson)
        {
            return ParseMessage<RoundStartMessage>(dataJson);
        }
        
        public static GameEndedMessage ParseGameEnded(string dataJson)
        {
            return ParseMessage<GameEndedMessage>(dataJson);
        }
        
        public static ErrorMessage ParseError(string dataJson)
        {
            return ParseMessage<ErrorMessage>(dataJson);
        }
    }
}