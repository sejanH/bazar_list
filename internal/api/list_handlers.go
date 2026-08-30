package api

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/sejan/bazarlist/internal/live"
	"github.com/sejan/bazarlist/internal/models"
	"github.com/sejan/bazarlist/internal/storage"
)

// ListHandler handles shopping list operations
type ListHandler struct {
	storage *storage.MySQLStorage
	hub     *live.Hub
}

// NewListHandler creates a new list handler
func NewListHandler(store *storage.MySQLStorage, hub *live.Hub) *ListHandler {
	return &ListHandler{
		storage: store,
		hub:     hub,
	}
}

func (h *ListHandler) getUserDisplayName(userID uint) string {
	user, err := h.storage.GetUserByID(userID)
	if err != nil || user == nil {
		return "Someone"
	}
	if user.Name != "" {
		return user.Name
	}
	return user.Email
}

// GetLists returns paginated lists for a specific month
func (h *ListHandler) GetLists(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	if page <= 0 {
		page = 1
	}
	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	var year, month int
	monthStr := query.Get("month")

	if monthStr == "" {
		// Default to latest month
		var err error
		year, month, err = h.storage.GetLatestMonth(userID)
		if err != nil {
			// No lists found, return empty response
			months, _ := h.storage.GetAvailableMonths(userID)
			if months == nil {
				months = []string{}
			}
			respondJSON(w, http.StatusOK, models.PaginatedListsResponse{
				AvailableMonths: months,
				Lists:           []models.ShoppingList{},
				Month:           "",
				Pagination: models.PaginationInfo{
					CurrentPage: 1,
					TotalPages:  1,
					TotalItems:  0,
					Limit:       limit,
				},
			})
			return
		}
		monthStr = fmt.Sprintf("%04d-%02d", year, month)
	} else {
		// Parse requested month (YYYY-MM)
		if _, err := fmt.Sscanf(monthStr, "%d-%d", &year, &month); err != nil {
			respondError(w, http.StatusBadRequest, "Invalid month format. Use YYYY-MM")
			return
		}
	}

	// Fetch data
	lists, total, err := h.storage.GetPaginatedListsByMonth(userID, year, month, page, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch lists")
		return
	}
	if lists == nil {
		lists = []models.ShoppingList{}
	}

	// Get all available months for grouping menu
	availableMonths, _ := h.storage.GetAvailableMonths(userID)
	if availableMonths == nil {
		availableMonths = []string{}
	}

	// Get total for the month across all lists
	monthTotal, _ := h.storage.GetMonthTotal(userID, year, month)

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	response := models.PaginatedListsResponse{
		Month: monthStr,
		Lists: lists,
		Pagination: models.PaginationInfo{
			CurrentPage: page,
			TotalPages:  totalPages,
			TotalItems:  total,
			Limit:       limit,
		},
		AvailableMonths: availableMonths,
		TotalAmount:     monthTotal,
	}

	respondJSON(w, http.StatusOK, response)
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

	canView, err := h.storage.CanViewList(userID, listID)
	if err != nil || !canView {
		respondError(w, http.StatusForbidden, "Access denied")
		return
	}

	list, err := h.storage.GetListByID(listID)
	if err != nil {
		respondError(w, http.StatusNotFound, "List not found")
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

	canEdit, err := h.storage.CanEditList(userID, listID)
	if err != nil || !canEdit {
		respondError(w, http.StatusForbidden, "Access denied")
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

	isOwner, err := h.storage.IsListOwner(userID, listID)
	if err != nil || !isOwner {
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

	canView, err := h.storage.CanViewList(userID, listID)
	if err != nil || !canView {
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

	canEdit, err := h.storage.CanEditList(userID, listID)
	if err != nil || !canEdit {
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

	userName := h.getUserDisplayName(userID)
	activity := &models.ListActivity{
		ListID:    listID,
		UserID:    userID,
		UserName:  userName,
		Action:    models.ActionItemAdded,
		ItemName:  item.Name,
		Details:   fmt.Sprintf("Added %s (৳%.2f)", item.Name, item.Price),
		CreatedAt: time.Now(),
	}
	_ = h.storage.LogActivity(activity)

	if h.hub != nil {
		h.hub.Broadcast(listID, "item_created", item)
		h.hub.Broadcast(listID, "activity_logged", activity)
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

	canEdit, err := h.storage.CanEditList(userID, item.ListID)
	if err != nil || !canEdit {
		respondError(w, http.StatusForbidden, "Access denied")
		return
	}

	var req models.ItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	wasPurchased := item.Purchased
	oldName := item.Name

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

	userName := h.getUserDisplayName(userID)
	var action, details string
	if !wasPurchased && item.Purchased {
		action = models.ActionItemPurchased
		details = fmt.Sprintf("Marked %s as purchased", item.Name)
	} else if wasPurchased && !item.Purchased {
		action = models.ActionItemUpdated
		details = fmt.Sprintf("Marked %s as unpurchased", item.Name)
	} else {
		action = models.ActionItemUpdated
		if oldName != item.Name {
			details = fmt.Sprintf("Renamed %s to %s", oldName, item.Name)
		} else {
			details = fmt.Sprintf("Updated %s (৳%.2f)", item.Name, item.Price)
		}
	}

	activity := &models.ListActivity{
		ListID:    item.ListID,
		UserID:    userID,
		UserName:  userName,
		Action:    action,
		ItemName:  item.Name,
		Details:   details,
		CreatedAt: time.Now(),
	}
	_ = h.storage.LogActivity(activity)

	if h.hub != nil {
		h.hub.Broadcast(item.ListID, "item_updated", item)
		h.hub.Broadcast(item.ListID, "activity_logged", activity)
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

	canEdit, err := h.storage.CanEditList(userID, item.ListID)
	if err != nil || !canEdit {
		respondError(w, http.StatusForbidden, "Access denied")
		return
	}

	if err := h.storage.DeleteItem(itemID); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete item")
		return
	}

	userName := h.getUserDisplayName(userID)
	activity := &models.ListActivity{
		ListID:    item.ListID,
		UserID:    userID,
		UserName:  userName,
		Action:    models.ActionItemDeleted,
		ItemName:  item.Name,
		Details:   fmt.Sprintf("Deleted %s", item.Name),
		CreatedAt: time.Now(),
	}
	_ = h.storage.LogActivity(activity)

	if h.hub != nil {
		h.hub.Broadcast(item.ListID, "item_deleted", map[string]any{
			"item_id": itemID,
			"list_id": item.ListID,
		})
		h.hub.Broadcast(item.ListID, "activity_logged", activity)
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Item deleted successfully"})
}
