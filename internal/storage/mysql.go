package storage

import (
	"fmt"
	"os"
	"time"

	"github.com/sejan/bazarlist/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MySQLStorage handles MySQL database operations
type MySQLStorage struct {
	db *gorm.DB
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

// NewMySQLStorage creates a new MySQL storage instance
func NewMySQLStorage(config DatabaseConfig) (*MySQLStorage, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate tables
	if err := db.AutoMigrate(&models.User{}, &models.ShoppingList{}, &models.Item{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &MySQLStorage{db: db}, nil
}

// NewMySQLStorageFromEnv creates MySQL storage from environment variables
func NewMySQLStorageFromEnv() (*MySQLStorage, error) {
	config := DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "3306"),
		User:     getEnv("DB_USER", "root"),
		Password: getEnv("DB_PASSWORD", ""),
		Database: getEnv("DB_NAME", "bazarlist"),
	}

	return NewMySQLStorage(config)
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// User Operations
func (s *MySQLStorage) CreateUser(user *models.User) error {
	return s.db.Create(user).Error
}

func (s *MySQLStorage) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	err := s.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *MySQLStorage) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := s.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Shopping List Operations
func (s *MySQLStorage) CreateList(list *models.ShoppingList) error {
	return s.db.Create(list).Error
}

func (s *MySQLStorage) GetListByID(id uint) (*models.ShoppingList, error) {
	var list models.ShoppingList
	err := s.db.Preload("Items").First(&list, id).Error
	if err != nil {
		return nil, err
	}
	return &list, nil
}

func (s *MySQLStorage) GetListsByUserID(userID uint) ([]models.ShoppingList, error) {
	var lists []models.ShoppingList
	err := s.db.Where("user_id = ?", userID).
		Order("date DESC, created_at DESC").
		Preload("Items").
		Find(&lists).Error
	return lists, err
}

// GetPaginatedListsByMonth fetches lists for a specific month with pagination
func (s *MySQLStorage) GetPaginatedListsByMonth(userID uint, year, month int, page, limit int) ([]models.ShoppingList, int64, error) {
	var lists []models.ShoppingList
	var total int64

	query := s.db.Model(&models.ShoppingList{}).
		Where("user_id = ? AND YEAR(date) = ? AND MONTH(date) = ?", userID, year, month)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch paginated data
	offset := (page - 1) * limit
	err := query.Order("date DESC, created_at DESC").
		Preload("Items").
		Offset(offset).
		Limit(limit).
		Find(&lists).Error

	return lists, total, err
}

// GetLatestMonth finds the most recent year and month that has data for the user
func (s *MySQLStorage) GetLatestMonth(userID uint) (int, int, error) {
	var result struct {
		Year  int
		Month int
	}

	err := s.db.Model(&models.ShoppingList{}).
		Select("YEAR(date) as year, MONTH(date) as month").
		Where("user_id = ?", userID).
		Order("date DESC").
		First(&result).Error

	if err != nil {
		return 0, 0, err
	}

	return result.Year, result.Month, nil
}

// GetMonthTotal calculates the total sum of all item prices for a specific month
func (s *MySQLStorage) GetMonthTotal(userID uint, year, month int) (float64, error) {
	var total float64
	
	// Join items and shopping_lists to sum prices by user and month
	err := s.db.Table("items").
		Joins("JOIN shopping_lists ON items.list_id = shopping_lists.id").
		Where("shopping_lists.user_id = ? AND YEAR(shopping_lists.date) = ? AND MONTH(shopping_lists.date) = ?", userID, year, month).
		Select("COALESCE(SUM(items.price), 0)").
		Scan(&total).Error
	
	return total, err
}

// GetAvailableMonths returns a list of "YYYY-MM" strings for all months with data
func (s *MySQLStorage) GetAvailableMonths(userID uint) ([]string, error) {
	months := []string{}
	
	// Use Pluck to get formatted distinct months
	err := s.db.Model(&models.ShoppingList{}).
		Select("DISTINCT DATE_FORMAT(date, '%Y-%m') as month_str").
		Where("user_id = ?", userID).
		Order("month_str DESC").
		Pluck("month_str", &months).Error
	
	if err != nil {
		return []string{}, err
	}

	return months, nil
}

func (s *MySQLStorage) UpdateList(list *models.ShoppingList) error {
	return s.db.Save(list).Error
}

func (s *MySQLStorage) DeleteList(id uint) error {
	return s.db.Delete(&models.ShoppingList{}, id).Error
}

// Item Operations
func (s *MySQLStorage) CreateItem(item *models.Item) error {
	return s.db.Create(item).Error
}

func (s *MySQLStorage) GetItemByID(id uint) (*models.Item, error) {
	var item models.Item
	err := s.db.First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *MySQLStorage) GetItemsByListID(listID uint) ([]models.Item, error) {
	var items []models.Item
	err := s.db.Where("list_id = ?", listID).Order("created_at ASC").Find(&items).Error
	return items, err
}

func (s *MySQLStorage) UpdateItem(item *models.Item) error {
	return s.db.Save(item).Error
}

func (s *MySQLStorage) DeleteItem(id uint) error {
	return s.db.Delete(&models.Item{}, id).Error
}

// Close closes the database connection
func (s *MySQLStorage) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GetDB returns the underlying GORM DB instance
func (s *MySQLStorage) GetDB() *gorm.DB {
	return s.db
}
