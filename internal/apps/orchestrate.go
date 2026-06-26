package apps

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/slashbinslashnoname/slashnode/internal/config"
	"github.com/slashbinslashnoname/slashnode/internal/orchestrator"
	"github.com/slashbinslashnoname/slashnode/internal/paths"
)

// Install resolves the dependency graph (installing any missing dependencies
// first), renders a Compose file for each app, launches it via Docker and
// records its exports in the service registry so consumers can wire to it.
func Install(dir, id string, inputs map[string]string) error {
	return InstallStream(dir, id, inputs, nil, io.Discard)
}

// InstallStream is Install with the Docker pull/up output (and progress lines)
// streamed live to out, so the UI can show install progress as it happens.
// imageTags optionally pins per-service image tags for the target app (chosen in
// the install form); pass nil to use the manifest defaults.
func InstallStream(dir, id string, inputs, imageTags map[string]string, out io.Writer) error {
	plan, err := installPlan(dir, id)
	if err != nil {
		return err
	}

	if orchestrator.Available() {
		if err := orchestrator.EnsureNetwork(); err != nil {
			return fmt.Errorf("docker network: %w", err)
		}
	}

	for _, appID := range plan {
		isTarget := appID == id
		var provided, tags map[string]string
		if isTarget {
			provided = inputs
			tags = imageTags
		}
		fmt.Fprintf(out, "\n==> installing %s\n", appID)
		if err := installOne(dir, appID, provided, tags, isTarget, out); err != nil {
			fmt.Fprintf(out, "error: %v\n", err)
			return fmt.Errorf("installing %s: %w", appID, err)
		}
	}
	fmt.Fprintln(out, "\n==> wiring reverse proxy & tor")
	_ = ReloadProxy() // best-effort: refresh reverse-proxy routes
	_ = ReloadTor(dir) // best-effort: refresh Tor hidden services
	fmt.Fprintln(out, "==> done")
	printServiceURLs(out, dir, id)
	return nil
}

// PruneImages reclaims disk by removing dangling images (best-effort, no-op
// without Docker). Run on updates, on bootstrap and by the daily timer — not on
// install. Stopped containers and volumes are preserved.
func PruneImages(out io.Writer) {
	if orchestrator.Available() {
		_ = orchestrator.Prune(out)
	}
}

// nodeExports returns the node's own coordinates, exposed to manifests as
// ${node.exports.host} so they can build their conventional reverse-proxy URL
// (e.g. jitsi.<host>) without prompting the operator.
func nodeExports() map[string]string {
	host := "slashnode.local"
	if cfg, err := config.Load(paths.ConfigFile()); err == nil {
		if cfg.Access.Mode == "server" && cfg.Access.Address != "" {
			host = cfg.Access.Address
		} else if cfg.Hostname != "" {
			host = cfg.Hostname
		}
	}
	return map[string]string{"host": host}
}

// selfExports returns the installing app's own public coordinates for
// ${self.exports.*}: the HTTPS URL Caddy serves it on (and host/port), so the
// app can advertise itself without the operator entering a URL.
func selfExports(man *Manifest) map[string]string {
	out := map[string]string{}
	cfg, err := config.Load(paths.ConfigFile())
	if err != nil {
		return out
	}
	host, _ := baseHost(cfg)
	out["host"] = host
	if man.Web != nil {
		out["url"] = AppURL(cfg, man)
	}
	return out
}

// printServiceURLs writes a bootstrap-style summary of how to reach the app: its
// web UI and every declared endpoint, in clearnet form and (when Tor is enabled
// and provisioned) over .onion.
func printServiceURLs(out io.Writer, dir, id string) {
	man, err := Find(dir, id)
	if err != nil || (man.Web == nil && len(man.Endpoints) == 0) {
		return
	}
	host := "slashnode.local"
	if cfg, cerr := config.Load(paths.ConfigFile()); cerr == nil {
		if cfg.Access.Mode == "server" && cfg.Access.Address != "" {
			host = cfg.Access.Address
		} else if cfg.Hostname != "" {
			host = cfg.Hostname
		}
	}
	onion := AppOnion(id)

	fmt.Fprintln(out, "\n==> service URLs")
	if man.Web != nil {
		fmt.Fprintf(out, "    Web UI       http://%s:%d%s\n", host, man.Web.Port, man.Web.Path)
		if onion != "" {
			// Tor maps the web UI to onion:80, not the published web port.
			fmt.Fprintf(out, "      .onion     http://%s%s\n", onion, man.Web.Path)
		}
	}
	for _, e := range man.Endpoints {
		// Endpoints are connection addresses (RPC, P2P, …) — a bare host:port,
		// no scheme; they're reached with a client, not a browser.
		path := e.Path
		if path == "/" {
			path = ""
		}
		fmt.Fprintf(out, "    %-12s %s:%d%s\n", e.Label, host, e.Port, path)
		if onion != "" {
			fmt.Fprintf(out, "      .onion     %s:%d%s\n", onion, e.Port, path)
		}
	}
}

