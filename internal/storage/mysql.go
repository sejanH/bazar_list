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
