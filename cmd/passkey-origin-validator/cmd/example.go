package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/developmeh/passkey-origin-validator/internal/counter"
	"net/url"
	"strings"
	"text/tabwriter"
	"os"
)

// normalizeJSON takes a JSON byte array and returns a normalized version with consistent indentation
func normalizeJSON(data []byte) ([]byte, error) {
	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "    ") // Use 4 spaces for indentation
	if err := encoder.Encode(obj); err != nil {
		return nil, err
	}

	// Remove trailing newline added by json.Encoder
	result := buf.Bytes()
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}

	return result, nil
}

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

	// Mock JSON with ccTLDs (country code top-level domains)
	mockJSON3 := []byte(`{
    "origins": [
        "https://example.co.uk",
        "https://example.de",
        "https://example-rewards.com",
        "https://shop.example.fr",
        "https://blog.example.jp",
        "https://support.example.ca",
        "https://news.example.au"
    ]
}`)

	// Test case 1: Under the limit
	fmt.Println("\nTest case 1: Under the limit (3 labels)")
	result1, err := parseAndCountLabels(mockJSON1)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	displaySideBySide(mockJSON1, result1)

	// Test case 2: Over the limit
	fmt.Println("\nTest case 2: Over the limit (6 labels)")
	result2, err := parseAndCountLabels(mockJSON2)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	displaySideBySide(mockJSON2, result2)

	// Test case 3: ccTLDs (country code top-level domains)
	fmt.Println("\nTest case 3: ccTLDs (country code top-level domains)")
	result3, err := parseAndCountLabels(mockJSON3)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	if debug {
		fmt.Println("Debug: ccTLD mock JSON:")
		fmt.Println(string(mockJSON3))
		fmt.Printf("Debug: Found %d unique labels\n", result3.Count)
		fmt.Printf("Debug: Labels: %v\n", result3.LabelsFound)
		fmt.Printf("Debug: Exceeds limit: %v\n", result3.ExceedsLimit)
	}
	displaySideBySide(mockJSON3, result3)

	// Test validation
	fmt.Println("\nTest case 4: Validation (success)")

	// Parse and count labels for test case 4
	result4, err := parseAndCountLabels(mockJSON1)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Debug logging for test case 4
	if debug {
		fmt.Printf("Debug: Test case 4 - Found %d unique labels\n", result4.Count)
		fmt.Printf("Debug: Labels: %v\n", result4.LabelsFound)
		fmt.Printf("Debug: Exceeds limit: %v\n", result4.ExceedsLimit)
	}

	// Display side by side
	displaySideBySide(mockJSON1, result4)

	// Show validation results
	fmt.Println("\nValidation Results:")
	status1 := counter.ValidateWellKnownJSON("https://example.com", mockJSON1)
	fmt.Printf("Validating caller origin: https://example.com\nStatus: %s\n", status1)

	fmt.Println("\nTest case 5: Validation (failure)")

	// Parse and count labels for test case 5
	result5, err := parseAndCountLabels(mockJSON1)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Debug logging for test case 5
	if debug {
		fmt.Printf("Debug: Test case 5 - Found %d unique labels\n", result5.Count)
		fmt.Printf("Debug: Labels: %v\n", result5.LabelsFound)
		fmt.Printf("Debug: Exceeds limit: %v\n", result5.ExceedsLimit)
	}

	// Display side by side
	displaySideBySide(mockJSON1, result5)

	// Show validation results
	fmt.Println("\nValidation Results:")
	status2 := counter.ValidateWellKnownJSON("https://unknown.com", mockJSON1)
	fmt.Printf("Validating caller origin: https://unknown.com\nStatus: %s\n", status2)
}

// displaySideBySide displays the WebAuthn response and label output side by side
func displaySideBySide(jsonData []byte, result *counter.LabelCount) {
	// Create a new tabwriter
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

	// Print header
	fmt.Fprintln(w, "WebAuthn Response\tLabel Analysis")
	fmt.Fprintln(w, "----------------\t-------------")

	// Normalize the JSON for consistent display
	normalizedJSON, err := normalizeJSON(jsonData)
	if err != nil {
		if debug {
			fmt.Printf("Debug: Error normalizing JSON: %v\n", err)
		}
		normalizedJSON = jsonData // Fallback to original if normalization fails
	}

	// Split the JSON and result into lines
	jsonLines := strings.Split(string(normalizedJSON), "\n")
	resultLines := strings.Split(counter.FormatResults(result), "\n")

	// Determine the maximum number of lines
	maxLines := len(jsonLines)
	if len(resultLines) > maxLines {
		maxLines = len(resultLines)
	}

	// Print each line side by side
	for i := 0; i < maxLines; i++ {
		jsonLine := ""
		if i < len(jsonLines) {
			jsonLine = jsonLines[i]
		}

		resultLine := ""
		if i < len(resultLines) {
			resultLine = resultLines[i]
		}

		fmt.Fprintf(w, "%s\t%s\n", jsonLine, resultLine)
	}

	// Flush the tabwriter
	w.Flush()

	if debug {
		fmt.Println("\nDebug: Side-by-side display")
		fmt.Printf("Debug: JSON has %d lines, Results has %d lines\n", len(jsonLines), len(resultLines))
		fmt.Printf("Debug: Found %d unique labels\n", result.Count)
		fmt.Printf("Debug: Labels: %v\n", result.LabelsFound)
		fmt.Printf("Debug: Exceeds limit: %v\n", result.ExceedsLimit)
		fmt.Println("Debug: Normalized JSON:")
		fmt.Println(string(normalizedJSON))
	}
}

// parseAndCountLabels parses mock JSON data and counts the labels.
func parseAndCountLabels(jsonData []byte) (*counter.LabelCount, error) {
	// Normalize the JSON for consistent display
	normalizedJSON, err := normalizeJSON(jsonData)
	if err != nil {
		if debug {
			fmt.Printf("Debug: Error normalizing JSON in parseAndCountLabels: %v\n", err)
		}
		normalizedJSON = jsonData // Fallback to original if normalization fails
	}

	// Create a mock result
	result := &counter.LabelCount{
		URL:          "https://mock-domain.com/.well-known/webauthn",
		UniqueLabels: make(map[string]bool),
		RawJSON:      string(normalizedJSON),
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
