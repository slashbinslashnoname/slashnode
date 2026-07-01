// Package backup syncs a SlashNode's data (state, secrets, Tor onion keys and
// app Docker volumes) to a configurable destination using rclone, incrementally
// and without staging a local archive. The rclone binary itself is never
// installed on the host: every transfer runs the official rclone/rclone Docker
// image, mounting the source (a Docker volume or a host dir) read-only plus the
// generated rclone config.
package backup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/slashbinslashnoname/slashnode/internal/paths"
)

// rcloneImage is the engine; pinned-ish to a moving tag (re-pulled on demand).
const rcloneImage = "rclone/rclone:latest"

// Destination describes where backups go. One of the supported kinds:
//   - s3:    any S3-compatible store (AWS, MinIO, Backblaze B2 via S3, R2, Wasabi…)
//   - sftp:  rsync-over-SSH-style incremental sync to a remote host
//   - local: a mounted path on the node (USB disk, NFS mount…)
//   - node:  another SlashNode reached over Tailscale — off-site, self-hosted
//     backup onto a peer you own (transported over SSH/SFTP, encrypted
//     end-to-end by WireGuard). Same wire format as sftp; Host holds the
//     peer's 100.x tailnet address (or MagicDNS name).
type Destination struct {
	Kind   string `json:"kind"`   // s3 | sftp | local | node
	Prefix string `json:"prefix"` // sub-path within the destination (e.g. "slashnode")

	// s3
	Provider  string `json:"provider,omitempty"` // rclone S3 provider (AWS, Minio, Other…)
	Endpoint  string `json:"endpoint,omitempty"`
	Region    string `json:"region,omitempty"`
	Bucket    string `json:"bucket,omitempty"`
	AccessKey string `json:"access_key,omitempty"`
	SecretKey string `json:"secret_key,omitempty"`

	// sftp / node (node reuses these; Host is the peer's tailnet address)
	Host string `json:"host,omitempty"`
	Port int    `json:"port,omitempty"`
	User string `json:"user,omitempty"`
	Pass string `json:"pass,omitempty"` // obscured into rclone.conf at write time
}

// Config is persisted at paths.BackupConfigFile() (non-secret bits) — the actual
// credentials are rendered into paths.RcloneConfigFile() (mode 0600).
type Config struct {
	Destination Destination `json:"destination"`
	All         bool        `json:"all"` // include large re-syncable chain volumes
	LastRun     string      `json:"last_run,omitempty"`
	LastResult  string      `json:"last_result,omitempty"`
}

// LoadConfig reads the backup config; a missing file yields a zero Config.
func LoadConfig() (Config, error) {
	var c Config
	b, err := os.ReadFile(paths.BackupConfigFile())
	if err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}
		return c, err
	}
	return c, json.Unmarshal(b, &c)
}

// SaveConfig persists the non-secret config. Credentials also live here (0600)
// so a restored node keeps its destination; the file is mode 0600.
func SaveConfig(c Config) error {
	if err := os.MkdirAll(paths.DataDir(), 0o700); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(c, "", "  ")
	return os.WriteFile(paths.BackupConfigFile(), append(b, '\n'), 0o600)
}

func (d Destination) Configured() bool {
	switch d.Kind {
	case "s3":
		return d.Bucket != "" && d.AccessKey != "" && d.SecretKey != ""
	case "sftp", "node":
		return d.Host != "" && d.User != ""
	case "local":
		return d.Prefix != ""
	}
	return false
}

// Validate rejects destinations whose free-text fields could inject extra
// directives into the rendered rclone.conf. Host/User for the SSH transports
// (sftp, node) end up on their own `key = value` lines, so a newline or control
// character in them must never be accepted. Empty (unconfigured) is allowed.
func (d Destination) Validate() error {
	if d.Kind == "sftp" || d.Kind == "node" {
		if strings.ContainsAny(d.Host, " \t\r\n\"") || hasControl(d.Host) {
			return fmt.Errorf("invalid host")
		}
		if strings.ContainsAny(d.User, " \t\r\n\"") || hasControl(d.User) {
			return fmt.Errorf("invalid user")
		}
	}
	return nil
}

func hasControl(s string) bool {
	for _, r := range s {
		if r < 0x20 || r == 0x7f {
			return true
		}
	}
	return false
}

// prefix returns the destination sub-path, defaulting to "slashnode".
func (d Destination) prefixOr() string {
	if d.Prefix == "" {
		return "slashnode"
	}
	return strings.Trim(d.Prefix, "/")
}

// targetRoot is the rclone target the sources are written under, e.g.
// "backup:my-bucket/slashnode" (remote) or "/dest" (a mounted local path).
func (d Destination) targetRoot() string {
	switch d.Kind {
	case "s3":
		return "backup:" + d.Bucket + "/" + d.prefixOr()
	case "sftp", "node":
		return "backup:" + d.prefixOr()
	case "local":
		return "/dest" // the host path is bind-mounted at /dest in the container
	}
	return ""
}

// extraMounts are docker -v args specific to the destination. Remote kinds mount
// the rclone config; local mounts the destination directory itself.
func (d Destination) extraMounts(rw bool) ([]string, error) {
	if d.Kind == "local" {
		host := filepath.Clean(d.Prefix)
		if err := os.MkdirAll(host, 0o700); err != nil {
			return nil, fmt.Errorf("create local destination %s: %w", host, err)
		}
		mode := ":ro"
		if rw {
			mode = ""
		}
		return []string{"-v", host + ":/dest" + mode}, nil
	}
	// remote: mount the generated rclone.conf read-only
	return []string{"-v", paths.RcloneConfigDir() + ":/config/rclone:ro"}, nil
}

// writeRcloneConf renders the [backup] remote section to RcloneConfigFile (0600).
// Local destinations need no config (they reference a bind-mounted path), so the
// file is written empty. sftp passwords are obscured via the rclone container.
func (d Destination) WriteRcloneConf() error {
	if err := os.MkdirAll(paths.RcloneConfigDir(), 0o700); err != nil {
		return err
	}
	var b strings.Builder
	switch d.Kind {
	case "s3":
		provider := d.Provider
		if provider == "" {
			provider = "Other"
		}
		fmt.Fprintf(&b, "[backup]\ntype = s3\nprovider = %s\naccess_key_id = %s\nsecret_access_key = %s\n",
			provider, d.AccessKey, d.SecretKey)
		if d.Endpoint != "" {
			fmt.Fprintf(&b, "endpoint = %s\n", d.Endpoint)
		}
		if d.Region != "" {
			fmt.Fprintf(&b, "region = %s\n", d.Region)
		}
	case "sftp", "node":
		port := d.Port
		if port == 0 {
			port = 22
		}
		obscured := ""
		if d.Pass != "" {
			o, err := obscure(d.Pass)
			if err != nil {
				return err
			}
			obscured = o
		}
		fmt.Fprintf(&b, "[backup]\ntype = sftp\nhost = %s\nport = %d\nuser = %s\n",
			d.Host, port, d.User)
		if obscured != "" {
			fmt.Fprintf(&b, "pass = %s\n", obscured)
		}
	case "local":
		// no remote section needed
	}
	return os.WriteFile(paths.RcloneConfigFile(), []byte(b.String()), 0o600)
}
