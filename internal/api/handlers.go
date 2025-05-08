// internal/api/handlers.go
package api

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"

    "github.com/gorilla/websocket"

    "github.com/bass4/dcs-ice/internal/rules"
    "github.com/bass4/dcs-ice/pkg/models"
)

// DCSEvent represents the JSON structure coming from DCS
type DCSEvent struct {
    EventType string                 `json:"event_type"`
    Timestamp int64                  `json:"timestamp"`
    Data      map[string]interface{} `json:"data"`
}

// DCSAction represents an action to be sent back to DCS
type DCSAction struct {
    ActionType string                 `json:"action_type"`
    SubType    string                 `json:"sub_type,omitempty"`
    Data       map[string]interface{} `json:"data"`
}

// DCSResponse represents the complete response to DCS
type DCSResponse struct {
    Status  string      `json:"status"`
    Actions []DCSAction `json:"actions"`
}

// DCSEventHandler handles incoming DCS events via HTTP
func DCSEventHandler(ruleEngine *rules.RuleEngine) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Parse the incoming JSON
        var dcsEvent DCSEvent
        if err := json.NewDecoder(r.Body).Decode(&dcsEvent); err != nil {
            http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
            return
        }

        fmt.Printf("Received event: %s\n", dcsEvent.EventType)

        // Convert DCS event to Message
        message := convertDCSEventToMessage(dcsEvent)

        // Process the message through the rules engine
        actions, err := ruleEngine.ProcessMessage(message)
        if err != nil {
            http.Error(w, "Rule processing failed: "+err.Error(), http.StatusInternalServerError)
            return
        }

        // Convert actions to DCS response format
        dcsResponse := convertActionsToDCSResponse(actions)

        // Send response back to DCS
        w.Header().Set("Content-Type", "application/json")
        if err := json.NewEncoder(w).Encode(dcsResponse); err != nil {
            http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
            return
        }
        
        fmt.Println("Response sent successfully")
    }
}

// WebSocket support
var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

// DCSWebSocketHandler handles WebSocket connections from DCS
func DCSWebSocketHandler(ruleEngine *rules.RuleEngine) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        conn, err := upgrader.Upgrade(w, r, nil)
        if err != nil {
            log.Println("WebSocket upgrade failed:", err)
            return
        }
        defer conn.Close()

        log.Println("WebSocket connection established")

        // WebSocket message handling loop
        for {
            messageType, messageData, err := conn.ReadMessage()
            if err != nil {
                if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                    log.Printf("WebSocket error: %v", err)
                }
                break
            }

            // Parse DCS event
            var dcsEvent DCSEvent
            if err := json.Unmarshal(messageData, &dcsEvent); err != nil {
                log.Printf("Invalid JSON: %v", err)
                continue
            }

            log.Printf("Received event: %s", dcsEvent.EventType)

            // Convert and process
            message := convertDCSEventToMessage(dcsEvent)
            actions, err := ruleEngine.ProcessMessage(message)
            if err != nil {
                log.Printf("Rule processing failed: %v", err)
                continue
            }

            // Convert actions to response
            dcsResponse := convertActionsToDCSResponse(actions)
            responseJSON, err := json.Marshal(dcsResponse)
            if err != nil {
                log.Printf("Failed to encode response: %v", err)
                continue
            }

            // Send response back to DCS
            if err := conn.WriteMessage(messageType, responseJSON); err != nil {
                log.Printf("Failed to send response: %v", err)
                break
            }

            log.Printf("Sent %d actions back to DCS", len(dcsResponse.Actions))
        }
    }
}

// internal/api/handlers.go

// ReloadRulesHandler provides an endpoint to reload rules
func ReloadRulesHandler(ruleEngine *rules.RuleEngine) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if err := ruleEngine.ReloadRules(); err != nil {
            http.Error(w, "Failed to reload rules: "+err.Error(), http.StatusInternalServerError)
            return
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Rules reloaded successfully"})
    }
}

// Helper functions for data conversion

