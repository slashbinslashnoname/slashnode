package cli

import (
	"flag"
	"fmt"

	"github.com/slashbinslashnoname/slashnode/internal/config"
	"github.com/slashbinslashnoname/slashnode/internal/paths"
	"github.com/slashbinslashnoname/slashnode/internal/updater"
)

// CheckUpdate implémente `slashnoded check-update` (appelé par le timer
// systemd). Politique notify : il vérifie et persiste l'état, sans appliquer.
func CheckUpdate(args []string) error {
	cfg := loadCfgOrDefaultChannel()
	info, err := updater.Check(Version, cfg.Update.Channel)
	if err != nil {
		return fmt.Errorf("vérification des mises à jour : %w", err)
	}
	if info.Available {
		fmt.Printf("mise à jour disponible : %s → %s\n", info.Current, colorize(info.Latest, ansiRed))
	} else {
		fmt.Printf("à jour (%s)\n", info.Current)
	}
	return nil
}

// Update implémente `slashnoded update` : applique la dernière version (ou
// --to). Remplace le binaire et redémarre le service.
func Update(args []string) error {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	to := fs.String("to", "latest", "version cible (tag) ou 'latest'")
	if err := fs.Parse(args); err != nil {
		return err
	}
	cfg := loadCfgOrDefaultChannel()

	fmt.Printf("→ application de la mise à jour (%s)…\n", *to)
	if err := updater.Apply(*to, cfg.Update.Channel); err != nil {
		return fmt.Errorf("mise à jour : %w", err)
	}
	fmt.Println(colorize("✓ binaire mis à jour, redémarrage du service.", ansiRed))
	return nil
}

// loadCfgOrDefaultChannel charge la config, ou renvoie des valeurs par défaut
// si le nœud n'est pas (encore) initialisé.
func loadCfgOrDefaultChannel() *config.Config {
	if cfg, err := config.Load(paths.ConfigFile()); err == nil {
		if cfg.Update.Channel == "" {
			cfg.Update.Channel = "stable"
		}
		return cfg
	}
	return &config.Config{Update: config.Update{Policy: "notify", Channel: "stable"}}
}
