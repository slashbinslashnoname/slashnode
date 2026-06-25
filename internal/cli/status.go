package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/slashbinslashnoname/slashnode/internal/config"
	"github.com/slashbinslashnoname/slashnode/internal/paths"
)

// Status implements `slashnoded status`. With --post-install, displays the
// banner, the access URL and the initial credentials.
func Status(args []string) error {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	postInstall := fs.Bool("post-install", false, "display URL + credentials after installation")
	asJSON := fs.Bool("json", false, "JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load(paths.ConfigFile())
	if err != nil {
		return fmt.Errorf("node not initialized (run `slashnoded init`): %w", err)
	}

	urls := accessURLs(cfg)

	if *asJSON {
		out := map[string]any{
			"version":  cfg.Version,
			"node_id":  cfg.NodeID,
			"hostname": cfg.Hostname,
			"port":     cfg.HTTP.Port,
			"urls":     urls,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	}

	if *postInstall {
		Banner()
	}

	fmt.Printf("Node ID  : %s\n", cfg.NodeID)
	fmt.Printf("Version  : %s\n", cfg.Version)
	fmt.Printf("Mode     : %s\n", cfg.Access.Mode)
	fmt.Printf("Auth     : %s\n", authLabel(cfg))
	fmt.Println("Access   :")
	for _, u := range urls {
		fmt.Printf("  %s\n", colorize(u, ansiRed))
	}

	if *postInstall {
		printInitialCredentials()
	}
	return nil
}

func printInitialCredentials() {
	pw, err := os.ReadFile(paths.InitialPasswordFile())
	if err != nil {
		// No more file => password already retrieved / changed.
		fmt.Println(colorize("\nInitial password already consumed (file absent).", ansiDim))
		return
	}
	fmt.Println()
	fmt.Println(colorize("Initial credentials:", ansiRed))
	fmt.Printf("  username: %s\n", "admin")
	fmt.Printf("  password: %s\n", strings.TrimSpace(string(pw)))
	fmt.Println(colorize("  ⚠ change it after the first login.", ansiDim))
}

// authLabel describes whether the web UI requires a login.
func authLabel(cfg *config.Config) string {
	if cfg.Access.PasswordProtected {
		return "password-protected"
	}
	return "open (LAN)"
}

// accessURLs builds the list of access URLs (server address + mDNS + local IPs).
func accessURLs(cfg *config.Config) []string {
	var urls []string
	if cfg.Access.Mode == "server" && cfg.Access.Address != "" {
		urls = append(urls, fmt.Sprintf("http://%s:%d", cfg.Access.Address, cfg.HTTP.Port))
	}
	urls = append(urls, fmt.Sprintf("http://%s:%d", cfg.Hostname, cfg.HTTP.Port))
	for _, ip := range localIPs() {
		urls = append(urls, fmt.Sprintf("http://%s:%d", ip, cfg.HTTP.Port))
	}
	return urls
}

func localIPs() []string {
	var out []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return out
	}
	for _, a := range addrs {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() {
			continue
		}
		if ip4 := ipnet.IP.To4(); ip4 != nil {
			out = append(out, ip4.String())
		}
	}
	return out
}
