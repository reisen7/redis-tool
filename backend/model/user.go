package model

import "time"

// UserRole represents user role type
type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

// User represents a system user
type User struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	Username      string    `json:"username" gorm:"uniqueIndex;size:50;not null"`
	PasswordHash  string    `json:"-" gorm:"size:100;not null"`
	Role          UserRole  `json:"role" gorm:"size:20;default:'user'"`
	Enabled       bool      `json:"enabled" gorm:"default:true"`
	TotpSecret    string    `json:"-" gorm:"size:100"`
	TotpEnabled   bool      `json:"totpEnabled" gorm:"default:false"`
	MustChangePwd bool      `json:"mustChangePassword" gorm:"default:false"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// TableName specifies table name for User
func (User) TableName() string {
	return "users"
}

// IsAdmin returns true if user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}
