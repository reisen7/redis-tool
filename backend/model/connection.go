package model

import "time"

// UserConnection represents a Redis connection owned by a user
// This is separate from the existing ConnectionConfig to support user isolation
type UserConnection struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"userId" gorm:"index;not null"`
	GroupID     string    `json:"groupId" gorm:"size:36;index"` // Optional group reference
	Name        string    `json:"name" gorm:"size:100;not null"`
	Host        string    `json:"host" gorm:"size:255;not null"`
	Port        int       `json:"port" gorm:"default:6379"`
	PasswordEnc string    `json:"-" gorm:"size:500"` // AES-256 encrypted
	DB          int       `json:"db" gorm:"default:0"`
	Timeout     int       `json:"timeout" gorm:"default:5"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	// Relationship
	User User `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName specifies table name for UserConnection
func (UserConnection) TableName() string {
	return "user_connections"
}
