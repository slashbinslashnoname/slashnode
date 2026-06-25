package cli

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/slashbinslashnoname/slashnode/internal/apps"
	"github.com/slashbinslashnoname/slashnode/internal/avahi"
	"github.com/slashbinslashnoname/slashnode/internal/config"
	"github.com/slashbinslashnoname/slashnode/internal/paths"
	"github.com/slashbinslashnoname/slashnode/internal/secrets"
	"github.com/slashbinslashnoname/slashnode/internal/systemd"
)

// Init implements `slashnoded init`. Idempotent: re-runnable without breakage.
// All the bootstrap logic (config, secrets, systemd, Avahi) lives here, in the
// testable Go binary, rather than in fragile bash.
func Init(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	force := fs.Bool("force", false, "regenerate config + secrets even if they exist")
	quiet := fs.Bool("quiet", false, "reduce the output")
	accessMode := fs.String("access", "", "access mode: local|server")
	address := fs.String("address", "", "public address for server mode (e.g. node.example.com)")
	passwordProtect := fs.Bool("password-protect", false, "require a login for the web UI")
	password := fs.String("password", "", "set the admin password (otherwise a random one is generated)")
	enableTor := fs.Bool("tor", false, "expose the UI and apps as Tor hidden services (.onion)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if !*quiet {
		Banner()
	}

	// 1. Directories.
	if err := os.MkdirAll(paths.ConfigDir(), 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", paths.ConfigDir(), err)
	}
	if err := os.MkdirAll(paths.DataDir(), 0o700); err != nil {
		return fmt.Errorf("creating %s: %w", paths.DataDir(), err)
	}

	// 2. Config (idempotent: kept if already present, except with --force).
	cfg, err := config.Load(paths.ConfigFile())
	switch {
	case err == nil && !*force:
		logf(*quiet, "→ existing config kept (%s)", paths.ConfigFile())
	default:
		cfg, err = config.Default(Version, paths.DataDir())
		if err != nil {
			return err
		}
		if err := cfg.Save(paths.ConfigFile()); err != nil {
			return fmt.Errorf("writing config: %w", err)
		}
		logf(*quiet, "→ config generated (%s)", paths.ConfigFile())
	}

	// 2b. Access settings (applied whenever an access flag is provided).
	if *accessMode != "" || *address != "" || *passwordProtect {
		if *accessMode != "" {
			cfg.Access.Mode = *accessMode
		}
		if *address != "" {
			cfg.Access.Address = *address
		}
		if *passwordProtect || cfg.Access.Mode == "server" {
			// Server mode always requires a password.
			cfg.Access.PasswordProtected = true
		}
		if err := cfg.Save(paths.ConfigFile()); err != nil {
			return fmt.Errorf("writing config: %w", err)
		}
		logf(*quiet, "→ access set (mode=%s, password=%v)", cfg.Access.Mode, cfg.Access.PasswordProtected)
	}

	// 2c. Tor exposure (applied when --tor is provided).
	if *enableTor && !cfg.Tor.Enabled {
		cfg.Tor.Enabled = true
		if err := cfg.Save(paths.ConfigFile()); err != nil {
			return fmt.Errorf("writing config: %w", err)
		}
		logf(*quiet, "→ Tor hidden services enabled")
	}

	// 3. Secrets (idempotent: regenerated only if absent or --force).
	switch _, serr := secrets.Load(paths.SecretsFile()); {
	case serr != nil || *force:
		sec, initialPassword, gerr := secrets.Generate()
		if gerr != nil {
			return gerr
		}
		if *password != "" {
			if err := sec.SetPassword(*password); err != nil {
				return err
			}
			initialPassword = "" // operator-chosen: nothing to reveal
		}
		if err := sec.Save(paths.SecretsFile()); err != nil {
			return fmt.Errorf("writing secrets: %w", err)
		}
		if err := writeOrClearInitialPassword(initialPassword); err != nil {
			return err
		}
		logf(*quiet, "→ secrets generated (%s)", paths.SecretsFile())
	case *password != "":
		// Secrets exist but the operator is (re)setting the admin password.
		sec, lerr := secrets.Load(paths.SecretsFile())
		if lerr != nil {
			return lerr
		}
		if err := sec.SetPassword(*password); err != nil {
			return err
		}
		if err := sec.Save(paths.SecretsFile()); err != nil {
			return fmt.Errorf("writing secrets: %w", err)
		}
		_ = os.Remove(paths.InitialPasswordFile())
		logf(*quiet, "→ admin password updated")
	default:
		logf(*quiet, "→ existing secrets kept (%s)", paths.SecretsFile())
	}

	// 4. System integration: Linux only (systemd + Avahi).
	if runtime.GOOS == "linux" {
		if err := systemd.WriteUnit(paths.SystemdUnit(), paths.BinPath()); err != nil {
			return fmt.Errorf("writing systemd unit: %w", err)
		}
		logf(*quiet, "→ systemd unit written (%s)", paths.SystemdUnit())

		if err := avahi.WriteService(paths.AvahiService(), cfg.HTTP.Port); err != nil {
			return fmt.Errorf("writing Avahi service: %w", err)
		}
		logf(*quiet, "→ Avahi service written (%s)", paths.AvahiService())

		if err := systemd.WriteUpdateUnits(paths.SystemdUpdateService(), paths.SystemdUpdateTimer(), paths.BinPath()); err != nil {
			return fmt.Errorf("writing update units: %w", err)
		}
		logf(*quiet, "→ update timer written (%s)", paths.SystemdUpdateTimer())

		// Initial reverse-proxy config (root host → front end).
		if err := apps.ReloadProxy(); err != nil {
			return fmt.Errorf("writing Caddyfile: %w", err)
		}
		logf(*quiet, "→ Caddyfile written (%s)", paths.CaddyfilePath())

		// Tor hidden services (best-effort; no-op unless enabled + tor present).
		if err := apps.ReloadTor(); err != nil {
			logf(*quiet, "→ Tor reload skipped (%v)", err)
		} else if cfg.Tor.Enabled {
			logf(*quiet, "→ Tor hidden services written (%s)", paths.TorrcPath())
		}
	} else {
		logf(*quiet, "→ %s detected: systemd/Avahi skipped (Linux only)", runtime.GOOS)
	}

	if !*quiet {
		fmt.Println()
		fmt.Println(colorize("✓ Initialization complete.", ansiRed))
		fmt.Println(colorize("  Enable the service:  systemctl enable --now slashnoded", ansiDim))
		fmt.Println(colorize("  Then:                slashnoded status --post-install", ansiDim))
	}
	return nil
}

// writeOrClearInitialPassword writes the one-time initial password file when a
// random password was generated, or removes it when the operator set their own.
func writeOrClearInitialPassword(initialPassword string) error {
	if initialPassword == "" {
		_ = os.Remove(paths.InitialPasswordFile())
		return nil
	}
	if err := os.WriteFile(paths.InitialPasswordFile(), []byte(initialPassword+"\n"), 0o600); err != nil {
		return fmt.Errorf("writing initial password: %w", err)
	}
	return nil
}

func logf(quiet bool, format string, a ...any) {
	if quiet {
		return
	}
	fmt.Printf(format+"\n", a...)
}
