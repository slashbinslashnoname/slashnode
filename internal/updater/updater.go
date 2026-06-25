// Package updater manages the checking and application of updates to the
// slashnoded binary.
//
// Default policy: "notify". The systemd timer calls Check() which writes the
// state to update.json; the UI displays a banner and the operator triggers
// Apply() via the "Apply" button (or `slashnoded update`).
package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/slashbinslashnoname/slashnode/internal/paths"
)

const repo = "slashbinslashnoname/slashnode"

// Info describes the update state (serialized into update.json and exposed to
// the UI via the Go API).
type Info struct {
	Current   string `json:"current"`
	Latest    string `json:"latest"`
	Available bool   `json:"available"`
	CheckedAt string `json:"checked_at"`
}

var httpClient = &http.Client{Timeout: 30 * time.Second}

// Check queries the latest version, computes availability and persists the
// state to update.json.
func Check(current, channel string) (*Info, error) {
	_, version, err := latestRelease(channel)
	if err != nil {
		return nil, err
	}
	info := &Info{
		Current:   normalize(current),
		Latest:    normalize(version),
		Available: version != "" && normalize(version) != normalize(current),
		CheckedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if err := save(info); err != nil {
		return nil, err
	}
	return info, nil
}

// LoadState reads the last known state. Returns a "not available" Info if the
// file does not exist yet.
func LoadState(current string) *Info {
	b, err := os.ReadFile(paths.UpdateStateFile())
	if err != nil {
		return &Info{Current: normalize(current), Latest: normalize(current), Available: false}
	}
	var info Info
	if json.Unmarshal(b, &info) != nil {
		return &Info{Current: normalize(current), Latest: normalize(current), Available: false}
	}
	return &info
}

// Apply downloads the target version (or the latest), verifies the checksum,
// atomically replaces the current binary then restarts the service.
func Apply(target, channel string) error {
	// downloadTag is the release tag to fetch assets from; for the rolling
	// release this is "latest" even though the displayed version differs.
	downloadTag := target
	if target == "" || target == "latest" {
		tag, _, err := latestRelease(channel)
		if err != nil {
			return err
		}
		downloadTag = tag
	}

	osName, arch := runtime.GOOS, runtime.GOARCH
	base := fmt.Sprintf("https://github.com/%s/releases/download/%s/slashnoded-%s-%s",
		repo, downloadTag, osName, arch)

	bin, err := download(base)
	if err != nil {
		return fmt.Errorf("binary download: %w", err)
	}
	sumFile, err := download(base + ".sha256")
	if err != nil {
		return fmt.Errorf("checksum download: %w", err)
	}
	if err := verifySHA256(bin, sumFile); err != nil {
		return err
	}

	self, err := os.Executable()
	if err != nil {
		return err
	}
	if err := replaceBinary(self, bin); err != nil {
		return err
	}

	// Best-effort restart (systemd will relaunch with the new binary).
	restart()
	return nil
}

// latestRelease returns the latest release's download tag and display version.
// For the rolling release the tag is "latest" while the version is the release
// name (e.g. 2026.06.25-ab12cd). SLASHNODE_LATEST short-circuits the network.
func latestRelease(channel string) (tag, version string, err error) {
	if v := os.Getenv("SLASHNODE_LATEST"); v != "" {
		return v, v, nil
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	if channel == "beta" {
		url = fmt.Sprintf("https://api.github.com/repos/%s/releases", repo)
	}
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("GitHub API: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	type release struct {
		TagName string `json:"tag_name"`
		Name    string `json:"name"`
	}
	pick := func(r release) (string, string) {
		v := r.Name
		if v == "" {
			v = r.TagName
		}
		return r.TagName, v
	}
	if channel == "beta" {
		var rels []release
		if err := json.Unmarshal(body, &rels); err != nil {
			return "", "", err
		}
		if len(rels) == 0 {
			return "", "", nil
		}
		t, v := pick(rels[0])
		return t, v, nil
	}
	var rel release
	if err := json.Unmarshal(body, &rel); err != nil {
		return "", "", err
	}
	t, v := pick(rel)
	return t, v, nil
}

func download(url string) ([]byte, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d for %s", resp.StatusCode, url)
	}
	return io.ReadAll(resp.Body)
}

// verifySHA256 verifies that sha256(bin) matches the sum from the .sha256 file
// (format `<hex>  <name>`).
func verifySHA256(bin, sumFile []byte) error {
	want := strings.Fields(string(sumFile))
	if len(want) == 0 {
		return fmt.Errorf("empty checksum file")
	}
	sum := sha256.Sum256(bin)
	got := hex.EncodeToString(sum[:])
	if !strings.EqualFold(got, want[0]) {
		return fmt.Errorf("invalid checksum: expected %s, got %s", want[0], got)
	}
	return nil
}

// replaceBinary atomically replaces self with newBin (write+rename in the same
// directory).
func replaceBinary(self string, newBin []byte) error {
	dir := filepath.Dir(self)
	tmp, err := os.CreateTemp(dir, ".slashnoded-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(newBin); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, 0o755); err != nil {
		return err
	}
	return os.Rename(tmpName, self)
}

func save(info *Info) error {
	if err := os.MkdirAll(filepath.Dir(paths.UpdateStateFile()), 0o700); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(info, "", "  ")
	return os.WriteFile(paths.UpdateStateFile(), append(b, '\n'), 0o644)
}

func normalize(v string) string { return strings.TrimPrefix(v, "v") }

// restart relaunches the service via systemd (best-effort). In test --root mode
// or outside Linux, it is a no-op.
func restart() {
	if runtime.GOOS != "linux" || os.Getenv("SLASHNODE_ROOT") != "" {
		return
	}
	if path, err := exec.LookPath("systemctl"); err == nil {
		_ = exec.Command(path, "restart", "slashnoded").Start()
	}
}
