package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// respondJSON writes a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError writes an error response
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// parseID parses an ID from a string
func parseID(idStr string) (uint, error) {
	var id uint
	_, err := fmt.Sscanf(idStr, "%d", &id)
	return id, err
}
