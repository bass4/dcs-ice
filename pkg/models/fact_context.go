// File: pkg/models/fact_context.go
package models

// FactContext holds all the necessary data for rule evaluation
type FactContext struct {
	Facts    []Fact       `json:"facts"`
	Response *RuleResponse `json:"response"`
}