// convertDCSEventToMessage converts a DCS event to a Message
func convertDCSEventToMessage(dcsEvent DCSEvent) *models.Message {
    message := models.NewMessage(dcsEvent.EventType)
    
    // Extract common fields
    getString := func(data map[string]interface{}, key string) string {
        if val, ok := data[key]; ok {
            if strVal, ok := val.(string); ok {
                return strVal
            }
        }
        return ""
    }
    
    // Set fields based on the event data
    message.Zone = getString(dcsEvent.Data, "zone")
    message.UnitType = getString(dcsEvent.Data, "unit_type")
    message.UnitName = getString(dcsEvent.Data, "unit_name")
    message.GroupName = getString(dcsEvent.Data, "group_name")
    
    // Handle specific event types
    switch dcsEvent.EventType {
    case "alert_level_change":
        message.Level = getString(dcsEvent.Data, "level")
    case "unit_detected":
        message.Count = getString(dcsEvent.Data, "count")
        if message.Count == "" {
            message.Count = "1" // Default to 1 if not specified
        }
    }
    
    fmt.Printf("Created message: Event=%s, Zone=%s, UnitType=%s\n", 
        message.Event, message.Zone, message.UnitType)
    
    return message
}


// convertActionsToDCSResponse converts internal actions to DCS response format
func convertActionsToDCSResponse(actions []models.Action) DCSResponse {
    response := DCSResponse{
        Status:  "success",
        Actions: make([]DCSAction, 0, len(actions)),
    }

    for _, action := range actions {
        dcsAction := DCSAction{
            ActionType: action.Type,
            SubType:    action.SubType,
            Data:       make(map[string]interface{}),
        }

        // Convert each action type to the appropriate DCS action format
        switch action.Type {
        case "spawn":
            dcsAction.Data["zone"] = action.Zone
            dcsAction.Data["unit_type"] = action.UnitType
            dcsAction.Data["count"] = action.Count

        case "alert":
            dcsAction.Data["level"] = action.Level
            dcsAction.Data["message"] = action.Message
            if action.Zone != "" {
                dcsAction.Data["zone"] = action.Zone
            }

        case "reinforce":
            dcsAction.Data["unit_type"] = action.UnitType
            dcsAction.Data["group_name"] = action.GroupName
            dcsAction.Data["zone"] = action.Zone
            dcsAction.Data["count"] = action.Count
        }

        response.Actions = append(response.Actions, dcsAction)
    }

    return response
}
// in internal/api/handlers.go
// Add a function to batch process messages

// BatchProcessEvents processes multiple events at once
func BatchProcessEvents(ruleEngine *rules.RuleEngine, dcsEvents []DCSEvent) ([]models.Action, error) {
    var messages []*models.Message
    
    // Convert all events to messages
    for _, event := range dcsEvents {
        message := convertDCSEventToMessage(event)
        messages = append(messages, message)
    }
    
    // Process all messages at once
    return ruleEngine.ProcessMessages(messages)
}

// Add to handlers.go
// BatchDCSEventHandler handles batches of DCS events
func BatchDCSEventHandler(ruleEngine *rules.RuleEngine) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Parse the incoming JSON array
        var dcsEvents []DCSEvent
        if err := json.NewDecoder(r.Body).Decode(&dcsEvents); err != nil {
            http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
            return
        }

        fmt.Printf("Received batch of %d events\n", len(dcsEvents))
        for i, event := range dcsEvents {
            fmt.Printf("Event %d: Type=%s, Zone=%s\n", i, event.EventType, 
                event.Data["zone"])
        }

        // Process all events at once
        actions, err := BatchProcessEvents(ruleEngine, dcsEvents)
        if err != nil {
            http.Error(w, "Rule processing failed: "+err.Error(), http.StatusInternalServerError)
            return
        }

        // Convert actions to DCS response format
        dcsResponse := convertActionsToDCSResponse(actions)
        
        fmt.Printf("Response: %+v\n", dcsResponse)

        // Send response back to DCS
        w.Header().Set("Content-Type", "application/json")
        if err := json.NewEncoder(w).Encode(dcsResponse); err != nil {
            http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
            return
        }
        
        fmt.Println("Response sent successfully")
    }
}