// Command slashnoded is the single binary of SlashNode.
//
// Subcommands:
//
//	init       generates config + secrets + systemd unit + Avahi service (idempotent)
//	serve      starts the HTTP daemon
//	status     displays the state (and the URL + creds with --post-install)
//	uninstall  removes the service and the binary (--purge also deletes data)
//	version    displays the version
package main

import (
	"fmt"
	"os"

	"github.com/slashbinslashnoname/slashnode/internal/cli"
)

// Version is injected at build time via -ldflags "-X main.Version=...".
var Version = "dev"

func main() {
	cli.Version = Version

	if len(os.Args) < 2 {
		cli.Usage()
		os.Exit(2)
	}

	args := os.Args[2:]
	var err error

	switch os.Args[1] {
	case "init":
		err = cli.Init(args)
	case "serve", "run":
		err = cli.Serve(args)
	case "status":
		err = cli.Status(args)
	case "apps":
		err = cli.Apps(args)
	case "update":
		err = cli.Update(args)
	case "check-update":
		err = cli.CheckUpdate(args)
	case "prune":
		err = cli.Prune(args)
	case "passwd":
		err = cli.Passwd(args)
	case "uninstall":
		err = cli.Uninstall(args)
	case "version", "-v", "--version":
		fmt.Printf("slashnoded %s\n", Version)
	case "help", "-h", "--help":
		cli.Usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %q\n\n", os.Args[1])
		cli.Usage()
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
