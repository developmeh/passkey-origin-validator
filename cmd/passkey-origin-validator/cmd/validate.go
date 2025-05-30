package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/developmeh/passkey-origin-validator/internal/counter"
	"github.com/spf13/cobra"
)

var (
	// Origin is the caller origin to validate
	origin string
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate [domain]",
	Short: "Validate if a caller origin is authorized by a domain's .well-known/webauthn file",
	Long: `Validate if a caller origin is authorized by a domain's .well-known/webauthn file.

This command fetches the .well-known/webauthn endpoint for a given domain,
parses the JSON response, and checks if the specified caller origin is authorized.

If no domain is provided, it uses the default domain (webauthn.io).
If the --file flag is provided, it reads from the specified file instead.`,
	Run: func(cmd *cobra.Command, args []string) {
		if origin == "" {
			fmt.Fprintf(os.Stderr, "Error: --origin flag is required\n")
			os.Exit(1)
		}

		var result *counter.LabelCount
		var err error

		// Check if we're reading from a file
		if file != "" {
			if debug {
				fmt.Printf("Debug: Reading from file: %s\n", file)
				fmt.Printf("Debug: Validating caller origin: %s\n", origin)
			}
			result, err = counter.CountLabelsFromFile(file)
		} else {
			// Get the domain from command-line arguments or use the default
			domain := "https://webauthn.io"
			if len(args) > 0 {
				domain = args[0]
			}

			if debug {
				fmt.Printf("Debug: Testing domain: %s\n", domain)
				fmt.Printf("Debug: Validating caller origin: %s\n", origin)
			}

			result, err = counter.CountLabels(domain)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if result.ErrorMessage != "" {
			fmt.Fprintf(os.Stderr, "Error: %s\n", result.ErrorMessage)
			os.Exit(1)
		}

		// Parse the JSON response
		var webAuthnResp counter.WebAuthnResponse
		if err := json.Unmarshal([]byte(result.RawJSON), &webAuthnResp); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
			os.Exit(1)
		}

		// Validate the caller origin
		status := counter.ValidateWellKnownJSON(origin, []byte(result.RawJSON))

		// Print the results
		fmt.Printf("Validating caller origin: %s against domain: %s\n", origin, result.URL)
		fmt.Printf("Status: %s\n", status)

		// Exit with non-zero status if the validation failed
		if status != counter.StatusSuccess {
			os.Exit(3)
		}
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)

	// Local flags
	validateCmd.Flags().StringVar(&origin, "origin", "", "The caller origin to validate (required)")
	validateCmd.MarkFlagRequired("origin")
}
