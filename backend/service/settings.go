package service

import (
	"errors"

	"gorm.io/gorm"
	"redis-web-manager/model"
)

// Settings service errors
var (
	ErrSettingNotFound = errors.New("setting not found")
)

// GetSetting retrieves a system setting by key
// Requirements: 8.2, 8.3
func GetSetting(key string) (string, error) {
	var setting model.SystemSetting
	err := db.Where("key = ?", key).First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrSettingNotFound
		}
		return "", err
	}
	return setting.Value, nil
}

// SetSetting creates or updates a system setting
// Requirements: 8.3, 8.4
func SetSetting(key, value string) error {
	var setting model.SystemSetting
	err := db.Where("key = ?", key).First(&setting).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new setting
		setting = model.SystemSetting{
			Key:   key,
			Value: value,
		}
		return db.Create(&setting).Error
	} else if err != nil {
		return err
	}

	// Update existing setting
	setting.Value = value
	return db.Save(&setting).Error
}

// GetAllSettings retrieves all system settings
// Requirements: 8.2
func GetAllSettings() ([]model.SystemSetting, error) {
	var settings []model.SystemSetting
	err := db.Find(&settings).Error
	return settings, err
}

// GetSettingsMap retrieves all settings as a key-value map
// Requirements: 8.2
func GetSettingsMap() (map[string]string, error) {
	settings, err := GetAllSettings()
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, s := range settings {
		result[s.Key] = s.Value
	}
	return result, nil
}

// IsRegistrationEnabled checks if user registration is enabled
// Requirements: 8.1
func IsRegistrationEnabled() bool {
	value, err := GetSetting(model.SettingRegistrationEnabled)
	if err != nil {
		// Default to true if setting not found
		return true
	}
	return value == "true"
}

// SetRegistrationEnabled enables or disables user registration
// Requirements: 8.1
func SetRegistrationEnabled(enabled bool) error {
	value := "false"
	if enabled {
		value = "true"
	}
	return SetSetting(model.SettingRegistrationEnabled, value)
}
