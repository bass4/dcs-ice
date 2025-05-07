package routes

import (
	"encoding/json"
	"net/http"

	"github.com/yourusername/dcs-ice/internal/rules"
	"github.com/yourusername/dcs-ice/pkg/models"
)

// FactsHandler handles fact evaluation requests
type FactsHandler struct {
	ruleEngine *rules.RuleEngine
}

// NewFactsHandler creates a new facts handler
func NewFactsHandler(ruleEngine *rules.RuleEngine) *FactsHandler {
	return &FactsHandler{
		ruleEngine: ruleEngine,
	}
}

// HandleFacts processes POST requests to evaluate facts
func (h *FactsHandler) HandleFacts(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var request models.FactsRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Evaluate facts against rules
	response, err := h.ruleEngine.EvaluateFacts(request.Facts)
	if err != nil {
		http.Error(w, "Error evaluating facts: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
