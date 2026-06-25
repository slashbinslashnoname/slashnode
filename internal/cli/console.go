package cli

import (
	"context"
	"crypto/subtle"
	"net/http"
	"os/exec"

	"github.com/coder/websocket"
	"github.com/creack/pty"

	"github.com/slashbinslashnoname/slashnode/internal/config"
	"github.com/slashbinslashnoname/slashnode/internal/secrets"
)

// consoleHandler serves an interactive shell into a container over a WebSocket
// (xterm.js on the front, `docker exec -it` + PTY on the back). It is reached
// through Caddy at /__console; auth mirrors the UI (session cookie when the node
// is password-protected).
func consoleHandler(cfg *config.Config, sec *secrets.Secrets) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg.Access.PasswordProtected {
			c, err := r.Cookie("slashnode_session")
			if err != nil || subtle.ConstantTimeCompare([]byte(c.Value), []byte(sec.SessionSecret)) != 1 {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
		}
		container := r.URL.Query().Get("c")
		if container == "" {
			http.Error(w, "missing container", http.StatusBadRequest)
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

func runConsole(conn *websocket.Conn, container string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// -it gives the container process a PTY; the pty package provides the
	// client-side PTY that `docker exec -t` requires.
	cmd := exec.Command("docker", "exec", "-it", container, "sh", "-c",
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

	// browser input → container
	for {
		_, data, rerr := conn.Read(ctx)
		if rerr != nil {
			return
		}
		if _, werr := ptmx.Write(data); werr != nil {
			return
		}
	}
}
