// Package apps loads the catalog of application manifests and manages the
// installation state.
//
// NOTE: the actual Docker orchestration (docker compose up, templating,
// exports/wiring resolution) arrives with the registry engine. For now,
// Install validates the inputs, separates the secrets and persists the
// "installed" state so that the App Store is fully browsable and the install
// form (generated from `inputs`) works end to end.
package apps

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/slashbinslashnoname/slashnode/internal/paths"
)

// Input describes a user input (→ container environment variable).
type Input struct {
	Key         string   `json:"key"`
	Label       string   `json:"label"`
	Type        string   `json:"type"`
	Required    bool     `json:"required,omitempty"`
	Default     any      `json:"default,omitempty"`
	Placeholder string   `json:"placeholder,omitempty"`
	Help        string   `json:"help,omitempty"`
	Secret      bool     `json:"secret,omitempty"`
	Generate    bool     `json:"generate,omitempty"`
	Options     []string `json:"options,omitempty"`
	MinLength   int      `json:"minLength,omitempty"`
	Min         *float64 `json:"min,omitempty"`
	Max         *float64 `json:"max,omitempty"`
}

// Web declares an app's primary web UI (host port + path), used by the
// frontend to build an "Open" link.
type Web struct {
	Port int    `json:"port"`
	Path string `json:"path,omitempty"`
}

// Endpoint declares a connection URL/address exposed by an app (RPC, REST, S3,
// a game server, …). The frontend renders these as copyable URLs built from the
// node host + Port; http/https endpoints also get an open link.
type Endpoint struct {
	Label  string `json:"label"`
	Scheme string `json:"scheme,omitempty"` // http, https, tcp, … (empty = host:port only)
	Port   int    `json:"port"`
	Path   string `json:"path,omitempty"`
}

// ProbeStat declares one display value: a Label, the result Field to read (empty
// = the scalar result, e.g. geth eth_blockNumber), and whether it is Hex.
type ProbeStat struct {
	Label string `json:"label"`
	Field string `json:"field,omitempty"`
	Hex   bool   `json:"hex,omitempty"`
}

// Probe declares how to check what an app is doing:
//   - "http"     : GET the host port/path, report reachability.
//   - "rpc"      : Bitcoin/Ethereum-style JSON-RPC (optional basic auth).
//   - "electrum" : Electrum protocol over TCP (blockchain.headers.subscribe).
//   - "lnd"      : LND REST /v1/getinfo using the admin macaroon.
//
// Display declares which fields to surface and how to label them — the frontend
// renders these generically rather than hardcoding per-app stats.
type Probe struct {
	Type       string      `json:"type"`
	Port       int         `json:"port"`
	Path       string      `json:"path,omitempty"`
	Method     string      `json:"method,omitempty"`
	UserInput  string      `json:"userInput,omitempty"`
	PassSecret string      `json:"passSecret,omitempty"`
	Display    []ProbeStat `json:"display,omitempty"`
}

// ConfigFile declares a config file rendered from a ${...} template and
// bind-mounted into a service at Path. Service defaults to the app id.
type ConfigFile struct {
	Service string `json:"service,omitempty"`
	Path    string `json:"path"`
	Content string `json:"content"`
}

// Manifest is an app manifest (slashnode-app.json).
type Manifest struct {
	ManifestVersion int             `json:"manifestVersion"`
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	Version         string          `json:"version"`
	Category        string          `json:"category"`
	Description     string          `json:"description"`
	Icon            string          `json:"icon"`
	Dependencies    []string        `json:"dependencies"`
	Inputs          []Input         `json:"inputs"`
	Services        json.RawMessage `json:"services,omitempty"`
	Exports         map[string]any  `json:"exports,omitempty"`
	Wiring          map[string]any  `json:"wiring,omitempty"`
	Web             *Web            `json:"web,omitempty"`
	Endpoints       []Endpoint      `json:"endpoints,omitempty"`
	Probe           *Probe          `json:"probe,omitempty"`
	Configs         []ConfigFile    `json:"configs,omitempty"`
	Notes           string          `json:"notes,omitempty"`
}

