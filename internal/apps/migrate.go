package apps

import (
	"fmt"
	"io"
	"sort"

	"github.com/slashbinslashnoname/slashnode/internal/migrate"
	"github.com/slashbinslashnoname/slashnode/internal/orchestrator"
	"github.com/slashbinslashnoname/slashnode/internal/paths"
)

// AppMigration transforms an installed app's stored state (and/or its container
// data) from version-1 to Version, so a breaking app upgrade stays compatible
// with what an older version wrote. Steps run in array order; migrations run in
// ascending Version order. Append only — never reorder or renumber.
type AppMigration struct {
	Version     int             `json:"version"`
	Description string          `json:"description,omitempty"`
	Steps       []MigrationStep `json:"steps"`
}

// MigrationStep is one declarative operation. Exactly one field should be set.
type MigrationStep struct {
	RenameInput  *renameOp `json:"renameInput,omitempty"`  // move a stored (non-secret) input value
	RenameSecret *renameOp `json:"renameSecret,omitempty"` // move a stored secret value
	RemoveInput  string    `json:"removeInput,omitempty"`  // drop a stored input/secret
	SetInput     *kvOp     `json:"setInput,omitempty"`     // set a stored input value
	Exec         *execOp   `json:"exec,omitempty"`         // run a command in a running service
	CopyVolume   *renameOp `json:"copyVolume,omitempty"`   // copy one docker volume into another
}

type renameOp struct {
	From string `json:"from"`
	To   string `json:"to"`
}
type kvOp struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
type execOp struct {
	Service string `json:"service"`
	Command string `json:"command"`
}

// MigrationsPending returns a summary of installed apps that have pending
// per-app migrations (for `slashnoded migrate --dry-run`).
func MigrationsPending(dir string) []string {
	var out []string
	for id, inst := range LoadState().Installed {
		man, err := Find(dir, id)
		if err != nil {
			continue
		}
		if p := pendingAppMigrations(man, inst.MigrationVersion); len(p) > 0 {
			out = append(out, fmt.Sprintf("%s: %d pending", id, len(p)))
		}
	}
	sort.Strings(out)
	return out
}

func appMigrationLatest(man *Manifest) int {
	max := 0
	for _, m := range man.Migrations {
		if m.Version > max {
			max = m.Version
		}
	}
	return max
}

func pendingAppMigrations(man *Manifest, applied int) []AppMigration {
	var out []AppMigration
	for _, m := range man.Migrations {
		if m.Version > applied {
			out = append(out, m)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Version < out[j].Version })
	return out
}

// runAppMigrations applies the installed app's pending per-app migrations (in
// order) BEFORE it is re-rendered/recreated. It snapshots the node state files
// first and restores them if a step fails (volume/exec side effects can't be
// rolled back — those are logged). A no-op for a not-yet-installed app or when
// nothing is pending.
func runAppMigrations(dir, id string, out io.Writer) error {
	man, err := Find(dir, id)
	if err != nil {
		return err
	}
	inst, ok := LoadState().Installed[id]
	if !ok {
		return nil
	}
	pending := pendingAppMigrations(man, inst.MigrationVersion)
	if len(pending) == 0 {
		return nil
	}
	snap, _ := migrate.Snapshot()
	for _, m := range pending {
		fmt.Fprintf(out, "  • %s migration → v%d (%s)\n", id, m.Version, m.Description)
		for _, step := range m.Steps {
			if err := applyStep(id, step, out); err != nil {
				if snap != "" {
					_ = migrate.Restore(snap)
				}
				return fmt.Errorf("%s migration v%d failed (state rolled back): %w", id, m.Version, err)
			}
		}
		// Record progress after each migration so a later failure doesn't re-run it.
		state := LoadState()
		cur := state.Installed[id]
		cur.MigrationVersion = m.Version
		state.Installed[id] = cur
		if err := saveState(state); err != nil {
			return err
		}
	}
	return nil
}

func applyStep(id string, s MigrationStep, out io.Writer) error {
	switch {
	case s.RenameInput != nil:
		st := LoadState()
		inst := st.Installed[id]
		if v, ok := inst.Inputs[s.RenameInput.From]; ok {
			if inst.Inputs == nil {
				inst.Inputs = map[string]string{}
			}
			inst.Inputs[s.RenameInput.To] = v
			delete(inst.Inputs, s.RenameInput.From)
			st.Installed[id] = inst
			return saveState(st)
		}
		return nil
	case s.SetInput != nil:
		st := LoadState()
		inst := st.Installed[id]
		if inst.Inputs == nil {
			inst.Inputs = map[string]string{}
		}
		inst.Inputs[s.SetInput.Key] = s.SetInput.Value
		st.Installed[id] = inst
		return saveState(st)
	case s.RenameSecret != nil:
		secs := loadAppSecrets(id)
		if v, ok := secs[s.RenameSecret.From]; ok {
			secs[s.RenameSecret.To] = v
			delete(secs, s.RenameSecret.From)
			return mergeAppSecrets(id, secs)
		}
		return nil
	case s.RemoveInput != "":
		st := LoadState()
		inst := st.Installed[id]
		delete(inst.Inputs, s.RemoveInput)
		st.Installed[id] = inst
		if err := saveState(st); err != nil {
			return err
		}
		secs := loadAppSecrets(id)
		if _, ok := secs[s.RemoveInput]; ok {
			delete(secs, s.RemoveInput)
			return mergeAppSecrets(id, secs)
		}
		return nil
	case s.CopyVolume != nil:
		if !orchestrator.Available() {
			return nil
		}
		return orchestrator.CopyVolume(s.CopyVolume.From, s.CopyVolume.To)
	case s.Exec != nil:
		if !orchestrator.Available() {
			return nil
		}
		return orchestrator.ExecStreamed(id, paths.AppComposeFile(id), s.Exec.Service, s.Exec.Command, out)
	}
	return nil
}
