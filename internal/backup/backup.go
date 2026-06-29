package backup

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/slashbinslashnoname/slashnode/internal/apps"
	"github.com/slashbinslashnoname/slashnode/internal/orchestrator"
	"github.com/slashbinslashnoname/slashnode/internal/paths"
)

// manifestFormat is bumped if the on-destination layout changes.
const manifestFormat = 1

// chainApps own large, re-syncable Docker volumes that are skipped by default
// (their data is re-downloadable). lnd is deliberately NOT here — its channel
// state is unrecoverable.
var chainApps = []string{"bitcoind", "dogecoind", "monerod", "geth", "electrs", "electrumx"}

// Manifest is written to <dest>/manifest.json each run.
type Manifest struct {
	Format    int      `json:"format"`
	CreatedAt string   `json:"created_at"`
	Version   string   `json:"slashnode_version"`
	Apps      []string `json:"apps"`
	Volumes   []string `json:"volumes"`
	Excluded  []string `json:"excluded"`
}

// Options controls a backup run.
type Options struct {
	All     bool   // include chain volumes
	Version string // slashnode version, recorded in the manifest
}

// Run backs up the node to the configured destination. Sources are mounted
// read-only and rclone-synced; nothing large is staged on the local disk.
func Run(opts Options, out io.Writer) error {
	if !orchestrator.Available() {
		return fmt.Errorf("docker is not available")
	}
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	d := cfg.Destination
	if !d.Configured() {
		return fmt.Errorf("no backup destination configured — set one first")
	}
	if err := d.WriteRcloneConf(); err != nil {
		return err
	}

	inc, exc, err := volumesToBackup(opts.All)
	if err != nil {
		return err
	}

	// state (everything under DataDir except backups/, rclone/, backup.json)
	fmt.Fprintf(out, "--> backing up state\n")
	if err := sync(d, out, paths.DataDir(), "state",
		"--exclude", "/backups/**", "--exclude", "/rclone/**", "--exclude", "/backup.json"); err != nil {
		return fmt.Errorf("state: %w", err)
	}

	// Tor onion keys (best-effort: needs root to read tor-owned keys)
	if _, err := os.Stat(paths.TorDataDir()); err == nil {
		fmt.Fprintf(out, "--> backing up Tor keys\n")
		if e := sync(d, out, paths.TorDataDir(), "tor"); e != nil {
			fmt.Fprintf(out, "    (skipped Tor keys: %v)\n", e)
		}
	}

	// app volumes
	for _, v := range inc {
		fmt.Fprintf(out, "--> backing up volume %s\n", v)
		if err := sync(d, out, v, "volumes/"+v); err != nil {
			return fmt.Errorf("volume %s: %w", v, err)
		}
	}
	for _, v := range exc {
		fmt.Fprintf(out, "    skipped chain volume %s (use --all to include)\n", v)
	}

	if err := writeManifest(d, out, Manifest{
		Format: manifestFormat, CreatedAt: now(), Version: opts.Version,
		Apps: installedIDs(), Volumes: inc, Excluded: exc,
	}); err != nil {
		return err
	}

	cfg.LastRun = now()
	cfg.LastResult = "ok"
	_ = SaveConfig(cfg)
	fmt.Fprintf(out, "--> backup complete (%d volume(s), %d skipped)\n", len(inc), len(exc))
	return nil
}

// Restore pulls the backup back onto this node: state, Tor keys and every
// volume, then brings the apps up and reconciles the proxy/Tor. Destructive —
// run with the daemon stopped.
func Restore(out io.Writer) error {
	if !orchestrator.Available() {
		return fmt.Errorf("docker is not available")
	}
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	d := cfg.Destination
	if !d.Configured() {
		return fmt.Errorf("no backup destination configured")
	}
	if err := d.WriteRcloneConf(); err != nil {
		return err
	}

	man, err := readManifest(d)
	if err != nil {
		return fmt.Errorf("read manifest (is there a backup at this destination?): %w", err)
	}
	fmt.Fprintf(out, "--> restoring backup from %s (%d volumes)\n", man.CreatedAt, len(man.Volumes))

	// state → DataDir (copy = additive, never deletes the live rclone creds)
	fmt.Fprintf(out, "--> restoring state\n")
	if err := os.MkdirAll(paths.DataDir(), 0o700); err != nil {
		return err
	}
	if err := copyDown(d, out, "state", paths.DataDir()); err != nil {
		return fmt.Errorf("state: %w", err)
	}

	// Tor keys → TorDataDir (best-effort + chown to the tor user)
	if err := os.MkdirAll(paths.TorDataDir(), 0o700); err == nil {
		fmt.Fprintf(out, "--> restoring Tor keys\n")
		if e := copyDown(d, out, "tor", paths.TorDataDir()); e != nil {
			fmt.Fprintf(out, "    (skipped Tor keys: %v)\n", e)
		} else {
			chownTor()
		}
	}

	// volumes: recreate then populate
	for _, v := range man.Volumes {
		fmt.Fprintf(out, "--> restoring volume %s\n", v)
		_ = exec.Command("docker", "volume", "create", v).Run()
		if err := copyDownVolume(d, out, "volumes/"+v, v); err != nil {
			return fmt.Errorf("volume %s: %w", v, err)
		}
	}

	// bring everything back up + reconcile (same path serve uses on startup)
	fmt.Fprintf(out, "--> starting apps\n")
	appsDir := paths.AppsDir()
	if err := apps.Reapply(appsDir); err != nil {
		fmt.Fprintf(out, "    (reapply reported: %v)\n", err)
	}
	_ = apps.ReloadProxy()
	_ = apps.ReloadTor(appsDir)
	fmt.Fprintf(out, "--> restore complete\n")
	return nil
}

