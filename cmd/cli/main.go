package main

import (
	"fmt"
	"os"

	"github.com/sejan/bazarlist/internal/handlers"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "add":
		handlers.HandleAdd()
	case "list", "ls":
		handlers.HandleList()
	case "complete", "done":
		handlers.HandleComplete()
	case "remove", "rm", "delete":
		handlers.HandleRemove()
	case "search":
		handlers.HandleSearch()
	case "help", "-h", "--help":
		printHelp()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	helpText := `
Bazar List - Personal Shopping List Manager

USAGE:
    bazarlist <command> [options]

COMMANDS:
    add <item> [--category <category>]    Add a new item to the list
    list, ls                             List all items
    complete <id>                        Mark an item as purchased
    remove, rm, delete <id>              Remove an item from the list
    search <term>                        Search for items
    help, -h, --help                     Show this help message

EXAMPLES:
    bazarlist add "Milk" --category dairy
    bazarlist add "Apples" --category produce
    bazarlist list
    bazarlist complete 1
    bazarlist remove 1
    bazarlist search milk

CATEGORIES:
    produce, dairy, meat, pantry, frozen, bakery, beverages, household, other

For more information, visit: https://github.com/sejan/bazarlist
`
	fmt.Println(helpText)
}