// Reapply re-renders and recreates every installed app from the current
// manifests (in dependency order), reusing the stored inputs/secrets. Used after
// an update so manifest/config changes take effect without manual reinstall.
func Reapply(dir string) error {
	state := LoadState()

	var order []string
	seen := map[string]bool{}
	var visit func(string)
	visit = func(id string) {
		if seen[id] {
			return
		}
		seen[id] = true
		if man, err := Find(dir, id); err == nil {
			for _, dep := range man.Dependencies {
				if _, ok := state.Installed[dep]; ok {
					visit(dep)
				}
			}
		}
		if _, ok := state.Installed[id]; ok {
			order = append(order, id)
		}
	}
	for id := range state.Installed {
		visit(id)
	}

	if orchestrator.Available() {
		if err := orchestrator.EnsureNetwork(); err != nil {
			return fmt.Errorf("docker network: %w", err)
		}
	}
	for _, id := range order {
		if err := runAppMigrations(dir, id, io.Discard); err != nil {
			return fmt.Errorf("migrate %s: %w", id, err)
		}
		inst := LoadState().Installed[id] // reload — a migration may have changed inputs
		provided := map[string]string{}
		for k, v := range inst.Inputs {
			provided[k] = v
		}
		for k, v := range loadAppSecrets(id) {
			provided[k] = v
		}
		if err := installOne(dir, id, provided, nil, true, io.Discard); err != nil {
			return fmt.Errorf("reapply %s: %w", id, err)
		}
	}
	_ = ReloadProxy()
	_ = ReloadTor(dir)
	PruneImages(io.Discard) // reclaim disk after pulling updated images
	return nil
}

// ReapplyOne re-renders and recreates a single installed app from the current
// manifest (reusing stored inputs/secrets) — used to update one app to the
// catalog's version without re-entering its settings.
func ReapplyOne(dir, id string) error {
	inst, ok := LoadState().Installed[id]
	if !ok {
		return fmt.Errorf("app not installed: %s", id)
	}
	if orchestrator.Available() {
		if err := orchestrator.EnsureNetwork(); err != nil {
			return fmt.Errorf("docker network: %w", err)
		}
	}
	// Apply any pending per-app migrations against the OLD state/containers before
	// re-rendering with the new manifest.
	if err := runAppMigrations(dir, id, io.Discard); err != nil {
		return err
	}
	inst = LoadState().Installed[id] // reload — a migration may have changed inputs
	provided := map[string]string{}
	for k, v := range inst.Inputs {
		provided[k] = v
	}
	for k, v := range loadAppSecrets(id) {
		provided[k] = v
	}
	if err := installOne(dir, id, provided, nil, true, io.Discard); err != nil {
		return err
	}
	_ = ReloadProxy()
	_ = ReloadTor(dir)
	PruneImages(io.Discard) // reclaim disk after pulling the updated image
	return nil
}

// SetImageTag switches one service of an installed app to a different image tag
// (e.g. bump bitcoind to v31) and re-applies it. Works for any docker image:
// the tag of the service's current image is replaced. Never auto-bumped — only
// changed through this call.
func SetImageTag(dir, id, service, tag string) error {
	if service == "" || tag == "" {
		return fmt.Errorf("service and tag required")
	}
	if !validImageTag(tag) {
		return fmt.Errorf("invalid image tag: %q", tag)
	}
	state := LoadState()
	inst, ok := state.Installed[id]
	if !ok {
		return fmt.Errorf("app not installed: %s", id)
	}
	if inst.ImageTags == nil {
		inst.ImageTags = map[string]string{}
	}
	inst.ImageTags[service] = tag
	state.Installed[id] = inst
	if err := saveState(state); err != nil {
		return err
	}
	return ReapplyOne(dir, id)
}

