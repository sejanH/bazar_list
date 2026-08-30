# Design Spec: Offline-First PWA for Bazar List

**Date:** 2026-08-30  
**Status:** Approved for Implementation  
**Topic:** Progressive Web App (PWA) with Offline-First Shopping and Outbox Sync Engine  

---

## 1. Executive Summary

Bazar List is an application for managing grocery and supermarket shopping trips. Users frequently experience poor or intermittent cellular reception inside physical grocery stores and basements.

This design introduces a **Full Offline-First Progressive Web Application (PWA)** architecture. Users will be able to:
1. Install Bazar List as a standalone desktop or mobile application.
2. Load and navigate the app instantly even without an internet connection.
3. Perform all shopping actions offline (create lists, add items, set prices, toggle purchased status, and delete items).
4. Automatically sync and reconcile all offline mutations with the Go REST API when internet connectivity returns.

---

## 2. Architecture & Components

```
+-------------------------------------------------------------------------+
|                              Browser UI                                 |
|  +---------------------+  +--------------------+  +------------------+  |
|  |  Shopping List UI   |  |   Network Status   |  | Install Prompt   |  |
|  |  (Optimistic DOM)   |  |  (Online/Offline)  |  | (PWA Banner)     |  |
|  +----------+----------+  +---------+----------+  +--------+---------+  |
+-------------|-----------------------|----------------------|------------+
              |                       |                      |
              v                       v                      v
+-------------------------------------------------------------------------+
|                     Client-Side Data & Sync Layer                       |
|                                                                         |
|  +---------------------------------+  +------------------------------+  |
|  |       IndexedDB Storage         |  |         Sync Engine          |  |
|  |  - lists (cached items/totals)  |  |  - Sequential Queue Worker   |  |
|  |  - outbox (mutation operations) |  |  - Temp ID to Server ID Map  |  |
|  +---------------------------------+  +--------------+---------------+  |
+------------------------------------------------------|------------------+
                                                       |
+------------------------------------------------------|------------------+
|               Service Worker (sw.js)                 |
|  - Static Shell Cache (HTML/CSS/JS/Icons)            |
|  - Stale-While-Revalidate Asset Strategy             |
|  - Bypass API Requests to Sync Engine                |
+------------------------------------------------------|------------------+
                                                       | (HTTP / JWT)
                                                       v
+-------------------------------------------------------------------------+
|                          Go Backend (REST API)                          |
|  - /api/auth/*                                                          |
|  - /api/lists                                                           |
|  - /api/lists/{id}/items                                                |
|  - Static file server with optimized no-cache headers for sw.js         |
+-------------------------------------------------------------------------+
```

---

## 3. Detailed Specifications

### 3.1 Web App Manifest (`web/static/manifest.json`)
- **Metadata:**
  - `name`: `"Bazar List - Smart Shopping"`
  - `short_name`: `"BazarList"`
  - `description`: `"Personal shopping list manager with budget tracking & offline supermarket mode"`
  - `start_url`: `"/"`
  - `display`: `"standalone"`
  - `background_color`: `"#ffffff"`
  - `theme_color`: `"#2563eb"`
  - `orientation`: `"portrait-primary"`
- **Icons:**
  - 192x192 PNG (standard)
  - 512x512 PNG (standard)
  - Maskable icon SVG/PNG

### 3.2 Service Worker (`web/static/sw.js`)
- **Cache Versioning:** `bazarlist-shell-v1`
- **Precaching List:**
  - `/`
  - `/index.html`
  - `/manifest.json`
- **Caching Strategies:**
  - **Static Shell Assets:** Stale-While-Revalidate (serves cache instantly, updates cache in background).
  - **HTML Navigation:** Network-first falling back to cached `/index.html`.
  - **API Requests (`/api/*`):** Not cached by Service Worker; managed exclusively by client-side Sync Engine and IndexedDB.

