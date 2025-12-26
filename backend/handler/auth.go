package handler

import (
	"net/http"
	"regexp"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"redis-web-manager/model"
	"redis-web-manager/service"
)

// ============================================================================
// Request/Response Types
// ============================================================================

// LoginRequest represents login request body
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents login response
type LoginResponse struct {
	Token      string      `json:"token,omitempty"`
	User       *model.User `json:"user,omitempty"`
	Requires2FA bool       `json:"requires2FA,omitempty"`
	TempToken  string      `json:"tempToken,omitempty"`
}

// RegisterRequest represents registration request body
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8"`
}

// Verify2FARequest represents 2FA verification request body
type Verify2FARequest struct {
	TempToken string `json:"tempToken" binding:"required"`
	Code      string `json:"code" binding:"required"`
}

// ============================================================================
// Login Handler (Task 5.1 - Requirements 3.2, 3.3, 9.1)
// ============================================================================

// Login handles user login
// POST /api/auth/login
// Returns JWT token or 2FA required flag
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "参数错误",
		})
		return
	}

	// Find user by username
	db := service.GetDB()
	var user model.User
	if err := db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Don't reveal which field is incorrect (Requirement 3.3)
			c.JSON(http.StatusUnauthorized, model.APIResponse{
				Success: false,
				Error:   "用户名或密码错误",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "服务器内部错误",
		})
		return
	}

	// Check if user is enabled
	if !user.Enabled {
		c.JSON(http.StatusUnauthorized, model.APIResponse{
			Success: false,
			Error:   "用户账户已被禁用",
		})
		return
	}

	// Verify password
	if !service.VerifyPassword(req.Password, user.PasswordHash) {
		// Don't reveal which field is incorrect (Requirement 3.3)
		c.JSON(http.StatusUnauthorized, model.APIResponse{
			Success: false,
			Error:   "用户名或密码错误",
		})
		return
	}

	// Check if 2FA is enabled
	if user.TotpEnabled {
		// Generate temporary token for 2FA verification
		tempToken, err := service.GenerateToken(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, model.APIResponse{
				Success: false,
				Error:   "服务器内部错误",
			})
			return
		}

		c.JSON(http.StatusOK, model.APIResponse{
			Success: true,
			Data: LoginResponse{
				Requires2FA: true,
				TempToken:   tempToken,
			},
		})
		return
	}

	// Generate JWT token
	token, err := service.GenerateToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "服务器内部错误",
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data: LoginResponse{
			Token: token,
			User:  &user,
		},
	})
}


// ============================================================================
// Registration Handler (Task 5.3 - Requirements 4.1, 4.2, 4.3, 4.4, 4.5, 9.2)
// ============================================================================

// ValidatePassword checks if password meets strength requirements
// Requirements: 4.5, 4.6 - at least 8 chars, 1 uppercase, 1 lowercase, 1 number
func ValidatePassword(password string) (bool, string) {
	if len(password) < 8 {
		return false, "密码长度至少为8个字符"
	}

	var hasUpper, hasLower, hasNumber bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		}
	}

	if !hasUpper {
		return false, "密码必须包含至少一个大写字母"
	}
	if !hasLower {
		return false, "密码必须包含至少一个小写字母"
	}
	if !hasNumber {
		return false, "密码必须包含至少一个数字"
	}

	return true, ""
}

// ValidateUsername checks if username is valid
func ValidateUsername(username string) (bool, string) {
	if len(username) < 3 {
		return false, "用户名长度至少为3个字符"
	}
	if len(username) > 50 {
		return false, "用户名长度不能超过50个字符"
	}

	// Only allow alphanumeric and underscore
	validUsername := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !validUsername.MatchString(username) {
		return false, "用户名只能包含字母、数字和下划线"
	}

	return true, ""
}

// Register handles user registration
// POST /api/auth/register
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "参数错误",
		})
		return
	}

	// Check if registration is enabled (Requirement 4.2)
	if !service.IsRegistrationEnabled() {
		c.JSON(http.StatusForbidden, model.APIResponse{
			Success: false,
			Error:   "注册功能已关闭",
		})
		return
	}

	db := service.GetDB()

	// Validate username
	if valid, msg := ValidateUsername(req.Username); !valid {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   msg,
		})
		return
	}

	// Validate password strength (Requirement 4.5)
	if valid, msg := ValidatePassword(req.Password); !valid {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   msg,
		})
		return
	}

	// Check if username already exists (Requirement 4.4)
	var existingUser model.User
	if err := db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, model.APIResponse{
			Success: false,
			Error:   "用户名已存在",
		})
		return
	}

	// Hash password
	hashedPassword, err := service.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "服务器内部错误",
		})
		return
	}

	// Create user
	user := model.User{
		Username:     req.Username,
		PasswordHash: hashedPassword,
		Role:         model.RoleUser,
		Enabled:      true,
	}

	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "服务器内部错误",
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data: gin.H{
			"message": "注册成功",
		},
	})
}


