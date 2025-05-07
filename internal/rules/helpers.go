// File: internal/rules/helpers.go
package rules

// Helpers contains helper functions for use in rules
type Helpers struct{}

// NewHelpers creates a new Helpers instance
func NewHelpers() *Helpers {
	return &Helpers{}
}

// ContainsString checks if a string slice contains a specific string
func (h *Helpers) ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
