package cli

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/slashbinslashnoname/slashnode/internal/avahi"
	"github.com/slashbinslashnoname/slashnode/internal/config"
	"github.com/slashbinslashnoname/slashnode/internal/paths"
	"github.com/slashbinslashnoname/slashnode/internal/secrets"
	"github.com/slashbinslashnoname/slashnode/internal/systemd"
)

// Init implémente `slashnoded init`. Idempotent : relançable sans casse. Toute
// la logique d'amorçage (config, secrets, systemd, Avahi) vit ici, dans le
// binaire Go testable, plutôt que dans du bash fragile.
func Init(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	force := fs.Bool("force", false, "régénère config + secrets même s'ils existent")
	quiet := fs.Bool("quiet", false, "réduit la sortie")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if !*quiet {
		Banner()
	}

	// 1. Répertoires.
	if err := os.MkdirAll(paths.ConfigDir(), 0o755); err != nil {
		return fmt.Errorf("création %s : %w", paths.ConfigDir(), err)
	}
	if err := os.MkdirAll(paths.DataDir(), 0o700); err != nil {
		return fmt.Errorf("création %s : %w", paths.DataDir(), err)
	}

	// 2. Config (idempotent : conservée si déjà présente, sauf --force).
	cfg, err := config.Load(paths.ConfigFile())
	switch {
	case err == nil && !*force:
		logf(*quiet, "→ config existante conservée (%s)", paths.ConfigFile())
	default:
		cfg, err = config.Default(Version, paths.DataDir())
		if err != nil {
			return err
		}
		if err := cfg.Save(paths.ConfigFile()); err != nil {
			return fmt.Errorf("écriture config : %w", err)
		}
		logf(*quiet, "→ config générée (%s)", paths.ConfigFile())
	}

	// 3. Secrets (idempotent : régénérés seulement si absents ou --force).
	if _, err := secrets.Load(paths.SecretsFile()); err != nil || *force {
		sec, initialPassword, gerr := secrets.Generate()
		if gerr != nil {
			return gerr
		}
		if err := sec.Save(paths.SecretsFile()); err != nil {
			return fmt.Errorf("écriture secrets : %w", err)
		}
		// Mot de passe initial : fichier lisible une fois, mode 0600.
		if err := os.WriteFile(paths.InitialPasswordFile(), []byte(initialPassword+"\n"), 0o600); err != nil {
			return fmt.Errorf("écriture mot de passe initial : %w", err)
		}
		logf(*quiet, "→ secrets générés (%s)", paths.SecretsFile())
	} else {
		logf(*quiet, "→ secrets existants conservés (%s)", paths.SecretsFile())
	}

	// 4. Intégration système : Linux uniquement (systemd + Avahi).
	if runtime.GOOS == "linux" {
		if err := systemd.WriteUnit(paths.SystemdUnit(), paths.BinPath()); err != nil {
			return fmt.Errorf("écriture unit systemd : %w", err)
		}
		logf(*quiet, "→ unit systemd écrite (%s)", paths.SystemdUnit())

		if err := avahi.WriteService(paths.AvahiService(), cfg.HTTP.Port); err != nil {
			return fmt.Errorf("écriture service Avahi : %w", err)
		}
		logf(*quiet, "→ service Avahi écrit (%s)", paths.AvahiService())

		if err := systemd.WriteUpdateUnits(paths.SystemdUpdateService(), paths.SystemdUpdateTimer(), paths.BinPath()); err != nil {
			return fmt.Errorf("écriture units de mise à jour : %w", err)
		}
		logf(*quiet, "→ timer de mise à jour écrit (%s)", paths.SystemdUpdateTimer())
	} else {
		logf(*quiet, "→ %s détecté : systemd/Avahi ignorés (Linux uniquement)", runtime.GOOS)
	}

	if !*quiet {
		fmt.Println()
		fmt.Println(colorize("✓ Initialisation terminée.", ansiRed))
		fmt.Println(colorize("  Active le service :  systemctl enable --now slashnoded", ansiDim))
		fmt.Println(colorize("  Puis :               slashnoded status --post-install", ansiDim))
	}
	return nil
}

func logf(quiet bool, format string, a ...any) {
	if quiet {
		return
	}
	fmt.Printf(format+"\n", a...)
}
