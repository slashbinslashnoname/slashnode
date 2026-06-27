package apps

import (
	"github.com/slashbinslashnoname/slashnode/internal/config"
	"github.com/slashbinslashnoname/slashnode/internal/paths"
	"github.com/slashbinslashnoname/slashnode/internal/tor"
)

// ReloadTor regenerates the torrc exposing the SlashNode UI and every installed
// app as Tor hidden services (.onion), then reloads tor. Each app gets a single
// onion forwarding its web UI (onion:80 → web port) and every endpoint it
// declares (onion:<port> → 127.0.0.1:<port>), so node services — RPC, P2P,
// Electrum, Lightning… — are reachable over Tor, not just web UIs. Best-effort
// and a no-op unless Tor is enabled and the tor binary is available.
func ReloadTor(dir string) error {
	cfg, err := config.Load(paths.ConfigFile())
	if err != nil {
		return nil
	}
	if !cfg.Tor.Enabled || !tor.Available() {
		return nil
	}

	services := []tor.Service{{
		Name:  "slashnode",
		Ports: []tor.PortForward{{Virtual: 80, Local: cfg.HTTP.Port}},
	}}
	for _, a := range LoadState().Installed {
		var ports []tor.PortForward
		seen := map[int]bool{}
		add := func(virtual, local int) {
			if virtual > 0 && local > 0 && !seen[virtual] {
				seen[virtual] = true
				ports = append(ports, tor.PortForward{Virtual: virtual, Local: local})
			}
		}
		if a.WebPort > 0 {
			add(80, a.WebPort)
		}
		if man, _, ferr := ResolveBase(dir, a.ID); ferr == nil {
			for _, e := range man.Endpoints {
				// Endpoints are reachable on their published host port; an extra
				// instance publishes them on a reassigned port.
				local := e.Port
				if mapped, ok := a.Ports[e.Port]; ok {
					local = mapped
				}
				add(e.Port, local)
			}
		}
		if len(ports) > 0 {
			services = append(services, tor.Service{Name: a.ID, Ports: ports})
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
