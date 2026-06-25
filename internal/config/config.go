// Package config définit la structure de configuration de SlashNode et son
// chargement/sauvegarde au format JSON (/etc/slashnode/config.json).
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

// HTTP regroupe les paramètres réseau.
//
//	Port    : port public du front Next.js (servi/lancé par le démon).
//	APIPort : port de l'API Go locale (liée à 127.0.0.1, consommée par le front).
type HTTP struct {
	Bind    string `json:"bind"`
	Port    int    `json:"port"`
	APIPort int    `json:"api_port"`
}

// Update contrôle la politique de mise à jour du démon.
//
//	Policy  : "notify" (signale, l'opérateur applique) ou "auto".
//	Channel : "stable" (par défaut) ou "beta".
type Update struct {
	Policy  string `json:"policy"`
	Channel string `json:"channel"`
}

// Theme contrôle l'apparence de l'UI.
type Theme struct {
	// Mode : "system", "light" ou "dark".
	Mode string `json:"mode"`
	// Primary : couleur d'accent (hex). Rouge par défaut.
	Primary string `json:"primary"`
}

// Config est la configuration persistée du nœud.
type Config struct {
	Version   string `json:"version"`
	NodeID    string `json:"node_id"`
	Hostname  string `json:"hostname"`
	DataDir   string `json:"data_dir"`
	HTTP      HTTP   `json:"http"`
	Theme     Theme  `json:"theme"`
	Update    Update `json:"update"`
	CreatedAt string `json:"created_at"`
}

// Default construit une configuration par défaut, avec un NodeID aléatoire.
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
		Theme:     Theme{Mode: "system", Primary: "#e5484d"}, // rouge
		Update:    Update{Policy: "notify", Channel: "stable"},
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// Load lit et décode la configuration depuis path.
func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("config invalide (%s) : %w", path, err)
	}
	return &c, nil
}

// Save écrit la configuration en JSON indenté (mode 0644).
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
