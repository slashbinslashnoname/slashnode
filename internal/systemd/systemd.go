// Package systemd génère l'unit systemd de slashnoded.
package systemd

import (
	"fmt"
	"os"
	"path/filepath"
)

// UnitContent rend le contenu de l'unit systemd. binPath est le chemin absolu
// du binaire (ExecStart).
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

// WriteUnit écrit l'unit à path (mode 0644).
func WriteUnit(path, binPath string) error {
	return write(path, UnitContent(binPath))
}

// UpdateServiceContent rend le service oneshot qui vérifie les mises à jour
// (politique notify : il signale, il n'applique pas).
func UpdateServiceContent(binPath string) string {
	return fmt.Sprintf(`[Unit]
Description=SlashNode — vérification des mises à jour
After=network-online.target
Wants=network-online.target

[Service]
Type=oneshot
ExecStart=%s check-update
`, binPath)
}

// UpdateTimerContent rend le timer quotidien (avec un délai aléatoire pour
// lisser la charge sur l'infra de release).
func UpdateTimerContent() string {
	return `[Unit]
Description=SlashNode — vérification quotidienne des mises à jour

[Timer]
OnCalendar=daily
RandomizedDelaySec=2h
Persistent=true

[Install]
WantedBy=timers.target
`
}

// WriteUpdateUnits écrit le service et le timer de vérification des MAJ.
func WriteUpdateUnits(servicePath, timerPath, binPath string) error {
	if err := write(servicePath, UpdateServiceContent(binPath)); err != nil {
		return err
	}
	return write(timerPath, UpdateTimerContent())
}

func write(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}
