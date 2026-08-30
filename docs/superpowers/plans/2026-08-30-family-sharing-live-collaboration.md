# Family Sharing & Live Collaboration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enable multi-user household collaboration on shopping lists with email invitations, real-time live sync (SSE broadcast), and an audit activity timeline.

**Architecture:** Extend the relational MySQL model with `list_members` (role-based access) and `list_activities` (event audit). Implement an in-memory thread-safe `LiveHub` pub/sub broker broadcasting Server-Sent Events (SSE) on list mutations. Build client-side SSE auto-reconnect listeners in `web/static/js/` that dynamically update items, spend totals, member counts, and activity logs without page reloads.

**Tech Stack:** Go 1.21+, Gorilla Mux, GORM (MySQL), Server-Sent Events (SSE HTTP/1.1 streaming), HTML5/ES6 PWA, IndexedDB offline cache.

**Spec:** Family Sharing & Live Collaboration Feature Requirements:
1. Household / Shared Lists: Invite members by email with roles (`editor`, `viewer`), list members, and revoke access.
2. Live Sync (SSE): Instant check-off, item additions, edits, and price updates across connected devices.
3. Activity Log: Who added, edited, or marked an item as purchased, displayed in a collapsible list activity drawer.

## Global Constraints
- Must remain compatible with existing PWA Offline Mode and Outbox synchronization.
- All REST API endpoints must require JWT authentication via `Bearer <token>` (or `?token=<token>` for EventSource streams).
- Real-time updates must gracefully fall back to HTTP polling/offline outbox when SSE connection drops.
- Follow existing codebase conventions in `internal/api/`, `internal/storage/`, and `web/static/`.

---

### File Structure Map

- **New Backend Files:**
  - `internal/live/hub.go`: Thread-safe in-memory pub/sub broker for SSE clients per shopping list.
  - `internal/live/hub_test.go`: Unit tests for subscription, broadcast, and cleanup.
  - `internal/api/sharing_handlers.go`: HTTP endpoints for list members, invitations, activity logs, and SSE streaming.
  - `internal/api/sharing_handlers_test.go`: Unit tests for sharing and activity log endpoints.
- **Modified Backend Files:**
  - `internal/models/database.go`: Add `ListMember` and `ListActivity` structs, invite request models, and update `ShoppingList`.
  - `internal/storage/mysql.go`: Add member CRUD, permission checks (`CanViewList`, `CanEditList`, `IsOwner`), and activity log persistence.
  - `internal/api/list_handlers.go`: Add permission guards and emit activity events + SSE broadcasts on item/list mutations.
  - `cmd/web/main.go`: Register `/api/lists/{id}/live`, `/api/lists/{id}/members`, and `/api/lists/{id}/activities` routes.
- **New Frontend Files:**
  - `web/static/js/live.js`: Client-side SSE stream connection manager with exponential backoff and DOM mutation triggers.
- **Modified Frontend Files:**
  - `web/static/index.html`: Add Family Share modal markup, Activity Drawer UI, and real-time DOM dispatchers.
  - `web/static/sw.js`: Precache `live.js` and bump cache version.

---

### Task 1: Database Models & Schema Migrations

**Files:**
- Modify: `internal/models/database.go`
- Test: `internal/models/database_test.go`

**Interfaces:**
- Produces: `ListMember`, `ListActivity`, `InviteMemberRequest`, `UpdateMemberRoleRequest`, `ListActivityResponse`

- [ ] **Step 1: Write test for ListMember and ListActivity model validation**

Create `internal/models/database_test.go`:
```go
package models_test

import (
	"testing"
	"time"

	"github.com/sejan/bazarlist/internal/models"
)

func TestListMemberAndActivityModels(t *testing.T) {
	now := time.Now()
	member := models.ListMember{
		ListID:    1,
		UserID:    2,
		Role:      models.RoleEditor,
		CreatedAt: now,
	}

	if member.Role != "editor" {
		t.Fatalf("expected role editor, got %s", member.Role)
	}

	activity := models.ListActivity{
		ListID:    1,
		UserID:    2,
		UserName:  "Sarah",
		Action:    models.ActionItemPurchased,
		ItemName:  "Eggs",
		Details:   "Marked as purchased",
		CreatedAt: now,
	}

	if activity.Action != "item_purchased" {
		t.Fatalf("expected action item_purchased, got %s", activity.Action)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/models/...`
