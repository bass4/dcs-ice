package routes

import (

	"github.com/gorilla/mux"
	"github.com/bass4/dcs-ice/internal/rules"
)

// SetupRoutes configures the HTTP routes
func SetupRoutes(ruleEngine *rules.RuleEngine) *mux.Router {
	router := mux.NewRouter()
	
	// Register handlers
	factsHandler := NewFactsHandler(ruleEngine)
	router.HandleFunc("/facts", factsHandler.HandleFacts).Methods("POST")
	
	return router
}
