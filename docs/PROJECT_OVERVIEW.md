# Bazar List - Project Overview

## 📋 What is Bazar List?

A personal shopping list manager built with Golang, designed for learning Go programming concepts through a practical, real-world application.

## 🎯 Project Goals

1. **Learn Golang Fundamentals**: Master Go syntax, types, and idioms
2. **Understand Go Project Structure**: Learn standard Go project organization
3. **Practice Clean Architecture**: Implement layered architecture patterns
4. **Build Something Useful**: Create a practical application for personal use

## 🏗️ Architecture Overview

### Layered Architecture

```
┌─────────────────────────────────────────┐
│  Presentation Layer (CLI)               │
│  - cmd/cli/main.go                      │
│  - internal/handlers/                   │
└─────────────────────────────────────────┘
                 ↓
┌─────────────────────────────────────────┐
│  Business Logic Layer                   │
│  - internal/service/                    │
│  - ShoppingService                      │
└─────────────────────────────────────────┘
                 ↓
┌─────────────────────────────────────────┐
│  Data Access Layer                      │
│  - internal/storage/                    │
│  - JSONStorage                          │
└─────────────────────────────────────────┘
                 ↓
┌─────────────────────────────────────────┐
│  Data Layer                             │
│  - JSON files                           │
│  - Future: Database                     │
└─────────────────────────────────────────┘
```

### Key Components

| Component | Purpose | Location |
|-----------|---------|----------|
| **Models** | Data structures and business entities | `internal/models/` |
| **Handlers** | Command-line interface handlers | `internal/handlers/` |
| **Service** | Business logic layer | `internal/service/` |
| **Storage** | Data persistence layer | `internal/storage/` |
| **Logger** | Logging utilities | `pkg/logger/` |
| **Validator** | Input validation | `pkg/validator/` |

## 🚀 Quick Start

### 1. Build the Application
```bash
make build
```

### 2. Add Items
```bash
./build/bazarlist add "Milk" --category dairy
./build/bazarlist add "Bread" --category bakery
```

### 3. List Items
```bash
./build/bazarlist list
```

### 4. Mark as Complete
```bash
./build/bazarlist complete 1
```

## 📁 Project Structure

```
bazarlist/
├── cmd/                    # Application entry points
│   └── cli/               # CLI application
│       └── main.go        # Main entry point
├── internal/              # Private application code
│   ├── handlers/         # CLI command handlers
│   ├── models/           # Data structures
│   ├── service/          # Business logic
│   ├── storage/          # Data persistence
│   └── utils/            # Utilities
├── pkg/                   # Public libraries
│   ├── logger/           # Logging
│   └── validator/        # Validation
├── docs/                  # Documentation
│   ├── ARCHITECTURE.md   # Architecture details
│   ├── GO_TUTORIAL.md    # Go learning guide
│   └── LEARNING_ROADMAP.md # Learning path
├── scripts/              # Build scripts
├── test/                 # Integration tests
├── README.md             # Project documentation
├── QUICKSTART.md         # Quick start guide
├── go.mod               # Go module definition
├── Makefile             # Build automation
└── .gitignore           # Git ignore rules
```

## 🎓 Learning Path

### Phase 1: Go Fundamentals (Current)
- ✅ Project structure setup
- ✅ Basic CLI application
- ✅ Data models and structs
- ✅ File I/O and JSON
- ✅ Error handling

### Phase 2: Testing
- ⏳ Unit tests
- ⏳ Integration tests
- ⏳ Code coverage

### Phase 3: REST API (Future)
- ⏳ HTTP server
- ⏳ JSON API
- ⏳ Middleware

### Phase 4: Database (Future)
- ⏳ PostgreSQL integration
- ⏳ GORM ORM
- ⏳ Migrations

## 🛠️ Available Commands

| Command | Description | Example |
|---------|-------------|---------|
| `add` | Add item to list | `bazarlist add "Milk" --category dairy` |
| `list` | List all items | `bazarlist list` |
| `complete` | Mark item as purchased | `bazarlist complete 1` |
| `remove` | Remove item from list | `bazarlist remove 1` |
| `search` | Search for items | `bazarlist search milk` |
| `help` | Show help | `bazarlist help` |

## 📚 Documentation

- **[README.md](README.md)** - Complete project documentation
- **[QUICKSTART.md](QUICKSTART.md)** - Get started in 5 minutes
- **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)** - Architecture details
- **[docs/GO_TUTORIAL.md](docs/GO_TUTORIAL.md)** - Go learning guide
- **[docs/LEARNING_ROADMAP.md](docs/LEARNING_ROADMAP.md)** - 5-week learning path

## 🧪 Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage
```

## 🎯 Features

### Current Features
- ✅ Add items to shopping list
- ✅ Remove items from list
- ✅ Mark items as purchased
- ✅ List all items (with filtering)
- ✅ Categorize items
- ✅ Search items
- ✅ JSON file persistence
- ✅ CLI interface
- ✅ Error handling
- ✅ Logging

### Planned Features
- ⏳ Item quantities and units
- ⏳ Price tracking and budgeting
- ⏳ Multiple shopping lists
- ⏳ Item notes and descriptions
- ⏳ Priority levels
- ⏳ REST API
- ⏳ Web UI
- ⏳ Mobile app
- ⏳ Share lists with family
- ⏳ Recipe integration

## 🔧 Development

### Build Commands
```bash
make build        # Build the application
make clean        # Clean build artifacts
make run          # Run the application
make test         # Run tests
make fmt          # Format code
make lint         # Run linter
make deps         # Download dependencies
```

### Environment Variables
- `BAZARLIST_DATA_DIR` - Directory for data files (default: current directory)
- `BAZARLIST_DEBUG` - Enable debug logging (default: false)

## 📖 Go Concepts You'll Learn

1. **Basics**: Variables, types, functions, methods
2. **Structs**: Data structures, methods, receivers
3. **Interfaces**: Abstraction, polymorphism
4. **Error Handling**: Idiomatic Go error handling
5. **File I/O**: Reading/writing files, JSON
6. **CLI Development**: Flag parsing, command handling
7. **Testing**: Unit tests, table-driven tests
8. **Modules**: Go modules, dependency management
9. **Concurrency**: Goroutines, channels (future)
10. **Web Development**: HTTP servers, APIs (future)

## 🎨 Design Patterns Used

1. **Service Layer Pattern** - Business logic separation
2. **Repository Pattern** - Data access abstraction
3. **Factory Pattern** - Object creation
4. **Dependency Injection** - Loose coupling
5. **Singleton Pattern** - Service instance

## 🔒 Security Considerations

- Input validation and sanitization
- Safe file path handling
- No hardcoded credentials
- Error message sanitization (production)

## 📊 Project Statistics

- **Lines of Code**: ~1000+
- **Packages**: 8
- **Files**: 20+
- **Tests**: 10+
- **Categories**: 9
- **Commands**: 6

## 🤝 Contributing

This is a personal learning project. Feel free to:
- Fork and modify for your own learning
- Study the code structure
- Extend with new features
- Share your learnings

## 📝 License

MIT License - Feel free to use for learning purposes.

## 🙏 Acknowledgments

- Go team for the amazing language
- Go community for excellent documentation
- All Go tutorials and resources available online

## 📞 Support

For issues or questions:
1. Check the documentation in `docs/`
2. Review the code comments
3. Experiment with the code
4. Learn by doing!

---

**Happy Coding! 🚀**

Remember: The best way to learn Go is to write Go code. Start small, experiment, and don't be afraid to make mistakes!
