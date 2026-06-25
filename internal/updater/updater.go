// Package updater gère la vérification et l'application des mises à jour du
// binaire slashnoded.
//
// Politique par défaut : "notify". Le timer systemd appelle Check() qui écrit
// l'état dans update.json ; l'UI affiche une bannière et l'opérateur déclenche
// Apply() via le bouton « Appliquer » (ou `slashnoded update`).
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

// Info décrit l'état des mises à jour (sérialisé dans update.json et exposé à
// l'UI via l'API Go).
type Info struct {
	Current   string `json:"current"`
	Latest    string `json:"latest"`
	Available bool   `json:"available"`
	CheckedAt string `json:"checked_at"`
}

var httpClient = &http.Client{Timeout: 30 * time.Second}

// Check interroge la dernière version, calcule la disponibilité et persiste
// l'état dans update.json.
func Check(current, channel string) (*Info, error) {
	latest, err := latestVersion(channel)
	if err != nil {
		return nil, err
	}
	info := &Info{
		Current:   normalize(current),
		Latest:    normalize(latest),
		Available: latest != "" && normalize(latest) != normalize(current),
		CheckedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if err := save(info); err != nil {
		return nil, err
	}
	return info, nil
}

// LoadState lit le dernier état connu. Renvoie un Info "non disponible" si le
// fichier n'existe pas encore.
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

// Apply télécharge la version cible (ou la dernière), vérifie le checksum,
// remplace le binaire courant de façon atomique puis redémarre le service.
func Apply(target, channel string) error {
	if target == "" || target == "latest" {
		v, err := latestVersion(channel)
		if err != nil {
			return err
		}
		target = v
	}

	osName, arch := runtime.GOOS, runtime.GOARCH
	base := fmt.Sprintf("https://github.com/%s/releases/download/%s/slashnoded-%s-%s",
		repo, target, osName, arch)

	bin, err := download(base)
	if err != nil {
		return fmt.Errorf("téléchargement binaire : %w", err)
	}
	sumFile, err := download(base + ".sha256")
	if err != nil {
		return fmt.Errorf("téléchargement checksum : %w", err)
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

	// Redémarrage best-effort (systemd relancera avec le nouveau binaire).
	restart()
	return nil
}

// latestVersion renvoie la dernière version publiée pour le canal.
// SLASHNODE_LATEST court-circuite le réseau (hook de test).
func latestVersion(channel string) (string, error) {
	if v := os.Getenv("SLASHNODE_LATEST"); v != "" {
		return v, nil
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	if channel == "beta" {
		url = fmt.Sprintf("https://api.github.com/repos/%s/releases", repo)
	}
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API GitHub : statut %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if channel == "beta" {
		var rels []struct {
			TagName string `json:"tag_name"`
		}
		if err := json.Unmarshal(body, &rels); err != nil {
			return "", err
		}
		if len(rels) == 0 {
			return "", nil
		}
		return rels[0].TagName, nil
	}
	var rel struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal(body, &rel); err != nil {
		return "", err
	}
	return rel.TagName, nil
}

func download(url string) ([]byte, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("statut %d pour %s", resp.StatusCode, url)
	}
	return io.ReadAll(resp.Body)
}

// verifySHA256 vérifie que sha256(bin) correspond à la somme du fichier
// .sha256 (format `<hex>  <nom>`).
func verifySHA256(bin, sumFile []byte) error {
	want := strings.Fields(string(sumFile))
	if len(want) == 0 {
		return fmt.Errorf("fichier checksum vide")
	}
	sum := sha256.Sum256(bin)
	got := hex.EncodeToString(sum[:])
	if !strings.EqualFold(got, want[0]) {
		return fmt.Errorf("checksum invalide : attendu %s, obtenu %s", want[0], got)
	}
	return nil
}

// replaceBinary remplace self par newBin de façon atomique (write+rename dans le
// même répertoire).
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

// restart relance le service via systemd (best-effort). En --root de test ou
// hors Linux, c'est un no-op.
func restart() {
	if runtime.GOOS != "linux" || os.Getenv("SLASHNODE_ROOT") != "" {
		return
	}
	if path, err := exec.LookPath("systemctl"); err == nil {
		_ = exec.Command(path, "restart", "slashnoded").Start()
	}
}
