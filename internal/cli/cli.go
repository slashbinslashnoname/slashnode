// Package cli implements the slashnoded subcommands.
package cli

import (
	"fmt"
	"os"
)

// Version is populated from main at startup.
var Version = "dev"

// ANSI: bold red / dim / reset. Disabled when NO_COLOR is set.
const (
	ansiRed   = "\x1b[1;31m"
	ansiDim   = "\x1b[2m"
	ansiReset = "\x1b[0m"
)

func colorize(s, color string) string {
	if os.Getenv("NO_COLOR") != "" {
		return s
	}
	return color + s + ansiReset
}

// asciiSkull is the red skull — the SlashNode mascot.
const asciiSkull = `     .-------.
    /         \
   |  ()   ()  |
   |     ^     |
   |  '-----'  |
    \  |||||  /
     '-------'`

// Banner prints the red skull and wordmark on stdout.
func Banner() {
	fmt.Println(colorize(asciiSkull, ansiRed))
	fmt.Printf("   %s%s\n", colorize("/", ansiRed), "SlashNode")
	fmt.Println(colorize("   your node, your rules\n", ansiDim))
}

// Usage prints the general help.
func Usage() {
	Banner()
	fmt.Print(`Usage: slashnoded <command> [options]

Commands:
  init         Generate config + secrets + systemd unit + Avahi service (idempotent)
  serve        Start the daemon (Go API + supervised Next.js front end)
  status       Show node status (--post-install for URL + credentials)
  update       Apply the latest binary update (--to <tag>)
  check-update Check whether an update is available (notify-only)
  uninstall    Remove the service and binary (--purge also removes data)
  version      Print the version
  help         Show this help

Environment variables:
  SLASHNODE_ROOT   Root prefix for all paths (testing without root)
  NO_COLOR         Disable colored output
`)
}
