// Package paths centralise tous les chemins du système de fichiers utilisés
// par SlashNode. Tous les chemins peuvent être préfixés via la variable
// d'environnement SLASHNODE_ROOT, ce qui permet de tester init/uninstall sans
// privilèges root et sans toucher au vrai système.
package paths

import (
	"os"
	"path/filepath"
)

// root renvoie le préfixe racine (vide => "/").
func root() string {
	if r := os.Getenv("SLASHNODE_ROOT"); r != "" {
		return r
	}
	return "/"
}

func p(parts ...string) string {
	return filepath.Join(append([]string{root()}, parts...)...)
}

// Répertoires et fichiers de configuration.
func ConfigDir() string  { return p("etc", "slashnode") }
func ConfigFile() string { return p("etc", "slashnode", "config.json") }

// Données et secrets (mode 0700).
func DataDir() string             { return p("var", "lib", "slashnode") }
func SecretsFile() string         { return p("var", "lib", "slashnode", "secrets.json") }
func InitialPasswordFile() string { return p("var", "lib", "slashnode", "initial-password") }
func UpdateStateFile() string     { return p("var", "lib", "slashnode", "update.json") }

// Front Next.js déployé sur le device (servi/lancé par le démon).
func WebDir() string { return p("usr", "share", "slashnode", "web") }

// Binaire et intégration système.
func BinPath() string              { return p("usr", "local", "bin", "slashnoded") }
func SystemdUnit() string          { return p("etc", "systemd", "system", "slashnoded.service") }
func SystemdUpdateService() string { return p("etc", "systemd", "system", "slashnoded-update.service") }
func SystemdUpdateTimer() string   { return p("etc", "systemd", "system", "slashnoded-update.timer") }
func AvahiDir() string             { return p("etc", "avahi", "services") }
func AvahiService() string         { return p("etc", "avahi", "services", "slashnode.service") }
