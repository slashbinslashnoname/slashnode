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

	case "status":
		if len(args) < 2 {
			return fmt.Errorf("usage: slashnoded apps status <id>")
		}
		if !apps.DockerAvailable() {
			fmt.Println("docker not available")
			return nil
		}
		st, err := apps.Status(args[1])
		if err != nil {
			return err
		}
		if len(st) == 0 {
			fmt.Println("no containers")
			return nil
		}
		for _, s := range st {
			fmt.Printf("%-14s %-10s %s\n", s.Service, s.State, s.Status)
		}
		return nil

	case "start", "stop", "restart":
		if len(args) < 2 {
			return fmt.Errorf("usage: slashnoded apps %s <id>", args[0])
		}
		var err error
		switch args[0] {
		case "start":
			err = apps.Start(args[1])
		case "stop":
			err = apps.Stop(args[1])
		case "restart":
			err = apps.Restart(args[1])
		}
		if err != nil {
			return err
		}
		fmt.Println(colorize("✓ "+args[0]+" "+args[1], ansiRed))
		return nil

	case "logs":
		if len(args) < 2 {
			return fmt.Errorf("usage: slashnoded apps logs <id>")
		}
		logs, err := apps.Logs(args[1], 200)
		if err != nil {
			return err
		}
		fmt.Print(logs)
		return nil

	case "probe":
		if len(args) < 2 {
			return fmt.Errorf("usage: slashnoded apps probe <id>")
		}
		res, err := apps.RunProbe(dir, args[1])
		if err != nil {
			return err
		}
		fmt.Printf("type=%s ok=%v %s\n", res.Type, res.OK, res.Detail)
		for _, s := range res.Stats {
			fmt.Printf("  %s: %s\n", s.Label, s.Value)
		}
		return nil

	default:
		return fmt.Errorf("unknown apps subcommand: %q", args[0])
	}
}