// Test verifies the destination is reachable with the current credentials.
func Test(out io.Writer) error {
	if !orchestrator.Available() {
		return fmt.Errorf("docker is not available")
	}
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	d := cfg.Destination
	if !d.Configured() {
		return fmt.Errorf("no backup destination configured")
	}
	if err := d.WriteRcloneConf(); err != nil {
		return err
	}
	mounts, err := d.extraMounts(false)
	if err != nil {
		return err
	}
	// lsd lists the destination root; succeeds on valid creds/reachability.
	return dockerRclone(out, mounts, "lsd", d.targetRoot(), "--low-level-retries=1")
}

// --- rclone runners -------------------------------------------------------

// sync mirrors a source (a Docker volume name or a host dir) into <dest>/<name>.
func sync(d Destination, out io.Writer, src, name string, extra ...string) error {
	mounts := []string{"-v", src + ":/src:ro"}
	dm, err := d.extraMounts(true)
	if err != nil {
		return err
	}
	mounts = append(mounts, dm...)
	args := append([]string{"sync", "/src", d.targetRoot() + "/" + name}, statsArgs...)
	args = append(args, extra...)
	return dockerRclone(out, mounts, args...)
}

// copyDown copies <name> from the destination into a host directory.
func copyDown(d Destination, out io.Writer, name, dstHostDir string) error {
	mounts := []string{"-v", dstHostDir + ":/dst"}
	dm, err := d.extraMounts(false)
	if err != nil {
		return err
	}
	mounts = append(mounts, dm...)
	args := append([]string{"copy", d.targetRoot() + "/" + name, "/dst"}, statsArgs...)
	return dockerRclone(out, mounts, args...)
}

// copyDownVolume copies <name> from the destination into a Docker volume.
func copyDownVolume(d Destination, out io.Writer, name, vol string) error {
	mounts := []string{"-v", vol + ":/dst"}
	dm, err := d.extraMounts(false)
	if err != nil {
		return err
	}
	mounts = append(mounts, dm...)
	args := append([]string{"copy", d.targetRoot() + "/" + name, "/dst"}, statsArgs...)
	return dockerRclone(out, mounts, args...)
}

var statsArgs = []string{"--transfers=4", "--stats=2s", "--stats-one-line", "--stats-log-level", "NOTICE"}

func dockerRclone(out io.Writer, mounts []string, rcloneArgs ...string) error {
	args := append([]string{"run", "--rm"}, mounts...)
	args = append(args, rcloneImage)
	args = append(args, rcloneArgs...)
	return streamCmd(out, "docker", args...)
}

// --- manifest -------------------------------------------------------------

func writeManifest(d Destination, out io.Writer, m Manifest) error {
	tmp, err := os.MkdirTemp("", "sn-backup-manifest")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	b, _ := json.MarshalIndent(m, "", "  ")
	if err := os.WriteFile(filepath.Join(tmp, "manifest.json"), b, 0o600); err != nil {
		return err
	}
	mounts := []string{"-v", tmp + ":/src:ro"}
	dm, err := d.extraMounts(true)
	if err != nil {
		return err
	}
	mounts = append(mounts, dm...)
	return dockerRclone(out, mounts, "copyto", "/src/manifest.json", d.targetRoot()+"/manifest.json")
}

func readManifest(d Destination) (Manifest, error) {
	var m Manifest
	mounts, err := d.extraMounts(false)
	if err != nil {
		return m, err
	}
	args := append([]string{"run", "--rm"}, mounts...)
	args = append(args, rcloneImage, "cat", d.targetRoot()+"/manifest.json")
	raw, err := captureCmd("docker", args...)
	if err != nil {
		return m, err
	}
	return m, json.Unmarshal([]byte(raw), &m)
}

// --- helpers --------------------------------------------------------------

func volumesToBackup(all bool) (include, exclude []string, err error) {
	raw, err := captureCmd("docker", "volume", "ls", "-q", "--filter", "name=slashnode-")
	if err != nil {
		return nil, nil, err
	}
	for _, v := range strings.Fields(raw) {
		if !all && isChainVolume(v) {
			exclude = append(exclude, v)
			continue
		}
		include = append(include, v)
	}
	return include, exclude, nil
}

func isChainVolume(v string) bool {
	for _, id := range chainApps {
		if strings.HasPrefix(v, "slashnode-"+id+"_") {
			return true
		}
	}
	return false
}

func installedIDs() []string {
	var ids []string
	for id := range apps.LoadState().Installed {
		ids = append(ids, id)
	}
	return ids
}

// obscure runs `rclone obscure` (used for sftp/webdav passwords in the conf).
func obscure(s string) (string, error) {
	o, err := captureCmd("docker", "run", "--rm", rcloneImage, "obscure", s)
	return strings.TrimSpace(o), err
}

// chownTor best-effort restores tor-user ownership of the onion key dir so the
// tor daemon can read it. Ignored when unprivileged or the user is absent.
func chownTor() {
	for _, u := range []string{"debian-tor:debian-tor", "tor:tor"} {
		if exec.Command("chown", "-R", u, paths.TorDataDir()).Run() == nil {
			return
		}
	}
}

func streamCmd(out io.Writer, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = out
	cmd.Stderr = out
	return cmd.Run()
}

func captureCmd(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var sb strings.Builder
	cmd.Stdout = &sb
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return sb.String(), nil
}

func now() string { return time.Now().UTC().Format(time.RFC3339) }
