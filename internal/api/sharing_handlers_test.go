package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/gorilla/mux"
	"github.com/sejan/bazarlist/internal/api"
	"github.com/sejan/bazarlist/internal/auth"
	"github.com/sejan/bazarlist/internal/live"
	"github.com/sejan/bazarlist/internal/models"
	"github.com/sejan/bazarlist/internal/storage"
	"gorm.io/gorm"
)

type testEnv struct {
	db             *gorm.DB
	store          *storage.MySQLStorage
	hub            *live.Hub
	authMgr        *auth.AuthManager
	sharingHandler *api.SharingHandler
	listHandler    *api.ListHandler
	authHandler    *api.AuthHandler
	router         *mux.Router
	user1          models.User
	user2          models.User
	user3          models.User
	list1          models.ShoppingList
}

func setupTestEnvironment(t *testing.T) *testEnv {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite db: %v", err)
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.ShoppingList{},
		&models.Item{},
		&models.ListMember{},
		&models.ListActivity{},
	)
	if err != nil {
		t.Fatalf("failed to auto-migrate schema: %v", err)
	}

	store := storage.NewStorageWithDB(db)
	hub := live.NewHub()
	sharingHandler := api.NewSharingHandler(store, hub)
	listHandler := api.NewListHandler(store, hub)
	authHandler := api.NewAuthHandler(store)
	authMgr := auth.NewAuthManager()

	u1 := models.User{Name: "Alice", Email: "alice@example.com", Password: "hashedpassword"}
	u2 := models.User{Name: "Bob", Email: "bob@example.com", Password: "hashedpassword"}
	u3 := models.User{Name: "Charlie", Email: "charlie@example.com", Password: "hashedpassword"}
	db.Create(&u1)
	db.Create(&u2)
	db.Create(&u3)

	l1 := models.ShoppingList{UserID: u1.ID, Name: "Family Groceries", Date: time.Now()}
	db.Create(&l1)

	r := mux.NewRouter()
	r.HandleFunc("/api/auth/register", authHandler.Register).Methods("POST")
	r.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST")

	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.Use(authHandler.AuthMiddleware)

	apiRouter.HandleFunc("/lists", listHandler.GetLists).Methods("GET")
	apiRouter.HandleFunc("/lists", listHandler.CreateList).Methods("POST")
	apiRouter.HandleFunc("/lists/{id}", listHandler.GetList).Methods("GET")
	apiRouter.HandleFunc("/lists/{id}", listHandler.UpdateList).Methods("PUT", "PATCH")
	apiRouter.HandleFunc("/lists/{id}", listHandler.DeleteList).Methods("DELETE")

	apiRouter.HandleFunc("/lists/{id}/items", listHandler.GetItems).Methods("GET")
	apiRouter.HandleFunc("/lists/{id}/items", listHandler.CreateItem).Methods("POST")
	apiRouter.HandleFunc("/lists/{id}/items/{itemId}", listHandler.UpdateItem).Methods("PUT", "PATCH")
	apiRouter.HandleFunc("/lists/{id}/items/{itemId}", listHandler.DeleteItem).Methods("DELETE")

	apiRouter.HandleFunc("/lists/{id}/live", sharingHandler.StreamLiveUpdates).Methods("GET")
	apiRouter.HandleFunc("/lists/{id}/members", sharingHandler.GetMembers).Methods("GET")
	apiRouter.HandleFunc("/lists/{id}/members", sharingHandler.InviteMember).Methods("POST")
	apiRouter.HandleFunc("/lists/{id}/members/{userId}", sharingHandler.RemoveMember).Methods("DELETE")
	apiRouter.HandleFunc("/lists/{id}/activities", sharingHandler.GetActivities).Methods("GET")

	return &testEnv{
		db:             db,
		store:          store,
		hub:            hub,
		authMgr:        authMgr,
		sharingHandler: sharingHandler,
		listHandler:    listHandler,
		authHandler:    authHandler,
		router:         r,
		user1:          u1,
		user2:          u2,
		user3:          u3,
		list1:          l1,
	}
}

