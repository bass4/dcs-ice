package models

// Fact represents a single fact received from DCS World
type Fact struct {
	Event      string            `json:"event"`
	Unit       string            `json:"unit"`
	Zone       string            `json:"zone"`
	AlertLevel string            `json:"alertLevel"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// FactsRequest represents the incoming request to evaluate facts
type FactsRequest struct {
	Facts []Fact `json:"facts"`
}
