# Quick Start Guide

Get started with your Bazar List application in just a few minutes!

## Step 1: Initialize the Project

```bash
# Download dependencies
go mod download

# Build the application
make build
```

## Step 2: Add Your First Items

```bash
# Add some groceries
./build/bazarlist add "Milk" --category dairy
./build/bazarlist add "Bread" --category bakery
./build/bazarlist add "Apples" --category produce
./build/bazarlist add "Chicken" --category meat
./build/bazarlist add "Rice" --category pantry
```

## Step 3: View Your Shopping List

```bash
# List all pending items
./build/bazarlist list

# List all items (including completed)
./build/bazarlist list --completed
```

## Step 4: Mark Items as Purchased

```bash
# Mark item with ID 1 as purchased
./build/bazarlist complete 1
```

## Step 5: Search for Items

```bash
# Search for specific items
./build/bazarlist search milk
./build/bazarlist search chicken
```

## Step 6: Remove Items

```bash
# Remove item with ID 1
./build/bazarlist remove 1
```

## Data Storage

Your shopping list is automatically saved to `./shopping_list.json` in the current directory.

To use a different data directory:

```bash
export BAZARLIST_DATA_DIR="/path/to/data"
./build/bazarlist list
```

## Enable Debug Logging

```bash
export BAZARLIST_DEBUG=true
./build/bazarlist add "Test item"
```

## Available Categories

- **produce** - Fruits and vegetables
- **dairy** - Milk, cheese, yogurt, etc.
- **meat** - Beef, chicken, pork, etc.
- **pantry** - Pasta, rice, canned goods, etc.
- **frozen** - Frozen foods
- **bakery** - Bread, pastries, etc.
- **beverages** - Drinks, juices, etc.
- **household** - Cleaning supplies, etc.
- **other** - Everything else

## Common Commands Reference

```bash
# Add item
./build/bazarlist add "<name>" --category <category>

# List items
./build/bazarlist list
./build/bazarlist list --category dairy
./build/bazarlist list --completed

# Complete item
./build/bazarlist complete <id>

# Remove item
./build/bazarlist remove <id>

# Search items
./build/bazarlist search <term>

# Help
./build/bazarlist help
```

## Next Steps

1. **Learn the Code**: Explore the source code to understand how it works
2. **Add Tests**: Write tests for the functions (see `internal/` directories)
3. **Extend Features**: Add new features like quantity, price, or notes
4. **REST API**: Start building a web API version

## Troubleshooting

**"command not found: ./build/bazarlist"**
- Make sure you've run `make build` first

**Permission denied**
- Make the binary executable: `chmod +x ./build/bazarlist`

**"Failed to create data directory"**
- Check that you have write permissions in the current directory

**Data not persisting**
- Check that the `shopping_list.json` file is being created
- Enable debug logging to see what's happening

Happy shopping! 🛒
