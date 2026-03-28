package handlers

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sejan/bazarlist/internal/models"
	"github.com/sejan/bazarlist/internal/service"
)

var (
	serviceInstance *service.ShoppingService
	initialized     bool
)

// getService gets or creates the service instance
func getService() (*service.ShoppingService, error) {
	if initialized {
		return serviceInstance, nil
	}

	// Determine data directory
	dataDir := os.Getenv("BAZARLIST_DATA_DIR")
	if dataDir == "" {
		// Use current directory
		dataDir = "."
	}

	svc, err := service.NewShoppingService(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize service: %w", err)
	}

	serviceInstance = svc
	initialized = true
	return serviceInstance, nil
}

// HandleAdd handles the add command
func HandleAdd() {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	categoryFlag := fs.String("category", "other", "Category of the item")

	if err := fs.Parse(os.Args[2:]); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	args := fs.Args()
	if len(args) < 1 {
		fmt.Println("Error: item name is required")
		fmt.Println("Usage: bazarlist add \"<item name>\" [--category <category>]")
		os.Exit(1)
	}

	name := args[0]
	category := models.Category(*categoryFlag)

	svc, err := getService()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	item, err := svc.AddItem(name, category)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Added: %s (ID: %d, Category: %s)\n", item.Name, item.ID, item.Category)
}

// HandleList handles the list command
func HandleList() {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	categoryFlag := fs.String("category", "", "Filter by category")
	showCompleted := fs.Bool("completed", false, "Show only completed items")

	if err := fs.Parse(os.Args[2:]); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	svc, err := getService()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	var items []*models.Item
	var title string

	switch {
	case *showCompleted:
		items = svc.GetCompletedItems()
		title = "Completed Items"
	case *categoryFlag != "":
		category := models.Category(*categoryFlag)
		items = svc.GetItemsByCategory(category)
		title = fmt.Sprintf("Items in %s", category)
	default:
		items = svc.GetPendingItems()
		title = "Shopping List"
	}

	if len(items) == 0 {
		fmt.Println("No items found.")
		return
	}

	fmt.Printf("\n%s\n", title)
	fmt.Println("=" + "="*len(title))
	printItems(items)
	fmt.Printf("\nTotal: %d item(s)\n", len(items))

	// Show storage location
	storagePath, _ := svc.GetRelativeStoragePath()
	if storagePath != "" {
		relPath, _ := filepath.Rel(".", storagePath)
		fmt.Printf("Data file: %s\n", relPath)
	}
}

// HandleComplete handles the complete command
func HandleComplete() {
	fs := flag.NewFlagSet("complete", flag.ExitOnError)

	if err := fs.Parse(os.Args[2:]); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	args := fs.Args()
	if len(args) < 1 {
		fmt.Println("Error: item ID is required")
		fmt.Println("Usage: bazarlist complete <id>")
		os.Exit(1)
	}

	var id int
	if _, err := fmt.Sscanf(args[0], "%d", &id); err != nil {
		fmt.Printf("Error: invalid item ID: %v\n", err)
		os.Exit(1)
	}

	svc, err := getService()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if err := svc.CompleteItem(id); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Marked item %d as completed\n", id)
}

// HandleRemove handles the remove command
func HandleRemove() {
	fs := flag.NewFlagSet("remove", flag.ExitOnError)

	if err := fs.Parse(os.Args[2:]); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	args := fs.Args()
	if len(args) < 1 {
		fmt.Println("Error: item ID is required")
		fmt.Println("Usage: bazarlist remove <id>")
		os.Exit(1)
	}

	var id int
	if _, err := fmt.Sscanf(args[0], "%d", &id); err != nil {
		fmt.Printf("Error: invalid item ID: %v\n", err)
		os.Exit(1)
	}

	svc, err := getService()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if err := svc.RemoveItem(id); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Removed item %d\n", id)
}

// HandleSearch handles the search command
func HandleSearch() {
	fs := flag.NewFlagSet("search", flag.ExitOnError)

	if err := fs.Parse(os.Args[2:]); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	args := fs.Args()
	if len(args) < 1 {
		fmt.Println("Error: search term is required")
		fmt.Println("Usage: bazarlist search <term>")
		os.Exit(1)
	}

	term := args[0]

	svc, err := getService()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	items := svc.SearchItems(term)

	if len(items) == 0 {
		fmt.Printf("No items found matching '%s'\n", term)
		return
	}

	fmt.Printf("\nSearch Results for '%s'\n", term)
	fmt.Println("=" + "="*(len(term)+16))
	printItems(items)
	fmt.Printf("\nFound: %d item(s)\n", len(items))
}

// printItems prints items in a formatted table
func printItems(items []*models.Item) {
	fmt.Printf("%-5s %-30s %-12s %-10s\n", "ID", "Name", "Category", "Status")
	fmt.Println("---------------------------------------------------")

	for _, item := range items {
		status := "Pending"
		if item.IsCompleted() {
			status = "Done ✓"
		}
		name := item.Name
		if len(name) > 27 {
			name = name[:27] + "..."
		}
		fmt.Printf("%-5d %-30s %-12s %-10s\n", item.ID, name, item.Category, status)
	}
}