Expected: FAIL due to missing types and constants.

- [ ] **Step 3: Define models in `internal/models/database.go`**

Add to `internal/models/database.go`:
```go
// Member Roles
const (
	RoleOwner  = "owner"
	RoleEditor = "editor"
	RoleViewer = "viewer"
)

// Activity Actions
const (
	ActionItemAdded     = "item_added"
	ActionItemUpdated   = "item_updated"
	ActionItemPurchased = "item_purchased"
	ActionItemDeleted   = "item_deleted"
	ActionMemberJoined  = "member_joined"
	ActionMemberRemoved = "member_removed"
)

type ListMember struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ListID    uint      `json:"list_id" gorm:"not null;index;uniqueIndex:idx_list_user"`
	UserID    uint      `json:"user_id" gorm:"not null;index;uniqueIndex:idx_list_user"`
	Role      string    `json:"role" gorm:"type:varchar(20);not null;default:'editor'"`
	User      *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	List      *ShoppingList `json:"-" gorm:"foreignKey:ListID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListActivity struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ListID    uint      `json:"list_id" gorm:"not null;index"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	UserName  string    `json:"user_name" gorm:"type:varchar(100);not null"`
	Action    string    `json:"action" gorm:"type:varchar(50);not null"`
	ItemName  string    `json:"item_name" gorm:"type:varchar(255)"`
	Details   string    `json:"details" gorm:"type:varchar(255)"`
	CreatedAt time.Time `json:"created_at"`
}

type InviteMemberRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role"` // "editor" or "viewer" (defaults to "editor")
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required"`
}
```

- [ ] **Step 4: Update GORM AutoMigrate in `internal/storage/mysql.go`**

In `internal/storage/mysql.go`, update line 47:
```go
if err := db.AutoMigrate(&models.User{}, &models.ShoppingList{}, &models.Item{}, &models.ListMember{}, &models.ListActivity{}); err != nil {
	return nil, fmt.Errorf("failed to migrate database: %w", err)
}
```

- [ ] **Step 5: Run tests and verify compile**

Run: `go test ./internal/models/...`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/models/ internal/storage/mysql.go
git commit -m "feat(models): add ListMember and ListActivity schema definitions"
```

---

### Task 2: Storage Layer for Sharing, Permissions, & Activity Log

**Files:**
- Modify: `internal/storage/mysql.go`
- Test: `internal/storage/mysql_sharing_test.go`

**Interfaces:**
- Produces:
  - `CanViewList(userID, listID uint) (bool, error)`
  - `CanEditList(userID, listID uint) (bool, error)`
  - `IsListOwner(userID, listID uint) (bool, error)`
  - `AddListMember(listID, userID uint, role string) error`
  - `RemoveListMember(listID, userID uint) error`
  - `GetListMembers(listID uint) ([]models.ListMember, error)`
  - `LogActivity(activity *models.ListActivity) error`
  - `GetListActivities(listID uint, limit int) ([]models.ListActivity, error)`

- [ ] **Step 1: Write integration tests for storage sharing methods**

Create `internal/storage/mysql_sharing_test.go` with mock/test database verification.

- [ ] **Step 2: Implement query filters in `GetPaginatedListsByMonth`**

Update `GetPaginatedListsByMonth` in `internal/storage/mysql.go` so shared lists appear for members:
```go
// User can see lists they own OR lists where they are a member
query := s.db.Model(&models.ShoppingList{}).
	Where("user_id = ? OR id IN (SELECT list_id FROM list_members WHERE user_id = ?)", userID, userID).
	Where("YEAR(date) = ? AND MONTH(date) = ?", year, month)
```

- [ ] **Step 3: Implement sharing and activity log storage methods in `internal/storage/mysql.go`**

