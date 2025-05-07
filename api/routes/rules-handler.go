// File: api/routes/rules_handler.go
package routes

import (
	"net/http"
	"encoding/json"

	"github.com/bass4/dcs-ice/internal/rules"
)

// RulesHandler handles rule management operations
type RulesHandler struct {
	ruleEngine *rules.RuleEngine
}

// NewRulesHandler creates a new rules handler
func NewRulesHandler(ruleEngine *rules.RuleEngine) *RulesHandler {
	return &RulesHandler{
		ruleEngine: ruleEngine,
	}
}

// HandleReloadRules handles requests to reload rules
func (h *RulesHandler) HandleReloadRules(w http.ResponseWriter, r *http.Request) {
	// Reload the rules
	err := h.ruleEngine.ReloadRules()
	
	// Create response
	response := map[string]interface{}{
		"success": err == nil,
	}
	
	if err != nil {
		response["error"] = err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	
	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