// ============================================================================
// 2FA Verification Handler (Task 5.6 - Requirements 5.3, 5.4, 9.4)
// ============================================================================

// Verify2FA handles 2FA code verification during login
// POST /api/auth/verify-2fa
func Verify2FA(c *gin.Context) {
	var req Verify2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "参数错误",
		})
		return
	}

	// Validate temp token
	claims, err := service.ValidateToken(req.TempToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.APIResponse{
			Success: false,
			Error:   "认证已过期",
		})
		return
	}

	// Get user from database
	db := service.GetDB()
	var user model.User
	if err := db.First(&user, claims.UserID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, model.APIResponse{
			Success: false,
			Error:   "用户不存在",
		})
		return
	}

	// Check if user is still enabled
	if !user.Enabled {
		c.JSON(http.StatusUnauthorized, model.APIResponse{
			Success: false,
			Error:   "用户账户已被禁用",
		})
		return
	}

	// Normalize code (remove spaces and dashes)
	code := strings.ReplaceAll(req.Code, " ", "")
	code = strings.ReplaceAll(code, "-", "")
	code = strings.ToUpper(code)

	// Try TOTP verification first
	if service.VerifyTOTP(code, user.TotpSecret) {
		// Generate full JWT token
		token, err := service.GenerateToken(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, model.APIResponse{
				Success: false,
				Error:   "服务器内部错误",
			})
			return
		}

		c.JSON(http.StatusOK, model.APIResponse{
			Success: true,
			Data: LoginResponse{
				Token: token,
				User:  &user,
			},
		})
		return
	}

	// Try recovery code verification
	var recoveryCodes []model.RecoveryCode
	if err := db.Where("user_id = ? AND used = ?", user.ID, false).Find(&recoveryCodes).Error; err == nil {
		// Extract hashed codes
		hashedCodes := make([]string, len(recoveryCodes))
		for i, rc := range recoveryCodes {
			hashedCodes[i] = rc.CodeHash
		}

		// Check if code matches any recovery code
		matchIndex := service.VerifyRecoveryCode(code, hashedCodes)
		if matchIndex >= 0 {
			// Mark recovery code as used
			recoveryCodes[matchIndex].Used = true
			db.Save(&recoveryCodes[matchIndex])

			// Generate full JWT token
			token, err := service.GenerateToken(&user)
			if err != nil {
				c.JSON(http.StatusInternalServerError, model.APIResponse{
					Success: false,
					Error:   "服务器内部错误",
				})
				return
			}

			c.JSON(http.StatusOK, model.APIResponse{
				Success: true,
				Data: LoginResponse{
					Token: token,
					User:  &user,
				},
			})
			return
		}
	}

	// Invalid code
	c.JSON(http.StatusUnauthorized, model.APIResponse{
		Success: false,
		Error:   "验证码无效",
	})
}

// ============================================================================
// Logout Handler (Task 5.7 - Requirement 3.5)
// ============================================================================

// Logout handles user logout
// POST /api/auth/logout
// Note: JWT tokens are stateless, so logout is handled client-side
// This endpoint exists for API completeness and potential future token blacklisting
func Logout(c *gin.Context) {
	// JWT is stateless, actual logout happens on client side by removing token
	// This endpoint can be used for:
	// 1. Logging logout events
	// 2. Future token blacklisting implementation
	// 3. Clearing server-side session data if any

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data: gin.H{
			"message": "登出成功",
		},
	})
}

// ============================================================================
// 2FA Setup Handlers (Task 6 - Requirements 5.1, 5.2, 5.5, 5.7)
// ============================================================================

// Setup2FARequest is empty as no input is needed
type Setup2FARequest struct{}

// Setup2FAResponse represents the response for 2FA setup initiation
type Setup2FAResponse struct {
	Secret        string   `json:"secret"`
	QRCode        string   `json:"qrCode"`
	RecoveryCodes []string `json:"recoveryCodes"`
}

// Activate2FARequest represents the request to activate 2FA
type Activate2FARequest struct {
	Code string `json:"code" binding:"required"`
}

// Disable2FARequest represents the request to disable 2FA
type Disable2FARequest struct {
	Password string `json:"password" binding:"required"`
	Code     string `json:"code" binding:"required"`
}