```go
func (s *MySQLStorage) CanViewList(userID, listID uint) (bool, error) {
	var count int64
	err := s.db.Model(&models.ShoppingList{}).
		Where("id = ? AND (user_id = ? OR id IN (SELECT list_id FROM list_members WHERE user_id = ?))", listID, userID, userID).
		Count(&count).Error
	return count > 0, err
}

func (s *MySQLStorage) CanEditList(userID, listID uint) (bool, error) {
	var count int64
	// Owner or editor member
	err := s.db.Model(&models.ShoppingList{}).
		Where("id = ? AND (user_id = ? OR id IN (SELECT list_id FROM list_members WHERE user_id = ? AND role IN ('owner', 'editor')))", listID, userID, userID).
		Count(&count).Error
	return count > 0, err
}

func (s *MySQLStorage) IsListOwner(userID, listID uint) (bool, error) {
	var count int64
	err := s.db.Model(&models.ShoppingList{}).
		Where("id = ? AND user_id = ?", listID, userID).
		Count(&count).Error
	return count > 0, err
}

func (s *MySQLStorage) AddListMember(listID, userID uint, role string) error {
	if role == "" {
		role = models.RoleEditor
	}
	member := models.ListMember{
		ListID: listID,
		UserID: userID,
		Role:   role,
	}
	return s.db.Where(models.ListMember{ListID: listID, UserID: userID}).
		Assign(models.ListMember{Role: role}).
		FirstOrCreate(&member).Error
}

func (s *MySQLStorage) RemoveListMember(listID, userID uint) error {
	return s.db.Where("list_id = ? AND user_id = ?", listID, userID).Delete(&models.ListMember{}).Error
}

func (s *MySQLStorage) GetListMembers(listID uint) ([]models.ListMember, error) {
	var members []models.ListMember
	err := s.db.Preload("User").Where("list_id = ?", listID).Find(&members).Error
	return members, err
}

func (s *MySQLStorage) LogActivity(activity *models.ListActivity) error {
	return s.db.Create(activity).Error
}

func (s *MySQLStorage) GetListActivities(listID uint, limit int) ([]models.ListActivity, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	var activities []models.ListActivity
	err := s.db.Where("list_id = ?", listID).Order("created_at desc").Limit(limit).Find(&activities).Error
	return activities, err
}
```

- [ ] **Step 4: Run tests and verify**

Run: `go test ./internal/storage/...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/storage/
git commit -m "feat(storage): implement list sharing, permissions, and activity log queries"
```

---

### Task 3: In-Memory LivePubSub Broadcast Hub (SSE Stream Manager)

**Files:**
- Create: `internal/live/hub.go`
- Create: `internal/live/hub_test.go`

**Interfaces:**
- Produces:
  - `NewHub() *Hub`
  - `(h *Hub) Subscribe(listID uint) (<-chan Event, func())`
  - `(h *Hub) Broadcast(listID uint, eventType string, payload any)`
  - `Event struct { Type string, Data []byte }`

- [ ] **Step 1: Write unit tests for `LiveHub`**

Create `internal/live/hub_test.go`:
```go
package live_test

import (
	"testing"
	"time"

	"github.com/sejan/bazarlist/internal/live"
)

func TestLiveHubBroadcast(t *testing.T) {
	hub := live.NewHub()
	ch, unsubscribe := hub.Subscribe(42)
	defer unsubscribe()

	go func() {
		time.Sleep(10 * time.Millisecond)
		hub.Broadcast(42, "item_purchased", map[string]any{"id": 101, "purchased": true})
	}()

	select {
	case event := <-ch:
		if event.Type != "item_purchased" {
			t.Fatalf("expected event item_purchased, got %s", event.Type)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for live hub broadcast")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/live/...`
Expected: FAIL (package does not exist).

- [ ] **Step 3: Implement `internal/live/hub.go`**

