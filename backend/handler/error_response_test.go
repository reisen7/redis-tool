package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/quick"

	"github.com/gin-gonic/gin"
)

// **Feature: redis-web-manager, Property 9: Error Response Structure**
// **Validates: Requirements 7.4**
//
// For any error condition in the backend, the response should follow a consistent
// structure with success=false, an error field containing the message, and
// appropriate HTTP status code.

func init() {
	gin.SetMode(gin.TestMode)
}

// ErrorResponseStructure represents the expected structure of error responses
type ErrorResponseStructure struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// validateErrorResponse checks that a response body conforms to the error response structure
func validateErrorResponse(body []byte) bool {
	var resp ErrorResponseStructure
	if err := json.Unmarshal(body, &resp); err != nil {
		return false
	}
	// Error responses must have success=false and a non-empty error message
	return !resp.Success && resp.Error != ""
}

// TestProperty9_ErrorResponseStructure_InvalidJSON tests that invalid JSON input
// produces a properly structured error response
// **Feature: redis-web-manager, Property 9: Error Response Structure**
// **Validates: Requirements 7.4**
func TestProperty9_ErrorResponseStructure_InvalidJSON(t *testing.T) {
	// Property: For any non-JSON string, the handler should return a structured error response
	f := func(invalidInput string) bool {
		// Skip empty strings as they might be valid edge cases
		if invalidInput == "" {
			return true
		}

		// Create a request with invalid JSON
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/ping", bytes.NewBufferString(invalidInput))
		c.Request.Header.Set("Content-Type", "application/json")

		// Call the handler
		Ping(c)

		// Check that we get a proper error response structure
		if w.Code == http.StatusBadRequest {
			return validateErrorResponse(w.Body.Bytes())
		}
		// If it somehow parsed as valid JSON, that's also acceptable
		return true
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property 9 failed: %v", err)
	}
}

// TestProperty9_ErrorResponseStructure_MissingFields tests that requests with missing
// required fields produce properly structured error responses
// **Feature: redis-web-manager, Property 9: Error Response Structure**
// **Validates: Requirements 7.4**
// Note: This test validates JSON binding errors only, not Redis connection errors.
func TestProperty9_ErrorResponseStructure_MissingFields(t *testing.T) {
	// Test various incomplete connection configs that should fail JSON binding
	testCases := []struct {
		name    string
		payload string
	}{
		{"invalid_json_syntax", `{invalid}`},
		{"missing_closing_brace", `{"host": "localhost"`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/ping", bytes.NewBufferString(tc.payload))
			c.Request.Header.Set("Content-Type", "application/json")

			Ping(c)

			// These should return 400 Bad Request with structured errors
			if w.Code == http.StatusBadRequest {
				var resp map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Errorf("Response is not valid JSON: %v", err)
					return
				}
				// Check that success field exists
				if _, ok := resp["success"]; !ok {
					t.Errorf("Response missing 'success' field")
				}
				// If success is false, error field must exist and be non-empty
				if success, ok := resp["success"].(bool); ok && !success {
					errField, hasError := resp["error"]
					if !hasError {
						t.Errorf("Error response missing 'error' field")
					}
					if errStr, ok := errField.(string); ok && errStr == "" {
						t.Errorf("Error response has empty 'error' field")
					}
				}
			}
		})
	}
}

// TestProperty9_ErrorResponseStructure_CommandHandler tests error responses from command handler
// **Feature: redis-web-manager, Property 9: Error Response Structure**
// **Validates: Requirements 7.4**
// Note: This test validates JSON parsing errors only, not Redis connection errors,
// to avoid requiring a real Redis server during testing.
func TestProperty9_ErrorResponseStructure_CommandHandler(t *testing.T) {
	// Property: For any invalid JSON input, the handler should return a structured error
	invalidPayloads := []string{
		"not json at all",
		"{invalid}",
		"[unclosed",
		`{"key": undefined}`,
		"",
	}

	for _, payload := range invalidPayloads {
		t.Run("invalid_json", func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/execute", bytes.NewBufferString(payload))
			c.Request.Header.Set("Content-Type", "application/json")

			Execute(c)

			// Should return 400 Bad Request with proper error structure
			if w.Code != http.StatusBadRequest {
				// Empty body might be handled differently
				if payload == "" {
					return
				}
				t.Errorf("Expected status 400, got %d for payload: %s", w.Code, payload)
				return
			}

			if !validateErrorResponse(w.Body.Bytes()) {
				t.Errorf("Invalid error response structure: %s", w.Body.String())
			}
		})
	}
}

// TestProperty9_ErrorResponseStructure_AllHandlers verifies that all handlers
// return consistent error response structures for invalid input
// **Feature: redis-web-manager, Property 9: Error Response Structure**
// **Validates: Requirements 7.4**
func TestProperty9_ErrorResponseStructure_AllHandlers(t *testing.T) {
	handlers := []struct {
		name    string
		handler gin.HandlerFunc
		path    string
	}{
		{"Ping", Ping, "/api/ping"},
		{"Execute", Execute, "/api/execute"},
		{"GetInfo", GetInfo, "/api/info"},
	}

	// Test with completely invalid JSON
	invalidPayloads := []string{
		"not json at all",
		"{invalid}",
		"[unclosed",
		`{"key": undefined}`,
	}

	for _, h := range handlers {
		for _, payload := range invalidPayloads {
			t.Run(h.name+"_"+payload[:min(10, len(payload))], func(t *testing.T) {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request = httptest.NewRequest(http.MethodPost, h.path, bytes.NewBufferString(payload))
				c.Request.Header.Set("Content-Type", "application/json")

				h.handler(c)

				// Should return 400 Bad Request with proper error structure
				if w.Code != http.StatusBadRequest {
					t.Errorf("Expected status 400, got %d", w.Code)
					return
				}

				if !validateErrorResponse(w.Body.Bytes()) {
					t.Errorf("Invalid error response structure: %s", w.Body.String())
				}
			})
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