// CatalogEntry enriches a manifest with its installation state for the UI.
type CatalogEntry struct {
	Manifest
	Installed        bool   `json:"installed"`
	InstalledVersion string `json:"installed_version,omitempty"`
	UpdateAvailable  bool   `json:"update_available"`
	URL              string `json:"url,omitempty"`       // reverse-proxy URL (set by the API layer)
	OnionURL         string `json:"onion_url,omitempty"` // Tor hidden-service URL (set by the API layer)
}

// LoadCatalog reads all manifests dir/*/slashnode-app.json, sorted by name.
func LoadCatalog(dir string) ([]Manifest, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*", "slashnode-app.json"))
	if err != nil {
		return nil, err
	}
	var out []Manifest
	for _, m := range matches {
		b, err := os.ReadFile(m)
		if err != nil {
			return nil, err
		}
		var man Manifest
		if err := json.Unmarshal(b, &man); err != nil {
			return nil, fmt.Errorf("invalid manifest (%s): %w", m, err)
		}
		out = append(out, man)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// Find returns the manifest with the given id.
func Find(dir, id string) (*Manifest, error) {
	cat, err := LoadCatalog(dir)
	if err != nil {
		return nil, err
	}
	for i := range cat {
		if cat[i].ID == id {
			return &cat[i], nil
		}
	}
	return nil, fmt.Errorf("unknown app: %s", id)
}

// InstalledApp is the persisted state of an installed app (secrets excluded).
type InstalledApp struct {
	ID          string            `json:"id"`
	Version     string            `json:"version"`
	InstalledAt string            `json:"installed_at"`
	Inputs      map[string]string `json:"inputs"`
	WebPort     int               `json:"web_port,omitempty"` // host port of the app's web UI (for the reverse proxy)
}

// State is the installation state (var/lib/slashnode/apps.json).
type State struct {
	Installed map[string]InstalledApp `json:"installed"`
}

// LoadState reads the installation state (empty if absent).
func LoadState() *State {
	st := &State{Installed: map[string]InstalledApp{}}
	b, err := os.ReadFile(paths.AppsStateFile())
	if err != nil {
		return st
	}
	_ = json.Unmarshal(b, st)
	if st.Installed == nil {
		st.Installed = map[string]InstalledApp{}
	}
	return st
}

// Catalog returns the catalog annotated with the installation state.
func Catalog(dir string) ([]CatalogEntry, error) {
	cat, err := LoadCatalog(dir)
	if err != nil {
		return nil, err
	}
	state := LoadState()
	out := make([]CatalogEntry, 0, len(cat))
	for _, m := range cat {
		inst, installed := state.Installed[m.ID]
		entry := CatalogEntry{Manifest: m, Installed: installed}
		if installed {
			entry.InstalledVersion = inst.Version
			entry.UpdateAvailable = inst.Version != m.Version
		}
		out = append(out, entry)
	}
	return out, nil
}

func saveState(s *State) error {
	if err := os.MkdirAll(filepath.Dir(paths.AppsStateFile()), 0o700); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(s, "", "  ")
	return os.WriteFile(paths.AppsStateFile(), append(b, '\n'), 0o644)
}

// loadAppSecrets returns the stored secret inputs for an app (empty if none).
func loadAppSecrets(id string) map[string]string {
	all := map[string]map[string]string{}
	if b, err := os.ReadFile(paths.AppSecretsFile()); err == nil {
		_ = json.Unmarshal(b, &all)
	}
	if s, ok := all[id]; ok {
		return s
	}
	return map[string]string{}
}

// mergeAppSecrets merges (or purges if values==nil) an app's secrets into
// app-secrets.json (mode 0600).
func mergeAppSecrets(id string, values map[string]string) error {
	all := map[string]map[string]string{}
	if b, err := os.ReadFile(paths.AppSecretsFile()); err == nil {
		_ = json.Unmarshal(b, &all)
	}
	if values == nil {
		delete(all, id)
	} else if len(values) > 0 {
		all[id] = values
	}
	if err := os.MkdirAll(filepath.Dir(paths.AppSecretsFile()), 0o700); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(all, "", "  ")
	return os.WriteFile(paths.AppSecretsFile(), append(b, '\n'), 0o600)
}
