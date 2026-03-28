# Bazar List Web Application

## 🌐 Overview

The Bazar List has been converted from a CLI application to a modern web application with a beautiful UI and REST API.

## 🏗️ Architecture

### New Components

```
┌─────────────────────────────────────────┐
│  Web Browser (Frontend)                │
│  - HTML/CSS/JavaScript                  │
│  - Beautiful UI                        │
└─────────────────────────────────────────┘
                 ↓
┌─────────────────────────────────────────┐
│  REST API (Backend)                    │
│  - HTTP/JSON endpoints                 │
│  - Gorilla Mux router                   │
└─────────────────────────────────────────┘
                 ↓
┌─────────────────────────────────────────┐
│  Service Layer                         │
│  - Business Logic                      │
│  - Same as CLI version                 │
└─────────────────────────────────────────┘
                 ↓
┌─────────────────────────────────────────┐
│  Storage Layer                         │
│  - JSON File Storage                   │
└─────────────────────────────────────────┘
```

## 📁 New Project Structure

```
bazarlist/
├── cmd/
│   ├── cli/main.go              # CLI application (still available!)
│   └── web/main.go              # Web application entry point ⭐ NEW
├── internal/
│   ├── api/                     # HTTP API handlers ⭐ NEW
│   │   └── handlers.go
│   ├── handlers/               # CLI handlers (original)
│   ├── models/                 # Data models (shared)
│   ├── service/                # Business logic (shared)
│   └── storage/                # Storage layer (shared)
├── web/                        # Frontend assets ⭐ NEW
│   ├── static/
│   │   └── index.html          # Single-page web UI
│   └── templates/              # Future: HTML templates
└── ... (rest of the structure)
```

## 🚀 Getting Started

### Prerequisites

1. **Go 1.21 or higher**
   ```bash
   # Check if Go is installed
   go version

   # If not installed, download from: https://golang.org/dl/
   ```

2. **Download Dependencies**
   ```bash
   go mod download
   ```

### Running the Web Application

#### Option 1: Run directly
```bash
make run-web
```

#### Option 2: Build and run
```bash
make build-web
./build/bazarlist-web
```

#### Option 3: Custom port
```bash
export PORT=3000
make run-web
```

### Accessing the Application

Open your browser and navigate to:
```
http://localhost:8080
```

## 📡 REST API Endpoints

### Items

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/items` | Get all items |
| POST | `/api/items` | Add new item |
| GET | `/api/items/{id}` | Get specific item |
| PUT/PATCH | `/api/items/{id}` | Update item |
| DELETE | `/api/items/{id}` | Delete item |
| POST | `/api/items/{id}/complete` | Mark as completed |
| POST | `/api/items/{id}/pending` | Mark as pending |

### Search & Filter

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/search?q=query` | Search items |
| GET | `/api/items/category/{category}` | Filter by category |

### Stats

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/stats` | Get statistics |

### Request/Response Examples

#### Add Item
```bash
curl -X POST http://localhost:8080/api/items \
  -H "Content-Type: application/json" \
  -d '{"name": "Milk", "category": "dairy"}'
```

**Response:**
```json
{
  "id": 1,
  "name": "Milk",
  "category": "dairy",
  "status": "pending",
  "created_at": "2024-03-28T10:00:00Z",
  "updated_at": "2024-03-28T10:00:00Z"
}
```

#### Get All Items
```bash
curl http://localhost:8080/api/items
```

**Response:**
```json
[
  {
    "id": 1,
    "name": "Milk",
    "category": "dairy",
    "status": "pending",
    "created_at": "2024-03-28T10:00:00Z",
    "updated_at": "2024-03-28T10:00:00Z"
  }
]
```

#### Complete Item
```bash
curl -X POST http://localhost:8080/api/items/1/complete
```

#### Delete Item
```bash
curl -X DELETE http://localhost:8080/api/items/1
```

#### Search Items
```bash
curl http://localhost:8080/api/search?q=milk
```

#### Get Stats
```bash
curl http://localhost:8080/api/stats
```

**Response:**
```json
{
  "total": 10,
  "pending": 7,
  "completed": 3
}
```

## 🎨 Web UI Features

### User Interface

- **Beautiful Design**: Modern gradient background with glassmorphism effects
- **Responsive**: Works on desktop, tablet, and mobile devices
- **Real-time Updates**: Changes reflect immediately
- **Statistics Dashboard**: Shows total, pending, and completed items

### Core Features

1. **Add Items**
   - Enter item name
   - Select from 9 categories
   - Instant feedback

2. **View Items**
   - List all items
   - Filter by status (All, Pending, Completed)
   - Search functionality

3. **Manage Items**
   - Mark as completed with checkbox
   - Delete items
   - Toggle between pending/completed

4. **Categories**
   - 🥬 Produce
   - 🥛 Dairy
   - 🥩 Meat
   - 🥫 Pantry
   - ❄️ Frozen
   - 🍞 Bakery
   - 🥤 Beverages
   - 🏠 Household
   - 📦 Other

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `BAZARLIST_DATA_DIR` | Data directory | `./data` |
| `BAZARLIST_DEBUG` | Enable debug logging | `false` |

### Example Configuration

```bash
# Set custom port
export PORT=3000

