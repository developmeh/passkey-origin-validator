package cmd

import (
	"fmt"
	"os"

	"github.com/developmeh/passkey-origin-validator/internal/counter"
	"github.com/spf13/cobra"
)

// countCmd represents the count command
var countCmd = &cobra.Command{
	Use:   "count [domain]",
	Short: "Count the unique labels in a .well-known/webauthn endpoint",
	Long: `Count the unique labels in a .well-known/webauthn endpoint.

This command fetches the .well-known/webauthn endpoint for a given domain,
parses the JSON response, and counts the number of unique labels.

If no domain is provided, it uses the default domain (webauthn.io).
If the --file flag is provided, it reads from the specified file instead.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if we're running with mock data
		if mock {
			runMockData()
			return
		}

		var result *counter.LabelCount
		var err error

		// Check if we're reading from a file
		if file != "" {
			if debug {
				fmt.Printf("Debug: Reading from file: %s\n", file)
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
				fmt.Printf("Debug: Max labels allowed: %d\n", counter.MaxLabels)
			}

			result, err = counter.CountLabels(domain)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Debug logging
		if debug && result.ErrorMessage == "" {
			fmt.Printf("Debug: Found %d unique labels\n", result.Count)
			fmt.Printf("Debug: Labels: %v\n", result.LabelsFound)
			fmt.Printf("Debug: Exceeds limit: %v\n", result.ExceedsLimit)
		}

		// Print the results
		fmt.Println(counter.FormatResults(result))

		// Exit with non-zero status if the number of labels exceeds the limit
		if result.ExceedsLimit {
			os.Exit(2)
		}
	},
}

func init() {
	rootCmd.AddCommand(countCmd)
}
