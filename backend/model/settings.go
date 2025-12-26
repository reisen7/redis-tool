package model

// SystemSetting represents a system-wide configuration setting
type SystemSetting struct {
	Key   string `json:"key" gorm:"primaryKey;size:50"`
	Value string `json:"value" gorm:"size:500"`
}

// TableName specifies table name for SystemSetting
func (SystemSetting) TableName() string {
	return "system_settings"
}

// Default setting keys
const (
	SettingRegistrationEnabled = "registration_enabled"
)

// DefaultSettings returns the default system settings
func DefaultSettings() []SystemSetting {
	return []SystemSetting{
		{Key: SettingRegistrationEnabled, Value: "true"},
	}
}
