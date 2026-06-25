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

	"github.com/slashbinslashnoname/slashnode/internal/config"
	"github.com/slashbinslashnoname/slashnode/internal/paths"
	"github.com/slashbinslashnoname/slashnode/internal/secrets"
	"github.com/slashbinslashnoname/slashnode/internal/updater"
)

// Serve démarre le démon : l'API Go locale (127.0.0.1, auth par token) ET le
// front Next.js lancé en sous-processus et supervisé.
func Serve(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	webDir := fs.String("web-dir", "", "dossier du front Next.js (défaut : auto)")
	noWeb := fs.Bool("no-web", false, "ne pas lancer le front Next.js (API seule)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load(paths.ConfigFile())
	if err != nil {
		return fmt.Errorf("nœud non initialisé (lance `slashnoded init`) : %w", err)
	}
	sec, err := secrets.Load(paths.SecretsFile())
	if err != nil {
		return fmt.Errorf("secrets introuvables (lance `slashnoded init`) : %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	Banner()

	// 1. API Go locale.
	apiAddr := fmt.Sprintf("127.0.0.1:%d", cfg.HTTP.APIPort)
	apiSrv := &http.Server{
		Addr:              apiAddr,
		Handler:           apiHandler(cfg, sec),
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		fmt.Printf("API     : %s\n", colorize("http://"+apiAddr, ansiDim))
		if err := apiSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintln(os.Stderr, "erreur API :", err)
			stop()
		}
	}()

	// 2. Front Next.js (supervisé).
	if !*noWeb {
		dir := resolveWebDir(*webDir)
		if dir == "" {
			fmt.Fprintln(os.Stderr, colorize("⚠ front Next.js introuvable — API seule. (--web-dir pour le préciser)", ansiDim))
		} else {
			fmt.Printf("Front   : %s\n", colorize(fmt.Sprintf("http://%s:%d", cfg.Hostname, cfg.HTTP.Port), ansiRed))
			go superviseWeb(ctx, cfg, sec, dir)
		}
	}

	<-ctx.Done()
	fmt.Println("\narrêt…")
	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return apiSrv.Shutdown(shutCtx)
}

// apiHandler construit le routeur de l'API Go (toutes les routes sont protégées
// par le token Bearer, sauf /healthz).
func apiHandler(cfg *config.Config, sec *secrets.Secrets) http.Handler {
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

	mux.Handle("/api/v1/update/apply", bearer(sec, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "POST requis"})
			return
		}
		// Lancé en arrière-plan : le redémarrage tuera ce processus.
		go func() { _ = updater.Apply("latest", cfg.Update.Channel) }()
		writeJSON(w, http.StatusAccepted, map[string]string{"status": "applying"})
	}))

	return mux
}

// resolveWebDir détermine le dossier du front : flag > env > chemin système >
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

// superviseWeb lance le serveur Next et le relance s'il meurt, jusqu'à
// l'annulation du contexte.
func superviseWeb(ctx context.Context, cfg *config.Config, sec *secrets.Secrets, dir string) {
	backoff := time.Second
	for ctx.Err() == nil {
		cmd := webCommand(ctx, cfg, sec, dir)
		fmt.Printf("→ front : %s (cwd %s)\n", strings.Join(cmd.Args, " "), dir)
		start := time.Now()
		if err := cmd.Run(); err != nil && ctx.Err() == nil {
			fmt.Fprintln(os.Stderr, "front arrêté :", err)
		}
		if ctx.Err() != nil {
			return
		}
		// Reset du backoff si le process a tenu un moment.
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

// webCommand construit la commande de lancement du front. Privilégie le build
// autonome (node server.js) sinon `npm run start`.
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
	)
	// Groupe de process pour pouvoir tuer tout l'arbre Next à l'arrêt.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
	}
	return cmd
}

// bearer protège un handler par token Bearer (== secrets.APIToken).
func bearer(sec *secrets.Secrets, next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const prefix = "Bearer "
		auth := r.Header.Get("Authorization")
		token := strings.TrimPrefix(auth, prefix)
		if !strings.HasPrefix(auth, prefix) ||
			subtle.ConstantTimeCompare([]byte(token), []byte(sec.APIToken)) != 1 {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "non autorisé"})
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