Create `internal/live/hub.go`:
```go
package live

import (
	"encoding/json"
	"sync"
)

type Event struct {
	Type string `json:"type"`
	Data []byte `json:"data"`
}

type Hub struct {
	mu          sync.RWMutex
	subscribers map[uint]map[chan Event]struct{}
}

func NewHub() *Hub {
	return &Hub{
		subscribers: make(map[uint]map[chan Event]struct{}),
	}
}

func (h *Hub) Subscribe(listID uint) (<-chan Event, func()) {
	h.mu.Lock()
	defer h.mu.Unlock()

	ch := make(chan Event, 32)
	if _, ok := h.subscribers[listID]; !ok {
		h.subscribers[listID] = make(map[chan Event]struct{})
	}
	h.subscribers[listID][ch] = struct{}{}

	unsubscribe := func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		if subs, ok := h.subscribers[listID]; ok {
			delete(subs, ch)
			close(ch)
			if len(subs) == 0 {
				delete(h.subscribers, listID)
			}
		}
	}

	return ch, unsubscribe
}

func (h *Hub) Broadcast(listID uint, eventType string, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	event := Event{
		Type: eventType,
		Data: data,
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	subs, ok := h.subscribers[listID]
	if !ok {
		return
	}

	for ch := range subs {
		select {
		case ch <- event:
		default:
			// Buffer full, drop non-critical event to prevent slow client head-of-line blocking
		}
	}
}
```

- [ ] **Step 4: Run unit tests**

Run: `go test -v ./internal/live/...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/live/
git commit -m "feat(live): implement in-memory pubsub live hub for SSE broadcasting"
```

---

### Task 4: API Handlers for Member Invitations, Activity Logs, and Real-Time SSE Stream

**Files:**
- Create: `internal/api/sharing_handlers.go`
- Modify: `internal/api/list_handlers.go`
- Modify: `cmd/web/main.go`
- Test: `internal/api/sharing_handlers_test.go`

**Interfaces:**
- Produces:
  - `GET /api/lists/{id}/live`: SSE stream endpoint with event flushing.
  - `GET /api/lists/{id}/members`: List all members and owner.
  - `POST /api/lists/{id}/members`: Invite member by email.
  - `DELETE /api/lists/{id}/members/{userId}`: Remove member from list.
  - `GET /api/lists/{id}/activities`: Fetch recent activity log entries.

- [ ] **Step 1: Write integration tests for sharing handlers**

Create `internal/api/sharing_handlers_test.go` testing invite, list members, and SSE connection.

- [ ] **Step 2: Implement `internal/api/sharing_handlers.go`**

```go
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/sejan/bazarlist/internal/live"
	"github.com/sejan/bazarlist/internal/models"
	"github.com/sejan/bazarlist/internal/storage"
)

type SharingHandler struct {
	storage *storage.MySQLStorage
	hub     *live.Hub
}

func NewSharingHandler(store *storage.MySQLStorage, hub *live.Hub) *SharingHandler {
	return &SharingHandler{
		storage: store,
		hub:     hub,
	}
}

// SSE Live Stream Endpoint
func (h *SharingHandler) StreamLiveUpdates(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	userID := GetUserID(r)
	vars := mux.Vars(r)
	listID, _ := strconv.Atoi(vars["id"])

	canView, err := h.storage.CanViewList(userID, uint(listID))
	if err != nil || !canView {
		respondError(w, http.StatusForbidden, "Access denied to list")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	events, unsubscribe := h.hub.Subscribe(uint(listID))
	defer unsubscribe()

	// Initial connected ping
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

// Get Members
func (h *SharingHandler) GetMembers(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	vars := mux.Vars(r)
	listID, _ := strconv.Atoi(vars["id"])

	canView, _ := h.storage.CanViewList(userID, uint(listID))
	if !canView {
		respondError(w, http.StatusForbidden, "Access denied")
		return
	}

	members, err := h.storage.GetListMembers(uint(listID))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to load members")
		return
	}

	list, _ := h.storage.GetListByID(uint(listID))
	owner, _ := h.storage.GetUserByID(list.UserID)

	respondJSON(w, http.StatusOK, map[string]any{
		"owner":   owner,
		"members": members,
	})
}

// Invite Member by Email
func (h *SharingHandler) InviteMember(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	vars := mux.Vars(r)
	listID, _ := strconv.Atoi(vars["id"])

	isOwner, _ := h.storage.IsListOwner(userID, uint(listID))
	if !isOwner {
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

	if err := h.storage.AddListMember(uint(listID), targetUser.ID, role); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to invite member")
		return
	}

	currentUser, _ := h.storage.GetUserByID(userID)
	h.storage.LogActivity(&models.ListActivity{
		ListID:    uint(listID),
		UserID:    userID,
		UserName:  currentUser.Name,
		Action:    models.ActionMemberJoined,
		Details:   fmt.Sprintf("Invited %s as %s", targetUser.Name, role),
		CreatedAt: time.Now(),
	})

	h.hub.Broadcast(uint(listID), "member_updated", map[string]any{
		"user_id": targetUser.ID,
		"name":    targetUser.Name,
		"role":    role,
	})

	respondJSON(w, http.StatusOK, map[string]any{
		"message": "Member invited successfully",
		"member":  targetUser,
		"role":    role,
	})
}

// Get Activities
func (h *SharingHandler) GetActivities(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	vars := mux.Vars(r)
	listID, _ := strconv.Atoi(vars["id"])

	canView, _ := h.storage.CanViewList(userID, uint(listID))
	if !canView {
		respondError(w, http.StatusForbidden, "Access denied")
		return
	}

	activities, err := h.storage.GetListActivities(uint(listID), 25)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to load activities")
		return
	}

	respondJSON(w, http.StatusOK, activities)
}
```

