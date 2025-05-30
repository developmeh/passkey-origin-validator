// Package counter provides functionality to count unique labels in a .well-known/webauthn endpoint.
package counter

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"
)

const (
	// MaxLabels is the maximum number of unique labels allowed in a .well-known/webauthn endpoint.
	MaxLabels = 5
	// WellKnownPath is the path to the .well-known/webauthn endpoint.
	WellKnownPath = "/.well-known/webauthn"
	// MaxBodySize is the maximum size of the response body in bytes.
	MaxBodySize = 1 << 18 // 256KB
	// Timeout is the timeout for the HTTP request.
	Timeout = 10 * time.Second
)

// AuthenticatorStatus represents the status of a WebAuthn authentication request.
type AuthenticatorStatus int

const (
	// StatusSuccess indicates that the authentication request was successful.
	StatusSuccess AuthenticatorStatus = iota
	// StatusBadRelyingPartyIDJSONParseError indicates that the relying party ID JSON could not be parsed.
	StatusBadRelyingPartyIDJSONParseError
	// StatusBadRelyingPartyIDNoJSONMatch indicates that the relying party ID JSON did not match the caller origin.
	StatusBadRelyingPartyIDNoJSONMatch
	// StatusBadRelyingPartyIDNoJSONMatchHitLimits indicates that the relying party ID JSON did not match the caller origin and hit the label limit.
	StatusBadRelyingPartyIDNoJSONMatchHitLimits
)

// String returns a string representation of the AuthenticatorStatus.
func (s AuthenticatorStatus) String() string {
	switch s {
	case StatusSuccess:
		return "SUCCESS"
	case StatusBadRelyingPartyIDJSONParseError:
		return "BAD_RELYING_PARTY_ID_JSON_PARSE_ERROR"
	case StatusBadRelyingPartyIDNoJSONMatch:
		return "BAD_RELYING_PARTY_ID_NO_JSON_MATCH"
	case StatusBadRelyingPartyIDNoJSONMatchHitLimits:
		return "BAD_RELYING_PARTY_ID_NO_JSON_MATCH_HIT_LIMITS"
	default:
		return fmt.Sprintf("UNKNOWN_STATUS(%d)", s)
	}
}

// WebAuthnResponse represents the JSON structure of a .well-known/webauthn response.
type WebAuthnResponse struct {
	Origins []string `json:"origins"`
}

// LabelCount represents the count of unique labels found in a .well-known/webauthn endpoint.
type LabelCount struct {
	URL          string
	UniqueLabels map[string]bool
	Count        int
	ExceedsLimit bool
	LabelsFound  []string
	ErrorMessage string
	RawJSON      string
}

// getLabel extracts the eTLD+1 label from a domain using the publicsuffix package.
// This mirrors the behavior of net::registry_controlled_domains::GetDomainAndRegistry in Chromium.
func getLabel(domain string) (string, error) {
	// Find the first dot in the eTLD+1
	dotIndex := strings.Index(domain, ".")
	if dotIndex == -1 {
		// If there's no dot, domain isn't valid and we don't care
		return domain, errors.New("Skip Domain not valid")
	}

	// Get the eTLD+1 using the publicsuffix package
	tld, _ := publicsuffix.PublicSuffix(domain)

	// Extract the label (the part before the first dot)
	label := strings.TrimSuffix(domain, tld)
	return label, nil
}

// CountLabels fetches the .well-known/webauthn endpoint for the given domain and counts the unique labels.
func CountLabels(domain string) (*LabelCount, error) {
	// Ensure domain is properly formatted
	if !strings.HasPrefix(domain, "https://") && !strings.HasPrefix(domain, "http://") {
		domain = "https://" + domain
	}

	// Parse the domain to ensure it's valid
	parsedURL, err := url.Parse(domain)
	if err != nil {
		return nil, fmt.Errorf("invalid domain: %w", err)
	}

	// Construct the well-known URL
	wellKnownURL := parsedURL.Scheme + "://" + parsedURL.Host + WellKnownPath

	// Create a client with a timeout
	client := &http.Client{
		Timeout: Timeout,
	}

	// Make the request
	resp, err := client.Get(wellKnownURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch well-known URL: %w", err)
	}
	defer resp.Body.Close()

	// Check if the response is successful
	if resp.StatusCode != http.StatusOK {
		return &LabelCount{
			URL:          wellKnownURL,
			ErrorMessage: fmt.Sprintf("HTTP request failed with status code: %d", resp.StatusCode),
		}, nil
	}

	// Check if the content type is JSON
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		return &LabelCount{
			URL:          wellKnownURL,
			ErrorMessage: fmt.Sprintf("unexpected content type: %s", contentType),
		}, nil
	}

	// Read the response body with a size limit
	bodyReader := io.LimitReader(resp.Body, MaxBodySize)
	body, err := io.ReadAll(bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Store the raw JSON
	rawJSON := string(body)

	// Parse the JSON
	var webAuthnResp WebAuthnResponse
	if err := json.Unmarshal(body, &webAuthnResp); err != nil {
		return &LabelCount{
			URL:          wellKnownURL,
			ErrorMessage: fmt.Sprintf("failed to parse JSON: %s", err),
			RawJSON:      rawJSON,
		}, nil
	}

	// Count unique labels
	result := &LabelCount{
		URL:          wellKnownURL,
		UniqueLabels: make(map[string]bool),
		RawJSON:      rawJSON,
	}

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

		// Extract the eTLD+1 label using publicsuffix package
		label, err := getLabel(domain)
		if err != nil {
			// Skip this origin if we can't extract the label
			continue
		}
		if !result.UniqueLabels[label] {
			result.UniqueLabels[label] = true
			result.LabelsFound = append(result.LabelsFound, label)
		}
	}

	result.Count = len(result.UniqueLabels)
	result.ExceedsLimit = result.Count > MaxLabels

	return result, nil
}

