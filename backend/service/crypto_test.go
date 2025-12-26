package service

import (
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	// Set up a test encryption key
	testKey := []byte("12345678901234567890123456789012") // 32 bytes
	if err := SetEncryptionKey(testKey); err != nil {
		t.Fatalf("Failed to set encryption key: %v", err)
	}

	testCases := []struct {
		name      string
		plaintext string
	}{
		{"empty string", ""},
		{"simple password", "mypassword123"},
		{"special characters", "p@$$w0rd!#$%^&*()"},
		{"unicode", "密码测试123"},
		{"long password", "this-is-a-very-long-password-that-exceeds-normal-length-requirements-for-testing-purposes"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Encrypt
			encrypted, err := Encrypt(tc.plaintext)
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}

			// Decrypt
			decrypted, err := Decrypt(encrypted)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}

			// Verify round-trip
			if decrypted != tc.plaintext {
				t.Errorf("Round-trip failed: expected %q, got %q", tc.plaintext, decrypted)
			}
		})
	}
}

func TestEncryptionKeyValidation(t *testing.T) {
	// Test invalid key lengths
	invalidKeys := [][]byte{
		[]byte("short"),                                                // too short
		[]byte("this-key-is-way-too-long-for-aes-256-encryption-test"), // too long
	}

	for _, key := range invalidKeys {
		err := SetEncryptionKey(key)
		if err == nil {
			t.Errorf("Expected error for key length %d, got nil", len(key))
		}
	}

	// Test valid key length
	validKey := []byte("12345678901234567890123456789012")
	if err := SetEncryptionKey(validKey); err != nil {
		t.Errorf("Expected no error for valid key, got: %v", err)
	}
}

func TestDecryptInvalidCiphertext(t *testing.T) {
	// Set up a test encryption key
	testKey := []byte("12345678901234567890123456789012")
	if err := SetEncryptionKey(testKey); err != nil {
		t.Fatalf("Failed to set encryption key: %v", err)
	}

	// Test invalid ciphertext
	_, err := Decrypt("invalid-base64!")
	if err == nil {
		t.Error("Expected error for invalid base64, got nil")
	}

	// Test too short ciphertext (valid base64 but too short for GCM)
	_, err = Decrypt("YWJj") // "abc" in base64
	if err == nil {
		t.Error("Expected error for too short ciphertext, got nil")
	}
}
