package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/developmeh/passkey-origin-validator/internal/counter"
	"net/url"
	"strings"
)

// runMockData demonstrates the functionality of the counter package with mock data.
func runMockData() {
	fmt.Println("Testing with mock data...")

	// Mock JSON with 3 unique labels (under the limit)
	mockJSON1 := []byte(`{
		"origins": [
			"https://example.com",
			"https://test.example.org",
			"https://another.example.net"
		]
	}`)

	// Mock JSON with 6 unique labels (over the limit)
	mockJSON2 := []byte(`{
		"origins": [
			"https://one.example.com",
			"https://two.example.org",
			"https://three.example.net",
			"https://four.example.io",
			"https://five.example.co",
			"https://six.example.dev"
		]
	}`)

	// Test case 1: Under the limit
	fmt.Println("\nTest case 1: Under the limit (3 labels)")
	result1, err := parseAndCountLabels(mockJSON1)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println(counter.FormatResults(result1))

	// Test case 2: Over the limit
	fmt.Println("\nTest case 2: Over the limit (6 labels)")
	result2, err := parseAndCountLabels(mockJSON2)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println(counter.FormatResults(result2))

	// Test validation
	fmt.Println("\nTest case 3: Validation (success)")
	status1 := counter.ValidateWellKnownJSON("https://example.com", mockJSON1)
	fmt.Printf("Validating caller origin: https://example.com\nStatus: %s\n", status1)

	fmt.Println("\nTest case 4: Validation (failure)")
	status2 := counter.ValidateWellKnownJSON("https://unknown.com", mockJSON1)
	fmt.Printf("Validating caller origin: https://unknown.com\nStatus: %s\n", status2)
}

// parseAndCountLabels parses mock JSON data and counts the labels.
func parseAndCountLabels(jsonData []byte) (*counter.LabelCount, error) {
	// Create a mock result
	result := &counter.LabelCount{
		URL:          "https://mock-domain.com/.well-known/webauthn",
		UniqueLabels: make(map[string]bool),
		RawJSON:      string(jsonData),
	}

	// Parse the JSON
	var webAuthnResp counter.WebAuthnResponse
	if err := json.Unmarshal(jsonData, &webAuthnResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Count unique labels
	for _, originStr := range webAuthnResp.Origins {
		originURL, err := url.Parse(originStr)
		if err != nil {
			continue
		}

		// Extract the domain
		domain := originURL.Host
		if domain == "" {
			continue
		}

		// Extract the eTLD+1 label (first part of the domain)
		parts := strings.Split(domain, ".")
		if len(parts) < 2 {
			continue
		}

		label := parts[0]
		if !result.UniqueLabels[label] {
			result.UniqueLabels[label] = true
			result.LabelsFound = append(result.LabelsFound, label)
		}
	}

	result.Count = len(result.UniqueLabels)
	result.ExceedsLimit = result.Count > counter.MaxLabels

	return result, nil
}
