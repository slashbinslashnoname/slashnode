package cli

import (
	"flag"
	"fmt"

	"github.com/slashbinslashnoname/slashnode/internal/config"
	"github.com/slashbinslashnoname/slashnode/internal/paths"
	"github.com/slashbinslashnoname/slashnode/internal/updater"
)

// CheckUpdate implements `slashnoded check-update` (called by the systemd
// timer). Notify policy: it checks and persists the state, without applying.
func CheckUpdate(args []string) error {
	cfg := loadCfgOrDefaultChannel()
	info, err := updater.Check(Version, cfg.Update.Channel)
	if err != nil {
		return fmt.Errorf("update check: %w", err)
	}
	if info.Available {
		fmt.Printf("update available: %s → %s\n", info.Current, colorize(info.Latest, ansiRed))
	} else {
		fmt.Printf("up to date (%s)\n", info.Current)
	}
	return nil
}

// Update implements `slashnoded update`: applies the latest version (or --to).
// Replaces the binary and restarts the service.
func Update(args []string) error {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	to := fs.String("to", "latest", "target version (tag) or 'latest'")
	if err := fs.Parse(args); err != nil {
		return err
	}
	cfg := loadCfgOrDefaultChannel()

	fmt.Printf("→ applying the update (%s)…\n", *to)
	if err := updater.Apply(*to, cfg.Update.Channel); err != nil {
		return fmt.Errorf("update: %w", err)
	}
	fmt.Println(colorize("✓ binary updated, restarting the service.", ansiRed))
	return nil
}

// loadCfgOrDefaultChannel loads the config, or returns default values if the
// node is not (yet) initialized.
func loadCfgOrDefaultChannel() *config.Config {
	if cfg, err := config.Load(paths.ConfigFile()); err == nil {
		if cfg.Update.Channel == "" {
			cfg.Update.Channel = "stable"
		}
		return cfg
	}
	return &config.Config{Update: config.Update{Policy: "notify", Channel: "stable"}}
}