- [ ] **Step 3: Integrate Hub broadcasting and Activity Logging into `internal/api/list_handlers.go`**

Update `ListHandler` to accept `*live.Hub`. On `CreateItem`, `UpdateItem`, `DeleteItem`:
1. Log activity (`ActionItemAdded`, `ActionItemUpdated`, `ActionItemPurchased`, `ActionItemDeleted`).
2. Call `h.hub.Broadcast(listID, eventType, itemPayload)`.

- [ ] **Step 4: Wire routes in `cmd/web/main.go`**

```go
liveHub := live.NewHub()
sharingHandler := api.NewSharingHandler(store, liveHub)
listHandler := api.NewListHandler(store, liveHub)

// Sharing & Live routes
apiRouter.HandleFunc("/lists/{id}/live", sharingHandler.StreamLiveUpdates).Methods("GET")
apiRouter.HandleFunc("/lists/{id}/members", sharingHandler.GetMembers).Methods("GET")
apiRouter.HandleFunc("/lists/{id}/members", sharingHandler.InviteMember).Methods("POST")
apiRouter.HandleFunc("/lists/{id}/activities", sharingHandler.GetActivities).Methods("GET")
```

- [ ] **Step 5: Run tests & verify server startup**

Run: `make build-web`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/api/ cmd/web/main.go
git commit -m "feat(api): add sharing, activity log, and real-time SSE stream endpoints"
```

---

### Task 5: Live Sync Client in Frontend (`live.js`)

**Files:**
- Create: `web/static/js/live.js`
- Modify: `web/static/index.html`
- Modify: `web/static/sw.js`

**Interfaces:**
- Produces: `window.liveClient = new LiveClient()`
- Methods:
  - `connect(listId)`
  - `disconnect()`
  - Event triggers: `onItemUpdated(fn)`, `onItemCreated(fn)`, `onItemDeleted(fn)`, `onActivityLogged(fn)`

- [ ] **Step 1: Write `web/static/js/live.js`**

```javascript
// Live Client for Real-Time List Synchronization via SSE
class LiveClient {
  constructor() {
    this.eventSource = null;
    this.currentListId = null;
    this.reconnectTimer = null;
    this.listeners = new Map();
  }

