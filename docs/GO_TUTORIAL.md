# Go Tutorial: Learning Through the Bazar List Project

This tutorial will help you learn Go by exploring the Bazar List codebase.

## Table of Contents

1. [Go Basics](#1-go-basics)
2. [Structs and Methods](#2-structs-and-methods)
3. [Interfaces](#3-interfaces)
4. [Error Handling](#4-error-handling)
5. [Packages and Modules](#5-packages-and-modules)
6. [File I/O and JSON](#6-file-io-and-json)
7. [CLI Development](#7-cli-development)
8. [Testing](#8-testing)

---

## 1. Go Basics

### Variables and Types

```go
// In internal/models/item.go
type Category string  // Custom type

const (
    CategoryProduce Category = "produce"  // Constants
    CategoryDairy   Category = "dairy"
)

// Variables
var (
    nextID int = 1
    items  []*Item
)
```

**Key Concepts:**
- Go is statically typed
- Types can be defined with `type`
- Constants use `const` keyword
- Multiple variables can be declared with `var`

### Functions

```go
// Function with multiple return values
func NewItem(id int, name string, category Category) *Item {
    now := time.Now()
    return &Item{
        ID:        id,
        Name:      name,
        Category:  category,
        CreatedAt: now,
    }
}
```

**Key Concepts:**
- Functions use `func` keyword
- Can return multiple values
- Use `&` to get a pointer to a struct

---

## 2. Structs and Methods

### Defining Structs

```go
type Item struct {
    ID        int       `json:"id"`        // Struct tags
    Name      string    `json:"name"`
    Category  Category  `json:"category"`
    Status    Status    `json:"status"`
    CreatedAt time.Time `json:"created_at"`
}
```

**Key Concepts:**
- Structs group related data
- Struct tags (`json:"id"`) define serialization
- Fields are exported if they start with capital letter

### Methods

```go
// Value receiver (read-only)
func (i *Item) IsCompleted() bool {
    return i.Status == StatusCompleted
}

// Pointer receiver (can modify)
func (i *Item) MarkCompleted() {
    i.Status = StatusCompleted
    i.UpdatedAt = time.Now()
}
```

**Key Concepts:**
- Methods are functions with a receiver
- Pointer receivers can modify the struct
- Value receivers are read-only

---

## 3. Interfaces

### Defining Interfaces

```go
// An interface defines behavior
type Storage interface {
    Save(list *ShoppingList) error
    Load() (*ShoppingList, error)
}
```

**Key Concepts:**
- Interfaces define a set of methods
- Types implement interfaces implicitly
- No explicit "implements" keyword needed

### Using Interfaces

```go
func NewShoppingService(storage Storage) *ShoppingService {
    return &ShoppingService{
        storage: storage,
    }
}
```

**Benefits:**
- Decoupling: Code depends on behavior, not implementation
- Testing: Can use mock implementations
- Flexibility: Easy to swap implementations

---

## 4. Error Handling

### Returning Errors

```go
func AddItem(name string, category Category) (*Item, error) {
    if name == "" {
        return nil, fmt.Errorf("item name cannot be empty")
    }
    // ... rest of function
}
```

### Handling Errors

```go
item, err := svc.AddItem(name, category)
if err != nil {
    fmt.Printf("Error: %v\n", err)
    os.Exit(1)
}
```

### Error Wrapping

```go
if err := os.WriteFile(path, data, 0644); err != nil {
    return fmt.Errorf("failed to write file: %w", err)
}
```

**Key Concepts:**
- Errors are values, not exceptions
- Multiple return values (result, error)
- Use `%w` to wrap errors with context
- Always handle errors explicitly

---

## 5. Packages and Modules

### Importing Packages

```go
import (
    "fmt"              // Standard library
    "os"               // Standard library
    "time"             // Standard library

    "github.com/sejan/bazarlist/internal/models"  // Your code
    "github.com/spf13/cobra"                       // External package
)
```

### Package Declaration

```go
package models  // All files in this directory belong to this package
```

### Go Modules

```go
// go.mod
module github.com/sejan/bazarlist

go 1.21

require (
    github.com/spf13/cobra v1.8.0
)
```

**Key Concepts:**
- One package per directory
- Exported names start with capital letter
- `go.mod` defines module and dependencies
- `go get` to add new dependencies

---

## 6. File I/O and JSON

### Reading JSON

```go
data, err := os.ReadFile(filePath)
if err != nil {
    return nil, fmt.Errorf("failed to read file: %w", err)
}

var list ShoppingList
if err := json.Unmarshal(data, &list); err != nil {
    return nil, fmt.Errorf("failed to unmarshal: %w", err)
}
```

### Writing JSON

```go
data, err := json.MarshalIndent(list, "", "  ")
if err != nil {
    return fmt.Errorf("failed to marshal: %w", err)
}

if err := os.WriteFile(filePath, data, 0644); err != nil {
    return fmt.Errorf("failed to write: %w", err)
}
```

**Key Concepts:**
- `os.ReadFile()` / `os.WriteFile()` for file operations
- `json.Unmarshal()` for JSON → Go
- `json.Marshal()` for Go → JSON
- `json.MarshalIndent()` for pretty formatting

---

## 7. CLI Development

### Parsing Flags

```go
fs := flag.NewFlagSet("add", flag.ExitOnError)
categoryFlag := fs.String("category", "other", "Category of the item")

if err := fs.Parse(os.Args[2:]); err != nil {
    fmt.Printf("Error parsing flags: %v\n", err)
    os.Exit(1)
}
```

### Command Handling

```go
switch command {
case "add":
    handlers.HandleAdd()
case "list":
    handlers.HandleList()
default:
    fmt.Printf("Unknown command: %s\n", command)
}
```

**Key Concepts:**
- `flag` package for CLI flags
- `os.Args` to access command-line arguments
- Switch statements for command routing

---

## 8. Testing

### Writing Tests

```go
func TestNewItem(t *testing.T) {
    item := NewItem(1, "Milk", CategoryDairy)

    if item.ID != 1 {
        t.Errorf("Expected ID 1, got %d", item.ID)
    }

    if item.Name != "Milk" {
        t.Errorf("Expected name 'Milk', got %s", item.Name)
    }
}
```

### Table-Driven Tests

```go
func TestValidateCategory(t *testing.T) {
    tests := []struct {
        name     string
        category Category
        wantErr  bool
    }{
        {"valid dairy", CategoryDairy, false},
        {"invalid category", "invalid", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validator.ValidateCategory(tt.category)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateCategory() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

**Key Concepts:**
- Tests go in `*_test.go` files
- Use `go test` to run tests
- Table-driven tests for multiple cases
- `t.Errorf()` to report failures

---

## Practice Exercises

1. **Add a field to Item**: Add a `Quantity` field and update the code
2. **Create a new command**: Add a `bazarlist clear` command to remove all completed items
3. **Add validation**: Ensure item names don't contain special characters
4. **Write tests**: Create tests for the `ShoppingList` methods
5. **Add filtering**: Add a flag to filter items by multiple categories

---

## Recommended Resources

- [Official Go Tour](https://tour.golang.org/)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go by Example](https://gobyexample.com/)
- [A Tour of Go](https://tour.golang.org/welcome/1)

---

Happy Learning! 🎓
