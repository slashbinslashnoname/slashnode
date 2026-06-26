// Package paths centralizes all the filesystem paths used by SlashNode. All
// paths can be prefixed via the SLASHNODE_ROOT environment variable, which
// allows testing init/uninstall without root privileges and without touching
// the real system.
package paths

import (
	"os"
	"path/filepath"
)

// root returns the root prefix (empty => "/").
func root() string {
	if r := os.Getenv("SLASHNODE_ROOT"); r != "" {
		return r
	}
	return "/"
}

func p(parts ...string) string {
	return filepath.Join(append([]string{root()}, parts...)...)
}

// Configuration directories and files.
func ConfigDir() string  { return p("etc", "slashnode") }
func ConfigFile() string { return p("etc", "slashnode", "config.json") }

// Data and secrets (mode 0700).
func DataDir() string             { return p("var", "lib", "slashnode") }
func SecretsFile() string         { return p("var", "lib", "slashnode", "secrets.json") }
func InitialPasswordFile() string { return p("var", "lib", "slashnode", "initial-password") }
func UpdateStateFile() string     { return p("var", "lib", "slashnode", "update.json") }
func AppsStateFile() string       { return p("var", "lib", "slashnode", "apps.json") }
func AppSecretsFile() string      { return p("var", "lib", "slashnode", "app-secrets.json") }
func RegistryFile() string        { return p("var", "lib", "slashnode", "registry.json") }

// Per-app runtime directory (generated compose file, etc.).
func AppRuntimeDir(id string) string { return p("var", "lib", "slashnode", "apps", id) }
func AppComposeFile(id string) string {
	return p("var", "lib", "slashnode", "apps", id, "docker-compose.yml")
}
func AppConfigDir(id string) string { return p("var", "lib", "slashnode", "apps", id, "config") }

// Next.js front deployed on the device (served/launched by the daemon).
func WebDir() string { return p("usr", "share", "slashnode", "web") }

// App catalog (JSON manifests) deployed on the device.
func AppsDir() string { return p("usr", "share", "slashnode", "apps") }

// Binary and system integration.
func BinPath() string              { return p("usr", "local", "bin", "slashnoded") }
func SystemdUnit() string          { return p("etc", "systemd", "system", "slashnoded.service") }
func SystemdUpdateService() string { return p("etc", "systemd", "system", "slashnoded-update.service") }
func SystemdUpdateTimer() string   { return p("etc", "systemd", "system", "slashnoded-update.timer") }
func SystemdPruneService() string  { return p("etc", "systemd", "system", "slashnoded-prune.service") }
func SystemdPruneTimer() string    { return p("etc", "systemd", "system", "slashnoded-prune.timer") }
func AvahiDir() string             { return p("etc", "avahi", "services") }
func AvahiService() string         { return p("etc", "avahi", "services", "slashnode.service") }
func CaddyfilePath() string        { return p("etc", "caddy", "Caddyfile") }
func TorrcPath() string            { return p("etc", "tor", "torrc") }
func TorDataDir() string           { return p("var", "lib", "tor", "slashnode") }
func TorHostnameFile(name string) string {
	return p("var", "lib", "tor", "slashnode", name, "hostname")
}
