package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sejan/bazarlist/internal/models"
	"github.com/sejan/bazarlist/internal/service"
)

// HTTPHandler handles HTTP requests
type HTTPHandler struct {
	service *service.ShoppingService
}

// NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(svc *service.ShoppingService) *HTTPHandler {
	return &HTTPHandler{
		service: svc,
	}
}

// RegisterRoutes registers all API routes
func (h *HTTPHandler) RegisterRoutes(router *mux.Router) {
	// API routes
	api := router.PathPrefix("/api").Subrouter()

	// Items CRUD
	api.HandleFunc("/items", h.GetItems).Methods("GET")
	api.HandleFunc("/items", h.AddItem).Methods("POST")
	api.HandleFunc("/items/{id}", h.GetItem).Methods("GET")
	api.HandleFunc("/items/{id}", h.UpdateItem).Methods("PUT", "PATCH")
	api.HandleFunc("/items/{id}", h.DeleteItem).Methods("DELETE")
	api.HandleFunc("/items/{id}/complete", h.CompleteItem).Methods("POST")
	api.HandleFunc("/items/{id}/pending", h.MakeItemPending).Methods("POST")

	// Search and filter
	api.HandleFunc("/search", h.SearchItems).Methods("GET")
	api.HandleFunc("/items/category/{category}", h.GetItemsByCategory).Methods("GET")

	// Stats
	api.HandleFunc("/stats", h.GetStats).Methods("GET")

	// Serve static files and templates
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/static/")))
}

// GetItems returns all items
func (h *HTTPHandler) GetItems(w http.ResponseWriter, r *http.Request) {
	items := h.service.GetAllItems()
	respondJSON(w, http.StatusOK, items)
}

// GetItem returns a single item by ID
func (h *HTTPHandler) GetItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid item ID")
		return
	}

	items := h.service.GetAllItems()
	for _, item := range items {
		if item.ID == id {
			respondJSON(w, http.StatusOK, item)
			return
		}
	}

	respondError(w, http.StatusNotFound, "Item not found")
}

// AddItemRequest represents the request to add an item
type AddItemRequest struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}

// AddItem adds a new item
func (h *HTTPHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	var req AddItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		respondError(w, http.StatusBadRequest, "Item name is required")
		return
	}

	category := models.Category(req.Category)
	if category == "" {
		category = models.CategoryOther
	}

	item, err := h.service.AddItem(req.Name, category)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, item)
}

// UpdateItemRequest represents the request to update an item
type UpdateItemRequest struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}

// UpdateItem updates an existing item
func (h *HTTPHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid item ID")
		return
	}

	var req UpdateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get the item
	items := h.service.GetAllItems()
	var targetItem *models.Item
	for _, item := range items {
		if item.ID == id {
			targetItem = item
			break
		}
	}

	if targetItem == nil {
		respondError(w, http.StatusNotFound, "Item not found")
		return
	}

	// Update fields if provided
	if req.Name != "" {
		targetItem.Name = req.Name
	}
	if req.Category != "" {
		targetItem.Category = models.Category(req.Category)
	}

	respondJSON(w, http.StatusOK, targetItem)
}

// DeleteItem deletes an item
func (h *HTTPHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid item ID")
		return
	}

	if err := h.service.RemoveItem(id); err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Item deleted successfully"})
}

// CompleteItem marks an item as completed
func (h *HTTPHandler) CompleteItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid item ID")
		return
	}

	if err := h.service.CompleteItem(id); err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	items := h.service.GetAllItems()
	for _, item := range items {
		if item.ID == id {
			respondJSON(w, http.StatusOK, item)
			return
		}
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Item completed successfully"})
}

// MakeItemPending marks an item as pending
func (h *HTTPHandler) MakeItemPending(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid item ID")
		return
	}

	// Get the item
	items := h.service.GetAllItems()
	var targetItem *models.Item
	for _, item := range items {
		if item.ID == id {
			targetItem = item
			break
		}
	}

	if targetItem == nil {
		respondError(w, http.StatusNotFound, "Item not found")
		return
	}

	targetItem.MarkPending()
	respondJSON(w, http.StatusOK, targetItem)
}

// SearchItems searches for items
func (h *HTTPHandler) SearchItems(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		respondError(w, http.StatusBadRequest, "Search query is required")
		return
	}

	items := h.service.SearchItems(query)
	respondJSON(w, http.StatusOK, items)
}

// GetItemsByCategory returns items filtered by category
func (h *HTTPHandler) GetItemsByCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	category := models.Category(vars["category"])

	items := h.service.GetItemsByCategory(category)
	respondJSON(w, http.StatusOK, items)
}

// Stats represents statistics about the shopping list
type Stats struct {
	Total     int `json:"total"`
	Pending   int `json:"pending"`
	Completed int `json:"completed"`
}

// GetStats returns shopping list statistics
func (h *HTTPHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	allItems := h.service.GetAllItems()
	pendingItems := h.service.GetPendingItems()
	completedItems := h.service.GetCompletedItems()

	stats := Stats{
		Total:     len(allItems),
		Pending:   len(pendingItems),
		Completed: len(completedItems),
	}

	respondJSON(w, http.StatusOK, stats)
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

// respondError sends an error response
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// EnableCORS enables CORS for the handler
func (h *HTTPHandler) EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware logs HTTP requests
func (h *HTTPHandler) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[%s] %s %s\n", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