# Set custom data directory
export BAZARLIST_DATA_DIR=/path/to/data

# Enable debug logging
export BAZARLIST_DEBUG=true

# Run server
make run-web
```

## 🧪 Testing the API

### Using curl

```bash
# Add item
curl -X POST http://localhost:8080/api/items \
  -H "Content-Type: application/json" \
  -d '{"name": "Apples", "category": "produce"}'

# List items
curl http://localhost:8080/api/items

# Complete item
curl -X POST http://localhost:8080/api/items/1/complete

# Delete item
curl -X DELETE http://localhost:8080/api/items/1

# Search items
curl http://localhost:8080/api/search?q=apple

# Get stats
curl http://localhost:8080/api/stats
```

### Using Postman/Insomnia

1. Import the API endpoints
2. Set base URL: `http://localhost:8080`
3. Use the examples above for request bodies

## 📊 Comparison: CLI vs Web

| Feature | CLI | Web |
|---------|-----|-----|
| Add items | ✅ | ✅ |
| List items | ✅ | ✅ |
| Complete items | ✅ | ✅ |
| Delete items | ✅ | ✅ |
| Search items | ✅ | ✅ |
| Filter by category | ✅ | ✅ |
| Visual UI | ❌ | ✅ |
| Statistics | ❌ | ✅ |
| Real-time updates | ❌ | ✅ |
| Mobile-friendly | ❌ | ✅ |
| Batch operations | ❌ | ⏳ |

## 🔒 Security Features

- **CORS Enabled**: Allows cross-origin requests
- **Input Validation**: All inputs are validated
- **XSS Prevention**: HTML escaping in frontend
- **Safe File Paths**: Sanitized file operations

## 🚀 Deployment

### Local Development

```bash
# Run locally
make run-web
```

### Building for Production

```bash
# Build optimized binary
go build -ldflags="-s -w" -o bazarlist-web cmd/web/main.go

# Run
./bazarlist-web
```

### Docker (Future)

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o bazarlist-web cmd/web/main.go

FROM alpine:latest
COPY --from=builder /app/bazarlist-web /usr/local/bin/
EXPOSE 8080
CMD ["bazarlist-web"]
```

## 📱 Mobile Access

The web application is fully responsive and works great on mobile devices. Simply access it from your phone's browser:

```
http://your-server-ip:8080
```

## 🔄 Data Sync

Both CLI and Web applications use the same JSON storage file, so you can:

1. Add items via the web UI
2. View them using the CLI
3. Complete items via CLI
4. See changes in the web UI

They share the same data source!

## 🎓 Learning Opportunities

By working with this web application, you'll learn:

### Backend
- HTTP server development in Go
- REST API design
- JSON handling
- Routing with Gorilla Mux
- Middleware (CORS, logging)

### Frontend
- Modern HTML/CSS
- JavaScript fetch API
- DOM manipulation
- Responsive design
- User experience design

### Full Stack
- Client-server communication
- API integration
- State management
- Error handling

## 🐛 Troubleshooting

### Port Already in Use

```bash
# Use a different port
export PORT=3000
make run-web
```

### Cannot Access from Other Devices

```bash
# Bind to all interfaces (0.0.0.0)
# Modify cmd/web/main.go:
server.Addr = ":8080"  # Already binds to all interfaces
```

### Data Not Persisting

- Check data directory permissions
- Verify `BAZARLIST_DATA_DIR` is set correctly
- Check logs for errors

## 📚 Next Steps

### Phase 2: Enhancements
- [ ] User authentication
- [ ] Multiple shopping lists
- [ ] Item quantity and units
- [ ] Price tracking
- [ ] Budget management

### Phase 3: Advanced Features
- [ ] Real-time updates (WebSockets)
- [ ] Push notifications
- [ ] Export to PDF
- [ ] Import/Export lists
- [ ] Recipe integration

### Phase 4: Production Ready
- [ ] Database migration (PostgreSQL)
- [ ] Caching (Redis)
- [ ] Rate limiting
- [ ] Monitoring and logging
- [ ] CI/CD pipeline

## 🤝 Contributing

Feel free to:
- Add new features
- Improve the UI
- Fix bugs
- Write tests
- Improve documentation

## 📝 License

MIT License - Feel free to use for learning purposes.

---

**Happy Shopping! 🛒**
