// pkg/models/message_collection.go
package models

import (
	"strconv"
)

// MessageCollection holds multiple messages for batch processing
type MessageCollection struct {
    Messages []*Message
}

// NewMessageCollection creates a new empty message collection
func NewMessageCollection() *MessageCollection {
    return &MessageCollection{
        Messages: make([]*Message, 0),
    }
}

// AddMessage adds a message to the collection
func (mc *MessageCollection) AddMessage(message *Message) {
    mc.Messages = append(mc.Messages, message)
}

// GetMessagesByEvent returns all messages of a specific event type
func (mc *MessageCollection) GetMessagesByEvent(eventType string) []*Message {
    var result []*Message
    for _, msg := range mc.Messages {
        if msg.Event == eventType {
            result = append(result, msg)
        }
    }
    return result
}

// GetMessagesFromZone returns all messages from a specific zone
func (mc *MessageCollection) GetMessagesFromZone(zone string) []*Message {
    var result []*Message
    for _, msg := range mc.Messages {
        if msg.Zone == zone {
            result = append(result, msg)
        }
    }
    return result
}

// GetMessagesByEventAndZone returns all messages of a specific event type from a specific zone
func (mc *MessageCollection) GetMessagesByEventAndZone(eventType, zone string) []*Message {
    var result []*Message
    for _, msg := range mc.Messages {
        if msg.Event == eventType && msg.Zone == zone {
            result = append(result, msg)
        }
    }
    return result
}

// CountMessagesByEvent returns the count of messages of a specific event type
func (mc *MessageCollection) CountMessagesByEvent(eventType string) int {
    count := 0
    for _, msg := range mc.Messages {
        if msg.Event == eventType {
            count++
        }
    }
    return count
}

// HasDetectionsInBothZones checks if there are unit_detected events in both specified zones
func (mc *MessageCollection) HasDetectionsInBothZones(zone1, zone2 string) bool {
    hasZone1 := false
    hasZone2 := false
    
    for _, msg := range mc.Messages {
        if msg.Event == "unit_detected" {
            if msg.Zone == zone1 {
                hasZone1 = true
            } else if msg.Zone == zone2 {
                hasZone2 = true
            }
            
            if hasZone1 && hasZone2 {
                return true
            }
        }
    }
    
    return false
}

// GetTotalDetectedUnits returns the total count of detected units across all messages
func (mc *MessageCollection) GetTotalDetectedUnits() int {
    total := 0
    for _, msg := range mc.Messages {
        if msg.Event == "unit_detected" {
            // Parse count to int, default to 1 if parsing fails
            count := 1
            if msg.Count != "" {
                if parsedCount, err := strconv.Atoi(msg.Count); err == nil {
                    count = parsedCount
                }
            }
            total += count
        }
    }
    return total
}
