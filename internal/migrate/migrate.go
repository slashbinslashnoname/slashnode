// Package migrate applies ordered, idempotent schema migrations to the node's
// on-disk state (config.json, apps.json, registry.json) so a "hard" update can
// transform state written by an older version. It snapshots the state before
// migrating and rolls back on failure. Run on bootstrap (init) and on every
// daemon start (so the freshly-updated binary migrates before serving).
package migrate

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/slashbinslashnoname/slashnode/internal/paths"
)

// Migration transforms the node's on-disk state from version-1 to Version. Up
// must be idempotent. NEVER reorder or renumber existing migrations — only
// append new ones with the next integer.
type Migration struct {
	Version int
	Name    string
	Up      func() error
}

// migrations is the ordered list of node-state migrations. Empty until the first
// breaking on-disk change ships; the framework records version 0 meanwhile.
var migrations = []Migration{
	// {Version: 1, Name: "example", Up: func() error { return nil }},
}

// State records the applied node schema version.
type State struct {
	Node int `json:"node"`
}

// Latest is the highest known node migration version.
func Latest() int {
	if len(migrations) == 0 {
		return 0
	}
	return migrations[len(migrations)-1].Version
}

// Pending returns the migrations not yet applied, in order.
func Pending() []Migration {
	cur := Load().Node
	var out []Migration
	for _, m := range migrations {
		if m.Version > cur {
			out = append(out, m)
		}
	}
	return out
}

// Run applies all pending node migrations in order. It first snapshots the
// state files; if any migration fails it restores the snapshot and returns the
// error, leaving the node on its previous schema version. A no-op when nothing
// is pending.
func Run(logw io.Writer) error {
	pending := Pending()
	if len(pending) == 0 {
		return nil
	}
	snap, err := Snapshot()
	if err != nil {
		return fmt.Errorf("migration snapshot: %w", err)
	}
	fmt.Fprintf(logw, "→ %d pending node migration(s); snapshot at %s\n", len(pending), snap)
	st := Load()
	for _, m := range pending {
		fmt.Fprintf(logw, "  • node schema → v%d (%s)\n", m.Version, m.Name)
		if err := m.Up(); err != nil {
			_ = restore(snap)
			return fmt.Errorf("node migration %d (%s) failed; rolled back: %w", m.Version, m.Name, err)
		}
		st.Node = m.Version
		if err := save(st); err != nil {
			_ = restore(snap)
			return fmt.Errorf("recording migration %d failed; rolled back: %w", m.Version, err)
		}
	}
	return nil
}

// Load reads the recorded schema state (zero-value if absent).
func Load() State {
	var s State
	if b, err := os.ReadFile(paths.SchemaFile()); err == nil {
		_ = json.Unmarshal(b, &s)
	}
	return s
}

func save(s State) error {
	if err := os.MkdirAll(filepath.Dir(paths.SchemaFile()), 0o700); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(s, "", "  ")
	return os.WriteFile(paths.SchemaFile(), append(b, '\n'), 0o600)
}

// snapshotFiles are the state files captured before migrating.
func snapshotFiles() []string {
	return []string{
		paths.ConfigFile(),
		paths.AppsStateFile(),
		paths.RegistryFile(),
		paths.AppSecretsFile(),
		paths.SchemaFile(),
	}
}

// Snapshot copies the current state files into a timestamped backup dir and
// returns its path. Exposed so per-app migrations can snapshot too.
func Snapshot() (string, error) {
	dir := filepath.Join(paths.BackupsDir(), time.Now().UTC().Format("20060102-150405"))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	for _, f := range snapshotFiles() {
		b, err := os.ReadFile(f)
		if err != nil {
			continue // absent file → nothing to back up
		}
		if err := os.WriteFile(filepath.Join(dir, filepath.Base(f)), b, 0o600); err != nil {
			return "", err
		}
	}
	pruneBackups(10)
	return dir, nil
}

// Restore copies a snapshot's state files back into place (rollback).
func Restore(dir string) error { return restore(dir) }

func restore(dir string) error {
	for _, f := range snapshotFiles() {
		b, err := os.ReadFile(filepath.Join(dir, filepath.Base(f)))
		if err != nil {
			continue
		}
		if err := os.WriteFile(f, b, 0o600); err != nil {
			return err
		}
	}
	return nil
}

// pruneBackups keeps only the most recent keep backup directories.
func pruneBackups(keep int) {
	entries, err := os.ReadDir(paths.BackupsDir())
	if err != nil || len(entries) <= keep {
		return
	}
	for _, e := range entries[:len(entries)-keep] { // ReadDir is name-sorted ⇒ oldest first
		_ = os.RemoveAll(filepath.Join(paths.BackupsDir(), e.Name()))
	}
}
