// Background Sync Engine & Network State Manager for Bazar List PWA

class SyncEngine {
  constructor() {
    this.isSyncing = false;
    this.tempIdMap = new Map(); // Maps client temp_ids to permanent server IDs
    this.onSyncCompleteCallbacks = [];
  }

  init() {
    // Network event listeners
    window.addEventListener('online', () => {
      this.updateStatusUI();
      this.sync();
    });

    window.addEventListener('offline', () => {
      this.updateStatusUI();
    });

    // Sync when tab becomes visible again
    document.addEventListener('visibilitychange', () => {
      if (!document.hidden && navigator.onLine) {
        this.sync();
      }
    });

    // Periodic check every 20 seconds
    setInterval(() => {
      if (navigator.onLine && !this.isSyncing) {
        this.sync();
      }
    }, 20000);

    this.updateStatusUI();
  }

  onSyncComplete(cb) {
    this.onSyncCompleteCallbacks.push(cb);
  }

  async updateStatusUI() {
    const statusDot = document.getElementById('networkStatusPill');
    if (!statusDot) return;

    const isOnline = navigator.onLine;
    const pendingCount = await window.appDB.countOutbox();

    if (!isOnline) {
      statusDot.className = 'status-indicator-dot dot-offline';
      statusDot.title = `Offline${pendingCount > 0 ? ` (${pendingCount} pending changes)` : ''}`;
    } else if (this.isSyncing) {
      statusDot.className = 'status-indicator-dot dot-syncing';
      statusDot.title = 'Syncing changes...';
    } else {
      statusDot.className = 'status-indicator-dot dot-online';
      statusDot.title = 'Online';
    }
  }

  async sync() {
    const token = localStorage.getItem('token');
    if (!token || !navigator.onLine || this.isSyncing) {
      this.updateStatusUI();
      return;
    }

    const items = await window.appDB.getOutboxItems();
    if (items.length === 0) {
      this.updateStatusUI();
      return;
    }

    this.isSyncing = true;
    this.updateStatusUI();

    try {
      for (const entry of items) {
        // Resolve mapped IDs if dependencies were created offline
        let targetListId = entry.target_list_id;
        if (this.tempIdMap.has(targetListId)) {
          targetListId = this.tempIdMap.get(targetListId);
        }

        let targetItemId = entry.target_item_id;
        if (this.tempIdMap.has(targetItemId)) {
          targetItemId = this.tempIdMap.get(targetItemId);
        }

        let success = false;
        let responseData = null;

        try {
          switch (entry.action) {
            case 'CREATE_LIST': {
              const res = await fetch('/api/lists', {
                method: 'POST',
                headers: {
                  'Content-Type': 'application/json',
                  'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify(entry.payload)
              });

              if (res.ok) {
                responseData = await res.json();
                if (entry.temp_id && responseData.id) {
                  this.tempIdMap.set(entry.temp_id, responseData.id);
                  // Update IndexedDB list key from temp to server ID
                  await window.appDB.deleteList(entry.temp_id);
                  await window.appDB.saveList(responseData);
                }
                success = true;
              } else if (res.status === 401) {
                this.handleAuthError();
                return;
              }
              break;
            }

            case 'UPDATE_LIST': {
              const res = await fetch(`/api/lists/${targetListId}`, {
                method: 'PUT',
                headers: {
                  'Content-Type': 'application/json',
                  'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify(entry.payload)
              });

              if (res.ok || res.status === 404) {
                success = true;
              } else if (res.status === 401) {
                this.handleAuthError();
                return;
              }
              break;
            }

            case 'DELETE_LIST': {
              const res = await fetch(`/api/lists/${targetListId}`, {
                method: 'DELETE',
                headers: { 'Authorization': `Bearer ${token}` }
              });

              if (res.ok || res.status === 404) {
                success = true;
              } else if (res.status === 401) {
                this.handleAuthError();
                return;
              }
              break;
            }

            case 'CREATE_ITEM': {
              const res = await fetch(`/api/lists/${targetListId}/items`, {
                method: 'POST',
                headers: {
                  'Content-Type': 'application/json',
                  'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify(entry.payload)
              });

              if (res.ok) {
                responseData = await res.json();
                if (entry.temp_id && responseData.id) {
                  this.tempIdMap.set(entry.temp_id, responseData.id);
                  // Update item in local list in IndexedDB
                  const localList = await window.appDB.getList(targetListId);
                  if (localList && localList.items) {
                    const itemIdx = localList.items.findIndex(i => i.id === entry.temp_id);
                    if (itemIdx !== -1) {
                      localList.items[itemIdx] = responseData;
                      await window.appDB.saveList(localList);
                    }
                  }
                }
                success = true;
              } else if (res.status === 404) {
                // List was deleted on server
                success = true;
              } else if (res.status === 401) {
                this.handleAuthError();
                return;
              }
              break;
            }

            case 'UPDATE_ITEM': {
              const res = await fetch(`/api/lists/${targetListId}/items/${targetItemId}`, {
                method: 'PUT',
                headers: {
                  'Content-Type': 'application/json',
                  'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify(entry.payload)
              });

              if (res.ok || res.status === 404) {
                success = true;
              } else if (res.status === 401) {
                this.handleAuthError();
                return;
              }
              break;
            }

            case 'DELETE_ITEM': {
              const res = await fetch(`/api/lists/${targetListId}/items/${targetItemId}`, {
                method: 'DELETE',
                headers: { 'Authorization': `Bearer ${token}` }
              });

              if (res.ok || res.status === 404) {
                success = true;
              } else if (res.status === 401) {
                this.handleAuthError();
                return;
              }
              break;
            }

            default:
              success = true;
          }
        } catch (netErr) {
          // Network interruption during fetch; break out of loop and wait for next online event
          console.warn('Network interruption during sync:', netErr);
          break;
        }

        if (success) {
          await window.appDB.deleteOutboxItem(entry.id);
        }
      }
    } finally {
      this.isSyncing = false;
      this.updateStatusUI();

      // Trigger completion callbacks if all items synced
      const remaining = await window.appDB.countOutbox();
      if (remaining === 0) {
        this.onSyncCompleteCallbacks.forEach(cb => {
          try { cb(); } catch (err) { console.error('Sync callback error:', err); }
        });
      }
    }
  }

  handleAuthError() {
    this.isSyncing = false;
    this.updateStatusUI();
    if (window.showToast) {
      window.showToast('Session expired. Please log in to sync changes.', 'error');
    }
  }
}

// Global Sync instance
window.syncEngine = new SyncEngine();
