package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sejan/bazarlist/internal/live"
	"github.com/sejan/bazarlist/internal/models"
	"github.com/sejan/bazarlist/internal/storage"
)

// SharingHandler handles list membership, invitations, activities, and live SSE streaming
type SharingHandler struct {
	storage *storage.MySQLStorage
	hub     *live.Hub
}

// NewSharingHandler creates a new SharingHandler instance
func NewSharingHandler(store *storage.MySQLStorage, hub *live.Hub) *SharingHandler {
	return &SharingHandler{
		storage: store,
		hub:     hub,
	}
}

func (h *SharingHandler) getUserDisplayName(userID uint) string {
	user, err := h.storage.GetUserByID(userID)
	if err != nil || user == nil {
		return "Someone"
	}
	if user.Name != "" {
		return user.Name
	}
	return user.Email
}

// StreamLiveUpdates handles the SSE stream endpoint for real-time list updates
func (h *SharingHandler) StreamLiveUpdates(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

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
		respondError(w, http.StatusForbidden, "Access denied to list")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	var events <-chan live.Event
	var unsubscribe func()
	if h.hub != nil {
		events, unsubscribe = h.hub.Subscribe(listID)
		defer unsubscribe()
	}

	// Send initial connected event
	fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"connected\",\"list_id\":%d}\n\n", listID)
	flusher.Flush()

	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		case event, ok := <-events:
			if !ok {
				return
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, string(event.Data))
			flusher.Flush()
		}
	}
}

// GetMembers returns the list owner and all members
func (h *SharingHandler) GetMembers(w http.ResponseWriter, r *http.Request) {
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

	members, err := h.storage.GetListMembers(listID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to load members")
		return
	}
	if members == nil {
		members = []models.ListMember{}
	}

	list, err := h.storage.GetListByID(listID)
	if err != nil {
		respondError(w, http.StatusNotFound, "List not found")
		return
	}

	owner, _ := h.storage.GetUserByID(list.UserID)

	respondJSON(w, http.StatusOK, map[string]any{
		"owner":   owner,
		"members": members,
	})
}

// InviteMember invites a new member to the list by email
func (h *SharingHandler) InviteMember(w http.ResponseWriter, r *http.Request) {
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
		respondError(w, http.StatusForbidden, "Only list owners can invite members")
		return
	}

	var req models.InviteMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		respondError(w, http.StatusBadRequest, "Valid email required")
		return
	}

	targetUser, err := h.storage.GetUserByEmail(req.Email)
	if err != nil {
		respondError(w, http.StatusNotFound, "User not found. They must register for Bazar List first.")
		return
	}

	if targetUser.ID == userID {
		respondError(w, http.StatusBadRequest, "You are already the owner of this list")
		return
	}

	role := req.Role
	if role != models.RoleViewer && role != models.RoleEditor {
		role = models.RoleEditor
	}

	if err := h.storage.AddListMember(listID, targetUser.ID, role); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to invite member")
		return
	}

	inviterName := h.getUserDisplayName(userID)
	targetDisplayName := targetUser.Name
	if targetDisplayName == "" {
		targetDisplayName = targetUser.Email
	}

	activity := &models.ListActivity{
		ListID:    listID,
		UserID:    userID,
		UserName:  inviterName,
		Action:    models.ActionMemberJoined,
		Details:   fmt.Sprintf("Invited %s as %s", targetDisplayName, role),
		CreatedAt: time.Now(),
	}
	_ = h.storage.LogActivity(activity)

	if h.hub != nil {
		h.hub.Broadcast(listID, "member_updated", map[string]any{
			"user_id": targetUser.ID,
			"name":    targetUser.Name,
			"email":   targetUser.Email,
			"role":    role,
		})
		h.hub.Broadcast(listID, "activity_logged", activity)
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"message": "Member invited successfully",
		"member":  targetUser,
		"role":    role,
	})
}

// RemoveMember removes a member from the list
func (h *SharingHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
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

	targetUserID, err := parseID(vars["userId"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid member user ID")
		return
	}

	isOwner, err := h.storage.IsListOwner(userID, listID)
	if err != nil || !isOwner {
		respondError(w, http.StatusForbidden, "Only list owners can remove members")
		return
	}

	targetUser, _ := h.storage.GetUserByID(targetUserID)
	targetDisplayName := fmt.Sprintf("User #%d", targetUserID)
	if targetUser != nil {
		if targetUser.Name != "" {
			targetDisplayName = targetUser.Name
		} else {
			targetDisplayName = targetUser.Email
		}
	}

	if err := h.storage.RemoveListMember(listID, targetUserID); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to remove member")
		return
	}

	removerName := h.getUserDisplayName(userID)
	activity := &models.ListActivity{
		ListID:    listID,
		UserID:    userID,
		UserName:  removerName,
		Action:    models.ActionMemberRemoved,
		Details:   fmt.Sprintf("Removed %s from list", targetDisplayName),
		CreatedAt: time.Now(),
	}
	_ = h.storage.LogActivity(activity)

	if h.hub != nil {
		h.hub.Broadcast(listID, "member_updated", map[string]any{
			"user_id": targetUserID,
			"removed": true,
		})
		h.hub.Broadcast(listID, "activity_logged", activity)
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Member removed successfully"})
}

// GetActivities returns recent activity history for a list
func (h *SharingHandler) GetActivities(w http.ResponseWriter, r *http.Request) {
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

	activities, err := h.storage.GetListActivities(listID, 25)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to load activities")
		return
	}
	if activities == nil {
		activities = []models.ListActivity{}
	}

	respondJSON(w, http.StatusOK, activities)
}