  connect(listId) {
    if (this.currentListId === listId && this.eventSource) return;
    this.disconnect();
    this.currentListId = listId;

    const token = localStorage.getItem('token');
    if (!token) return;

    const streamUrl = `/api/lists/${listId}/live`;
    
    // Use EventSource with custom reconnect
    try {
      this.eventSource = new EventSource(streamUrl);

      this.eventSource.addEventListener('connected', (e) => {
        console.log(`[LiveSync] Connected to list #${listId}`);
      });

      this.eventSource.addEventListener('item_updated', (e) => {
        const item = JSON.parse(e.data);
        this.emit('item_updated', item);
      });

      this.eventSource.addEventListener('item_created', (e) => {
        const item = JSON.parse(e.data);
        this.emit('item_created', item);
      });

      this.eventSource.addEventListener('item_deleted', (e) => {
        const data = JSON.parse(e.data);
        this.emit('item_deleted', data);
      });

      this.eventSource.addEventListener('activity_logged', (e) => {
        const activity = JSON.parse(e.data);
        this.emit('activity_logged', activity);
      });

      this.eventSource.addEventListener('member_updated', (e) => {
        const data = JSON.parse(e.data);
        this.emit('member_updated', data);
      });

      this.eventSource.onerror = (err) => {
        console.warn('[LiveSync] Connection lost, retrying in 5s...', err);
        this.disconnect();
        clearTimeout(this.reconnectTimer);
        this.reconnectTimer = setTimeout(() => {
          if (this.currentListId === listId) {
            this.connect(listId);
          }
        }, 5000);
      };
    } catch (err) {
      console.error('[LiveSync] Could not establish EventSource:', err);
    }
  }

  disconnect() {
    if (this.eventSource) {
      this.eventSource.close();
      this.eventSource = null;
    }
    clearTimeout(this.reconnectTimer);
  }

  on(event, callback) {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, []);
    }
    this.listeners.get(event).push(callback);
  }

  emit(event, data) {
    const callbacks = this.listeners.get(event) || [];
    callbacks.forEach(cb => {
      try { cb(data); } catch (e) { console.error(e); }
    });
  }
}

window.liveClient = new LiveClient();
```

- [ ] **Step 2: Precache `live.js` in `web/static/sw.js`**

Add `'/js/live.js'` to `PRECACHE_ASSETS` and bump cache version.

- [ ] **Step 3: Include `live.js` in `web/static/index.html`**

Add `<script src="/js/live.js"></script>` in `<head>`.

- [ ] **Step 4: Commit**

```bash
git add web/static/js/live.js web/static/sw.js web/static/index.html
git commit -m "feat(live): implement client-side SSE live sync engine"
```

---

### Task 6: Family Sharing Modal & Member Management UI

**Files:**
- Modify: `web/static/index.html`

**Interfaces:**
- Produces:
  - Header Action: Family / Share button `👥` on List Detail view.
  - Modal: `#shareModal` with email input, role select, and current member list pills.
  - Functions: `showShareModal()`, `hideShareModal()`, `inviteMember()`, `removeMember()`

- [ ] **Step 1: Add Share Button to List Detail Header in `web/static/index.html`**

```html
<button class="logout-btn" onclick="showShareModal()" title="Family Sharing" style="background: #eef2ff; color: var(--primary);">
    <i data-lucide="users" size="18"></i>
    <span id="memberCountBadge" class="filter-count-badge" style="margin-left: 4px; font-size: 0.7rem;">1</span>
</button>
```

- [ ] **Step 2: Add Family Sharing Modal Markup**

```html
<!-- Family Sharing Modal -->
<div id="shareModal" class="modal hidden">
    <div class="modal-content">
        <div class="modal-header">
            <h3 class="modal-title" style="display: flex; align-items: center; gap: 8px;">
                <i data-lucide="users" class="text-primary" size="20"></i>
                Family & Sharing
            </h3>
            <button class="modal-close" onclick="hideShareModal()">&times;</button>
        </div>
        
        <!-- Invite Form -->
        <form id="inviteMemberForm" onsubmit="inviteMember(event)" style="margin-bottom: 20px;">
            <div class="form-group" style="margin-bottom: 10px;">
                <label style="font-size: 0.8rem;">Invite Member by Email</label>
                <div style="display: flex; gap: 8px;">
                    <input type="email" id="inviteEmail" placeholder="family@example.com" required style="flex: 1; padding: 10px 14px;">
                    <select id="inviteRole" style="padding: 10px; border-radius: 12px; border: 1px solid var(--border); font-weight: 600;">
                        <option value="editor">Editor</option>
                        <option value="viewer">Viewer</option>
                    </select>
                </div>
            </div>
            <button type="submit" class="btn" style="padding: 10px; font-size: 0.875rem;">
                <i data-lucide="user-plus" size="16"></i>
                Send Invitation
            </button>
        </form>

        <!-- Member List -->
        <div>
            <div style="font-size: 0.8rem; font-weight: 700; color: var(--text-dim); text-transform: uppercase; margin-bottom: 8px;">
                Current Members
            </div>
            <div id="sharedMembersList" style="display: flex; flex-direction: column; gap: 8px; max-height: 200px; overflow-y: auto;">
                <!-- Rendered by JS -->
            </div>
        </div>
    </div>
</div>
```

