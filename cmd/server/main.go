// File: cmd/server/main.go
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/bass4/dcs-ice/api/routes"
	"github.com/bass4/dcs-ice/internal/actions"
	"github.com/bass4/dcs-ice/internal/rules"
	"github.com/bass4/dcs-ice/internal/websocket"
)

func main() {
	// Parse command line flags
	port := flag.String("port", "8080", "HTTP server port")
	ruleDir := flag.String("rules", "./config/rules", "Directory containing rule files")
	flag.Parse()

	// Initialize the action manager
	actionManager := actions.NewManager()

	// Initialize the rule engine with the action manager
	ruleEngine, err := rules.NewRuleEngine(*ruleDir, actionManager)
	if err != nil {
		log.Fatalf("Failed to initialize rule engine: %v", err)
	}

	// Initialize the WebSocket server
	wsServer := websocket.NewServer(actionManager)

	// Set up HTTP router
	router := setupRoutes(ruleEngine, wsServer)
	
	// Start HTTP server
	server := &http.Server{
		Addr:         ":" + *port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		fmt.Printf("DCS-ICE server listening on port %s\n", *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down server...")
	
	// Clean shutdown of the WebSocket server
	wsServer.Close()
}

// setupRoutes configures all HTTP routes
func setupRoutes(ruleEngine *rules.RuleEngine, wsServer *websocket.Server) *mux.Router {
	router := mux.NewRouter()
	
	// API routes
	apiRouter := router.PathPrefix("/api").Subrouter()
	
	// Facts endpoint for DCS World
	factsHandler := routes.NewFactsHandler(ruleEngine)
	apiRouter.HandleFunc("/facts", factsHandler.HandleFacts).Methods("POST")
	
	// WebSocket endpoint
	router.HandleFunc("/ws", wsServer.HandleWebSocket)
	
	// Rule management endpoints
	rulesHandler := routes.NewRulesHandler(ruleEngine)
	apiRouter.HandleFunc("/rules/reload", rulesHandler.HandleReloadRules).Methods("POST")
	
	// Static file server for web UI (if needed)
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/static")))
	
	return router
}
