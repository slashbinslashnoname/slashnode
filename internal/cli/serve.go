package cli

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/slashbinslashnoname/slashnode/internal/apps"
	"github.com/slashbinslashnoname/slashnode/internal/config"
	"github.com/slashbinslashnoname/slashnode/internal/paths"
	"github.com/slashbinslashnoname/slashnode/internal/secrets"
	"github.com/slashbinslashnoname/slashnode/internal/updater"
)

// Serve starts the daemon: the local Go API (127.0.0.1, token auth) AND the
// Next.js front launched as a supervised subprocess.
func Serve(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	webDir := fs.String("web-dir", "", "Next.js front directory (default: auto)")
	noWeb := fs.Bool("no-web", false, "do not launch the Next.js front (API only)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load(paths.ConfigFile())
	if err != nil {
		return fmt.Errorf("node not initialized (run `slashnoded init`): %w", err)
	}
	sec, err := secrets.Load(paths.SecretsFile())
	if err != nil {
		return fmt.Errorf("secrets not found (run `slashnoded init`): %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	Banner()

	appsDir := resolveAppsDir()

	// 1. API Go locale.
	apiAddr := fmt.Sprintf("127.0.0.1:%d", cfg.HTTP.APIPort)
	apiSrv := &http.Server{
		Addr:              apiAddr,
		Handler:           apiHandler(cfg, sec, appsDir),
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		fmt.Printf("API     : %s\n", colorize("http://"+apiAddr, ansiDim))
		if err := apiSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintln(os.Stderr, "API error:", err)
			stop()
		}
	}()

	// 2. Next.js front (supervised).
	if !*noWeb {
		dir := resolveWebDir(*webDir)
		if dir == "" {
			fmt.Fprintln(os.Stderr, colorize("⚠ Next.js front not found — API only. (--web-dir to specify it)", ansiDim))
		} else {
			fmt.Printf("Front   : %s\n", colorize(fmt.Sprintf("http://%s:%d", cfg.Hostname, cfg.HTTP.Port), ansiRed))
			go superviseWeb(ctx, cfg, sec, dir)
		}
	}

	<-ctx.Done()
	fmt.Println("\nshutting down…")
	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return apiSrv.Shutdown(shutCtx)
}

// apiHandler builds the Go API router (all routes are protected by the Bearer
// token, except /healthz).
func apiHandler(cfg *config.Config, sec *secrets.Secrets, appsDir string) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "node": cfg.NodeID})
	})

	mux.Handle("/api/v1/status", bearer(sec, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"node_id":  cfg.NodeID,
			"version":  Version,
			"hostname": cfg.Hostname,
			"port":     cfg.HTTP.Port,
		})
	}))

	mux.Handle("/api/v1/update", bearer(sec, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, updater.LoadState(Version))
	}))

	mux.Handle("POST /api/v1/update/apply", bearer(sec, func(w http.ResponseWriter, r *http.Request) {
		// Runs in the background: the restart will kill this process.
		go func() { _ = updater.Apply("latest", cfg.Update.Channel) }()
		writeJSON(w, http.StatusAccepted, map[string]string{"status": "applying"})
	}))

	// Verifies the admin password (called server-side by the Next /api/login
	// route, which then sets the session cookie).
	mux.Handle("POST /api/v1/auth", bearer(sec, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Password string `json:"password"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if !sec.Verify(body.Password) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid password"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}))

	// --- App Store ---
	mux.Handle("GET /api/v1/apps", bearer(sec, func(w http.ResponseWriter, r *http.Request) {
		cat, err := apps.Catalog(appsDir)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"apps": cat})
	}))

	mux.Handle("GET /api/v1/apps/{id}", bearer(sec, func(w http.ResponseWriter, r *http.Request) {
		man, err := apps.Find(appsDir, r.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		_, installed := apps.LoadState().Installed[man.ID]
		writeJSON(w, http.StatusOK, apps.CatalogEntry{Manifest: *man, Installed: installed})
	}))

	mux.Handle("POST /api/v1/apps/{id}/install", bearer(sec, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Inputs map[string]string `json:"inputs"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if err := apps.Install(appsDir, r.PathValue("id"), body.Inputs); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "installed"})
	}))

	mux.Handle("POST /api/v1/apps/{id}/uninstall", bearer(sec, func(w http.ResponseWriter, r *http.Request) {
		purge := r.URL.Query().Get("purge") == "true"
		if err := apps.Uninstall(r.PathValue("id"), purge); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "uninstalled"})
	}))

	mux.Handle("GET /api/v1/apps/{id}/status", bearer(sec, func(w http.ResponseWriter, r *http.Request) {
		st, err := apps.Status(r.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"docker": apps.DockerAvailable(), "services": st})
	}))

	mux.Handle("GET /api/v1/apps/{id}/logs", bearer(sec, func(w http.ResponseWriter, r *http.Request) {
		tail := 200
		if t := r.URL.Query().Get("tail"); t != "" {
			if n, err := strconv.Atoi(t); err == nil {
				tail = n
			}
		}
		logs, err := apps.Logs(r.PathValue("id"), tail)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"logs": logs})
	}))

	lifecycle := func(action func(string) error, ok string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if err := action(r.PathValue("id")); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"status": ok})
		}
	}
	mux.Handle("POST /api/v1/apps/{id}/start", bearer(sec, lifecycle(apps.Start, "started")))
	mux.Handle("POST /api/v1/apps/{id}/stop", bearer(sec, lifecycle(apps.Stop, "stopped")))
	mux.Handle("POST /api/v1/apps/{id}/restart", bearer(sec, lifecycle(apps.Restart, "restarted")))

	return mux
}

