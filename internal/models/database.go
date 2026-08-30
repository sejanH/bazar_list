package models

import (
	"time"
)

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:""` // Optional name field
	Email     string    `json:"email" gorm:"type:varchar(255);unique;not null"`
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

type PaginationInfo struct {
	CurrentPage int   `json:"current_page"`
	TotalPages  int   `json:"total_pages"`
	TotalItems  int64 `json:"total_items"`
	Limit       int   `json:"limit"`
}

type PaginatedListsResponse struct {
	Month           string         `json:"month"`
	Lists           []ShoppingList `json:"lists"`
	Pagination      PaginationInfo `json:"pagination"`
	AvailableMonths []string       `json:"available_months"`
	TotalAmount     float64        `json:"total_amount"`
}

// Member Roles
const (
	RoleOwner  = "owner"
	RoleEditor = "editor"
	RoleViewer = "viewer"
)

// Activity Actions
const (
	ActionItemAdded     = "item_added"
	ActionItemUpdated   = "item_updated"
	ActionItemPurchased = "item_purchased"
	ActionItemDeleted   = "item_deleted"
	ActionMemberJoined  = "member_joined"
	ActionMemberRemoved = "member_removed"
)

type ListMember struct {
	ID        uint          `json:"id" gorm:"primaryKey"`
	ListID    uint          `json:"list_id" gorm:"not null;index;uniqueIndex:idx_list_user"`
	UserID    uint          `json:"user_id" gorm:"not null;index;uniqueIndex:idx_list_user"`
	Role      string        `json:"role" gorm:"type:varchar(20);not null;default:'editor'"`
	User      *User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
	List      *ShoppingList `json:"-" gorm:"foreignKey:ListID"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type ListActivity struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ListID    uint      `json:"list_id" gorm:"not null;index"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	UserName  string    `json:"user_name" gorm:"type:varchar(100);not null"`
	Action    string    `json:"action" gorm:"type:varchar(50);not null"`
	ItemName  string    `json:"item_name" gorm:"type:varchar(255)"`
	Details   string    `json:"details" gorm:"type:varchar(255)"`
	CreatedAt time.Time `json:"created_at"`
}

type InviteMemberRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role"` // "editor" or "viewer" (defaults to "editor")
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

