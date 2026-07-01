package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/slashbinslashnoname/slashnode/internal/config"
	"github.com/slashbinslashnoname/slashnode/internal/paths"
	"github.com/slashbinslashnoname/slashnode/internal/tailscale"
)

// Tailscale implements `slashnoded tailscale <up|down|status>`. It joins the
// node to a Tailscale tailnet so two SlashNodes in different locations can reach
// each other privately and back one up onto the other off-site.
//
//	tailscale up --authkey <key> [--hostname <name>]   join / re-authenticate
//	tailscale down                                     leave (identity preserved)
//	tailscale status                                   show tailnet address + peers
func Tailscale(args []string) error {
	if len(args) == 0 {
		return tailscaleStatus()
	}
	sub, rest := args[0], args[1:]
	switch sub {
	case "up":
		return tailscaleUp(rest)
	case "down":
		return tailscaleDown()
	case "status":
		return tailscaleStatus()
	default:
		return fmt.Errorf("unknown tailscale subcommand %q (up | down | status)", sub)
	}
}

func tailscaleUp(args []string) error {
	fs := flag.NewFlagSet("tailscale up", flag.ContinueOnError)
	authKey := fs.String("authkey", "", "Tailscale auth key (first join / re-auth; reused from state afterwards)")
	hostname := fs.String("hostname", "", "machine name shown on the tailnet (optional)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if !tailscale.ValidHostname(*hostname) {
		return fmt.Errorf("invalid hostname (DNS label: letters, digits, hyphens)")
	}
	if err := tailscale.Up(*authKey, *hostname, os.Stdout); err != nil {
		return err
	}
	// Persist the enabled state so serve reconciles the tailnet on startup.
	return saveTailscaleConfig(true, *hostname)
}

func tailscaleDown() error {
	if err := tailscale.Down(os.Stdout); err != nil {
		return err
	}
	cfg, err := config.Load(paths.ConfigFile())
	if err != nil {
		return nil // not initialised — nothing to persist
	}
	return saveTailscaleConfig(false, cfg.Tailscale.Hostname)
}

func tailscaleStatus() error {
	st := tailscale.GetStatus()
	b, _ := json.MarshalIndent(st, "", "  ")
	fmt.Println(string(b))
	return nil
}

// saveTailscaleConfig updates the node config's tailscale block, tolerating an
// uninitialised node (the container still runs; the setting just isn't recorded).
func saveTailscaleConfig(enabled bool, hostname string) error {
	cfg, err := config.Load(paths.ConfigFile())
	if err != nil {
		return nil
	}
	cfg.Tailscale.Enabled = enabled
	if hostname != "" {
		cfg.Tailscale.Hostname = hostname
	}
	return cfg.Save(paths.ConfigFile())
}