- [ ] **Step 3: Add Sharing JavaScript handlers in `index.html`**

Implement `loadMembers()`, `inviteMember()`, and `removeMember()`.

- [ ] **Step 4: Test in browser**

Verify modal opens, fetches members, sends invitation, and updates list.

- [ ] **Step 5: Commit**

```bash
git add web/static/index.html
git commit -m "feat(ui): add family sharing modal and member management interface"
```

---

### Task 7: Activity Log Timeline & Real-Time Item Check-Off UI

**Files:**
- Modify: `web/static/index.html`

**Interfaces:**
- Produces:
  - Collapsible Activity Log drawer on List Detail view.
  - Live check-off DOM updates on incoming SSE `item_updated` events.
  - Toast notification when collaborator modifies an item.

- [ ] **Step 1: Add Activity Log Drawer to List Detail View**

```html
<!-- Activity Log Section -->
<div class="summary-section-card" style="margin-top: 16px;">
    <div class="summary-section-header" style="cursor: pointer;" onclick="toggleActivityLog()">
        <div class="summary-section-title">
            <i data-lucide="history" class="text-primary" size="18"></i>
            Activity Log
        </div>
        <i id="activityChevron" data-lucide="chevron-down" size="16"></i>
    </div>
    <div id="activityLogList" class="hidden" style="display: flex; flex-direction: column; gap: 8px;">
        <!-- Rendered by JS -->
    </div>
</div>
```

- [ ] **Step 2: Connect `liveClient` listeners on `showListPage(listId)`**

In `showListPage(listId)`:
```javascript
window.liveClient.connect(listId);

window.liveClient.on('item_updated', (updatedItem) => {
    // Find DOM element and animate checkbox / price
    const itemEl = document.querySelector(`[data-item-id="${updatedItem.id}"]`);
    if (itemEl) {
        const checkbox = itemEl.querySelector('input[type="checkbox"]');
        if (checkbox) checkbox.checked = updatedItem.purchased;
        itemEl.classList.toggle('item-purchased', updatedItem.purchased);
    }
    updateListSummarySpend();
    showToast(`Updated "${updatedItem.name}"`);
});

window.liveClient.on('activity_logged', (activity) => {
    prependActivityLog(activity);
});
```

- [ ] **Step 3: Clean up stream on navigation back**

In `goBack()`:
```javascript
window.liveClient.disconnect();
```

- [ ] **Step 4: Commit**

```bash
git add web/static/index.html
git commit -m "feat(ui): implement real-time live check-off and activity timeline drawer"
```

---

### Task 8: End-to-End Multi-Device Live Sync Verification

**Files:**
- Test: `cmd/web/main.go`, `web/static/index.html`

- [x] **Step 1: Build and start local server**

Run: `make build-web && ./build/bazarlist-web`

- [x] **Step 2: Test Sharing & Live Broadcast via curl / browser tabs**

1. Register User A (`alice@example.com`) and User B (`bob@example.com`).
2. Alice creates list "Weekend Bazar".
3. Alice invites Bob as `editor`.
4. Bob logs in, sees "Weekend Bazar" in his lists.
5. Alice opens list in Tab 1, Bob opens in Tab 2.
6. Alice marks "Eggs" as purchased in Tab 1.
7. Verify Tab 2 instantly updates "Eggs" as purchased and logs activity *"Alice marked Eggs as purchased"*.

- [x] **Step 3: Commit and tag feature**

```bash
git commit -am "feat: family sharing and real-time live collaboration verified"
```
