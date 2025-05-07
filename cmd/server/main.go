package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourusername/dcs-ice/api/routes"
	"github.com/yourusername/dcs-ice/internal/rules"
)

func main() {
	// Parse command line flags
	port := flag.String("port", "8080", "HTTP server port")
	ruleDir := flag.String("rules", "./config/rules", "Directory containing rule files")
	flag.Parse()

	// Initialize the rule engine
	ruleEngine, err := rules.NewRuleEngine(*ruleDir)
	if err != nil {
		log.Fatalf("Failed to initialize rule engine: %v", err)
	}

	// Set up HTTP router
	router := routes.SetupRoutes(ruleEngine)
	
	// Start HTTP server
	server := &http.Server{
		Addr:    ":" + *port,
		Handler: router,
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
}
