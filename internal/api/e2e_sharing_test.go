package api_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/gorilla/mux"
	"github.com/sejan/bazarlist/internal/api"
	"github.com/sejan/bazarlist/internal/live"
	"github.com/sejan/bazarlist/internal/models"
	"github.com/sejan/bazarlist/internal/storage"
	"gorm.io/gorm"
)

type sseEvent struct {
	Type string
	Data string
}

func TestEndToEndMultiDeviceLiveSync(t *testing.T) {
	// 1. Setup in-memory sqlite DB and GORM
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
		t.Fatalf("failed to auto-migrate database: %v", err)
	}

	store := storage.NewStorageWithDB(db)
	hub := live.NewHub()

	authHandler := api.NewAuthHandler(store)
	listHandler := api.NewListHandler(store, hub)
	sharingHandler := api.NewSharingHandler(store, hub)

	router := mux.NewRouter()
	router.Use(api.CORSMiddleware)

	// Auth routes
	router.HandleFunc("/api/auth/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST")

	// Protected routes
	apiRouter := router.PathPrefix("/api").Subrouter()
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

	// Start real HTTP test server for streaming SSE verification
	ts := httptest.NewServer(router)
	defer ts.Close()

	client := ts.Client()

	// Helper for authenticated HTTP requests
	sendReq := func(method, path string, body any, token string) (*http.Response, []byte) {
		var bodyReader io.Reader
		if body != nil {
			jsonData, _ := json.Marshal(body)
			bodyReader = bytes.NewReader(jsonData)
		}
		req, err := http.NewRequest(method, ts.URL+path, bodyReader)
		if err != nil {
			t.Fatalf("failed to create request %s %s: %v", method, path, err)
		}
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("failed to execute request %s %s: %v", method, path, err)
		}
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		return resp, respBody
	}

	// -------------------------------------------------------------
	// Step 1: Register User A (Alice) and User B (Bob)
	// -------------------------------------------------------------
	aliceEmail := "alice_e2e@example.com"
	bobEmail := "bob_e2e@example.com"

	respA, bodyA := sendReq("POST", "/api/auth/register", models.RegisterRequest{
		Name:     "Alice",
		Email:    aliceEmail,
		Password: "password123",
	}, "")
	if respA.StatusCode != http.StatusCreated {
		t.Fatalf("failed to register Alice: status %d, body: %s", respA.StatusCode, string(bodyA))
	}
	var aliceTokenResp models.TokenResponse
	if err := json.Unmarshal(bodyA, &aliceTokenResp); err != nil {
		t.Fatalf("failed to unmarshal Alice token: %v", err)
	}

	respB, bodyB := sendReq("POST", "/api/auth/register", models.RegisterRequest{
		Name:     "Bob",
		Email:    bobEmail,
		Password: "password123",
	}, "")
	if respB.StatusCode != http.StatusCreated {
		t.Fatalf("failed to register Bob: status %d, body: %s", respB.StatusCode, string(bodyB))
	}
	var bobTokenResp models.TokenResponse
	if err := json.Unmarshal(bodyB, &bobTokenResp); err != nil {
		t.Fatalf("failed to unmarshal Bob token: %v", err)
	}

	// -------------------------------------------------------------
	// Step 2: Alice creates list "Family Groceries"
	// -------------------------------------------------------------
	respList, bodyList := sendReq("POST", "/api/lists", models.ListRequest{
		Name: "Family Groceries",
	}, aliceTokenResp.Token)
	if respList.StatusCode != http.StatusCreated {
		t.Fatalf("Alice failed to create list: status %d, body: %s", respList.StatusCode, string(bodyList))
	}
	var createdList models.ShoppingList
	if err := json.Unmarshal(bodyList, &createdList); err != nil {
		t.Fatalf("failed to unmarshal created list: %v", err)
	}
	listID := createdList.ID
	if listID == 0 {
		t.Fatalf("expected non-zero listID, got %d", listID)
	}

	// -------------------------------------------------------------
	// Step 3: Alice invites Bob as editor to "Family Groceries"
	// -------------------------------------------------------------
	respInvite, bodyInvite := sendReq("POST", fmt.Sprintf("/api/lists/%d/members", listID), models.InviteMemberRequest{
		Email: bobEmail,
		Role:  models.RoleEditor,
	}, aliceTokenResp.Token)
	if respInvite.StatusCode != http.StatusOK {
		t.Fatalf("Alice failed to invite Bob: status %d, body: %s", respInvite.StatusCode, string(bodyInvite))
	}

	// -------------------------------------------------------------
	// Step 4: Bob retrieves his lists via GET /api/lists
	// -------------------------------------------------------------
	respBobLists, bodyBobLists := sendReq("GET", "/api/lists", nil, bobTokenResp.Token)
	if respBobLists.StatusCode != http.StatusOK {
		t.Fatalf("Bob failed to get lists: status %d, body: %s", respBobLists.StatusCode, string(bodyBobLists))
	}
	var bobListsResp models.PaginatedListsResponse
	if err := json.Unmarshal(bodyBobLists, &bobListsResp); err != nil {
		t.Fatalf("failed to unmarshal Bob's lists response: %v", err)
	}
	foundSharedList := false
	for _, l := range bobListsResp.Lists {
		if l.ID == listID && l.Name == "Family Groceries" {
			foundSharedList = true
			break
		}
	}
	if !foundSharedList {
		t.Fatalf("Bob's lists did not contain shared list #%d 'Family Groceries'", listID)
	}

	// -------------------------------------------------------------
	// Step 5: Bob connects to SSE stream GET /api/lists/{id}/live?token=...
	// -------------------------------------------------------------
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sseReq, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/lists/%d/live?token=%s", ts.URL, listID, bobTokenResp.Token), nil)
	if err != nil {
		t.Fatalf("failed to create SSE stream request: %v", err)
	}

	sseResp, err := client.Do(sseReq)
	if err != nil {
		t.Fatalf("Bob failed to connect to live SSE stream: %v", err)
	}
	defer sseResp.Body.Close()

	if sseResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for SSE stream, got %d", sseResp.StatusCode)
	}
	if sseResp.Header.Get("Content-Type") != "text/event-stream" {
		t.Fatalf("expected Content-Type text/event-stream, got %s", sseResp.Header.Get("Content-Type"))
	}

	eventsChan := make(chan sseEvent, 50)
	go func() {
		scanner := bufio.NewScanner(sseResp.Body)
		var currentEvent string
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "event:") {
				currentEvent = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			} else if strings.HasPrefix(line, "data:") {
				data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
				eventsChan <- sseEvent{
					Type: currentEvent,
					Data: data,
				}
				currentEvent = ""
			}
		}
	}()

	// Verify Bob receives initial "connected" event
	select {
	case ev := <-eventsChan:
		if ev.Type != "connected" {
			t.Fatalf("expected initial event 'connected', got '%s'", ev.Type)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for Bob's initial SSE 'connected' event")
	}

	// -------------------------------------------------------------
	// Step 6: Alice creates item "Milk" and updates it to purchased=true
	// -------------------------------------------------------------
	// 6a: Create item "Milk"
	respItem, bodyItem := sendReq("POST", fmt.Sprintf("/api/lists/%d/items", listID), models.ItemRequest{
		Name:      "Milk",
		Price:     85.00,
		Purchased: false,
	}, aliceTokenResp.Token)
	if respItem.StatusCode != http.StatusCreated {
		t.Fatalf("Alice failed to create item: status %d, body: %s", respItem.StatusCode, string(bodyItem))
	}
	var createdItem models.Item
	if err := json.Unmarshal(bodyItem, &createdItem); err != nil {
		t.Fatalf("failed to unmarshal created item: %v", err)
	}
	if createdItem.Name != "Milk" || createdItem.Purchased != false {
		t.Fatalf("unexpected item created: %+v", createdItem)
	}

	// 6b: Update item "Milk" to purchased=true
	respUpdate, bodyUpdate := sendReq("PUT", fmt.Sprintf("/api/lists/%d/items/%d", listID, createdItem.ID), models.ItemRequest{
		Name:      "Milk",
		Price:     85.00,
		Purchased: true,
	}, aliceTokenResp.Token)
	if respUpdate.StatusCode != http.StatusOK {
		t.Fatalf("Alice failed to update item: status %d, body: %s", respUpdate.StatusCode, string(bodyUpdate))
	}

	// -------------------------------------------------------------
	// Step 7: Verify Bob's SSE stream receives item_created, item_updated, and activity_logged events
	// -------------------------------------------------------------
	receivedEvents := make(map[string]int)
	timeout := time.After(2 * time.Second)

	// We expect: item_created, activity_logged (for add), item_updated, activity_logged (for purchase)
	for len(receivedEvents) < 3 || receivedEvents["activity_logged"] < 2 {
		select {
		case ev := <-eventsChan:
			receivedEvents[ev.Type]++
			if ev.Type == "item_created" {
				var item models.Item
				if err := json.Unmarshal([]byte(ev.Data), &item); err != nil || item.Name != "Milk" {
					t.Fatalf("invalid item_created event payload: %s", ev.Data)
				}
			} else if ev.Type == "item_updated" {
				var item models.Item
				if err := json.Unmarshal([]byte(ev.Data), &item); err != nil || item.Name != "Milk" || !item.Purchased {
					t.Fatalf("invalid item_updated event payload: %s", ev.Data)
				}
			}
		case <-timeout:
			t.Fatalf("timed out waiting for SSE events. Received so far: %+v", receivedEvents)
		}
	}

	if receivedEvents["item_created"] < 1 {
		t.Fatalf("expected item_created event in Bob's SSE stream")
	}
	if receivedEvents["item_updated"] < 1 {
		t.Fatalf("expected item_updated event in Bob's SSE stream")
	}
	if receivedEvents["activity_logged"] < 2 {
		t.Fatalf("expected at least 2 activity_logged events in Bob's SSE stream, got %d", receivedEvents["activity_logged"])
	}

	// -------------------------------------------------------------
	// Step 8: Bob fetches GET /api/lists/{id}/activities and verifies logs for Alice's actions
	// -------------------------------------------------------------
	respAct, bodyAct := sendReq("GET", fmt.Sprintf("/api/lists/%d/activities", listID), nil, bobTokenResp.Token)
	if respAct.StatusCode != http.StatusOK {
		t.Fatalf("Bob failed to get activities: status %d, body: %s", respAct.StatusCode, string(bodyAct))
	}
	var activities []models.ListActivity
	if err := json.Unmarshal(bodyAct, &activities); err != nil {
		t.Fatalf("failed to unmarshal activities: %v", err)
	}

	// Verify activities contains item_added and item_purchased
	hasItemAdded := false
	hasItemPurchased := false
	for _, a := range activities {
		if a.Action == models.ActionItemAdded && a.ItemName == "Milk" && a.UserName == "Alice" {
			hasItemAdded = true
		}
		if a.Action == models.ActionItemPurchased && a.ItemName == "Milk" && a.UserName == "Alice" {
			hasItemPurchased = true
		}
	}

	if !hasItemAdded {
		t.Fatalf("Bob's activity list did not contain item_added action for Milk by Alice: %+v", activities)
	}
	if !hasItemPurchased {
		t.Fatalf("Bob's activity list did not contain item_purchased action for Milk by Alice: %+v", activities)
	}
}
