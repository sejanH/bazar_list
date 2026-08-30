# Implementation Plan: Offline-First PWA for Bazar List

**Spec:** `docs/superpowers/specs/2026-08-30-pwa-offline-mode-design.md`  
**Branch:** `feature/pwa-offline-mode`  
**Status:** Completed  

---

## Tasks

- [x] **Task 1: Backend Static Serving & Cache-Control Headers**
  - File: `cmd/web/main.go`
  - Ensure `/sw.js` and `/manifest.json` are served with `Cache-Control: no-cache, no-store, must-revalidate` and correct `Content-Type`.
  - Verify: Run `go build ./cmd/web` to confirm compilation.

- [x] **Task 2: PWA Web App Manifest & App Icons**
  - Files: `web/static/manifest.json`, `web/static/icons/icon.svg`
  - Define application identity, standalone display mode, theme colors, and icons.
  - Verify: Check manifest syntax and ensure icons exist.

- [x] **Task 3: Service Worker Implementation**
  - File: `web/static/sw.js`
  - Implement precaching of shell assets (`/`, `/index.html`, `/manifest.json`, icons).
  - Implement cache invalidation on activate (`bazarlist-shell-v1`).
  - Implement Stale-While-Revalidate strategy for static resources and pass-through for API requests.
  - Verify: Syntax validation and Service Worker lifecycle checks.

- [x] **Task 4: IndexedDB Storage & Outbox Mutation Layer**
  - File: `web/static/js/db.js`
  - Implement IndexedDB wrapper (`bazarlist_db`) with `lists`, `outbox`, and `meta` stores.
  - Implement functions: `getListsByMonth()`, `saveLists()`, `saveList()`, `getList()`, `deleteList()`, `queueOutbox()`, `getOutboxItems()`, `deleteOutboxItem()`, `countOutbox()`.
  - Verify: Test IndexedDB initialization and data insertion/retrieval.

- [x] **Task 5: Sync Engine & Temporary ID Reconciliation**
  - File: `web/static/js/sync.js`
  - Implement `SyncEngine` class:
    - Sequential processing of outbox queue.
    - Temporary ID mapping (`temp_id -> server_id`) for lists and items.
    - Event listeners for `online`, `offline`, and `visibilitychange`.
    - Auto-retry on reconnect.
    - Network status badge updater (🟢 Online, 🟡 Offline, 🔄 Syncing...).
  - Verify: Test outbox resolution logic and error handling.

- [x] **Task 6: Frontend Integration & Optimistic UI**
  - File: `web/static/index.html`
  - Link manifest and register Service Worker.
  - Add network status pill and PWA Install button in navigation.
  - Update all list & item actions (Create, Toggle, Add, Delete) to:
    1. Update IndexedDB & DOM optimistically with temporary IDs and pending badges.
    2. Queue action in `outbox`.
    3. Trigger `SyncEngine.sync()` if online.
  - Verify: Check app rendering, offline storage, and UI status updates.

- [x] **Task 7: Build, End-to-End Verification & Commit**
  - Compile the Go web binary (`make build-web`).
  - Run comprehensive verification tests.
  - Review git diff, verify all features work cleanly, and commit.
