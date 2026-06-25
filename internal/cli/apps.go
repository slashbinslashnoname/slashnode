package cli

import (
	"fmt"
	"strings"

	"github.com/slashbinslashnoname/slashnode/internal/apps"
)

// Apps implements `slashnoded apps <list|install|uninstall>`.
//
//	apps list
//	apps install <id> [KEY=VALUE ...]
//	apps uninstall <id> [--purge]
func Apps(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: slashnoded apps <list|install|uninstall>")
	}
	dir := resolveAppsDir()

	switch args[0] {
	case "list":
		cat, err := apps.Catalog(dir)
		if err != nil {
			return err
		}
		for _, a := range cat {
			mark := " "
			if a.Installed {
				mark = colorize("✓", ansiRed)
			}
			fmt.Printf("%s %-10s %-14s %s\n", mark, a.ID, a.Version, a.Description)
		}
		return nil

	case "install":
		if len(args) < 2 {
			return fmt.Errorf("usage: slashnoded apps install <id> [KEY=VALUE ...]")
		}
		id := args[1]
		inputs := map[string]string{}
		for _, kv := range args[2:] {
			k, v, ok := strings.Cut(kv, "=")
			if !ok {
				return fmt.Errorf("invalid input %q (expected KEY=VALUE)", kv)
			}
			inputs[k] = v
		}
		if err := apps.Install(dir, id, inputs); err != nil {
			return err
		}
		fmt.Println(colorize("✓ installed "+id, ansiRed))
		return nil

	case "uninstall":
		if len(args) < 2 {
			return fmt.Errorf("usage: slashnoded apps uninstall <id> [--purge]")
		}
		id := args[1]
		purge := false
		for _, a := range args[2:] {
			if a == "--purge" {
				purge = true
			}
		}
		if err := apps.Uninstall(id, purge); err != nil {
			return err
		}
		fmt.Println(colorize("✓ uninstalled "+id, ansiRed))
		return nil

	default:
		return fmt.Errorf("unknown apps subcommand: %q", args[0])
	}
}