// resolveAppsDir locates the app catalog: env > system path > ./examples (dev).
func resolveAppsDir() string {
	candidates := []string{os.Getenv("SLASHNODE_APPS_DIR"), paths.AppsDir(), "examples"}
	for _, c := range candidates {
		if c == "" {
			continue
		}
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return "examples"
}

// resolveWebDir determines the front directory: flag > env > system path >
// ./web (dev).
func resolveWebDir(flagVal string) string {
	candidates := []string{flagVal, os.Getenv("SLASHNODE_WEB_DIR"), paths.WebDir(), "web"}
	for _, c := range candidates {
		if c == "" {
			continue
		}
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

// superviseWeb launches the Next server and relaunches it if it dies, until
// the context is cancelled.
func superviseWeb(ctx context.Context, cfg *config.Config, sec *secrets.Secrets, dir string) {
	backoff := time.Second
	for ctx.Err() == nil {
		cmd := webCommand(ctx, cfg, sec, dir)
		fmt.Printf("→ front: %s (cwd %s)\n", strings.Join(cmd.Args, " "), dir)
		start := time.Now()
		if err := cmd.Run(); err != nil && ctx.Err() == nil {
			fmt.Fprintln(os.Stderr, "front stopped:", err)
		}
		if ctx.Err() != nil {
			return
		}
		// Reset the backoff if the process held up for a while.
		if time.Since(start) > 30*time.Second {
			backoff = time.Second
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
}

// webCommand builds the front launch command. Prefers the standalone build
// (node server.js) otherwise `npm run start`.
func webCommand(ctx context.Context, cfg *config.Config, sec *secrets.Secrets, dir string) *exec.Cmd {
	var cmd *exec.Cmd
	if _, err := os.Stat(filepath.Join(dir, "server.js")); err == nil {
		cmd = exec.CommandContext(ctx, "node", "server.js")
	} else {
		cmd = exec.CommandContext(ctx, "npm", "run", "start")
	}
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(),
		"NODE_ENV=production",
		"PORT="+strconv.Itoa(cfg.HTTP.Port),
		"HOSTNAME="+cfg.HTTP.Bind,
		fmt.Sprintf("SLASHNODE_API_URL=http://127.0.0.1:%d", cfg.HTTP.APIPort),
		"SLASHNODE_API_TOKEN="+sec.APIToken,
		fmt.Sprintf("SLASHNODE_PASSWORD_PROTECTED=%t", cfg.Access.PasswordProtected),
		"SLASHNODE_SESSION_SECRET="+sec.SessionSecret,
		"SLASHNODE_ACCESS_MODE="+cfg.Access.Mode,
	)
	// Process group so we can kill the whole Next tree on shutdown.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
	}
	return cmd
}

// bearer protects a handler with a Bearer token (== secrets.APIToken).
func bearer(sec *secrets.Secrets, next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const prefix = "Bearer "
		auth := r.Header.Get("Authorization")
		token := strings.TrimPrefix(auth, prefix)
		if !strings.HasPrefix(auth, prefix) ||
			subtle.ConstantTimeCompare([]byte(token), []byte(sec.APIToken)) != 1 {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		next(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