### 3.3 Client-Side Storage (`IndexedDB`: `bazarlist_db`)
- **Database Version:** 1
- **Object Stores:**
  1. `lists`:
     - Key: `id` (Supports both numeric server IDs and `temp_...` strings)
     - Indexes: `month` (`YYYY-MM`), `user_id`
     - Fields: `id`, `user_id`, `name`, `date`, `items`, `total_amount`, `updated_at`
  2. `outbox`:
     - Key: `id` (Auto-increment integer)
     - Fields:
       - `action`: `"CREATE_LIST"` | `"UPDATE_LIST"` | `"DELETE_LIST"` | `"CREATE_ITEM"` | `"UPDATE_ITEM"` | `"DELETE_ITEM"`
       - `temp_id`: Unique client string identifier (e.g. `temp_1725000000123_45`)
       - `target_list_id`: Server list ID or temporary list ID
       - `target_item_id`: Optional server/temp item ID (for item updates/deletions)
       - `payload`: JSON object with request parameters
       - `created_at`: Timestamp
       - `retry_count`: Number of retry attempts

### 3.4 Sync Engine & Reconciliation
- **Lifecycle Triggers:**
  - `window.addEventListener('online', runSync)`
  - `document.addEventListener('visibilitychange', runSync)`
  - Periodic polling (every 15s when `outbox.length > 0` and `navigator.onLine == true`)
  - User-initiated "Sync Now" button
- **Reconciliation Algorithm:**
  1. Retrieve pending mutations from `outbox` sorted by ascending `id`.
  2. Maintain an in-memory session translation map: `tempIdMap = new Map()`.
  3. For each operation:
     - If `target_list_id` exists in `tempIdMap`, rewrite it with the mapped permanent server ID.
     - If `target_item_id` exists in `tempIdMap`, rewrite it with the mapped permanent server ID.
     - Dispatch corresponding HTTP request with current JWT token.
     - Upon HTTP 200/201:
       - If server returns a new ID for a temporary entity, update `tempIdMap`, replace the temporary ID in IndexedDB `lists`, and update corresponding DOM element attributes.
       - Delete the operation from `outbox`.
     - Upon HTTP 401: Pause sync and present authentication prompt.
     - Upon HTTP 404 (resource deleted on server): Remove orphaned local operation and update local cache.
     - Upon Network Failure: Stop sync cycle and wait for next `online` event.
  4. Fetch fresh month data `/api/lists?month=...` upon complete queue drainage to ensure full parity.

### 3.5 UI & Network Status Indicators
- **Status Indicator Header Pill:**
  - 🟢 **Online** (0 pending changes)
  - 🟡 **Offline (N changes saved locally)**
  - 🔄 **Syncing... (X/N)**
- **Item State Indicators:**
  - Optimistic items show a subtle clock icon `⏳` until resolved to `✓`.
- **Install App Button:**
  - Listens to `beforeinstallprompt` event.
  - Prominently displays an "📱 Install App" button in navigation.

---

## 4. Backend Adjustments (Go)

- **Header Management:**
  - Update `cmd/web/main.go` static file server handler to inject `Cache-Control: no-cache, no-store, must-revalidate` for `sw.js` and `manifest.json`.
  - Retain existing REST API routes without breaking changes.

---

## 5. Verification & Test Plan

1. **Service Worker Registration:** Verify registration succeeds on first load with no console warnings.
2. **PWA Lighthouse Audit:** Run Chrome Lighthouse PWA audit to ensure manifest and installability pass.
3. **Simulated Offline Flow:**
   - Disconnect network in browser DevTools.
   - Reload page $\rightarrow$ confirm shell loads.
   - Create a list named *"Supermarket Run"*.
   - Add items *"Rice 5kg"* (price: 350) and *"Cooking Oil 2L"* (price: 380).
   - Check off *"Rice 5kg"*.
   - Confirm immediate UI update and IndexedDB outbox record creation.
4. **Reconnection & Sync Flow:**
   - Reconnect network in DevTools.
   - Verify outbox processes sequentially.
   - Verify server receives valid IDs and database has consistent rows.
   - Verify UI badges switch to synced.