// ValidateWellKnownJSON validates if a caller origin is authorized by a relying party's .well-known/webauthn file.
// This function is based on the Chromium implementation of ValidateWellKnownJSON.
// It checks if the caller origin is in the list of authorized origins in the .well-known/webauthn file.
// It also enforces a limit on the number of unique eTLD+1 labels (MaxLabels) that can be processed.
// If the limit is reached before finding the caller origin, it returns StatusBadRelyingPartyIDNoJSONMatchHitLimits.
func ValidateWellKnownJSON(callerOrigin string, jsonData []byte) AuthenticatorStatus {
	// Parse the JSON
	var webAuthnResp WebAuthnResponse
	if err := json.Unmarshal(jsonData, &webAuthnResp); err != nil {
		return StatusBadRelyingPartyIDJSONParseError
	}

	// Check if the origins array exists
	if webAuthnResp.Origins == nil {
		return StatusBadRelyingPartyIDJSONParseError
	}

	// Parse the caller origin
	callerURL, err := url.Parse(callerOrigin)
	if err != nil {
		return StatusBadRelyingPartyIDNoJSONMatch
	}

	// Count unique labels and check if the caller origin is authorized
	uniqueLabels := make(map[string]bool)
	hitLimits := false

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

		// Extract the eTLD+1 label using publicsuffix package
		etldPlus1Label, err := getLabel(domain)
		if err != nil {
			// Skip this origin if we can't extract the label
			continue
		}

		if !uniqueLabels[etldPlus1Label] {
			if len(uniqueLabels) >= MaxLabels {
				hitLimits = true
				continue
			}
			uniqueLabels[etldPlus1Label] = true
		}

		// Check if the origin matches the caller origin
		if originURL.Scheme == callerURL.Scheme && originURL.Host == callerURL.Host {
			return StatusSuccess
		}
	}

	if hitLimits {
		return StatusBadRelyingPartyIDNoJSONMatchHitLimits
	}
	return StatusBadRelyingPartyIDNoJSONMatch
}

// CountLabelsFromFile reads a JSON file and counts the unique labels.
func CountLabelsFromFile(filePath string) (*LabelCount, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read the file content with a size limit
	bodyReader := io.LimitReader(file, MaxBodySize)
	body, err := io.ReadAll(bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Store the raw JSON
	rawJSON := string(body)

	// Parse the JSON
	var webAuthnResp WebAuthnResponse
	if err := json.Unmarshal(body, &webAuthnResp); err != nil {
		return &LabelCount{
			URL:          filePath,
			ErrorMessage: fmt.Sprintf("failed to parse JSON: %s", err),
			RawJSON:      rawJSON,
		}, nil
	}

	// Count unique labels
	result := &LabelCount{
		URL:          filePath,
		UniqueLabels: make(map[string]bool),
		RawJSON:      rawJSON,
	}

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

		// Extract the eTLD+1 label using publicsuffix package
		label, err := getLabel(domain)
		if err != nil {
			// Skip this origin if we can't extract the label
			continue
		}

		if !result.UniqueLabels[label] {
			result.UniqueLabels[label] = true
			result.LabelsFound = append(result.LabelsFound, label)
		}
	}

	result.Count = len(result.UniqueLabels)
	result.ExceedsLimit = result.Count > MaxLabels

	return result, nil
}

// FormatResults formats the label count results into a human-readable string.
func FormatResults(result *LabelCount) string {
	if result.ErrorMessage != "" {
		return fmt.Sprintf("Error: %s\nURL: %s", result.ErrorMessage, result.URL)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("URL: %s\n", result.URL))
	sb.WriteString(fmt.Sprintf("Unique labels found: %d\n", result.Count))

	if result.ExceedsLimit {
		sb.WriteString(fmt.Sprintf("WARNING: The number of unique labels exceeds the maximum limit of %d!\n", MaxLabels))
	}

	sb.WriteString("Labels found:\n")
	for _, label := range result.LabelsFound {
		sb.WriteString(fmt.Sprintf("- %s\n", label))
	}

	return sb.String()
}
