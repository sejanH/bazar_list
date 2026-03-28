# Architecture Documentation

## Project Architecture

This project follows a clean architecture pattern with clear separation of concerns.

## Directory Structure

```
bazarlist/
├── cmd/                    # Application entry points
│   └── cli/               # CLI application (main.go)
├── internal/              # Private application code
│   ├── models/           # Data structures and domain models
│   ├── storage/          # Data persistence layer
│   ├── service/          # Business logic layer
│   ├── handlers/         # CLI command handlers
│   └── utils/            # Utility functions
├── pkg/                   # Public libraries (reusable)
│   ├── logger/           # Logging utilities
│   └── validator/        # Validation utilities
├── api/                   # API definitions (OpenAPI/Swagger)
├── web/                   # Web assets (for future web UI)
├── scripts/              # Build and deployment scripts
├── test/                 # Integration and E2E tests
├── docs/                 # Documentation
├── build/                # Build output
├── configs/              # Configuration files
└── init/                 # Initialization scripts
```

## Architecture Layers

### 1. Presentation Layer (cmd/, internal/handlers/)
- **Purpose**: Handle user input and display output
- **Components**:
  - `cmd/cli/main.go`: Application entry point
  - `internal/handlers/handlers.go`: Command handlers

### 2. Business Logic Layer (internal/service/)
- **Purpose**: Contain core business rules and operations
- **Components**:
  - `ShoppingService`: Main service for shopping list operations
  - Operations: Add, Remove, Complete, List, Search

### 3. Data Access Layer (internal/storage/)
- **Purpose**: Handle data persistence
- **Components**:
  - `JSONStorage`: JSON file-based storage implementation
  - Future: Database storage implementations

### 4. Domain Layer (internal/models/)
- **Purpose**: Define domain entities and business rules
- **Components**:
  - `Item`: Shopping list item entity
  - `ShoppingList`: Collection of items
  - `Category`, `Status`: Type definitions

### 5. Infrastructure Layer (pkg/)
- **Purpose**: Provide reusable utilities and infrastructure
- **Components**:
  - `logger`: Logging functionality
  - `validator`: Input validation

## Data Flow

```
User Input (CLI)
    ↓
Handlers (Parse & Validate)
    ↓
Service (Business Logic)
    ↓
Storage (Persistence)
    ↓
File System (JSON)
```

## Key Design Patterns

### 1. Dependency Injection
- Services receive storage as a parameter
- Makes testing easier and code more flexible

### 2. Repository Pattern
- Storage layer abstracts data access
- Easy to swap implementations (JSON → Database)

### 3. Service Layer Pattern
- Business logic separated from presentation
- Reusable across different interfaces (CLI, API, Web)

### 4. Factory Pattern
- `NewShoppingList()`, `NewItem()`, `NewJSONStorage()`
- Consistent object creation

## Error Handling Strategy

1. **Always return errors**: Functions return errors explicitly
2. **Wrap errors**: Use `fmt.Errorf` with `%w` for error wrapping
3. **Handle at boundaries**: Handle errors at the CLI/handler layer
4. **Log errors**: Use logger for debugging

## Go Best Practices Applied

1. **Package organization**:
   - `cmd/` for executables
   - `internal/` for private code
   - `pkg/` for public libraries

2. **Naming conventions**:
   - Exported names start with capital letter
   - Private names start with lowercase
   - Interface names end with -er

3. **Error handling**:
   - Explicit error returns
   - Error wrapping for context
   - Early returns for error cases

4. **Testing**:
   - Table-driven tests
   - Test files named `*_test.go`
   - Mock implementations for dependencies

## Future Extensions

### Phase 2: Testing
- Unit tests for all layers
- Integration tests
- Mock storage for testing

### Phase 3: REST API
- HTTP server using `net/http`
- JSON API endpoints
- API versioning

### Phase 4: Database
- SQL database (PostgreSQL/SQLite)
- GORM ORM
- Database migrations

### Phase 5: Web UI
- Frontend framework (React/Vue)
- Real-time updates (WebSockets)
- Mobile app support

## Performance Considerations

1. **File I/O**: Minimize reads/writes, load once per session
2. **Memory**: Use pointers for large structs
3. **Caching**: Cache frequently accessed data
4. **Concurrency**: Use goroutines for async operations (future)

## Security Considerations

1. **Input Validation**: Validate all user input
2. **Path Traversal**: Sanitize file paths
3. **Injection**: Use parameterized queries (when using DB)
4. **Secrets**: Never hardcode credentials

## Learning Outcomes

By studying this project, you'll learn:

1. **Go basics**: Syntax, types, functions, methods
2. **Structs & interfaces**: Data modeling and abstractions
3. **Error handling**: Idiomatic Go error handling
4. **File I/O**: Reading and writing JSON files
5. **CLI development**: Flag parsing and command handling
6. **Testing**: Go testing framework
7. **Project structure**: Standard Go project layout
8. **Clean architecture**: Layered architecture patterns
