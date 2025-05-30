# Passkey Origin Validator

A tool for validating passkey/WebAuthn origin constraints in .well-known/webauthn endpoints. This tool is based on the Chromium project's implementation of WebAuthn security checking and helps ensure that your WebAuthn implementation follows the same constraints as browsers.

For detailed information about the project background, label definition, and technical details, please see [project_status.md](project_status.md).

## Installation and Dependencies

### Dependencies

- [Go](https://golang.org/) (version 1.24 or later)
- [Cobra](https://github.com/spf13/cobra) - A Commander for modern Go CLI interactions
- [Viper](https://github.com/spf13/viper) - Go configuration with fangs

These dependencies will be automatically installed when running `make deps`.

### Building

```bash
# Get dependencies
make deps

# Build the application
make build

# Clean build artifacts
make clean

# Run tests
make test
```

## Command Reference

The tool provides several commands and flags to validate WebAuthn origin constraints. Here's a comprehensive guide to using each command:

### Global Flags

These flags can be used with any command:

| Flag | Description |
|------|-------------|
| `--config <file>` | Config file (default is $HOME/.passkey-origin-validator.yaml) |
| `--debug` | Enable debug logging |
| `--file <file>` | Use a local JSON file instead of fetching from a domain |
| `--example` | Run with example data for testing |
| `--version`, `-v` | Print version information and exit |

### Count Command

The `count` command fetches the .well-known/webauthn endpoint for a given domain, parses the JSON response, and counts the number of unique labels.

**Usage:**
```
passkey-origin-validator count [domain]
```

**Arguments:**
- `domain` (optional): The domain to check. If not provided, defaults to webauthn.io.

**Examples:**
```bash
# Count labels for default domain (webauthn.io)
./build/passkey-origin-validator count

# Count labels with debug logging
./build/passkey-origin-validator count --debug

# Count labels for specific domain
./build/passkey-origin-validator count example.com

# Count labels from local file
./build/passkey-origin-validator count --file ./test.json
```

**Using with Makefile:**
```bash
# Count labels for default domain (webauthn.io)
make run

# Count labels with debug logging
make run DEBUG=true

# Count labels for specific domain
make run DOMAIN=google.com

# Count labels from local file
make run FILE=./test.json
```

### Validate Command

The `validate` command checks if a caller origin is authorized by a domain's .well-known/webauthn file.

**Usage:**
```
passkey-origin-validator validate [domain] --origin <origin>
```

**Arguments:**
- `domain` (optional): The domain to check. If not provided, defaults to webauthn.io.

**Required Flags:**
- `--origin <origin>`: The caller origin to validate (e.g., https://example.com)

**Examples:**
```bash
# Validate origin against default domain
./build/passkey-origin-validator validate --origin https://example.com

# Validate origin against specific domain
./build/passkey-origin-validator validate --origin https://example.com google.com

# Validate origin against local file
./build/passkey-origin-validator validate --origin https://example.com --file ./test.json
```

**Using with Makefile:**
```bash
# Validate origin against default domain
make validate ORIGIN=https://example.com

# Validate origin against specific domain
make validate ORIGIN=https://example.com DOMAIN=google.com

# Validate origin against local file
make validate ORIGIN=https://example.com FILE=./test.json
```

### Example Data

You can run the tool with example data to see how it works without making actual HTTP requests:

```bash
# Run with example data
./build/passkey-origin-validator --example

# Using Makefile
make mock
```

This will demonstrate the functionality with predefined test cases, showing both successful and failed validations.

## Configuration

The tool can be configured using a YAML configuration file. By default, it looks for a file named `.passkey-origin-validator.yaml` in your home directory. You can specify a different configuration file using the `--config` flag.

### Configuration File Format

The configuration file uses YAML format and supports the following options:

| Option | Type | Description |
|--------|------|-------------|
| `debug` | boolean | Enable debug logging |
| `default_domain` | string | Default domain to check if not specified |
| `file` | string | Use a local JSON file instead of fetching from a domain |
| `example` | boolean | Run with example data for testing |
| `origin` | string | Default caller origin to validate (for validate command) |
| `timeout` | integer | HTTP request timeout in seconds |
| `max_labels` | integer | Maximum number of labels allowed |

### Sample Configuration File

A sample configuration file is provided in the repository as `sample-config.yaml`. You can copy this file to your home directory and customize it:

```bash
# Copy the sample config to your home directory
cp sample-config.yaml ~/.passkey-origin-validator.yaml
```

Here's the content of the sample configuration file:

```yaml
# Sample configuration file for passkey-origin-validator
# Save this as $HOME/.passkey-origin-validator.yaml or specify with --config flag

# Enable debug logging
debug: false

# Default domain to check if not specified
default_domain: "https://webauthn.io"

# Use a local JSON file instead of fetching from a domain
# file: "./test.json"

# Run with example data for testing
example: false

# Default caller origin to validate (for validate command)
# origin: "https://example.com"

# HTTP request timeout in seconds
timeout: 10

# Maximum number of labels allowed
max_labels: 5
```

### Using the Configuration File

To use the configuration file:

1. Create a YAML file with your desired configuration options
2. Save it as `.passkey-origin-validator.yaml` in your home directory, or
3. Specify the path to your config file with the `--config` flag:

```bash
./build/passkey-origin-validator --config /path/to/your/config.yaml count
```

Configuration values in the file can be overridden by command-line flags. For example, if your config file has `debug: false` but you run with `--debug`, debug logging will be enabled for that run.

## Debugging

The tool provides debug logging that can be enabled with the `--debug` flag or by setting `DEBUG=true` when using the Makefile. Debug logging provides additional information about:

- The domain being tested
- The maximum number of labels allowed
- The number of unique labels found
- The list of labels found
- Whether the number of labels exceeds the limit
- JSON parsing details

Example:
```bash
# Enable debug logging with direct command
./build/passkey-origin-validator count --debug

# Enable debug logging with Makefile
make run DEBUG=true
```

## Exit Status

The tool returns different exit codes depending on the result:

| Exit Code | Description |
|-----------|-------------|
| `0` | Success (number of labels is within the limit) |
| `1` | Error (failed to fetch or parse the .well-known/webauthn endpoint) |
| `2` | Warning (number of labels exceeds the limit) |
| `3` | Validation failure (caller origin is not authorized) |

## CI/CD Pipeline

This project uses GitHub Actions for automated testing and releasing. The workflow is configured in the `.github/workflows/ci.yml` file and consists of two jobs:

1. **Test Job**: Runs the project's tests on every pull request and commit to the master branch.
2. **Release Job**: Creates a release binary using GoReleaser when a tag is pushed to the repository.

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

## License

This project is licensed under the MIT License - see the LICENSE file for details.
