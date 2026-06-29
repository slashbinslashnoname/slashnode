package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/slashbinslashnoname/slashnode/internal/backup"
)

// Backup implements `slashnoded backup`. Syncs node data to the configured
// destination (state, Tor keys, app volumes). --all includes the large
// re-syncable chain volumes; --test only checks the destination is reachable.
func Backup(args []string) error {
	fs := flag.NewFlagSet("backup", flag.ContinueOnError)
	all := fs.Bool("all", false, "include large re-syncable chain volumes (bitcoind, monerod…)")
	test := fs.Bool("test", false, "only verify the destination is reachable")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if _, err := backup.LoadConfig(); err != nil {
		return err
	}
	if *test {
		return backup.Test(os.Stdout)
	}
	return backup.Run(backup.Options{All: *all, Version: Version}, os.Stdout)
}

// Restore implements `slashnoded restore`. Pulls the backup back onto this node
// (state, Tor keys, volumes) and brings the apps up. Destructive — confirmation
// required unless --yes, and it should run with the daemon stopped.
func Restore(args []string) error {
	fs := flag.NewFlagSet("restore", flag.ContinueOnError)
	yes := fs.Bool("yes", false, "do not ask for confirmation")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if !*yes {
		fmt.Print("⚠ restore overwrites this node's state and volumes. Run with the daemon stopped. Confirm? [y/N] ")
		var resp string
		fmt.Scanln(&resp)
		if resp != "o" && resp != "O" && resp != "y" && resp != "Y" {
			return fmt.Errorf("cancelled")
		}
	}
	return backup.Restore(os.Stdout)
}
