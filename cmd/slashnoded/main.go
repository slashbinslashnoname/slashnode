// Command slashnoded est le binaire unique de SlashNode.
//
// Sous-commandes :
//
//	init       génère config + secrets + unit systemd + service Avahi (idempotent)
//	serve      démarre le démon HTTP
//	status     affiche l'état (et l'URL + creds avec --post-install)
//	uninstall  retire le service et le binaire (--purge supprime aussi les données)
//	version    affiche la version
package main

import (
	"fmt"
	"os"

	"github.com/slashbinslashnoname/slashnode/internal/cli"
)

// Version est injectée au build via -ldflags "-X main.Version=...".
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
	case "update":
		err = cli.Update(args)
	case "check-update":
		err = cli.CheckUpdate(args)
	case "uninstall":
		err = cli.Uninstall(args)
	case "version", "-v", "--version":
		fmt.Printf("slashnoded %s\n", Version)
	case "help", "-h", "--help":
		cli.Usage()
	default:
		fmt.Fprintf(os.Stderr, "commande inconnue : %q\n\n", os.Args[1])
		cli.Usage()
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "erreur :", err)
		os.Exit(1)
	}
}
