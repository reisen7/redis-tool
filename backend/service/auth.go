package service

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"errors"
	"fmt"
	"image/png"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
	"redis-web-manager/model"
)

// Password hashing constants
const (
	// BcryptCost is the cost factor for bcrypt hashing (Requirement 9.6)
	BcryptCost = 12
)

// JWT constants
const (
	// JWTExpiration is the token expiration time (Requirement 9.7)
	JWTExpiration = 24 * time.Hour
)

// TOTP constants
const (
	// TOTPIssuer is the issuer name shown in authenticator apps
	TOTPIssuer = "Redis Web Manager"
	// TOTPPeriod is the time step in seconds
	TOTPPeriod = 30
	// TOTPDigits is the number of digits in the TOTP code
	TOTPDigits = 6
)

// Recovery code constants
const (
	// RecoveryCodeLength is the length of each recovery code
	RecoveryCodeLength = 8
	// RecoveryCodeCount is the number of recovery codes to generate
	RecoveryCodeCount = 10
)

var (
	// ErrInvalidCredentials is returned when login credentials are invalid
	ErrInvalidCredentials = errors.New("invalid username or password")
	// ErrInvalidToken is returned when JWT token is invalid
	ErrInvalidToken = errors.New("invalid or expired token")
	// ErrUserDisabled is returned when user account is disabled
	ErrUserDisabled = errors.New("user account is disabled")
	// ErrInvalidTOTP is returned when TOTP code is invalid
	ErrInvalidTOTP = errors.New("invalid TOTP code")
	// ErrInvalidRecoveryCode is returned when recovery code is invalid
	ErrInvalidRecoveryCode = errors.New("invalid or already used recovery code")
)

// jwtSecret is the secret key for signing JWT tokens
var jwtSecret []byte

// SetJWTSecret sets the JWT signing secret
func SetJWTSecret(secret []byte) {
	jwtSecret = make([]byte, len(secret))
	copy(jwtSecret, secret)
}

// JWTClaims represents the claims in a JWT token
type JWTClaims struct {
	UserID   uint           `json:"userId"`
	Username string         `json:"username"`
	Role     model.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// ============================================================================
// Password Hashing (Task 2.1 - Requirement 9.6)
// ============================================================================

// HashPassword hashes a password using bcrypt with cost factor 12
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword verifies a password against a bcrypt hash
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ============================================================================
// JWT Token Management (Task 2.2 - Requirements 9.7, 3.2)
// ============================================================================

// GenerateToken generates a JWT token for a user with 24-hour expiration
func GenerateToken(user *model.User) (string, error) {
	if len(jwtSecret) == 0 {
		return "", errors.New("JWT secret not configured")
	}

	now := time.Now()
	claims := JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(JWTExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString string) (*JWTClaims, error) {
	if len(jwtSecret) == 0 {
		return nil, errors.New("JWT secret not configured")
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ============================================================================
// TOTP Management (Task 2.3 - Requirements 5.1, 5.2, 5.4)
// ============================================================================

// TOTPSetupResult contains the result of TOTP setup
type TOTPSetupResult struct {
	Secret        string   `json:"secret"`
	QRCode        string   `json:"qrCode"`        // Base64 encoded PNG image
	RecoveryCodes []string `json:"recoveryCodes"` // Plain text codes (only shown once)
}

// GenerateTOTPSecret generates a new TOTP secret for a user
func GenerateTOTPSecret(username string) (string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      TOTPIssuer,
		AccountName: username,
		Period:      TOTPPeriod,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate TOTP secret: %w", err)
	}

	return key.Secret(), nil
}

// GenerateQRCode generates a QR code image for TOTP setup
// Returns base64 encoded PNG image with data URI prefix
func GenerateQRCode(secret, username string) (string, error) {
	key, err := otp.NewKeyFromURL(fmt.Sprintf(
		"otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=%d&period=%d",
		TOTPIssuer,
		username,
		secret,
		TOTPIssuer,
		TOTPDigits,
		TOTPPeriod,
	))
	if err != nil {
		return "", fmt.Errorf("failed to create TOTP key: %w", err)
	}

	// Generate QR code image
	img, err := key.Image(200, 200)
	if err != nil {
		return "", fmt.Errorf("failed to generate QR code image: %w", err)
	}

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("failed to encode QR code to PNG: %w", err)
	}

	// Return as data URI
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// VerifyTOTP verifies a TOTP code against a secret
func VerifyTOTP(code, secret string) bool {
	return totp.Validate(code, secret)
}

// ============================================================================
// Recovery Codes (Task 2.5 - Requirements 5.6, 5.7)
// ============================================================================

// GenerateRecoveryCodes generates 10 single-use recovery codes
// Returns plain text codes (to show to user) and hashed codes (to store)
func GenerateRecoveryCodes() (plainCodes []string, hashedCodes []string, err error) {
	plainCodes = make([]string, RecoveryCodeCount)
	hashedCodes = make([]string, RecoveryCodeCount)

	for i := 0; i < RecoveryCodeCount; i++ {
		// Generate random bytes
		randomBytes := make([]byte, RecoveryCodeLength)
		if _, err := rand.Read(randomBytes); err != nil {
			return nil, nil, fmt.Errorf("failed to generate random bytes: %w", err)
		}

		// Encode to base32 and take first RecoveryCodeLength characters
		code := strings.ToUpper(base32.StdEncoding.EncodeToString(randomBytes))[:RecoveryCodeLength]
		plainCodes[i] = code

		// Hash the code for storage
		hash, err := HashPassword(code)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to hash recovery code: %w", err)
		}
		hashedCodes[i] = hash
	}

	return plainCodes, hashedCodes, nil
}

// VerifyRecoveryCode verifies a recovery code against stored hashes
// Returns the index of the matching code, or -1 if not found
func VerifyRecoveryCode(code string, hashedCodes []string) int {
	code = strings.ToUpper(strings.TrimSpace(code))
	for i, hash := range hashedCodes {
		if VerifyPassword(code, hash) {
			return i
		}
	}
	return -1
}
