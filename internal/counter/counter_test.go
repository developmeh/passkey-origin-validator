package counter

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestCountLabels(t *testing.T) {
	// Test case 1: Valid JSON with 3 unique labels
	t.Run("Valid JSON with 3 unique labels", func(t *testing.T) {
		// Create a test server that returns a valid JSON response
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"origins": [
					"https://example.com",
					"https://test.example.org",
					"https://another.example.net",
					"https://subdomain.example.com"
				]
			}`))
		}))
		defer server.Close()

		// Call the CountLabels function with the test server URL
		result, err := CountLabels(server.URL)
		if err != nil {
			t.Fatalf("CountLabels returned an error: %v", err)
		}

		// Check the results
		if result.Count != 4 {
			t.Errorf("Expected 4 unique labels, got %d", result.Count)
		}
		if result.ExceedsLimit {
			t.Errorf("Expected ExceedsLimit to be false, got true")
		}
		expectedLabels := []string{"example", "test", "another", "subdomain"}
		for _, label := range expectedLabels {
			if !result.UniqueLabels[label] {
				t.Errorf("Expected label %s to be in UniqueLabels", label)
			}
		}
	})

	// Test case 2: Valid JSON with more than 5 unique labels
	t.Run("Valid JSON with more than 5 unique labels", func(t *testing.T) {
		// Create a test server that returns a valid JSON response with more than 5 unique labels
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"origins": [
					"https://one.example.com",
					"https://two.example.org",
					"https://three.example.net",
					"https://four.example.io",
					"https://five.example.co",
					"https://six.example.dev"
				]
			}`))
		}))
		defer server.Close()

		// Call the CountLabels function with the test server URL
		result, err := CountLabels(server.URL)
		if err != nil {
			t.Fatalf("CountLabels returned an error: %v", err)
		}

		// Check the results
		if result.Count != 6 {
			t.Errorf("Expected 6 unique labels, got %d", result.Count)
		}
		if !result.ExceedsLimit {
			t.Errorf("Expected ExceedsLimit to be true, got false")
		}
		expectedLabels := []string{"one", "two", "three", "four", "five", "six"}
		for _, label := range expectedLabels {
			if !result.UniqueLabels[label] {
				t.Errorf("Expected label %s to be in UniqueLabels", label)
			}
		}
	})

	// Test case 3: Invalid JSON
	t.Run("Invalid JSON", func(t *testing.T) {
		// Create a test server that returns an invalid JSON response
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"origins": [
					"https://example.com",
					"https://test.example.org",
				]`)) // Missing closing brace
		}))
		defer server.Close()

		// Call the CountLabels function with the test server URL
		result, err := CountLabels(server.URL)
		if err != nil {
			t.Fatalf("CountLabels returned an error: %v", err)
		}

		// Check the results
		if result.ErrorMessage == "" {
			t.Errorf("Expected an error message, got empty string")
		}
	})

	// Test case 4: Non-JSON content type
	t.Run("Non-JSON content type", func(t *testing.T) {
		// Create a test server that returns a non-JSON content type
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<html><body>Not JSON</body></html>`))
		}))
		defer server.Close()

		// Call the CountLabels function with the test server URL
		result, err := CountLabels(server.URL)
		if err != nil {
			t.Fatalf("CountLabels returned an error: %v", err)
		}

		// Check the results
		if result.ErrorMessage == "" {
			t.Errorf("Expected an error message, got empty string")
		}
		if !contains(result.ErrorMessage, "content type") {
			t.Errorf("Expected error message to contain 'content type', got %s", result.ErrorMessage)
		}
	})

	// Test case 5: HTTP error
	t.Run("HTTP error", func(t *testing.T) {
		// Create a test server that returns an HTTP error
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		// Call the CountLabels function with the test server URL
		result, err := CountLabels(server.URL)
		if err != nil {
			t.Fatalf("CountLabels returned an error: %v", err)
		}

		// Check the results
		if result.ErrorMessage == "" {
			t.Errorf("Expected an error message, got empty string")
		}
		if !contains(result.ErrorMessage, "404") {
			t.Errorf("Expected error message to contain '404', got %s", result.ErrorMessage)
		}
	})
}

