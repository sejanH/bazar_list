package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/sejan/bazarlist/internal/api"
	"github.com/sejan/bazarlist/internal/service"
)

func main() {
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Get data directory from environment or use default
	dataDir := os.Getenv("BAZARLIST_DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}

	// Initialize service
	svc, err := service.NewShoppingService(dataDir)
	if err != nil {
		log.Fatalf("Failed to initialize service: %v", err)
	}

	// Create HTTP handler
	handler := api.NewHTTPHandler(svc)

	// Create router
	router := mux.NewRouter()

	// Add middleware
	router.Use(handler.LoggingMiddleware)
	router.Use(handler.EnableCORS)

	// Register routes
	handler.RegisterRoutes(router)

	// Create server
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server
	fmt.Printf("🛒 Bazar List Web Server\n")
	fmt.Printf("📦 Data directory: %s\n", dataDir)
	fmt.Printf("🌐 Server running at: http://localhost:%s\n", port)
	fmt.Printf("📄 API docs: http://localhost:%s/api\n", port)
	fmt.Printf("\nPress Ctrl+C to stop the server\n\n")

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
