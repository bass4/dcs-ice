// File: pkg/models/actions.go
package models

// Action represents an action to be taken in response to a rule match
type Action struct {
	Type       string            `json:"type"`
	Target     string            `json:"target,omitempty"`
	Parameters map[string]string `json:"parameters,omitempty"`
	Timestamp  int64             `json:"timestamp,omitempty"` // Unix timestamp in milliseconds
}

// RuleResponse represents the response from rule evaluation
type RuleResponse struct {
	MatchedRules []string `json:"matchedRules"`
	Actions      []Action `json:"actions"`
}
