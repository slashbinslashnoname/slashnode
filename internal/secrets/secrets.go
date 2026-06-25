// Package secrets manages the generation and storage of SlashNode's secrets
// (admin password, API token, session secret) in a JSON file with mode 0600.
//
// SECURITY NOTE: the password hashing currently uses a PBKDF2-like approach
// via salted SHA-256 over several iterations, in stdlib only (zero
// dependencies). For a crypto-conscious audience, we will migrate to argon2id /
// bcrypt (golang.org/x/crypto) — see TODO and the open question to the maintainer.
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

// Secrets contains the persisted secrets of the node.
type Secrets struct {
	AdminPasswordHash string `json:"admin_password_hash"`
	AdminPasswordSalt string `json:"admin_password_salt"`
	SessionSecret     string `json:"session_secret"`
	APIToken          string `json:"api_token"`
}

// Generate produces a new set of secrets along with the initial admin password
// in clear text (to be displayed only once, never persisted in clear text on
// the secrets.json side).
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

// SetPassword sets the admin password to pw (fresh salt + hash).
func (s *Secrets) SetPassword(pw string) error {
	salt, err := randomToken(16)
	if err != nil {
		return err
	}
	s.AdminPasswordSalt = salt
	s.AdminPasswordHash = hashPassword(pw, salt)
	return nil
}

// Verify compares a candidate password against the stored hash (constant time).
func (s *Secrets) Verify(password string) bool {
	want, _ := hex.DecodeString(s.AdminPasswordHash)
	got, _ := hex.DecodeString(hashPassword(password, s.AdminPasswordSalt))
	return subtle.ConstantTimeCompare(want, got) == 1
}

// Load reads the secrets from path.
func Load(path string) (*Secrets, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s Secrets
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, fmt.Errorf("invalid secrets (%s): %w", path, err)
	}
	return &s, nil
}

// Save writes the secrets in mode 0600.
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

// hashPassword applies salted SHA-256 over hashIterations rounds (basic
// stretching, stdlib only). TODO: migrate to argon2id.
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
