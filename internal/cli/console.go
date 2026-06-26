package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"

	"github.com/coder/websocket"
	"github.com/creack/pty"

	"github.com/slashbinslashnoname/slashnode/internal/config"
	"github.com/slashbinslashnoname/slashnode/internal/secrets"
)

// consoleHandler serves an interactive shell into a container over a WebSocket
// (xterm.js on the front, `docker exec -it` + PTY on the back). It is reached
// through Caddy at /__console. A container shell is the highest-privilege
// operation, so it ALWAYS requires a valid session cookie — even in open
// (non-password) mode, where the front issues the cookie to same-origin loads.
// This blocks cross-site and non-browser clients from opening a shell.
func consoleHandler(cfg *config.Config, sec *secrets.Secrets) http.HandlerFunc {
	_ = cfg
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("slashnode_session")
		if err != nil || !sec.VerifySession(c.Value) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		container := r.URL.Query().Get("c")
		if container == "" {
			http.Error(w, "missing container", http.StatusBadRequest)
			return
		}
		// Only allow shelling into containers slashnode manages (compose project
		// "slashnode-*"), not arbitrary containers on the host's Docker daemon.
		if !isManagedContainer(container) {
			http.Error(w, "unknown container", http.StatusForbidden)
			return
		}

		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{})
		if err != nil {
			return
		}
		defer conn.CloseNow()
		conn.SetReadLimit(1 << 20)

		runConsole(conn, container)
	}
}

// isManagedContainer reports whether the named container belongs to a slashnode
// compose project (label com.docker.compose.project = "slashnode-*").
func isManagedContainer(name string) bool {
	out, err := exec.Command("docker", "inspect", "--format",
		`{{index .Config.Labels "com.docker.compose.project"}}`, "--", name).Output()
	if err != nil {
		return false
	}
	return strings.HasPrefix(strings.TrimSpace(string(out)), "slashnode-")
}

func runConsole(conn *websocket.Conn, container string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// -it gives the container process a PTY; the pty package provides the
	// client-side PTY that `docker exec -t` requires.
	// `--` stops flag parsing so a container name starting with '-' can't be
	// interpreted as a docker exec flag (e.g. --privileged, --user).
	cmd := exec.Command("docker", "exec", "-it", "--", container, "sh", "-c",
		"exec $(command -v bash >/dev/null 2>&1 && echo bash || echo sh)")
	ptmx, err := pty.Start(cmd)
	if err != nil {
		_ = conn.Write(ctx, websocket.MessageText, []byte("\r\nfailed to open console: "+err.Error()+"\r\n"))
		return
	}
	_ = pty.Setsize(ptmx, &pty.Winsize{Rows: 32, Cols: 120})
	defer func() {
		_ = ptmx.Close()
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	}()

	// container output → browser
	go func() {
		buf := make([]byte, 4096)
		for {
			n, rerr := ptmx.Read(buf)
			if n > 0 {
				if werr := conn.Write(ctx, websocket.MessageBinary, buf[:n]); werr != nil {
					cancel()
					return
				}
			}
			if rerr != nil {
				cancel()
				return
			}
		}
	}()

	// browser input → container. Binary messages are stdin; text messages are
	// control (terminal resize).
	for {
		typ, data, rerr := conn.Read(ctx)
		if rerr != nil {
			return
		}
		if typ == websocket.MessageText {
			var ctrl struct {
				Resize *struct {
					Cols uint16 `json:"cols"`
					Rows uint16 `json:"rows"`
				} `json:"resize"`
			}
			if json.Unmarshal(data, &ctrl) == nil && ctrl.Resize != nil && ctrl.Resize.Cols > 0 {
				_ = pty.Setsize(ptmx, &pty.Winsize{Rows: ctrl.Resize.Rows, Cols: ctrl.Resize.Cols})
			}
			continue
		}
		if _, werr := ptmx.Write(data); werr != nil {
			return
		}
	}
}
