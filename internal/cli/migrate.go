package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/slashbinslashnoname/slashnode/internal/apps"
	"github.com/slashbinslashnoname/slashnode/internal/migrate"
)

// Migrate implements `slashnoded migrate [--dry-run]`: applies pending node-state
// migrations (snapshotting + rolling back on failure) and re-applies installed
// apps so their per-app migrations run. Runs automatically on init and on serve
// startup; this command is for a manual trigger or a dry-run.
func Migrate(args []string) error {
	fs := flag.NewFlagSet("migrate", flag.ContinueOnError)
	dry := fs.Bool("dry-run", false, "list pending migrations without applying them")
	if err := fs.Parse(args); err != nil {
		return err
	}
	dir := resolveAppsDir()

	if *dry {
		node := migrate.Pending()
		appsPending := apps.MigrationsPending(dir)
		if len(node) == 0 && len(appsPending) == 0 {
			fmt.Println("up to date — no pending migrations")
			return nil
		}
		for _, m := range node {
			fmt.Printf("node  → v%d (%s)\n", m.Version, m.Name)
		}
		for _, a := range appsPending {
			fmt.Printf("app   %s\n", a)
		}
		return nil
	}

	if err := migrate.Run(os.Stdout); err != nil {
		return err
	}
	// Re-apply installed apps so their per-app migrations run, then recreate.
	if err := apps.Reapply(dir); err != nil {
		return fmt.Errorf("applying app migrations: %w", err)
	}
	fmt.Println(colorize("✓ migrations applied", ansiRed))
	return nil
}
