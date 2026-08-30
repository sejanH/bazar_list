// Live Client for Real-Time List Synchronization via SSE in Bazar List PWA

class LiveClient {
  constructor() {
    this.eventSource = null;
    this.currentListId = null;
    this.reconnectTimer = null;
    this.retryDelay = 2000;
    this.maxRetryDelay = 15000;
    this.listeners = new Map();
  }

  connect(listId) {
    if (!listId) return;

    // If already connected to this list with an active EventSource, do nothing
    if (this.currentListId === listId && this.eventSource && this.eventSource.readyState !== EventSource.CLOSED) {
      return;
    }

    this.disconnect(false);
    this.currentListId = listId;

    const token = localStorage.getItem('token');
    if (!token) {
      console.warn('[LiveSync] No auth token found, skipping live connection');
      return;
    }

    const streamUrl = `/api/lists/${encodeURIComponent(listId)}/live?token=${encodeURIComponent(token)}`;

    try {
      this.eventSource = new EventSource(streamUrl);

      this.eventSource.addEventListener('open', () => {
        this.retryDelay = 2000; // Reset backoff upon successful connection
      });

      this.eventSource.addEventListener('connected', (e) => {
        this.retryDelay = 2000;
        try {
          const data = JSON.parse(e.data);
          this.emit('connected', data);
        } catch {
          this.emit('connected', e.data);
        }
      });

      this.eventSource.addEventListener('item_created', (e) => {
        try {
          const item = JSON.parse(e.data);
          this.emit('item_created', item);
        } catch (err) {
          console.error('[LiveSync] Failed parsing item_created event:', err);
        }
      });

      this.eventSource.addEventListener('item_updated', (e) => {
        try {
          const item = JSON.parse(e.data);
          this.emit('item_updated', item);
        } catch (err) {
          console.error('[LiveSync] Failed parsing item_updated event:', err);
        }
      });

      this.eventSource.addEventListener('item_deleted', (e) => {
        try {
          const data = JSON.parse(e.data);
          this.emit('item_deleted', data);
        } catch (err) {
          console.error('[LiveSync] Failed parsing item_deleted event:', err);
        }
      });

      this.eventSource.addEventListener('activity_logged', (e) => {
        try {
          const activity = JSON.parse(e.data);
          this.emit('activity_logged', activity);
        } catch (err) {
          console.error('[LiveSync] Failed parsing activity_logged event:', err);
        }
      });

      this.eventSource.addEventListener('member_updated', (e) => {
        try {
          const data = JSON.parse(e.data);
          this.emit('member_updated', data);
        } catch (err) {
          console.error('[LiveSync] Failed parsing member_updated event:', err);
        }
      });

      this.eventSource.onerror = (err) => {
        console.warn(`[LiveSync] Connection lost for list #${listId}, retrying in ${this.retryDelay / 1000}s...`, err);
        if (this.eventSource) {
          this.eventSource.close();
          this.eventSource = null;
        }

        const delay = this.retryDelay;
        this.retryDelay = Math.min(this.retryDelay * 2, this.maxRetryDelay);

        clearTimeout(this.reconnectTimer);
        this.reconnectTimer = setTimeout(() => {
          if (this.currentListId === listId) {
            this.connect(listId);
          }
        }, delay);
      };
    } catch (err) {
      console.error('[LiveSync] Could not instantiate EventSource:', err);
    }
  }

  disconnect(resetList = true) {
    if (this.eventSource) {
      this.eventSource.close();
      this.eventSource = null;
    }
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    this.retryDelay = 2000;
    if (resetList) {
      this.currentListId = null;
    }
  }

  on(event, callback) {
    if (typeof callback !== 'function') return;
    if (!this.listeners.has(event)) {
      this.listeners.set(event, []);
    }
    this.listeners.get(event).push(callback);
  }

  off(event, callback) {
    if (!this.listeners.has(event)) return;
    if (!callback) {
      this.listeners.delete(event);
      return;
    }
    const filtered = this.listeners.get(event).filter(cb => cb !== callback);
    this.listeners.set(event, filtered);
  }

  emit(event, data) {
    const callbacks = this.listeners.get(event) || [];
    callbacks.forEach(cb => {
      try {
        cb(data);
      } catch (e) {
        console.error(`[LiveSync] Error in listener for event "${event}":`, e);
      }
    });
  }
}

// Attach global singleton instance
window.liveClient = new LiveClient();
