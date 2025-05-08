package models

// Action represents an action to be performed as a result of rule evaluation
type Action struct {
    Type      string `json:"type"`
    SubType   string `json:"sub_type,omitempty"`
    Zone      string `json:"zone,omitempty"`
    UnitType  string `json:"unit_type,omitempty"`
    Count     string `json:"count,omitempty"`
    Level     string `json:"level,omitempty"`
    Message   string `json:"message,omitempty"`
    GroupName string `json:"group_name,omitempty"`
}
