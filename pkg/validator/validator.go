package validator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sejan/bazarlist/internal/models"
)

// Validator provides validation functions
type Validator struct{}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateItemName validates an item name
func (v *Validator) ValidateItemName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("item name cannot be empty")
	}

	if len(name) > 200 {
		return fmt.Errorf("item name too long (max 200 characters)")
	}

	return nil
}

// ValidateCategory validates a category
func (v *Validator) ValidateCategory(category models.Category) error {
	validCategories := map[models.Category]bool{
		models.CategoryProduce:   true,
		models.CategoryDairy:     true,
		models.CategoryMeat:      true,
		models.CategoryPantry:    true,
		models.CategoryFrozen:    true,
		models.CategoryBakery:    true,
		models.CategoryBeverages: true,
		models.CategoryHousehold: true,
		models.CategoryOther:     true,
	}

	if !validCategories[category] {
		return fmt.Errorf("invalid category: %s", category)
	}

	return nil
}

// ParseCategory parses a string into a Category
func (v *Validator) ParseCategory(categoryStr string) (models.Category, error) {
	categoryStr = strings.ToLower(strings.TrimSpace(categoryStr))

	switch categoryStr {
	case "produce":
		return models.CategoryProduce, nil
	case "dairy":
		return models.CategoryDairy, nil
	case "meat":
		return models.CategoryMeat, nil
	case "pantry":
		return models.CategoryPantry, nil
	case "frozen":
		return models.CategoryFrozen, nil
	case "bakery":
		return models.CategoryBakery, nil
	case "beverages", "beverage":
		return models.CategoryBeverages, nil
	case "household":
		return models.CategoryHousehold, nil
	case "other":
		return models.CategoryOther, nil
	default:
		return "", fmt.Errorf("unknown category: %s", categoryStr)
	}
}

// ValidateID validates an item ID
func (v *Validator) ValidateID(id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid ID: must be positive")
	}
	return nil
}

// SanitizeInput sanitizes user input
func (v *Validator) SanitizeInput(input string) string {
	// Remove potentially dangerous characters
	re := regexp.MustCompile(`[<>\"'&]`)
	return re.ReplaceAllString(input, "")
}
