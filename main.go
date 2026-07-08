// Command cloudwrap fetches configuration and secrets from AWS SSM Parameter
// Store and either prints them or injects them into a wrapped command as
// environment variables.
package main

import (
	"os"

	"github.com/pennsieve/cloudwrap/cmd"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	cmd.SetVersion(version)
	os.Exit(cmd.Execute())
}