func (env *testEnv) makeRequest(method, url string, body any, userID uint) *httptest.ResponseRecorder {
	var bodyReader *bytes.Reader
	if body != nil {
		jsonBytes, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(jsonBytes)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	req := httptest.NewRequest(method, url, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	if userID > 0 {
		token, _ := env.authMgr.GenerateToken(userID)
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	env.router.ServeHTTP(rec, req)
	return rec
}

func TestInviteAndGetMembers(t *testing.T) {
	env := setupTestEnvironment(t)

	// 1. Non-owner (user2) tries to invite user3 -> 403 Forbidden
	inviteReq := models.InviteMemberRequest{
		Email: env.user3.Email,
		Role:  models.RoleEditor,
	}
	rec := env.makeRequest("POST", fmt.Sprintf("/api/lists/%d/members", env.list1.ID), inviteReq, env.user2.ID)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for non-owner invite, got %d: %s", rec.Code, rec.Body.String())
	}

	// 2. Owner invites non-existent email -> 404 Not Found
	rec = env.makeRequest("POST", fmt.Sprintf("/api/lists/%d/members", env.list1.ID), models.InviteMemberRequest{Email: "missing@example.com"}, env.user1.ID)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found for non-existent user invite, got %d: %s", rec.Code, rec.Body.String())
	}

	// 3. Owner invites self -> 400 Bad Request
	rec = env.makeRequest("POST", fmt.Sprintf("/api/lists/%d/members", env.list1.ID), models.InviteMemberRequest{Email: env.user1.Email}, env.user1.ID)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request when inviting owner, got %d", rec.Code)
	}

	// 4. Owner invites user2 as Editor -> 200 OK
	rec = env.makeRequest("POST", fmt.Sprintf("/api/lists/%d/members", env.list1.ID), models.InviteMemberRequest{
		Email: env.user2.Email,
		Role:  models.RoleEditor,
	}, env.user1.ID)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for successful invite, got %d: %s", rec.Code, rec.Body.String())
	}

	// 5. GetMembers as member (user2) -> 200 OK with owner and members
	rec = env.makeRequest("GET", fmt.Sprintf("/api/lists/%d/members", env.list1.ID), nil, env.user2.ID)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK getting members, got %d: %s", rec.Code, rec.Body.String())
	}

	var membersResp struct {
		Owner   models.User         `json:"owner"`
		Members []models.ListMember `json:"members"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&membersResp); err != nil {
		t.Fatalf("failed to decode members response: %v", err)
	}
	if membersResp.Owner.ID != env.user1.ID {
		t.Fatalf("expected owner ID %d, got %d", env.user1.ID, membersResp.Owner.ID)
	}
	if len(membersResp.Members) != 1 || membersResp.Members[0].UserID != env.user2.ID {
		t.Fatalf("expected 1 member (user2), got %d members", len(membersResp.Members))
	}

	// 6. Non-member (user3) GetMembers -> 403 Forbidden
	rec = env.makeRequest("GET", fmt.Sprintf("/api/lists/%d/members", env.list1.ID), nil, env.user3.ID)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for non-member GetMembers, got %d", rec.Code)
	}
}

func TestRemoveMember(t *testing.T) {
	env := setupTestEnvironment(t)

	// Add user2 as editor
	_ = env.store.AddListMember(env.list1.ID, env.user2.ID, models.RoleEditor)

	// Non-owner cannot remove member -> 403
	rec := env.makeRequest("DELETE", fmt.Sprintf("/api/lists/%d/members/%d", env.list1.ID, env.user2.ID), nil, env.user2.ID)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for non-owner removing member, got %d", rec.Code)
	}

	// Owner removes user2 -> 200 OK
	rec = env.makeRequest("DELETE", fmt.Sprintf("/api/lists/%d/members/%d", env.list1.ID, env.user2.ID), nil, env.user1.ID)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK removing member, got %d: %s", rec.Code, rec.Body.String())
	}

	// Verify user2 no longer has access
	canView, _ := env.store.CanViewList(env.user2.ID, env.list1.ID)
	if canView {
		t.Fatalf("expected removed member user2 to no longer have view access")
	}
}

func TestActivitiesAndLiveHubSync(t *testing.T) {
	env := setupTestEnvironment(t)

	// Add user2 as Editor
	_ = env.store.AddListMember(env.list1.ID, env.user2.ID, models.RoleEditor)

	// Subscribe to live hub to verify broadcasts
	events, unsubscribe := env.hub.Subscribe(env.list1.ID)
	defer unsubscribe()

	// 1. User2 creates an item -> 201 Created
	itemReq := models.ItemRequest{
		Name:      "Fresh Apples",
		Price:     150.00,
		Purchased: false,
	}
	rec := env.makeRequest("POST", fmt.Sprintf("/api/lists/%d/items", env.list1.ID), itemReq, env.user2.ID)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for item creation, got %d: %s", rec.Code, rec.Body.String())
	}

	var createdItem models.Item
	_ = json.NewDecoder(rec.Body).Decode(&createdItem)

	// Verify live hub broadcast received for item_created
	select {
	case event := <-events:
		if event.Type != "item_created" {
			t.Fatalf("expected event item_created, got %s", event.Type)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for item_created broadcast")
	}

	// Consume activity_logged broadcast
	select {
	case event := <-events:
		if event.Type != "activity_logged" {
			t.Fatalf("expected event activity_logged, got %s", event.Type)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for activity_logged broadcast")
	}

	// 2. User2 marks item as purchased -> 200 OK
	updateReq := models.ItemRequest{
		Name:      "Fresh Apples",
		Price:     150.00,
		Purchased: true,
	}
	rec = env.makeRequest("PUT", fmt.Sprintf("/api/lists/%d/items/%d", env.list1.ID, createdItem.ID), updateReq, env.user2.ID)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for item update, got %d: %s", rec.Code, rec.Body.String())
	}

	// Check broadcast for item_updated
	select {
	case event := <-events:
		if event.Type != "item_updated" {
			t.Fatalf("expected event item_updated, got %s", event.Type)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for item_updated broadcast")
	}

	// Consume activity_logged broadcast from update
	select {
	case event := <-events:
		if event.Type != "activity_logged" {
			t.Fatalf("expected event activity_logged, got %s", event.Type)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for activity_logged broadcast")
	}

	// 3. User2 deletes item -> 200 OK
	rec = env.makeRequest("DELETE", fmt.Sprintf("/api/lists/%d/items/%d", env.list1.ID, createdItem.ID), nil, env.user2.ID)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for item delete, got %d: %s", rec.Code, rec.Body.String())
	}

	// Check broadcast for item_deleted
	select {
	case event := <-events:
		if event.Type != "item_deleted" {
			t.Fatalf("expected event item_deleted, got %s", event.Type)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for item_deleted broadcast")
	}

	// Consume activity_logged broadcast from delete
	select {
	case event := <-events:
		if event.Type != "activity_logged" {
			t.Fatalf("expected event activity_logged, got %s", event.Type)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for activity_logged broadcast")
	}

	// 4. GetActivities -> should contain activities logged
	rec = env.makeRequest("GET", fmt.Sprintf("/api/lists/%d/activities", env.list1.ID), nil, env.user2.ID)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for GetActivities, got %d: %s", rec.Code, rec.Body.String())
	}

	var activities []models.ListActivity
	if err := json.NewDecoder(rec.Body).Decode(&activities); err != nil {
		t.Fatalf("failed to decode activities: %v", err)
	}
	if len(activities) < 3 {
		t.Fatalf("expected at least 3 activities, got %d", len(activities))
	}
}

func TestStreamLiveUpdates(t *testing.T) {
	env := setupTestEnvironment(t)

	// User3 (non-member) tries to stream -> 403 Forbidden
	rec := env.makeRequest("GET", fmt.Sprintf("/api/lists/%d/live", env.list1.ID), nil, env.user3.ID)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for non-member stream, got %d", rec.Code)
	}

	// User1 (owner) connects to SSE stream
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	token, _ := env.authMgr.GenerateToken(env.user1.ID)
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/lists/%d/live", env.list1.ID), nil).WithContext(ctx)
	req.Header.Set("Authorization", "Bearer "+token)
	recStreaming := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		env.router.ServeHTTP(recStreaming, req)
		close(done)
	}()

	// Broadcast an event to list1
	time.Sleep(20 * time.Millisecond)
	env.hub.Broadcast(env.list1.ID, "test_event", map[string]string{"foo": "bar"})

	// Allow event to be written
	time.Sleep(30 * time.Millisecond)
	cancel()
	<-done

	body := recStreaming.Body.String()
	if !strings.Contains(body, "event: connected") {
		t.Fatalf("expected stream to contain connected event, got:\n%s", body)
	}
	if !strings.Contains(body, "event: test_event") {
		t.Fatalf("expected stream to contain test_event, got:\n%s", body)
	}
	if !strings.Contains(body, `"foo":"bar"`) {
		t.Fatalf("expected stream to contain foo:bar payload, got:\n%s", body)
	}
}

func TestAuthMiddlewareTokenQueryParam(t *testing.T) {
	env := setupTestEnvironment(t)

	// Register a new user through auth endpoint to get a valid token
	regReq := models.RegisterRequest{
		Name:     "Test Auth",
		Email:    "testauth@example.com",
		Password: "password123",
	}
	regRec := env.makeRequest("POST", "/api/auth/register", regReq, 0)
	if regRec.Code != http.StatusCreated {
		t.Fatalf("failed to register user: %s", regRec.Body.String())
	}
	var tokenResp models.TokenResponse
	_ = json.NewDecoder(regRec.Body).Decode(&tokenResp)

	// Access with query param ?token=
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/lists?token=%s", tokenResp.Token), nil)
	rec := httptest.NewRecorder()
	env.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK using ?token= query parameter, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestViewerAndEditorPermissions(t *testing.T) {
	env := setupTestEnvironment(t)

	// Add user2 as Editor, user3 as Viewer
	_ = env.store.AddListMember(env.list1.ID, env.user2.ID, models.RoleEditor)
	_ = env.store.AddListMember(env.list1.ID, env.user3.ID, models.RoleViewer)

	// 1. Viewer (user3) can view list and items -> 200 OK
	rec := env.makeRequest("GET", fmt.Sprintf("/api/lists/%d", env.list1.ID), nil, env.user3.ID)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for viewer GetList, got %d: %s", rec.Code, rec.Body.String())
	}
	rec = env.makeRequest("GET", fmt.Sprintf("/api/lists/%d/items", env.list1.ID), nil, env.user3.ID)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for viewer GetItems, got %d: %s", rec.Code, rec.Body.String())
	}

	// 2. Viewer (user3) CANNOT create item -> 403 Forbidden
	rec = env.makeRequest("POST", fmt.Sprintf("/api/lists/%d/items", env.list1.ID), models.ItemRequest{
		Name:  "Bananas",
		Price: 60.00,
	}, env.user3.ID)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for viewer CreateItem, got %d: %s", rec.Code, rec.Body.String())
	}

	// 3. Editor (user2) CAN create item -> 201 Created
	rec = env.makeRequest("POST", fmt.Sprintf("/api/lists/%d/items", env.list1.ID), models.ItemRequest{
		Name:  "Bananas",
		Price: 60.00,
	}, env.user2.ID)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for editor CreateItem, got %d: %s", rec.Code, rec.Body.String())
	}

	var createdItem models.Item
	_ = json.NewDecoder(rec.Body).Decode(&createdItem)

	// 4. Viewer (user3) CANNOT update or delete item -> 403 Forbidden
	rec = env.makeRequest("PUT", fmt.Sprintf("/api/lists/%d/items/%d", env.list1.ID, createdItem.ID), models.ItemRequest{
		Name:      "Bananas",
		Price:     60.00,
		Purchased: true,
	}, env.user3.ID)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for viewer UpdateItem, got %d", rec.Code)
	}

	rec = env.makeRequest("DELETE", fmt.Sprintf("/api/lists/%d/items/%d", env.list1.ID, createdItem.ID), nil, env.user3.ID)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for viewer DeleteItem, got %d", rec.Code)
	}

	// 5. Editor (user2) CANNOT delete list (only owner can) -> 403 Forbidden
	rec = env.makeRequest("DELETE", fmt.Sprintf("/api/lists/%d", env.list1.ID), nil, env.user2.ID)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for editor DeleteList, got %d", rec.Code)
	}

	// 6. Owner (user1) CAN delete list -> 200 OK
	rec = env.makeRequest("DELETE", fmt.Sprintf("/api/lists/%d", env.list1.ID), nil, env.user1.ID)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for owner DeleteList, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestStreamLiveUpdatesTokenQueryParam(t *testing.T) {
	env := setupTestEnvironment(t)

	token, _ := env.authMgr.GenerateToken(env.user1.ID)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/lists/%d/live?token=%s", env.list1.ID, token), nil).WithContext(ctx)
	recStreaming := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		env.router.ServeHTTP(recStreaming, req)
		close(done)
	}()

	time.Sleep(20 * time.Millisecond)
	env.hub.Broadcast(env.list1.ID, "sync_test", map[string]int{"count": 42})

	time.Sleep(30 * time.Millisecond)
	cancel()
	<-done

	body := recStreaming.Body.String()
	if !strings.Contains(body, "event: connected") {
		t.Fatalf("expected stream to contain connected event, got:\n%s", body)
	}
	if !strings.Contains(body, "event: sync_test") {
		t.Fatalf("expected stream to contain sync_test event, got:\n%s", body)
	}
	if !strings.Contains(body, `"count":42`) {
		t.Fatalf("expected stream to contain count:42, got:\n%s", body)
	}
}
