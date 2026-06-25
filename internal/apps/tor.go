package apps

import (
	"github.com/slashbinslashnoname/slashnode/internal/config"
	"github.com/slashbinslashnoname/slashnode/internal/paths"
	"github.com/slashbinslashnoname/slashnode/internal/tor"
)

// ReloadTor regenerates the torrc exposing the SlashNode UI and every installed
// web app as Tor hidden services (.onion), then reloads tor. Best-effort and a
// no-op unless Tor is enabled in the config and the tor binary is available.
func ReloadTor() error {
	cfg, err := config.Load(paths.ConfigFile())
	if err != nil {
		return nil
	}
	if !cfg.Tor.Enabled || !tor.Available() {
		return nil
	}

	// The UI itself, plus a hidden service per app with a web UI. The hidden
	// service forwards onion:80 → 127.0.0.1:<port> (ports published on the host).
	services := []tor.Service{{Name: "slashnode", Port: cfg.HTTP.Port}}
	for _, a := range LoadState().Installed {
		if a.WebPort > 0 {
			services = append(services, tor.Service{Name: a.ID, Port: a.WebPort})
		}
	}

	if err := tor.Write(paths.TorrcPath(), paths.TorDataDir(), services); err != nil {
		return err
	}
	return tor.Reload()
}

// AppOnion returns the .onion address of an installed app's web UI (empty if Tor
// is disabled or the hidden service isn't provisioned yet).
func AppOnion(id string) string {
	cfg, err := config.Load(paths.ConfigFile())
	if err != nil || !cfg.Tor.Enabled {
		return ""
	}
	return tor.Onion(paths.TorHostnameFile(id))
}

// NodeOnion returns the .onion address of the SlashNode UI itself.
func NodeOnion() string {
	cfg, err := config.Load(paths.ConfigFile())
	if err != nil || !cfg.Tor.Enabled {
		return ""
	}
	return tor.Onion(paths.TorHostnameFile("slashnode"))
}