func TestFormatResults(t *testing.T) {
	// Test case 1: Successful result
	t.Run("Successful result", func(t *testing.T) {
		result := &LabelCount{
			URL:          "https://example.com/.well-known/webauthn",
			UniqueLabels: map[string]bool{"example": true, "test": true, "another": true},
			Count:        3,
			ExceedsLimit: false,
			LabelsFound:  []string{"example", "test", "another"},
		}

		output := FormatResults(result)
		if !contains(output, "Unique labels found: 3") {
			t.Errorf("Expected output to contain 'Unique labels found: 3', got %s", output)
		}
		if contains(output, "WARNING") {
			t.Errorf("Expected output not to contain 'WARNING', got %s", output)
		}
	})

	// Test case 2: Result exceeding limit
	t.Run("Result exceeding limit", func(t *testing.T) {
		result := &LabelCount{
			URL:          "https://example.com/.well-known/webauthn",
			UniqueLabels: map[string]bool{"one": true, "two": true, "three": true, "four": true, "five": true, "six": true},
			Count:        6,
			ExceedsLimit: true,
			LabelsFound:  []string{"one", "two", "three", "four", "five", "six"},
		}

		output := FormatResults(result)
		if !contains(output, "Unique labels found: 6") {
			t.Errorf("Expected output to contain 'Unique labels found: 6', got %s", output)
		}
		if !contains(output, "WARNING") {
			t.Errorf("Expected output to contain 'WARNING', got %s", output)
		}
	})

	// Test case 3: Error result
	t.Run("Error result", func(t *testing.T) {
		result := &LabelCount{
			URL:          "https://example.com/.well-known/webauthn",
			ErrorMessage: "HTTP request failed with status code: 404",
		}

		output := FormatResults(result)
		if !contains(output, "Error") {
			t.Errorf("Expected output to contain 'Error', got %s", output)
		}
		if !contains(output, "404") {
			t.Errorf("Expected output to contain '404', got %s", output)
		}
	})
}

// TestValidateWellKnownJSON tests the ValidateWellKnownJSON function.
func TestValidateWellKnownJSON(t *testing.T) {
	tests := []struct {
		name         string
		callerOrigin string
		json         string
		expected     AuthenticatorStatus
	}{
		{
			name:         "Empty JSON",
			callerOrigin: "https://foo.com",
			json:         "[]",
			expected:     StatusBadRelyingPartyIDJSONParseError,
		},
		{
			name:         "Empty object",
			callerOrigin: "https://foo.com",
			json:         "{}",
			expected:     StatusBadRelyingPartyIDJSONParseError,
		},
		{
			name:         "Missing origins key",
			callerOrigin: "https://foo.com",
			json:         `{"foo": "bar"}`,
			expected:     StatusBadRelyingPartyIDJSONParseError,
		},
		{
			name:         "Origins not an array",
			callerOrigin: "https://foo.com",
			json:         `{"origins": "bar"}`,
			expected:     StatusBadRelyingPartyIDJSONParseError,
		},
		{
			name:         "Empty origins array",
			callerOrigin: "https://foo.com",
			json:         `{"origins": []}`,
			expected:     StatusBadRelyingPartyIDNoJSONMatch,
		},
		{
			name:         "Origins array with non-string",
			callerOrigin: "https://foo.com",
			json:         `{"origins": [1]}`,
			expected:     StatusBadRelyingPartyIDJSONParseError,
		},
		{
			name:         "Origins array with matching origin",
			callerOrigin: "https://foo.com",
			json:         `{"origins": ["https://foo.com"]}`,
			expected:     StatusSuccess,
		},
		{
			name:         "Origins array with non-matching origin",
			callerOrigin: "https://foo.com",
			json:         `{"origins": ["https://foo2.com"]}`,
			expected:     StatusBadRelyingPartyIDNoJSONMatch,
		},
		{
			name:         "Origins array with invalid domain",
			callerOrigin: "https://foo.com",
			json:         `{"origins": ["https://com"]}`,
			expected:     StatusBadRelyingPartyIDNoJSONMatch,
		},
		{
			name:         "Origins array with different scheme",
			callerOrigin: "https://foo.com",
			json:         `{"origins": ["other://foo.com"]}`,
			expected:     StatusBadRelyingPartyIDNoJSONMatch,
		},
		{
			name:         "Origins array with 5 different labels and matching origin",
			callerOrigin: "https://foo.com",
			json:         `{"origins": ["https://a.com", "https://b.com", "https://c.com", "https://d.com", "https://foo.com"]}`,
			expected:     StatusSuccess,
		},
		{
			name:         "Origins array with 6 different labels and matching origin at the end",
			callerOrigin: "https://foo.com",
			json:         `{"origins": ["https://a.com", "https://b.com", "https://c.com", "https://d.com", "https://e.com", "https://foo.com"]}`,
			expected:     StatusBadRelyingPartyIDNoJSONMatchHitLimits,
		},
		{
			name:         "Origins array with 6 different labels and matching origin in the middle",
			callerOrigin: "https://foo.com",
			json:         `{"origins": ["https://a.com", "https://b.com", "https://c.com", "https://d.com", "https://foo.com", "https://e.com"]}`,
			expected:     StatusSuccess,
		},
		{
			name:         "Origins array with different TLDs but same eTLD+1 label",
			callerOrigin: "https://foo.com",
			json:         `{"origins": ["https://foo.co.uk", "https://foo.de", "https://foo.in", "https://foo.net", "https://foo.org", "https://foo.com"]}`,
			expected:     StatusSuccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateWellKnownJSON(tt.callerOrigin, []byte(tt.json))
			if result != tt.expected {
				t.Errorf("ValidateWellKnownJSON(%q, %q) = %v, want %v", tt.callerOrigin, tt.json, result, tt.expected)
			}
		})
	}
}

