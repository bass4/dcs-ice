// pkg/models/message.go
package models

// Message represents a direct message event from DCS
type Message struct {
    Event     string `json:"event"`
    Zone      string `json:"zone"`
    UnitType  string `json:"unit_type"`
    UnitName  string `json:"unit_name"`
    GroupName string `json:"group_name"`
    Level     string `json:"level"`
    Count     string `json:"count"`
}

func NewMessage(event string) *Message {
    return &Message{
        Event: event,
    }
}
