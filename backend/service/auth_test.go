package service

import (
	"strings"
	"testing"
	"time"

	"redis-web-manager/model"
)

// ============================================================================
// Password Hashing Tests (Task 2.1)
// ============================================================================

func TestHashPassword(t *testing.T) {
	password := "TestPassword123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Error("HashPassword returned empty hash")
	}

	if hash == password {
		t.Error("HashPassword returned unhashed password")
	}

	// Hash should be bcrypt format (starts with $2a$ or $2b$)
	if !strings.HasPrefix(hash, "$2") {
		t.Errorf("Hash doesn't appear to be bcrypt format: %s", hash)
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "TestPassword123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Correct password should verify
	if !VerifyPassword(password, hash) {
		t.Error("VerifyPassword failed for correct password")
	}

	// Wrong password should not verify
	if VerifyPassword("WrongPassword", hash) {
		t.Error("VerifyPassword succeeded for wrong password")
	}
}

func TestHashPasswordDifferentHashes(t *testing.T) {
	password := "TestPassword123!"

	hash1, _ := HashPassword(password)
	hash2, _ := HashPassword(password)

	// Same password should produce different hashes (due to salt)
	if hash1 == hash2 {
		t.Error("HashPassword produced identical hashes for same password")
	}

	// Both hashes should verify
	if !VerifyPassword(password, hash1) {
		t.Error("First hash doesn't verify")
	}
	if !VerifyPassword(password, hash2) {
		t.Error("Second hash doesn't verify")
	}
}

// ============================================================================
// JWT Token Tests (Task 2.2)
// ============================================================================

func TestGenerateAndValidateToken(t *testing.T) {
	// Set up JWT secret
	SetJWTSecret([]byte("test-secret-key-for-jwt-signing-32"))

	user := &model.User{
		ID:       1,
		Username: "testuser",
		Role:     model.RoleUser,
	}

	token, err := GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if token == "" {
		t.Error("GenerateToken returned empty token")
	}

	// Validate the token
	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("UserID mismatch: got %d, want %d", claims.UserID, user.ID)
	}

	if claims.Username != user.Username {
		t.Errorf("Username mismatch: got %s, want %s", claims.Username, user.Username)
	}

	if claims.Role != user.Role {
		t.Errorf("Role mismatch: got %s, want %s", claims.Role, user.Role)
	}
}

func TestValidateInvalidToken(t *testing.T) {
	SetJWTSecret([]byte("test-secret-key-for-jwt-signing-32"))

	// Invalid token should fail
	_, err := ValidateToken("invalid-token")
	if err == nil {
		t.Error("ValidateToken should fail for invalid token")
	}
}

func TestValidateTokenWithWrongSecret(t *testing.T) {
	// Generate token with one secret
	SetJWTSecret([]byte("secret-key-one-for-jwt-signing-32"))

	user := &model.User{
		ID:       1,
		Username: "testuser",
		Role:     model.RoleUser,
	}

	token, _ := GenerateToken(user)

	// Try to validate with different secret
	SetJWTSecret([]byte("secret-key-two-for-jwt-signing-32"))

	_, err := ValidateToken(token)
	if err == nil {
		t.Error("ValidateToken should fail with wrong secret")
	}
}

func TestTokenExpiration(t *testing.T) {
	SetJWTSecret([]byte("test-secret-key-for-jwt-signing-32"))

	user := &model.User{
		ID:       1,
		Username: "testuser",
		Role:     model.RoleUser,
	}

	token, _ := GenerateToken(user)
	claims, _ := ValidateToken(token)

	// Check expiration is approximately 24 hours from now
	expectedExpiry := time.Now().Add(JWTExpiration)
	actualExpiry := claims.ExpiresAt.Time

	diff := actualExpiry.Sub(expectedExpiry)
	if diff < -time.Minute || diff > time.Minute {
		t.Errorf("Token expiration not within expected range: got %v, want ~%v", actualExpiry, expectedExpiry)
	}
}

// ============================================================================
// TOTP Tests (Task 2.3)
// ============================================================================

func TestGenerateTOTPSecret(t *testing.T) {
	secret, err := GenerateTOTPSecret("testuser")
	if err != nil {
		t.Fatalf("GenerateTOTPSecret failed: %v", err)
	}

	if secret == "" {
		t.Error("GenerateTOTPSecret returned empty secret")
	}

	// Secret should be base32 encoded
	if len(secret) < 16 {
		t.Error("TOTP secret seems too short")
	}
}

func TestGenerateQRCode(t *testing.T) {
	secret, _ := GenerateTOTPSecret("testuser")

	qrCode, err := GenerateQRCode(secret, "testuser")
	if err != nil {
		t.Fatalf("GenerateQRCode failed: %v", err)
	}

	// QR code should be a data URI
	if !strings.HasPrefix(qrCode, "data:image/png;base64,") {
		t.Error("QR code should be a PNG data URI")
	}
}

func TestVerifyTOTP(t *testing.T) {
	// Generate a secret
	secret, _ := GenerateTOTPSecret("testuser")

	// Invalid code should fail
	if VerifyTOTP("000000", secret) {
		t.Error("VerifyTOTP should fail for invalid code")
	}

	// Note: We can't easily test valid TOTP without generating a real code
	// which would require time-based generation
}

// ============================================================================
// Recovery Code Tests (Task 2.5)
// ============================================================================

func TestGenerateRecoveryCodes(t *testing.T) {
	plainCodes, hashedCodes, err := GenerateRecoveryCodes()
	if err != nil {
		t.Fatalf("GenerateRecoveryCodes failed: %v", err)
	}

	// Should generate exactly 10 codes
	if len(plainCodes) != RecoveryCodeCount {
		t.Errorf("Expected %d plain codes, got %d", RecoveryCodeCount, len(plainCodes))
	}

	if len(hashedCodes) != RecoveryCodeCount {
		t.Errorf("Expected %d hashed codes, got %d", RecoveryCodeCount, len(hashedCodes))
	}

	// Each code should be unique
	seen := make(map[string]bool)
	for _, code := range plainCodes {
		if seen[code] {
			t.Error("Duplicate recovery code generated")
		}
		seen[code] = true

		// Code should be uppercase alphanumeric
		if len(code) != RecoveryCodeLength {
			t.Errorf("Recovery code length should be %d, got %d", RecoveryCodeLength, len(code))
		}
	}

	// Hashed codes should not equal plain codes
	for i, hash := range hashedCodes {
		if hash == plainCodes[i] {
			t.Error("Hashed code equals plain code")
		}
	}
}

func TestVerifyRecoveryCode(t *testing.T) {
	plainCodes, hashedCodes, _ := GenerateRecoveryCodes()

	// Each plain code should verify against its hash
	for i, code := range plainCodes {
		idx := VerifyRecoveryCode(code, hashedCodes)
		if idx != i {
			t.Errorf("VerifyRecoveryCode returned wrong index: got %d, want %d", idx, i)
		}
	}

	// Invalid code should return -1
	idx := VerifyRecoveryCode("INVALID1", hashedCodes)
	if idx != -1 {
		t.Errorf("VerifyRecoveryCode should return -1 for invalid code, got %d", idx)
	}

	// Lowercase should also work
	idx = VerifyRecoveryCode(strings.ToLower(plainCodes[0]), hashedCodes)
	if idx != 0 {
		t.Error("VerifyRecoveryCode should accept lowercase codes")
	}
}
