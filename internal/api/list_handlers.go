package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sejan/bazarlist/internal/models"
	"github.com/sejan/bazarlist/internal/storage"
)

// ListHandler handles shopping list operations
type ListHandler struct {
	storage *storage.MySQLStorage
}

// NewListHandler creates a new list handler
func NewListHandler(store *storage.MySQLStorage) *ListHandler {
	return &ListHandler{
		storage: store,
	}
}

// GetLists returns all lists for the authenticated user
func (h *ListHandler) GetLists(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	lists, err := h.storage.GetListsByUserID(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch lists")
		return
	}

	respondJSON(w, http.StatusOK, lists)
}

// GetList returns a specific list with items
func (h *ListHandler) GetList(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	listID, err := parseID(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid list ID")
		return
	}

	list, err := h.storage.GetListByID(listID)
	if err != nil {
		respondError(w, http.StatusNotFound, "List not found")
		return
	}

	// Verify user owns the list
	if list.UserID != userID {
		respondError(w, http.StatusForbidden, "Access denied")
		return
	}

	respondJSON(w, http.StatusOK, list)
}

// CreateList creates a new shopping list
func (h *ListHandler) CreateList(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.ListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Set date to today if not provided
	listDate := time.Now()
	if req.Date != "" {
		parsedDate, err := time.Parse("2006-01-02", req.Date)
		if err == nil {
			listDate = parsedDate
		}
	}

	list := &models.ShoppingList{
		UserID: userID,
		Name:   req.Name,
		Date:   listDate,
	}

	if err := h.storage.CreateList(list); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create list")
		return
	}

	// Reload to get items
	createdList, _ := h.storage.GetListByID(list.ID)
	respondJSON(w, http.StatusCreated, createdList)
}

// UpdateList updates a shopping list
func (h *ListHandler) UpdateList(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	listID, err := parseID(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid list ID")
		return
	}

	var req models.ListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	list, err := h.storage.GetListByID(listID)
	if err != nil {
		respondError(w, http.StatusNotFound, "List not found")
		return
	}

	// Verify user owns the list
	if list.UserID != userID {
		respondError(w, http.StatusForbidden, "Access denied")
		return
	}

	// Update fields
	if req.Name != "" {
		list.Name = req.Name
	}
	if req.Date != "" {
		parsedDate, err := time.Parse("2006-01-02", req.Date)
		if err == nil {
			list.Date = parsedDate
		}
	}

	if err := h.storage.UpdateList(list); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update list")
		return
	}

	respondJSON(w, http.StatusOK, list)
}

// DeleteList deletes a shopping list
func (h *ListHandler) DeleteList(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	listID, err := parseID(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid list ID")
		return
	}

	list, err := h.storage.GetListByID(listID)
	if err != nil {
		respondError(w, http.StatusNotFound, "List not found")
		return
	}

	// Verify user owns the list
	if list.UserID != userID {
		respondError(w, http.StatusForbidden, "Access denied")
		return
	}

	if err := h.storage.DeleteList(listID); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete list")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "List deleted successfully"})
}

// GetItems returns items for a specific list
func (h *ListHandler) GetItems(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	listID, err := parseID(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid list ID")
		return
	}

	// Verify user owns the list
	list, err := h.storage.GetListByID(listID)
	if err != nil {
		respondError(w, http.StatusNotFound, "List not found")
		return
	}

	if list.UserID != userID {
		respondError(w, http.StatusForbidden, "Access denied")
		return
	}

	items, err := h.storage.GetItemsByListID(listID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch items")
		return
	}

	respondJSON(w, http.StatusOK, items)
}

// CreateItem creates a new item in a list
func (h *ListHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	listID, err := parseID(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid list ID")
		return
	}

	// Verify user owns the list
	list, err := h.storage.GetListByID(listID)
	if err != nil {
		respondError(w, http.StatusNotFound, "List not found")
		return
	}

	if list.UserID != userID {
		respondError(w, http.StatusForbidden, "Access denied")
		return
	}

	var req models.ItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	item := &models.Item{
		ListID:    listID,
		Name:      req.Name,
		Price:     req.Price,
		Purchased: req.Purchased,
	}

	if err := h.storage.CreateItem(item); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create item")
		return
	}

	respondJSON(w, http.StatusCreated, item)
}

// UpdateItem updates an item
func (h *ListHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	itemID, err := parseID(vars["itemId"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid item ID")
		return
	}

	item, err := h.storage.GetItemByID(itemID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Item not found")
		return
	}

	// Verify user owns the list
	list, err := h.storage.GetListByID(item.ListID)
	if err != nil || list.UserID != userID {
		respondError(w, http.StatusForbidden, "Access denied")
		return
	}

	var req models.ItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Update fields
	if req.Name != "" {
		item.Name = req.Name
	}
	item.Price = req.Price
	item.Purchased = req.Purchased

	if err := h.storage.UpdateItem(item); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update item")
		return
	}

	respondJSON(w, http.StatusOK, item)
}

// DeleteItem deletes an item
func (h *ListHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	itemID, err := parseID(vars["itemId"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid item ID")
		return
	}

	item, err := h.storage.GetItemByID(itemID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Item not found")
		return
	}

	// Verify user owns the list
	list, err := h.storage.GetListByID(item.ListID)
	if err != nil || list.UserID != userID {
		respondError(w, http.StatusForbidden, "Access denied")
		return
	}

	if err := h.storage.DeleteItem(itemID); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete item")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Item deleted successfully"})
}
