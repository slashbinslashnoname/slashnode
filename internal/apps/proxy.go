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

// appHTTPSPort is the dedicated port on which Caddy terminates TLS for an app
// whose plain-HTTP backend is published on webPort. Derived deterministically so
// the URL is stable across reinstalls.
func appHTTPSPort(webPort int) int { return webPort + 10000 }

// ReloadProxy regenerates the Caddyfile from the installed apps (root host →
// front end on 443, plus a dedicated HTTPS port per app with a web UI) and
// reloads Caddy. Each app is reached at https://<host>:<appHTTPSPort> — works on
// a domain, an mDNS name or a bare IP, with no per-app DNS. Best-effort: a no-op
// when the node isn't initialized; writes the file even without Caddy installed.
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
				Host:         host,
				ListenPort:   appHTTPSPort(a.WebPort),
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

// AppURL returns the HTTPS URL for an app (Caddy terminates TLS on the app's
// dedicated port), or "" if it has no web UI.
func AppURL(cfg *config.Config, m *Manifest) string {
	if m.Web == nil {
		return ""
	}
	host, _ := baseHost(cfg)
	return fmt.Sprintf("https://%s:%d", host, appHTTPSPort(m.Web.Port))
}
