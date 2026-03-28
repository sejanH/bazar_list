package models

import (
	"time"
)

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:""` // Optional name field
	Email     string    `json:"email" gorm:"uniqueIndex;not null"`
	Password  string    `json:"-" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ShoppingList struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	Name      string    `json:"name" gorm:"not null"`
	Date      time.Time `json:"date" gorm:"not null;index"`
	User      *User     `json:"-" gorm:"foreignKey:UserID"`
	Items     []Item    `json:"items" gorm:"foreignKey:ListID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (sl *ShoppingList) GetFullName() string {
	dateStr := sl.Date.Format("2006-01-02")
	if sl.Name == "" {
		return dateStr
	}
	return sl.Name + "-" + dateStr
}

// NewShoppingList creates a new empty shopping list
func NewShoppingList() *ShoppingList {
	return &ShoppingList{
		Items: make([]Item, 0),
		Date:  time.Now(),
	}
}

type Item struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	ListID    uint   `json:"list_id" gorm:"not null;index"`
	Name      string `json:"name" gorm:"not null"`
	Price     float64 `json:"price" gorm:"type:decimal(10,2);default:0"`
	Purchased bool   `json:"purchased" gorm:"default:false"`
	List      *ShoppingList `json:"-" gorm:"foreignKey:ListID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Name     string `json:"name"` // Optional name field
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type TokenResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type ListRequest struct {
	Name string `json:"name" binding:"required"`
	Date string `json:"date"`
}

type ItemRequest struct {
	Name      string  `json:"name" binding:"required"`
	Price     float64 `json:"price"`
	Purchased bool    `json:"purchased"`
}
