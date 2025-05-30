// Package main provides a test function for the counter package using mock data.
package main

import (
	"fmt"
	"github.com/developmeh/passkey-origin-validator/internal/counter"
	"os"
)

// TestWithMockData demonstrates the functionality of the counter package with mock data.
func TestWithMockData() {
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

// parseAndCountLabels parses JSON data and counts the labels using the counter package.
func parseAndCountLabels(jsonData []byte) (*counter.LabelCount, error) {
	// Create a temporary file to store the JSON data
	tempFile, err := os.CreateTemp("", "webauthn-*.json")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tempFile.Name()) // Clean up the temporary file when done

	// Write the JSON to the temporary file
	if _, err := tempFile.Write(jsonData); err != nil {
		return nil, fmt.Errorf("failed to write to temporary file: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return nil, fmt.Errorf("failed to close temporary file: %w", err)
	}

	// Use the counter package to count labels from the file
	result, err := counter.CountLabelsFromFile(tempFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to count labels: %w", err)
	}

	// Override the URL to indicate this is from example data
	result.URL = "https://example-data/.well-known/webauthn"

	return result, nil
}
