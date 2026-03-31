// Package app provides the public entry point for building custom ColonyOS binaries.
// Import this package along with your desired plugins to create a custom server.
package app

import "github.com/colonyos/colonies/internal/cli"

// Execute runs the ColonyOS CLI application. Call this from your main() function
// after importing your desired plugins via blank imports.
func Execute() {
	cli.Execute()
}
