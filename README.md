# Passkey Origin Validator

A tool for validating passkey/WebAuthn origin constraints in .well-known/webauthn endpoints. This tool is based on the Chromium project's implementation of WebAuthn security checking and helps ensure that your WebAuthn implementation follows the same constraints as browsers.

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

## Background

The Chromium project's WebAuthn implementation includes a security check that limits the number of unique labels in a .well-known/webauthn endpoint to 5. This is to prevent potential security issues with domains that try to authorize too many different origins.

### Label Definition

According to the passkey specification, a label is defined as the name directly preceding the Effective Top-Level Domain (ETLD). In other words, the label is the +1 part of the ETLD+1.

For example:
- For "example.com", the ETLD is ".com", and the label is "example"
- For "test.example.org", the ETLD is ".org", and the label is "example"
- For "one.thing.com", the ETLD is ".com", and the label is "thing"
- For "one.anotherthing.com", the ETLD is ".com", and the label is "anotherthing"

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

## Dependencies

- [Go](https://golang.org/) (version 1.24 or later)
- [Cobra](https://github.com/spf13/cobra) - A Commander for modern Go CLI interactions
- [Viper](https://github.com/spf13/viper) - Go configuration with fangs

These dependencies will be automatically installed when running `make deps`.

## Building and Running

The project uses a Makefile for building and running:

```bash
# Get dependencies
make deps

# Build the application
make build

# Count labels for default domain (webauthn.io)
make run

# Count labels with debug logging
make run DEBUG=true

# Count labels for specific domain
make run DOMAIN=google.com

# Count labels from local file
make run FILE=./test.json

# Validate origin against default domain
make validate ORIGIN=https://example.com

# Validate origin against specific domain
make validate ORIGIN=https://example.com DOMAIN=google.com

# Validate origin against local file
make validate ORIGIN=https://example.com FILE=./test.json

# Run with mock data for testing
make mock

# Clean build artifacts
make clean

# Run tests
make test

# Show help
make help
```

You can also run the built binary directly:

```bash
# Show help
./build/passkey-origin-validator --help

# Count labels for default domain (webauthn.io)
./build/passkey-origin-validator count

# Count labels with debug logging
./build/passkey-origin-validator count --debug

# Count labels for specific domain
./build/passkey-origin-validator count example.com

# Count labels from local file
./build/passkey-origin-validator count --file ./test.json

# Validate origin against default domain
./build/passkey-origin-validator validate --origin https://example.com

# Validate origin against specific domain
./build/passkey-origin-validator validate --origin https://example.com google.com

# Validate origin against local file
./build/passkey-origin-validator validate --origin https://example.com --file ./test.json

# Run with mock data for testing
./build/passkey-origin-validator --mock
```

## CI/CD Pipeline

This project uses GitHub Actions for automated testing and releasing. The workflow is configured in the `.github/workflows/ci.yml` file and consists of two jobs:

1. **Test Job**: Runs the project's tests on every pull request and commit to the master branch.
2. **Release Job**: Creates a release binary using GoReleaser when a tag is pushed to the repository.

The workflow requires write permissions to the repository contents to create releases. These permissions are configured in the workflow file.

### Creating a Release

To create a new release:

1. Ensure all your changes are committed and pushed to the master branch
2. Create and push a new tag:

```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

This will trigger the release job in the GitHub Actions workflow, which will:
- Build binaries for multiple platforms (Linux, macOS, Windows)
- Create archives of the binaries
- Generate checksums
- Create a GitHub release with the binaries and changelog

The release configuration is defined in the `.goreleaser.yml` file.

## Project Structure

- `cmd/passkey-origin-validator/main.go` - Entry point for the application
- `cmd/passkey-origin-validator/cmd/` - Command-line interface using Cobra
  - `root.go` - Root command and global flags
  - `count.go` - Command for counting labels
  - `validate.go` - Command for validating origins
  - `mock.go` - Mock data functionality
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

## License

This project is licensed under the MIT License - see the LICENSE file for details.
