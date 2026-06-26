// Package config defines the configuration structure of SlashNode and its
// loading/saving in JSON format (/etc/slashnode/config.json).
package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// HTTP groups the network parameters.
//
//	Port    : public port of the Next.js front (served/launched by the daemon).
//	APIPort : port of the local Go API (bound to 127.0.0.1, consumed by the front).
type HTTP struct {
	Bind    string `json:"bind"`
	Port    int    `json:"port"`
	APIPort int    `json:"api_port"`
}

// Update controls the daemon's update policy.
//
//	Policy  : "notify" (signals, the operator applies) or "auto".
//	Channel : "stable" (default) or "beta".
type Update struct {
	Policy  string `json:"policy"`
	Channel string `json:"channel"`
}

// Access controls how the node is reached and whether the UI is protected.
//
//	Mode              : "local" (LAN, slashnode.local) or "server" (public address).
//	Address           : public host/domain used in server mode (e.g. node.example.com).
//	PasswordProtected : when true, the web UI requires a login (admin password).
//	                    Optional in local mode, always on in server mode.
type Access struct {
	Mode              string `json:"mode"`
	Address           string `json:"address"`
	PasswordProtected bool   `json:"password_protected"`
}

// Tor controls Tor hidden-service exposure for the UI and apps.
type Tor struct {
	Enabled bool `json:"enabled"`
}

// Theme controls the appearance of the UI.
type Theme struct {
	// Mode : "system", "light" or "dark".
	Mode string `json:"mode"`
	// Primary : accent color (hex). Red by default.
	Primary string `json:"primary"`
}

// Config is the persisted configuration of the node.
type Config struct {
	Version   string `json:"version"`
	NodeID    string `json:"node_id"`
	Hostname  string `json:"hostname"`
	DataDir   string `json:"data_dir"`
	HTTP      HTTP   `json:"http"`
	Access    Access `json:"access"`
	Tor       Tor    `json:"tor"`
	Theme     Theme  `json:"theme"`
	Update    Update `json:"update"`
	CreatedAt string `json:"created_at"`
}

// Default builds a default configuration, with a random NodeID.
func Default(version, dataDir string) (*Config, error) {
	id, err := randomHex(8)
	if err != nil {
		return nil, err
	}
	return &Config{
		Version:   version,
		NodeID:    id,
		Hostname:  "slashnode.local",
		DataDir:   dataDir,
		HTTP:      HTTP{Bind: "0.0.0.0", Port: 8080, APIPort: 8081},
		Access:    Access{Mode: "local", Address: "", PasswordProtected: false},
		Tor:       Tor{Enabled: true}, // Tor on by default: UI + apps reachable over .onion
		Theme:     Theme{Mode: "system", Primary: "#e5484d"}, // red
		Update:    Update{Policy: "notify", Channel: "stable"},
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// Load reads and decodes the configuration from path.
func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("invalid config (%s): %w", path, err)
	}
	return &c, nil
}

// Save writes the configuration as indented JSON (mode 0644).
func (c *Config) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0o644)
}

func randomHex(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
