# Passkey Origin Validator - Project Status

## Background

The Chromium project's WebAuthn implementation includes a security check that limits the number of unique labels in a .well-known/webauthn endpoint to 5. This is to prevent potential security issues with domains that try to authorize too many different origins.

### Label Definition

According to the passkey specification, a label is defined as the name directly preceding the Effective Top-Level Domain (ETLD). In other words, the label is the +1 part of the ETLD+1.

For example:
- For "example.com", the TLD is ".com", and the label is "example"
- For "test.example.org", the TLD is ".org", and the label is "test.example"
- For "one.thing.com", the TLD is ".com", and the label is "one.thing"
- For "one.anotherthing.com", the TLD is ".com", and the label is "one.anotherthing"

The tool uses the `golang.org/x/net/publicsuffix` package to determine the ETLD+1 for a domain, and then extracts the label from it.

From the Chromium code:
```cpp
constexpr size_t kMaxLabels = 5;
bool hit_limits = false;
base::flat_set<std::string> labels_seen;
// ...
if (!base::Contains(labels_seen, etld_plus_1_label)) {
  if (labels_seen.size() >= kMaxLabels) {
    hit_limits = true;
    continue;
  }
  labels_seen.insert(etld_plus_1_label);
}
```

This tool helps test whether a domain's .well-known/webauthn endpoint would pass this security check.

## Features

- Fetches the .well-known/webauthn endpoint for a given domain
- Parses the JSON response and extracts unique labels
- Counts the number of unique labels and warns if it exceeds the maximum limit (5)
- Provides detailed output with the list of labels found
- Validates if a caller origin is authorized by a relying party's .well-known/webauthn file
- Implements the same validation logic as the Chromium project
- Supports reading from local JSON files instead of fetching from domains
- Modern CLI with subcommands and flags using Cobra
- Configuration management using Viper
- Debug logging with feature flag support
- Non-zero exit status if the number of labels exceeds the limit

## Project Structure

- `cmd/passkey-origin-validator/main.go` - Entry point for the application
- `cmd/passkey-origin-validator/cmd/` - Command-line interface using Cobra
  - `root.go` - Root command and global flags
  - `count.go` - Command for counting labels
  - `validate.go` - Command for validating origins
  - `example.go` - Mock data functionality
- `internal/counter/` - Package for fetching and analyzing .well-known/webauthn endpoints
  - `counter.go` - Core functionality for counting labels and validating origins
  - `counter_test.go` - Tests for the counter package

## API Reference

### CountLabels

```
func CountLabels(domain string) (*LabelCount, error)
```

Fetches the .well-known/webauthn endpoint for the given domain and counts the unique labels.

### CountLabelsFromFile

```
func CountLabelsFromFile(filePath string) (*LabelCount, error)
```

Reads a JSON file and counts the unique labels, similar to CountLabels but with a file as input instead of a URL.

### ValidateWellKnownJSON

```
func ValidateWellKnownJSON(callerOrigin string, jsonData []byte) AuthenticatorStatus
```

Validates if a caller origin is authorized by a relying party's .well-known/webauthn file. This function is based on the Chromium implementation of `ValidateWellKnownJSON` and follows the same logic:

1. Parses the JSON data
2. Checks if the origins array exists
3. Parses the caller origin
4. Counts unique labels (eTLD+1) and checks if the caller origin is authorized
5. Returns the appropriate AuthenticatorStatus:
   - `StatusSuccess` - The caller origin is authorized
   - `StatusBadRelyingPartyIDJSONParseError` - The JSON could not be parsed
   - `StatusBadRelyingPartyIDNoJSONMatch` - The caller origin is not authorized
   - `StatusBadRelyingPartyIDNoJSONMatchHitLimits` - The caller origin is not authorized and the number of unique labels exceeds the limit

This function matches the test cases in the Chromium project's `WebAuthRequestSecurityCheckerWellKnownJSONTest`.

## Exit Status

- `0` - Success (number of labels is within the limit)
- `1` - Error (failed to fetch or parse the .well-known/webauthn endpoint)
- `2` - Warning (number of labels exceeds the limit)
- `3` - Validation failure (caller origin is not authorized)