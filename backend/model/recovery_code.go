package model

// RecoveryCode represents a single-use 2FA recovery code
type RecoveryCode struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	UserID   uint   `json:"userId" gorm:"index;not null"`
	CodeHash string `json:"-" gorm:"size:100;not null"`
	Used     bool   `json:"used" gorm:"default:false"`

	// Relationship
	User User `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName specifies table name for RecoveryCode
func (RecoveryCode) TableName() string {
	return "recovery_codes"
}

// RecoveryCodeCount is the number of recovery codes generated for 2FA
const RecoveryCodeCount = 10
