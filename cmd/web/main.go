package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sejan/bazarlist/internal/api"
	"github.com/sejan/bazarlist/internal/storage"
)

func main() {
	// Load environment variables from .env file
	loadEnv(".env")

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize MySQL storage
	store, err := storage.NewMySQLStorageFromEnv()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer store.Close()

	log.Println("✅ Database connected successfully")

	// Create handlers
	authHandler := api.NewAuthHandler(store)
	listHandler := api.NewListHandler(store)

	// Create router
	router := mux.NewRouter()

	// CORS middleware
	router.Use(api.CORSMiddleware)

	// Authentication routes (no auth required)
	router.HandleFunc("/api/auth/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST")

	// Shopping list routes (auth required)
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.Use(authHandler.AuthMiddleware)

	// Lists CRUD
	apiRouter.HandleFunc("/lists", listHandler.GetLists).Methods("GET")
	apiRouter.HandleFunc("/lists", listHandler.CreateList).Methods("POST")
	apiRouter.HandleFunc("/lists/{id}", listHandler.GetList).Methods("GET")
	apiRouter.HandleFunc("/lists/{id}", listHandler.UpdateList).Methods("PUT", "PATCH")
	apiRouter.HandleFunc("/lists/{id}", listHandler.DeleteList).Methods("DELETE")

	// Items CRUD
	apiRouter.HandleFunc("/lists/{id}/items", listHandler.GetItems).Methods("GET")
	apiRouter.HandleFunc("/lists/{id}/items", listHandler.CreateItem).Methods("POST")
	apiRouter.HandleFunc("/lists/{id}/items/{itemId}", listHandler.UpdateItem).Methods("PUT", "PATCH")
	apiRouter.HandleFunc("/lists/{id}/items/{itemId}", listHandler.DeleteItem).Methods("DELETE")

	// Serve static files
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/static/")))

	// Create server
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	log.Printf("🛒 Bazar List Web Server (MySQL + Auth)")
	log.Printf("📦 Database: MySQL")
	log.Printf("🌐 Server running at: http://localhost:%s", port)
	log.Printf("📄 API: http://localhost:%s/api", port)
	log.Printf("\nPress Ctrl+C to stop the server\n")

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// loadEnv loads environment variables from a file
func loadEnv(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		os.Setenv(key, value)
	}
}
