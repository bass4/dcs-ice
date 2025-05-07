// File: api/routes/facts_handler.go
package routes

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/bass4/dcs-ice/internal/rules"
	"github.com/bass4/dcs-ice/pkg/models"
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
// This is optimized for performance to minimize delay for DCS World
func (h *FactsHandler) HandleFacts(w http.ResponseWriter, r *http.Request) {
	// Quick check for content type
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	// Limit request body size to prevent abuse
	r.Body = http.MaxBytesReader(w, r.Body, 1048576) // 1MB limit
	
	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	
	var request models.FactsRequest
	if err := json.Unmarshal(body, &request); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}
	
	// Validate request
	if len(request.Facts) == 0 {
		http.Error(w, "No facts provided", http.StatusBadRequest)
		return
	}
	
	// Start timing
	startTime := time.Now()
	
	// Evaluate facts against rules
	response, err := h.ruleEngine.EvaluateFacts(request.Facts)
	if err != nil {
		http.Error(w, "Error evaluating facts: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Prepare response
	respBytes, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error serializing response", http.StatusInternalServerError)
		return
	}
	
	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(respBytes)
	
	// Log processing time after response is sent
	// This won't block the client
	go func(elapsedTime time.Duration) {
		// Log the timing information asynchronously
		// This could be replaced with proper structured logging
		// fmt.Printf("Facts evaluation completed in %v ms\n", elapsedTime.Milliseconds())
	}(time.Since(startTime))
}