var subdomainRe = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

// domainRe matches a full domain name (at least two labels and a letter TLD),
// e.g. app.example.com or my-node.org.
var domainRe = regexp.MustCompile(`^([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,63}$`)

// imageTagRe matches a valid docker image tag (the part after ':').
var imageTagRe = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9._-]{0,127}$`)

func validImageTag(t string) bool { return imageTagRe.MatchString(t) }

// AppSubdomain returns the effective reverse-proxy subdomain label for an app.
func AppSubdomain(id string) string { return appSubdomain(id) }

// appSubdomain returns the reverse-proxy subdomain label for an installed app:
// the operator's override, or the app id by default.
func appSubdomain(id string) string {
	if inst, ok := LoadState().Installed[id]; ok && inst.Subdomain != "" {
		return inst.Subdomain
	}
	return id
}

// SetSubdomain changes the reverse-proxy subdomain an app is served under
// (https://<sub>.<host>) and reloads Caddy. An empty value (or the app id)
// resets to the default. Validated as a DNS label.
func SetSubdomain(dir, id, sub string) error {
	sub = strings.ToLower(strings.TrimSpace(sub))
	state := LoadState()
	inst, ok := state.Installed[id]
	if !ok {
		return fmt.Errorf("app not installed: %s", id)
	}
	if sub == id {
		sub = ""
	}
	if sub != "" && !subdomainRe.MatchString(sub) {
		return fmt.Errorf("invalid subdomain: use lowercase letters, digits and hyphens")
	}
	inst.Subdomain = sub
	state.Installed[id] = inst
	if err := saveState(state); err != nil {
		return err
	}
	return ReloadProxy()
}

// SetDomain assigns a full custom domain to an app (served in addition to the
// default <subdomain>.<host>), and reloads Caddy so it requests a certificate
// for it. An empty value removes the custom domain. The operator must point the
// domain's DNS at this node for the HTTPS certificate to be issued.
func SetDomain(dir, id, domain string) error {
	domain = strings.ToLower(strings.TrimSpace(strings.TrimSuffix(domain, ".")))
	domain = strings.TrimPrefix(strings.TrimPrefix(domain, "https://"), "http://")
	state := LoadState()
	inst, ok := state.Installed[id]
	if !ok {
		return fmt.Errorf("app not installed: %s", id)
	}
	if domain != "" && !domainRe.MatchString(domain) {
		return fmt.Errorf("invalid domain: use a full name like app.example.com")
	}
	// Reject a domain already claimed by another app.
	for otherID, other := range state.Installed {
		if otherID != id && other.Domain != "" && other.Domain == domain {
			return fmt.Errorf("domain already used by %q", otherID)
		}
	}
	inst.Domain = domain
	state.Installed[id] = inst
	if err := saveState(state); err != nil {
		return err
	}
	return ReloadProxy()
}

// SetImageTags pins several services' image tags at once (e.g. "update all to
// latest") and re-applies the app a single time. Each tag is validated.
func SetImageTags(dir, id string, tags map[string]string) error {
	if len(tags) == 0 {
		return nil
	}
	state := LoadState()
	inst, ok := state.Installed[id]
	if !ok {
		return fmt.Errorf("app not installed: %s", id)
	}
	if inst.ImageTags == nil {
		inst.ImageTags = map[string]string{}
	}
	for svc, tag := range tags {
		if !validImageTag(tag) {
			return fmt.Errorf("invalid image tag for %s: %q", svc, tag)
		}
		inst.ImageTags[svc] = tag
	}
	state.Installed[id] = inst
	if err := saveState(state); err != nil {
		return err
	}
	return ReapplyOne(dir, id)
}

// ResolvedImages returns an installed app's per-service image refs (with any
// stored tag overrides applied), for display in the version picker.
func ResolvedImages(man *Manifest, id string) map[string]string {
	inst := LoadState().Installed[id]
	imgs, _ := resolveImages(man, inst.ImageTags)
	return imgs
}

// resolveImages returns each service's currently resolved image ref (the
// compose image with any stored per-service tag override applied), parsed from
// the app's compose document.
func resolveImages(man *Manifest, tags map[string]string) (map[string]string, error) {
	out := map[string]string{}
	for name, img := range orchestrator.ParseComposeImages(man.Compose) {
		if t := tags[name]; t != "" {
			img = replaceTag(img, t)
		}
		out[name] = img
	}
	return out, nil
}

// loopbackPort rewrites a compose ports entry that publishes hostPort (in the
// bare "hostPort:containerPort[/proto]" form) so the host side binds 127.0.0.1
// only. Entries that already pin a host IP, or publish a different port, are left
// untouched.
func loopbackPort(content string, hostPort int) string {
	p := strconv.Itoa(hostPort)
	re := regexp.MustCompile(`(?m)^(\s*-\s*)"?` + p + `:(\d+)(/[A-Za-z]+)?"?\s*$`)
	return re.ReplaceAllString(content, `${1}"127.0.0.1:`+p+`:${2}${3}"`)
}

// replaceTag swaps the tag of a docker image reference, preserving the registry
// (which may itself contain a ':' for a port) and stripping any digest.
func replaceTag(image, tag string) string {
	name := image
	if at := strings.Index(name, "@"); at >= 0 {
		name = name[:at]
	}
	prefix := ""
	last := name
	if slash := strings.LastIndex(name, "/"); slash >= 0 {
		prefix = name[:slash+1]
		last = name[slash+1:]
	}
	if colon := strings.Index(last, ":"); colon >= 0 {
		last = last[:colon]
	}
	return prefix + last + ":" + tag
}

// Uninstall stops the app's containers and removes it from the state/registry.
// With purge, container volumes and the runtime directory are removed too.
// Refuses to remove an app that another installed app depends on.
func Uninstall(dir, id string, purge bool) error {
	state := LoadState()
	var blockers []string
	for other := range state.Installed {
		if other == id {
			continue
		}
		man, err := Find(dir, other)
		if err != nil {
			continue
		}
		for _, dep := range man.Dependencies {
			if dep == id {
				blockers = append(blockers, other)
			}
		}
	}
	if len(blockers) > 0 {
		sort.Strings(blockers)
		return fmt.Errorf("cannot remove %s: required by %s", id, strings.Join(blockers, ", "))
	}

	if orchestrator.Available() {
		_ = orchestrator.Down(id, paths.AppComposeFile(id), purge)
	}

	delete(state.Installed, id)
	if err := saveState(state); err != nil {
		return err
	}

	registry := loadRegistry()
	delete(registry, id)
	_ = saveRegistry(registry)
	_ = mergeAppSecrets(id, nil)

	if purge {
		_ = os.RemoveAll(paths.AppRuntimeDir(id))
	}
	_ = ReloadProxy() // best-effort: refresh reverse-proxy routes
	_ = ReloadTor(dir)   // best-effort: refresh Tor hidden services
	return nil
}

// installPlan returns the install order: dependencies first (post-order DFS),
// skipping ones already installed, with the target always included last.
func installPlan(dir, id string) ([]string, error) {
	state := LoadState()
	var order []string
	seen := map[string]bool{}

	var visit func(string) error
	visit = func(appID string) error {
		if seen[appID] {
			return nil
		}
		seen[appID] = true
		man, err := Find(dir, appID)
		if err != nil {
			return err
		}
		for _, dep := range man.Dependencies {
			if err := visit(dep); err != nil {
				return err
			}
		}
		if _, installed := state.Installed[appID]; !installed || appID == id {
			order = append(order, appID)
		}
		return nil
	}
	if err := visit(id); err != nil {
		return nil, err
	}
	return order, nil
}

// installOne installs a single app. For the target, missing required inputs are
// an error; for auto-installed dependencies, required secrets are generated and
// required non-secret fields must have a default (else we bail and ask the
// operator to install the dependency manually).
func installOne(dir, appID string, provided, imageTagOverride map[string]string, isTarget bool, out io.Writer) error {
	man, err := Find(dir, appID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(man.Compose) == "" {
		return fmt.Errorf("app %s has no compose document", appID)
	}

	nonSecret := map[string]string{}
	secret := map[string]string{}
	for _, in := range man.Inputs {
		val, ok := provided[in.Key]
		if !ok || val == "" {
			switch {
			case in.Default != nil:
				val = fmt.Sprint(in.Default)
			case in.Generate && (in.Secret || in.Type == "password"):
				if val, err = randomToken(16); err != nil {
					return err
				}
			case in.Required && (in.Secret || in.Type == "password"):
				if isTarget {
					return fmt.Errorf("missing required field: %s", in.Key)
				}
				if val, err = randomToken(16); err != nil {
					return err
				}
			case in.Required:
				if isTarget {
					return fmt.Errorf("missing required field: %s", in.Key)
				}
				return fmt.Errorf("dependency needs %q (no default) — install %s manually first", in.Key, appID)
			default:
				continue
			}
		}
		if in.MinLength > 0 && len(val) < in.MinLength {
			return fmt.Errorf("%s: %d characters minimum", in.Key, in.MinLength)
		}
		if err := validateInput(in, val); err != nil {
			return err
		}
		if in.Secret || in.Type == "password" {
			secret[in.Key] = val
		} else {
			nonSecret[in.Key] = val
		}
	}

	registry := loadRegistry()

	// Expose the node's own coordinates and this app's own public URL so
	// manifests can set their public-origin env (PUBLIC_ORIGIN, …) to the HTTPS
	// URL Caddy serves — ${self.exports.url} — without the operator typing it and
	// without the app ever terminating TLS itself. Templating-only: both are
	// stripped before saveRegistry.
	registry["node"] = nodeExports()
	registry["self"] = selfExports(man)

	// Resolve this app's exports from its own inputs/secrets.
	exports := map[string]string{}
	for k, v := range man.Exports {
		rv, rerr := resolveValue(fmt.Sprint(v), nonSecret, secret, registry)
		if rerr != nil {
			continue // skip exports that reference unset inputs/secrets
		}
		exports[k] = rv
	}

	// Make sure each already-installed dependency has its exports in the
	// registry (older installs predate the registry — backfill them).
	for _, dep := range man.Dependencies {
		ensureExports(dir, dep, registry)
	}

	// Per-service image tag overrides (chosen in the version picker / install
	// form). Explicit override wins; otherwise reuse the stored choice so a
	// reapply keeps the operator's pinned versions. Never auto-bumped.
	imageTags := imageTagOverride
	if len(imageTags) == 0 {
		if inst, ok := LoadState().Installed[appID]; ok {
			imageTags = inst.ImageTags
		}
	}

	// Render config-file templates next to the compose file so the compose can
	// bind-mount them with a relative path (e.g. ./config/lnd.conf:/path:ro).
	if len(man.Configs) > 0 {
		if err := os.MkdirAll(paths.AppConfigDir(appID), 0o700); err != nil {
			return err
		}
		for _, c := range man.Configs {
			content, rerr := resolveValue(c.Content, nonSecret, secret, registry)
			if rerr != nil {
				return fmt.Errorf("config %s: %w", c.Path, rerr)
			}
			// 0644 so the (often non-root) container user can read the bind-mounted
			// config. The file lives under root-owned /var/lib/slashnode.
			hostFile := filepath.Join(paths.AppConfigDir(appID), filepath.Base(c.Path))
			if err := os.WriteFile(hostFile, []byte(content), 0o644); err != nil {
				return err
			}
		}
	}

	// Render the compose document: template ${input}/${secret}/${dep.exports.key}
	// references (leaving any compose-native ${VAR} untouched), then apply the
	// per-service image tag overrides by replacing each service's image ref.
	content := templateRefs(man.Compose, nonSecret, secret, registry)
	for svc, img := range orchestrator.ParseComposeImages(content) {
		if t := imageTags[svc]; t != "" {
			// The tag is written verbatim into the compose `image:` line, so it
			// must be a valid docker tag — otherwise a crafted tag could break out
			// of the YAML scalar and inject sibling keys.
			if !validImageTag(t) {
				return fmt.Errorf("invalid image tag for %s: %q", svc, t)
			}
			content = strings.Replace(content, img, replaceTag(img, t), 1)
		}
	}
	// In server mode (Caddy fronts each app at https://<id>.<host>), bind the
	// web-UI backend to loopback so the plain-HTTP port isn't publicly exposed —
	// only Caddy (HTTPS) and Tor reach it. In local mode the app subdomain isn't
	// resolvable over mDNS, so the web port stays published for direct access.
	if man.Web != nil && man.Web.Port > 0 {
		if cfg, cerr := config.Load(paths.ConfigFile()); cerr == nil && cfg.Access.Mode == "server" {
			content = loopbackPort(content, man.Web.Port)
		}
	}
	if err := os.MkdirAll(paths.AppRuntimeDir(appID), 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(paths.AppComposeFile(appID), []byte(content), 0o600); err != nil {
		return err
	}

	// Launch (only when a Docker daemon is reachable). Pull first so updates
	// fetch newer images for the same tag. Output is streamed to out (io.Discard
	// for non-interactive callers).
	if orchestrator.Available() {
		fmt.Fprintf(out, "--> pulling images for %s\n", appID)
		_ = orchestrator.PullStreamed(appID, paths.AppComposeFile(appID), out)
		fmt.Fprintf(out, "--> starting %s\n", appID)
		if err := orchestrator.UpStreamed(appID, paths.AppComposeFile(appID), out); err != nil {
			return err
		}
	}

	// Persist registry, secrets and installed state. The synthetic "node"/"self"
	// entries are templating-only — never persist them.
	delete(registry, "node")
	delete(registry, "self")
	registry[appID] = exports
	if err := saveRegistry(registry); err != nil {
		return err
	}
	if err := mergeAppSecrets(appID, secret); err != nil {
		return err
	}
	webPort := 0
	if man.Web != nil {
		webPort = man.Web.Port
	}
	state := LoadState()
	prev := state.Installed[appID]
	state.Installed[appID] = InstalledApp{
		ID:               appID,
		Version:          man.Version,
		ImageTags:        imageTags,
		Subdomain:        prev.Subdomain,          // preserve the subdomain override across updates
		Domain:           prev.Domain,             // preserve the custom domain across updates
		MigrationVersion: appMigrationLatest(man), // current after a successful install/reapply
		InstalledAt:      time.Now().UTC().Format(time.RFC3339),
		Inputs:           nonSecret,
		WebPort:          webPort,
	}
	return saveState(state)
}

// ensureExports backfills a dependency's exports into the registry if missing,
// by re-deriving them from the dependency's manifest and its stored
// inputs/secrets (no container changes). Handles deps installed before the
// registry existed.
func ensureExports(dir, depID string, registry map[string]map[string]string) {
	if _, ok := registry[depID]; ok {
		return
	}
	dep, err := Find(dir, depID)
	if err != nil {
		return
	}
	inst, ok := LoadState().Installed[depID]
	if !ok {
		return
	}
	inputs := inst.Inputs
	secrets := loadAppSecrets(depID)
	exp := map[string]string{}
	for k, v := range dep.Exports {
		if rv, rerr := resolveValue(fmt.Sprint(v), inputs, secrets, registry); rerr == nil {
			exp[k] = rv
		}
	}
	registry[depID] = exp
	_ = saveRegistry(registry)
}

var refRe = regexp.MustCompile(`\$\{([^}]+)\}`)

// resolveValue expands ${input.KEY}, ${secret.KEY} and ${app.exports.key}
// references in s. Returns an error if any reference cannot be resolved.
func resolveValue(s string, inputs, secrets map[string]string, registry map[string]map[string]string) (string, error) {
	var resErr error
	out := refRe.ReplaceAllStringFunc(s, func(m string) string {
		inner := m[2 : len(m)-1]
		switch {
		case strings.HasPrefix(inner, "input."):
			if v, ok := inputs[inner[len("input."):]]; ok {
				return v
			}
			resErr = fmt.Errorf("unknown input %q", inner)
		case strings.HasPrefix(inner, "secret."):
			if v, ok := secrets[inner[len("secret."):]]; ok {
				return v
			}
			resErr = fmt.Errorf("unknown secret %q", inner)
		case strings.Contains(inner, ".exports."):
			parts := strings.SplitN(inner, ".exports.", 2)
			if e, ok := registry[parts[0]]; ok {
				if v, ok := e[parts[1]]; ok {
					return v
				}
			}
			resErr = fmt.Errorf("unresolved reference %q (is %s installed?)", inner, parts[0])
		default:
			resErr = fmt.Errorf("unknown reference %q", inner)
		}
		return m
	})
	return out, resErr
}

// templateRefs expands ${input.KEY}, ${secret.KEY} and ${app.exports.key}
// references in s, leaving any other ${...} (e.g. a compose file's native
// ${VARIABLE} interpolation) untouched. Used to render compose documents so
// docker-compose-only projects stay compatible.
func templateRefs(s string, inputs, secrets map[string]string, registry map[string]map[string]string) string {
	return refRe.ReplaceAllStringFunc(s, func(m string) string {
		inner := m[2 : len(m)-1]
		// Values are escaped for the double-quoted YAML scalar context manifests
		// use, so a value can't break out of the compose structure.
		switch {
		case strings.HasPrefix(inner, "input."):
			if v, ok := inputs[inner[len("input."):]]; ok {
				return composeEscape(v)
			}
		case strings.HasPrefix(inner, "secret."):
			if v, ok := secrets[inner[len("secret."):]]; ok {
				return composeEscape(v)
			}
		case strings.Contains(inner, ".exports."):
			parts := strings.SplitN(inner, ".exports.", 2)
			if e, ok := registry[parts[0]]; ok {
				if v, ok := e[parts[1]]; ok {
					return composeEscape(v)
				}
			}
		}
		return m // leave unknown refs untouched
	})
}

// validateInput enforces an input's declared constraints server-side. This is a
// security boundary, not just UX: a value is substituted verbatim into the
// app's docker-compose document and config files, so a crafted value must not
// be able to break out of the structure it lands in. Rejecting control
// characters removes the newline primitive used to inject sibling YAML keys
// (e.g. `privileged: true`, a host bind-mount) or extra config directives, and
// option/number checks stop off-menu values.
func validateInput(in Input, val string) error {
	for _, r := range val {
		if r < 0x20 || r == 0x7f {
			return fmt.Errorf("%s: control characters are not allowed", in.Key)
		}
	}
	switch in.Type {
	case "select":
		if len(in.Options) > 0 {
			for _, o := range in.Options {
				if o == val {
					return nil
				}
			}
			return fmt.Errorf("%s: %q is not an allowed option", in.Key, val)
		}
	case "number":
		f, err := strconv.ParseFloat(strings.TrimSpace(val), 64)
		if err != nil {
			return fmt.Errorf("%s: must be a number", in.Key)
		}
		if in.Min != nil && f < *in.Min {
			return fmt.Errorf("%s: must be >= %v", in.Key, *in.Min)
		}
		if in.Max != nil && f > *in.Max {
			return fmt.Errorf("%s: must be <= %v", in.Key, *in.Max)
		}
	case "boolean":
		if val != "true" && val != "false" {
			return fmt.Errorf("%s: must be true or false", in.Key)
		}
	}
	return nil
}

// composeEscape escapes a value for safe inclusion inside a double-quoted YAML
// scalar in a compose document (the convention manifests use, e.g.
// KEY: "${input.X}"). Backslash and quote are escaped so a value containing them
// can't break the document or close the scalar early; `$` is doubled so docker
// compose's own ${VAR}/$VAR interpolation leaves the literal value intact (e.g.
// a password like `pa$$w` survives). Control characters are already rejected by
// validateInput, so structural newline injection is impossible.
func composeEscape(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, `$`, `$$`)
	return s
}

func randomToken(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func loadRegistry() map[string]map[string]string {
	reg := map[string]map[string]string{}
	if b, err := os.ReadFile(paths.RegistryFile()); err == nil {
		_ = json.Unmarshal(b, &reg)
	}
	if reg == nil {
		reg = map[string]map[string]string{}
	}
	return reg
}

func saveRegistry(reg map[string]map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(paths.RegistryFile()), 0o700); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(reg, "", "  ")
	return os.WriteFile(paths.RegistryFile(), append(b, '\n'), 0o600)
}
