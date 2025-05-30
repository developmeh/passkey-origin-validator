package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Used for flags
	cfgFile string
	debug   bool
	file    string
	mock    bool

 // rootCmd represents the base command when called without any subcommands
 rootCmd = &cobra.Command{
 	Use:   "passkey-origin-validator",
 	Short: "A tool for validating passkey/WebAuthn origin constraints in .well-known/webauthn endpoints",
 	Long: `A tool for validating passkey/WebAuthn origin constraints in .well-known/webauthn endpoints.
 This tool is based on the Chromium project's implementation of WebAuthn security checking.

 It can fetch the .well-known/webauthn endpoint for a given domain, parse the JSON response,
 and count the number of unique labels. It can also validate if a caller origin is authorized
 by a relying party's .well-known/webauthn file, following the same constraints as browsers.`,
 	Run: func(cmd *cobra.Command, args []string) {
 		// Check if we're running with mock data
 		if mock {
 			runMockData()
 			return
 		}

 		// If no command is specified, show help
 		cmd.Help()
 	},
 }
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.passkey-origin-validator.yaml)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging")
	rootCmd.PersistentFlags().StringVar(&file, "file", "", "Use a local JSON file instead of fetching from a domain")
	rootCmd.PersistentFlags().BoolVar(&mock, "mock", false, "Run with mock data for testing")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".passkey-origin-validator" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".passkey-origin-validator")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if debug {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		}
	}
}