// Setup2FA initiates 2FA setup for the current user
// POST /api/2fa/setup
// Generates TOTP secret, QR code, and recovery codes
// Requirements: 5.1, 5.7
func Setup2FA(c *gin.Context) {
	// Get user ID from context (set by JWTAuth middleware)
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, model.APIResponse{
			Success: false,
			Error:   "认证已过期",
		})
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.APIResponse{
			Success: false,
			Error:   "认证已过期",
		})
		return
	}

	// Get user from database
	db := service.GetDB()
	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Error:   "用户不存在",
		})
		return
	}

	// Check if 2FA is already enabled
	if user.TotpEnabled {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "双重认证已启用，请先禁用后再重新设置",
		})
		return
	}

	// Generate TOTP secret
	secret, err := service.GenerateTOTPSecret(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "服务器内部错误",
		})
		return
	}

	// Generate QR code
	qrCode, err := service.GenerateQRCode(secret, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "服务器内部错误",
		})
		return
	}

	// Generate recovery codes
	plainCodes, hashedCodes, err := service.GenerateRecoveryCodes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "服务器内部错误",
		})
		return
	}

	// Start a transaction to save secret and recovery codes
	tx := db.Begin()

	// Save TOTP secret to user (but don't enable yet)
	if err := tx.Model(&user).Update("totp_secret", secret).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "服务器内部错误",
		})
		return
	}

	// Delete any existing recovery codes for this user
	if err := tx.Where("user_id = ?", userID).Delete(&model.RecoveryCode{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "服务器内部错误",
		})
		return
	}

	// Save new recovery codes
	for _, hashedCode := range hashedCodes {
		recoveryCode := model.RecoveryCode{
			UserID:   userID,
			CodeHash: hashedCode,
			Used:     false,
		}
		if err := tx.Create(&recoveryCode).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, model.APIResponse{
				Success: false,
				Error:   "服务器内部错误",
			})
			return
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "服务器内部错误",
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data: Setup2FAResponse{
			Secret:        secret,
			QRCode:        qrCode,
			RecoveryCodes: plainCodes,
		},
	})
}

// Activate2FA activates 2FA after verifying the initial TOTP code
// POST /api/2fa/verify
// Requirements: 5.2
func Activate2FA(c *gin.Context) {
	var req Activate2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "参数错误",
		})
		return
	}

	// Get user ID from context
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, model.APIResponse{
			Success: false,
			Error:   "认证已过期",
		})
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.APIResponse{
			Success: false,
			Error:   "认证已过期",
		})
		return
	}

	// Get user from database
	db := service.GetDB()
	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Error:   "用户不存在",
		})
		return
	}

	// Check if TOTP secret exists (setup was initiated)
	if user.TotpSecret == "" {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "请先启动双重认证设置",
		})
		return
	}

	// Check if 2FA is already enabled
	if user.TotpEnabled {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "双重认证已启用",
		})
		return
	}

	// Normalize code
	code := strings.ReplaceAll(req.Code, " ", "")
	code = strings.ReplaceAll(code, "-", "")

	// Verify TOTP code
	if !service.VerifyTOTP(code, user.TotpSecret) {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "验证码无效",
		})
		return
	}

	// Enable 2FA
	if err := db.Model(&user).Update("totp_enabled", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "服务器内部错误",
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data: gin.H{
			"message": "双重认证已启用",
		},
	})
}

// Disable2FA disables 2FA for the current user
// POST /api/2fa/disable
// Requires password and TOTP code verification
// Requirements: 5.5
func Disable2FA(c *gin.Context) {
	var req Disable2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "参数错误",
		})
		return
	}

	// Get user ID from context
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, model.APIResponse{
			Success: false,
			Error:   "认证已过期",
		})
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.APIResponse{
			Success: false,
			Error:   "认证已过期",
		})
		return
	}

	// Get user from database
	db := service.GetDB()
	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Error:   "用户不存在",
		})
		return
	}

	// Check if 2FA is enabled
	if !user.TotpEnabled {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "双重认证未启用",
		})
		return
	}

	// Verify password
	if !service.VerifyPassword(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, model.APIResponse{
			Success: false,
			Error:   "密码错误",
		})
		return
	}

	// Normalize code
	code := strings.ReplaceAll(req.Code, " ", "")
	code = strings.ReplaceAll(code, "-", "")

	// Verify TOTP code
	if !service.VerifyTOTP(code, user.TotpSecret) {
		c.JSON(http.StatusUnauthorized, model.APIResponse{
			Success: false,
			Error:   "验证码无效",
		})
		return
	}

	// Start transaction to disable 2FA and remove recovery codes
	tx := db.Begin()

	// Disable 2FA and clear TOTP secret
	if err := tx.Model(&user).Updates(map[string]interface{}{
		"totp_enabled": false,
		"totp_secret":  "",
	}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "服务器内部错误",
		})
		return
	}

	// Delete recovery codes
	if err := tx.Where("user_id = ?", userID).Delete(&model.RecoveryCode{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "服务器内部错误",
		})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "服务器内部错误",
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data: gin.H{
			"message": "双重认证已禁用",
		},
	})
}

// ============================================================================
// Get Current User Handler
// ============================================================================

// GetCurrentUser returns the current authenticated user
// GET /api/auth/me
func GetCurrentUser(c *gin.Context) {
	// Get user ID from context (set by JWTAuth middleware)
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, model.APIResponse{
			Success: false,
			Error:   "认证已过期",
		})
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.APIResponse{
			Success: false,
			Error:   "认证已过期",
		})
		return
	}

	// Get user from database
	db := service.GetDB()
	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Error:   "用户不存在",
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    user,
	})
}