// TestCountLabelsFromFile tests the CountLabelsFromFile function.
func TestCountLabelsFromFile(t *testing.T) {
	// Create a temporary file with valid JSON
	validJSON := `{
		"origins": [
			"https://example.com",
			"https://test.example.org",
			"https://another.example.net"
		]
	}`
	validFile, err := os.CreateTemp("", "valid-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(validFile.Name())
	if _, err := validFile.Write([]byte(validJSON)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	validFile.Close()

	// Create a temporary file with invalid JSON
	invalidJSON := `{
		"origins": [
			"https://example.com",
			"https://test.example.org",
		]` // Missing closing brace
	invalidFile, err := os.CreateTemp("", "invalid-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(invalidFile.Name())
	if _, err := invalidFile.Write([]byte(invalidJSON)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	invalidFile.Close()

	// Test case 1: Valid JSON file
	t.Run("Valid JSON file", func(t *testing.T) {
		result, err := CountLabelsFromFile(validFile.Name())
		if err != nil {
			t.Fatalf("CountLabelsFromFile returned an error: %v", err)
		}

		// Check the results
		if result.Count != 3 {
			t.Errorf("Expected 3 unique labels, got %d", result.Count)
		}
		if result.ExceedsLimit {
			t.Errorf("Expected ExceedsLimit to be false, got true")
		}
		expectedLabels := []string{"example", "test", "another"}
		for _, label := range expectedLabels {
			if !result.UniqueLabels[label] {
				t.Errorf("Expected label %s to be in UniqueLabels", label)
			}
		}
	})

	// Test case 2: Invalid JSON file
	t.Run("Invalid JSON file", func(t *testing.T) {
		result, err := CountLabelsFromFile(invalidFile.Name())
		if err != nil {
			t.Fatalf("CountLabelsFromFile returned an error: %v", err)
		}

		// Check the results
		if result.ErrorMessage == "" {
			t.Errorf("Expected an error message, got empty string")
		}
		if !contains(result.ErrorMessage, "parse JSON") {
			t.Errorf("Expected error message to contain 'parse JSON', got %s", result.ErrorMessage)
		}
	})

	// Test case 3: Non-existent file
	t.Run("Non-existent file", func(t *testing.T) {
		_, err := CountLabelsFromFile("non-existent-file.json")
		if err == nil {
			t.Errorf("Expected an error, got nil")
		}
		if !contains(err.Error(), "failed to open file") {
			t.Errorf("Expected error message to contain 'failed to open file', got %s", err.Error())
		}
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s != substr && s != "" && substr != "" && strings.Contains(s, substr)
}
