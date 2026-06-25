package apps

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/slashbinslashnoname/slashnode/internal/orchestrator"
	"github.com/slashbinslashnoname/slashnode/internal/paths"
)

// Install resolves the dependency graph (installing any missing dependencies
// first), renders a Compose file for each app, launches it via Docker and
// records its exports in the service registry so consumers can wire to it.
func Install(dir, id string, inputs map[string]string) error {
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
		var provided map[string]string
		if isTarget {
			provided = inputs
		}
		if err := installOne(dir, appID, provided, isTarget); err != nil {
			return fmt.Errorf("installing %s: %w", appID, err)
		}
	}
	_ = ReloadProxy() // best-effort: refresh reverse-proxy routes
	return nil
}

// Uninstall stops the app's containers and removes it from the state/registry.
// With purge, container volumes and the runtime directory are removed too.
func Uninstall(id string, purge bool) error {
	if orchestrator.Available() {
		_ = orchestrator.Down(id, paths.AppComposeFile(id), purge)
	}

	state := LoadState()
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
func installOne(dir, appID string, provided map[string]string, isTarget bool) error {
	man, err := Find(dir, appID)
	if err != nil {
		return err
	}

	nonSecret := map[string]string{}
	secret := map[string]string{}
	for _, in := range man.Inputs {
		val, ok := provided[in.Key]
		if !ok || val == "" {
			switch {
			case in.Default != nil:
				val = fmt.Sprint(in.Default)
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
		if in.Secret || in.Type == "password" {
			secret[in.Key] = val
		} else {
			nonSecret[in.Key] = val
		}
	}

	registry := loadRegistry()

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

	// Build the environment: inputs + wiring (dependency exports).
	env := map[string]string{}
	for k, v := range nonSecret {
		env[k] = v
	}
	for k, v := range secret {
		env[k] = v
	}
	for k, v := range man.Wiring {
		rv, rerr := resolveValue(fmt.Sprint(v), nonSecret, secret, registry)
		if rerr != nil {
			return fmt.Errorf("wiring %s: %w", k, rerr)
		}
		env[envName(k)] = rv
	}

	// Render and write the Compose file.
	services, err := orchestrator.ParseServices(man.Services)
	if err != nil {
		return err
	}
	// Resolve ${input}/${secret}/${dep.exports.key} references inside each
	// service's environment values (config templating).
	for name, s := range services {
		for k, v := range s.Environment {
			if rv, rerr := resolveValue(v, nonSecret, secret, registry); rerr == nil {
				s.Environment[k] = rv
			}
		}
		if rv, rerr := resolveValue(s.Command, nonSecret, secret, registry); rerr == nil {
			s.Command = rv
		}
		services[name] = s
	}

	// Render config-file templates to host files and bind-mount them.
	configMounts := map[string][]string{}
	for i, c := range man.Configs {
		content, rerr := resolveValue(c.Content, nonSecret, secret, registry)
		if rerr != nil {
			return fmt.Errorf("config %s: %w", c.Path, rerr)
		}
		if err := os.MkdirAll(paths.AppConfigDir(appID), 0o700); err != nil {
			return err
		}
		hostFile := filepath.Join(paths.AppConfigDir(appID),
			fmt.Sprintf("%d-%s", i, filepath.Base(c.Path)))
		if err := os.WriteFile(hostFile, []byte(content), 0o600); err != nil {
			return err
		}
		svc := c.Service
		if svc == "" {
			svc = appID
		}
		configMounts[svc] = append(configMounts[svc], hostFile+":"+c.Path+":ro")
	}

	compose, err := orchestrator.BuildCompose(appID, services, env, configMounts)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(paths.AppRuntimeDir(appID), 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(paths.AppComposeFile(appID), compose, 0o600); err != nil {
		return err
	}

	// Launch (only when a Docker daemon is reachable).
	if orchestrator.Available() {
		if err := orchestrator.Up(appID, paths.AppComposeFile(appID)); err != nil {
			return err
		}
	}

	// Persist registry, secrets and installed state.
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
	state.Installed[appID] = InstalledApp{
		ID:          appID,
		Version:     man.Version,
		InstalledAt: time.Now().UTC().Format(time.RFC3339),
		Inputs:      nonSecret,
		WebPort:     webPort,
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

// envName converts a wiring key (e.g. "bitcoind.rpchost") into an environment
// variable name (e.g. "BITCOIND_RPCHOST").
func envName(k string) string {
	var b strings.Builder
	for _, r := range strings.ToUpper(k) {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	return b.String()
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
