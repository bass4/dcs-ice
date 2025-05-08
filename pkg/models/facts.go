package models

// Fact represents a piece of information for the rules engine
type Fact struct {
    ID        string `json:"id,omitempty"`
    Type      string `json:"type"`
    Value     string `json:"value"`
    Zone      string `json:"zone,omitempty"`
    UnitType  string `json:"unit_type,omitempty"`
    GroupName string `json:"group_name,omitempty"`
    Count     string `json:"count,omitempty"`
}
