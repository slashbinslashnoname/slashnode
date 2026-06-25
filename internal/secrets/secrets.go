// Package secrets gère la génération et le stockage des secrets de SlashNode
// (mot de passe admin, jeton d'API, secret de session) dans un fichier JSON en
// mode 0600.
//
// NOTE sécurité : le hachage du mot de passe utilise pour l'instant PBKDF2-like
// via SHA-256 salé sur plusieurs itérations, en stdlib uniquement (zéro
// dépendance). Pour un public crypto-conscient, on migrera vers argon2id /
// bcrypt (golang.org/x/crypto) — voir TODO et la question ouverte au mainteneur.
package secrets

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const hashIterations = 200_000

// Secrets contient les secrets persistés du nœud.
type Secrets struct {
	AdminPasswordHash string `json:"admin_password_hash"`
	AdminPasswordSalt string `json:"admin_password_salt"`
	SessionSecret     string `json:"session_secret"`
	APIToken          string `json:"api_token"`
}

// Generate produit un nouveau jeu de secrets ainsi que le mot de passe admin
// initial en clair (à afficher une seule fois, jamais persisté en clair côté
// secrets.json).
func Generate() (s *Secrets, initialPassword string, err error) {
	initialPassword, err = randomToken(12)
	if err != nil {
		return nil, "", err
	}
	salt, err := randomToken(16)
	if err != nil {
		return nil, "", err
	}
	session, err := randomToken(32)
	if err != nil {
		return nil, "", err
	}
	apiToken, err := randomToken(32)
	if err != nil {
		return nil, "", err
	}
	return &Secrets{
		AdminPasswordHash: hashPassword(initialPassword, salt),
		AdminPasswordSalt: salt,
		SessionSecret:     session,
		APIToken:          apiToken,
	}, initialPassword, nil
}

// Verify compare un mot de passe candidat au hash stocké (temps constant).
func (s *Secrets) Verify(password string) bool {
	want, _ := hex.DecodeString(s.AdminPasswordHash)
	got, _ := hex.DecodeString(hashPassword(password, s.AdminPasswordSalt))
	return subtle.ConstantTimeCompare(want, got) == 1
}

// Load lit les secrets depuis path.
func Load(path string) (*Secrets, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s Secrets
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, fmt.Errorf("secrets invalides (%s) : %w", path, err)
	}
	return &s, nil
}

// Save écrit les secrets en mode 0600.
func (s *Secrets) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0o600)
}

// hashPassword applique SHA-256 salé sur hashIterations tours (stretching
// basique, stdlib uniquement). TODO: migrer vers argon2id.
func hashPassword(password, salt string) string {
	data := []byte(salt + password)
	for i := 0; i < hashIterations; i++ {
		sum := sha256.Sum256(data)
		data = sum[:]
	}
	return hex.EncodeToString(data)
}

func randomToken(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
