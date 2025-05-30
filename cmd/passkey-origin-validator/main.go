// Command passkey-origin-validator is a tool for validating passkey/WebAuthn origin constraints in .well-known/webauthn endpoints.
package main

import (
	"github.com/developmeh/passkey-origin-validator/cmd/passkey-origin-validator/cmd"
)

// Version information set by goreleaser
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.SetVersionInfo(version, commit, date)
	cmd.Execute()
}
