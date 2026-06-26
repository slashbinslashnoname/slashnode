// Package systemd generates the systemd unit of slashnoded.
package systemd

import (
	"fmt"
	"os"
	"path/filepath"
)

// UnitContent renders the content of the systemd unit. binPath is the absolute
// path of the binary (ExecStart).
func UnitContent(binPath string) string {
	return fmt.Sprintf(`[Unit]
Description=SlashNode daemon
Documentation=https://github.com/slashbinslashnoname/slashnode
After=network-online.target docker.service
Wants=network-online.target

[Service]
Type=simple
ExecStart=%s serve
Restart=on-failure
RestartSec=5
User=root
NoNewPrivileges=true
AmbientCapabilities=CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
`, binPath)
}

// WriteUnit writes the unit to path (mode 0644).
func WriteUnit(path, binPath string) error {
	return write(path, UnitContent(binPath))
}

// UpdateServiceContent renders the oneshot service that checks for updates
// (notify policy: it signals, it does not apply).
func UpdateServiceContent(binPath string) string {
	return fmt.Sprintf(`[Unit]
Description=SlashNode — update check
After=network-online.target
Wants=network-online.target

[Service]
Type=oneshot
ExecStart=%s check-update
`, binPath)
}

// UpdateTimerContent renders the daily timer (with a random delay to smooth out
// the load on the release infrastructure).
func UpdateTimerContent() string {
	return `[Unit]
Description=SlashNode — daily update check

[Timer]
OnCalendar=daily
RandomizedDelaySec=2h
Persistent=true

[Install]
WantedBy=timers.target
`
}

// WriteUpdateUnits writes the update-check service and timer.
func WriteUpdateUnits(servicePath, timerPath, binPath string) error {
	if err := write(servicePath, UpdateServiceContent(binPath)); err != nil {
		return err
	}
	return write(timerPath, UpdateTimerContent())
}

// PruneServiceContent renders the oneshot service that reclaims disk by removing
// dangling docker images (stopped containers and volumes are preserved).
func PruneServiceContent(binPath string) string {
	return fmt.Sprintf(`[Unit]
Description=SlashNode — docker image prune
After=docker.service
Wants=docker.service

[Service]
Type=oneshot
ExecStart=%s prune
`, binPath)
}

// PruneTimerContent renders the daily prune timer.
func PruneTimerContent() string {
	return `[Unit]
Description=SlashNode — daily docker image prune

[Timer]
OnCalendar=daily
RandomizedDelaySec=1h
Persistent=true

[Install]
WantedBy=timers.target
`
}

// WritePruneUnits writes the prune service and daily timer.
func WritePruneUnits(servicePath, timerPath, binPath string) error {
	if err := write(servicePath, PruneServiceContent(binPath)); err != nil {
		return err
	}
	return write(timerPath, PruneTimerContent())
}

func write(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}
