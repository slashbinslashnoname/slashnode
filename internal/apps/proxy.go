package apps

import (
	"fmt"

	"github.com/slashbinslashnoname/slashnode/internal/caddy"
	"github.com/slashbinslashnoname/slashnode/internal/config"
	"github.com/slashbinslashnoname/slashnode/internal/paths"
)

// baseHost returns the host apps live under: the public address in server mode,
// otherwise the mDNS hostname. internalTLS is true for local mode (.local names
// use Caddy's internal CA; a public domain gets Let's Encrypt).
func baseHost(cfg *config.Config) (host string, internalTLS bool) {
	if cfg.Access.Mode == "server" && cfg.Access.Address != "" {
		return cfg.Access.Address, false
	}
	return cfg.Hostname, true
}

// ReloadProxy regenerates the Caddyfile from the installed apps (root host →
// front end, plus <id>.<host> → each app's web port, all over HTTPS) and reloads
// Caddy. Best-effort: a no-op when the node isn't initialized; writes the file
// even without Caddy installed so it's ready once it is.
func ReloadProxy() error {
	cfg, err := config.Load(paths.ConfigFile())
	if err != nil {
		return nil
	}
	host, internalTLS := baseHost(cfg)

	routes := []caddy.Route{{Host: host, UpstreamPort: cfg.HTTP.Port}}
	for _, a := range LoadState().Installed {
		if a.WebPort > 0 {
			routes = append(routes, caddy.Route{
				Host:         appSubdomain(a.ID) + "." + host,
				UpstreamPort: a.WebPort,
			})
		}
	}

	if err := caddy.Write(paths.CaddyfilePath(), routes, internalTLS); err != nil {
		return err
	}
	if caddy.Available() {
		return caddy.Reload(paths.CaddyfilePath())
	}
	return nil
}

// AppURL returns the HTTPS reverse-proxy URL for an app (https://<id>.<host>),
// or "" if it has no web UI.
func AppURL(cfg *config.Config, m *Manifest) string {
	if m.Web == nil {
		return ""
	}
	host, _ := baseHost(cfg)
	return fmt.Sprintf("https://%s.%s", appSubdomain(m.ID), host)
}
