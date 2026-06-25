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

// Status implémente `slashnoded status`. Avec --post-install, affiche la
// bannière, l'URL d'accès et les identifiants initiaux.
func Status(args []string) error {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	postInstall := fs.Bool("post-install", false, "affiche URL + identifiants après installation")
	asJSON := fs.Bool("json", false, "sortie JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load(paths.ConfigFile())
	if err != nil {
		return fmt.Errorf("nœud non initialisé (lance `slashnoded init`) : %w", err)
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
	fmt.Println("Accès    :")
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
		// Plus de fichier => mot de passe déjà récupéré / changé.
		fmt.Println(colorize("\nMot de passe initial déjà consommé (fichier absent).", ansiDim))
		return
	}
	fmt.Println()
	fmt.Println(colorize("Identifiants initiaux :", ansiRed))
	fmt.Printf("  utilisateur : %s\n", "admin")
	fmt.Printf("  mot de passe: %s\n", strings.TrimSpace(string(pw)))
	fmt.Println(colorize("  ⚠ change-le après la première connexion.", ansiDim))
}

// accessURLs construit la liste des URL d'accès (mDNS + IP locales).
func accessURLs(cfg *config.Config) []string {
	urls := []string{fmt.Sprintf("http://%s:%d", cfg.Hostname, cfg.HTTP.Port)}
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
