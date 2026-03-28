# Bazar List - Personal Shopping List Manager

A modern Golang web application for managing your home bazar (shopping) list. Built to learn Golang fundamentals and web development.

## 🎉 Now Available as a Web Application!

Bazar List has been upgraded from a CLI tool to a full-featured web application with a beautiful UI!

## Features

### Web Application ⭐ NEW
- 🎨 Beautiful, modern UI with responsive design
- 🌐 RESTful API
- 📱 Mobile-friendly interface
- 📊 Real-time statistics dashboard
- 🔍 Instant search and filtering
- ✨ Smooth animations and transitions

### Core Features
- ✅ Add items to your shopping list
- ✅ Remove items from the list
- ✅ Mark items as purchased
- ✅ Categorize items (produce, dairy, meat, pantry, etc.)
- ✅ Search and filter items
- 💾 Persistent storage using JSON
- 🖥️ CLI interface (still available!)
- 🌐 Web interface (NEW!)

## Project Structure

```
bazarlist/
├── cmd/                 # Application entry points
│   └── cli/            # CLI application
├── internal/           # Private application code
│   ├── handlers/       # CLI handlers
│   ├── models/         # Data structures
│   ├── service/        # Business logic
│   ├── storage/        # Data persistence
│   └── utils/          # Utility functions
├── pkg/                # Public libraries
│   ├── logger/         # Logging utilities
│   └── validator/      # Validation utilities
├── api/                # API definitions (for future REST API)
├── web/                # Web assets (for future web UI)
├── scripts/            # Build and deployment scripts
├── test/               # Additional test files
├── docs/               # Documentation
├── build/              # Build output
├── configs/            # Configuration files
└── init/               # Initialization files
```

## 🚀 Getting Started

### Prerequisites

- Go 1.21 or higher
- Git

### Installation

1. Clone the repository:
```bash
git clone https://github.com/sejan/bazarlist.git
cd bazarlist
```

2. Install dependencies:
```bash
go mod download
```

3. Build the application:
```bash
make build
```

## 🌐 Web Application (Recommended)

### Running the Web App

```bash
# Run directly
make run-web

# Or build and run
make build-web
./build/bazarlist-web
```

### Access the Web App

Open your browser and navigate to:
```
http://localhost:8080
```

### Using the Web App

1. **Add Items**: Enter item name and select category
2. **View Items**: See all items in your shopping list
3. **Complete Items**: Click the checkbox to mark as purchased
4. **Delete Items**: Click the delete button to remove
5. **Search**: Use the search bar to find items
6. **Filter**: Filter by All, Pending, or Completed

### Custom Configuration

```bash
# Set custom port
export PORT=3000
make run-web

# Set custom data directory
export BAZARLIST_DATA_DIR=/path/to/data
make run-web
```

## 🖥️ CLI Application (Still Available!)

The CLI version is still available for those who prefer command-line tools.

### Running the CLI App

```bash
# Run directly
make run

# Or build and run
make build-cli
./build/bazarlist
```

### Add an item
```bash
./build/bazarlist add "Milk" --category dairy
```

### List all items
```bash
./build/bazarlist list
```

### Mark item as purchased
```bash
./build/bazarlist complete 1
```

### Remove an item
```bash
./build/bazarlist remove 1
```

## 📡 REST API

The web application includes a full REST API for integration with other tools.

### Quick API Examples

```bash
# Add item via API
curl -X POST http://localhost:8080/api/items \
  -H "Content-Type: application/json" \
  -d '{"name": "Milk", "category": "dairy"}'

# Get all items
curl http://localhost:8080/api/items

# Complete an item
curl -X POST http://localhost:8080/api/items/1/complete

# Search items
curl http://localhost:8080/api/search?q=milk

# Get statistics
cur📚 Documentation

- **[PROJECT_OVERVIEW.md](PROJECT_OVERVIEW.md)** - Complete project overview
- **[QUICKSTART.md](QUICKSTART.md)** - Quick start guide
- **[docs/WEB_APPLICATION.md](docs/WEB_APPLICATION.md)** - Web application guide
- **[docs/API_REFERENCE.md](docs/API_REFERENCE.md)** - Complete API documentation
- **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)** - Architecture details
- **[docs/GO_TUTORIAL.md](docs/GO_TUTORIAL.md)** - Go learning guide
- **[docs/LEARNING_ROADMAP.md](docs/LEARNING_ROADMAP.md)** - 5-week learning path

## 🎓 Learning Path

This project is designed to help you learn Golang progressively:

1. **Phase 1**: Fundamentals (Complete) ✅
   - Basic Go syntax
   - Structs and methods
   - File I/O and JSON
   - CLI flag handling
   - Error handling

2. **Phase 2**: Testing (In Progress) 🔄
   - Go testing framework
   - Table-driven tests
   - Mocking

3. **Phase 3**: Web Development (New!) 🌐
   - HTTP servers
   - REST APIs
   - Middleware
   - Frontend integration
```

### Lint code
```bash
make lint
```

## Learning Path

This project is designed to help you learn Golang progressively:

1. **Phase 1**: CLI Application (Current)
   - Basic Go syntax
   - Structs and methods
   - File I/O and JSON
   - CLI flag handling
   - Error handling

2. **Phase 2**: Testing
   - Go testing framework
   - Table-driven tests
   - Mocking
🔄 Data Synchronization

Both the CLI and Web applications share the same JSON storage file. This means you can:

- Add items via the web UI
- View and manage them using the CLI
- Complete items via CLI
- See changes reflected in the web UI instantly

Perfect for using on desktop (web) and mobile (CLI)!

## 🌟 Features Comparison

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
| Scriptable | ✅ | ❌ |
| API access | ❌ | ✅ |

## 🤝 Contributing

This is a personal learning project, but feel free to fork and modify for your own learning!

## 📝 - Middleware

4. **Phase 4**: Database (Future)
   - Database connections
   - SQL vs NoSQL
   - ORM/GORM

## Contributing

This is a personal learning project, but feel free to fork and modify for your own learning!

## License

MIT License - feel free to use for learning purposes.
