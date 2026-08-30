// IndexedDB Data Layer for Bazar List Offline-First PWA
const DB_NAME = 'bazarlist_db';
const DB_VERSION = 1;

class AppDB {
  constructor() {
    this.db = null;
    this.initPromise = this.init();
  }

  async init() {
    return new Promise((resolve, reject) => {
      const request = indexedDB.open(DB_NAME, DB_VERSION);

      request.onupgradeneeded = (event) => {
        const db = event.target.result;

        // Lists Store
        if (!db.objectStoreNames.contains('lists')) {
          const listStore = db.createObjectStore('lists', { keyPath: 'id' });
          listStore.createIndex('month', 'month', { unique: false });
          listStore.createIndex('user_id', 'user_id', { unique: false });
        }

        // Outbox Store (Operations Queue)
        if (!db.objectStoreNames.contains('outbox')) {
          db.createObjectStore('outbox', { keyPath: 'id', autoIncrement: true });
        }

        // Meta Store
        if (!db.objectStoreNames.contains('meta')) {
          db.createObjectStore('meta', { keyPath: 'key' });
        }
      };

      request.onsuccess = (event) => {
        this.db = event.target.result;
        resolve(this.db);
      };

      request.onerror = (event) => {
        console.error('IndexedDB open error:', event.target.error);
        reject(event.target.error);
      };
    });
  }

  async getDB() {
    if (!this.db) {
      await this.initPromise;
    }
    return this.db;
  }

  // --- Lists Store Methods ---

  async saveLists(lists, month, userId) {
    const db = await this.getDB();
    return new Promise((resolve, reject) => {
      const tx = db.transaction(['lists'], 'readwrite');
      const store = tx.objectStore('lists');

      lists.forEach(list => {
        // Normalize list month attribute
        const listDate = new Date(list.date);
        const listMonth = !isNaN(listDate.getTime()) 
          ? `${listDate.getFullYear()}-${String(listDate.getMonth() + 1).padStart(2, '0')}`
          : month;

        store.put({
          ...list,
          month: listMonth,
          user_id: userId || list.user_id,
          _updated_at: Date.now()
        });
      });

      tx.oncomplete = () => resolve(true);
      tx.onerror = (e) => reject(e.target.error);
    });
  }

  async saveList(list) {
    const db = await this.getDB();
    return new Promise((resolve, reject) => {
      const tx = db.transaction(['lists'], 'readwrite');
      const store = tx.objectStore('lists');
      
      const listDate = new Date(list.date);
      const listMonth = !isNaN(listDate.getTime()) 
        ? `${listDate.getFullYear()}-${String(listDate.getMonth() + 1).padStart(2, '0')}`
        : '';

      store.put({
        ...list,
        month: list.month || listMonth,
        _updated_at: Date.now()
      });

      tx.oncomplete = () => resolve(true);
      tx.onerror = (e) => reject(e.target.error);
    });
  }

  async getList(id) {
    const db = await this.getDB();
    return new Promise((resolve, reject) => {
      const tx = db.transaction(['lists'], 'readonly');
      const store = tx.objectStore('lists');
      const request = store.get(id);

      request.onsuccess = () => resolve(request.result || null);
      request.onerror = (e) => reject(e.target.error);
    });
  }

  async getAllLists(userId) {
    const db = await this.getDB();
    return new Promise((resolve, reject) => {
      const tx = db.transaction(['lists'], 'readonly');
      const store = tx.objectStore('lists');
      const request = store.getAll();

      request.onsuccess = () => {
        let results = request.result || [];
        if (userId) {
          results = results.filter(l => l.user_id === userId);
        }
        // Sort descending by date
        results.sort((a, b) => new Date(b.date) - new Date(a.date));
        resolve(results);
      };
      request.onerror = (e) => reject(e.target.error);
    });
  }

  async getListsByMonth(month, userId) {
    const db = await this.getDB();
    return new Promise((resolve, reject) => {
      const tx = db.transaction(['lists'], 'readonly');
      const store = tx.objectStore('lists');
      const request = store.getAll();

      request.onsuccess = () => {
        let results = request.result || [];
        if (userId) {
          results = results.filter(l => l.user_id === userId);
        }
        if (month) {
          results = results.filter(l => l.month === month || (l.date && l.date.startsWith(month)));
        }
        // Sort descending by date
        results.sort((a, b) => new Date(b.date) - new Date(a.date));
        resolve(results);
      };
      request.onerror = (e) => reject(e.target.error);
    });
  }

  async deleteList(id) {
    const db = await this.getDB();
    return new Promise((resolve, reject) => {
      const tx = db.transaction(['lists'], 'readwrite');
      const store = tx.objectStore('lists');
      store.delete(id);

      tx.oncomplete = () => resolve(true);
      tx.onerror = (e) => reject(e.target.error);
    });
  }

  // --- Outbox Queue Methods ---

  async queueOutbox(action, payload, tempId = null, targetListId = null, targetItemId = null) {
    const db = await this.getDB();
    return new Promise((resolve, reject) => {
      const tx = db.transaction(['outbox'], 'readwrite');
      const store = tx.objectStore('outbox');

      const entry = {
        action,
        temp_id: tempId,
        target_list_id: targetListId,
        target_item_id: targetItemId,
        payload,
        created_at: Date.now(),
        retry_count: 0
      };

      const request = store.add(entry);
      request.onsuccess = () => resolve(request.result);
      request.onerror = (e) => reject(e.target.error);
    });
  }

  async getOutboxItems() {
    const db = await this.getDB();
    return new Promise((resolve, reject) => {
      const tx = db.transaction(['outbox'], 'readonly');
      const store = tx.objectStore('outbox');
      const request = store.getAll();

      request.onsuccess = () => {
        const items = request.result || [];
        items.sort((a, b) => a.id - b.id);
        resolve(items);
      };
      request.onerror = (e) => reject(e.target.error);
    });
  }

  async deleteOutboxItem(id) {
    const db = await this.getDB();
    return new Promise((resolve, reject) => {
      const tx = db.transaction(['outbox'], 'readwrite');
      const store = tx.objectStore('outbox');
      store.delete(id);

      tx.oncomplete = () => resolve(true);
      tx.onerror = (e) => reject(e.target.error);
    });
  }

  async countOutbox() {
    const db = await this.getDB();
    return new Promise((resolve, reject) => {
      const tx = db.transaction(['outbox'], 'readonly');
      const store = tx.objectStore('outbox');
      const request = store.count();

      request.onsuccess = () => resolve(request.result || 0);
      request.onerror = (e) => reject(e.target.error);
    });
  }

  // --- Meta Store Methods ---

  async setMeta(key, value) {
    const db = await this.getDB();
    return new Promise((resolve, reject) => {
      const tx = db.transaction(['meta'], 'readwrite');
      const store = tx.objectStore('meta');
      store.put({ key, value });

      tx.oncomplete = () => resolve(true);
      tx.onerror = (e) => reject(e.target.error);
    });
  }

  async getMeta(key) {
    const db = await this.getDB();
    return new Promise((resolve, reject) => {
      const tx = db.transaction(['meta'], 'readonly');
      const store = tx.objectStore('meta');
      const request = store.get(key);

      request.onsuccess = () => resolve(request.result ? request.result.value : null);
      request.onerror = (e) => reject(e.target.error);
    });
  }
}

// Global DB instance
window.appDB = new AppDB();
