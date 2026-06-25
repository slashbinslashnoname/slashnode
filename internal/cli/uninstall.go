package cli

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/slashbinslashnoname/slashnode/internal/paths"
)

// Uninstall implémente `slashnoded uninstall`. Retire l'unit systemd, le
// service Avahi et le binaire. Avec --purge, supprime aussi config + données
// (confirmation requise sauf --yes).
func Uninstall(args []string) error {
	fs := flag.NewFlagSet("uninstall", flag.ContinueOnError)
	purge := fs.Bool("purge", false, "supprime aussi /etc/slashnode et /var/lib/slashnode")
	yes := fs.Bool("yes", false, "ne pas demander de confirmation")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *purge && !*yes {
		fmt.Print("⚠ --purge supprime config + secrets + données. Confirmer ? [o/N] ")
		var resp string
		fmt.Scanln(&resp)
		if resp != "o" && resp != "O" && resp != "y" && resp != "Y" {
			return fmt.Errorf("annulé")
		}
	}

	// Arrêt + désactivation des services (best effort, Linux).
	if runtime.GOOS == "linux" {
		runBestEffort("systemctl", "disable", "--now", "slashnoded")
		runBestEffort("systemctl", "disable", "--now", "slashnoded-update.timer")
	}

	remove(paths.SystemdUnit())
	remove(paths.SystemdUpdateService())
	remove(paths.SystemdUpdateTimer())
	remove(paths.AvahiService())
	remove(paths.BinPath())

	if runtime.GOOS == "linux" {
		runBestEffort("systemctl", "daemon-reload")
	}

	if *purge {
		removeAll(paths.ConfigDir())
		removeAll(paths.DataDir())
	}

	fmt.Println(colorize("✓ SlashNode désinstallé.", ansiRed))
	if !*purge {
		fmt.Println(colorize("  (config + données conservées ; --purge pour tout supprimer)", ansiDim))
	}
	return nil
}

func remove(path string) {
	if err := os.Remove(path); err == nil {
		fmt.Printf("→ supprimé %s\n", path)
	}
}

func removeAll(path string) {
	if err := os.RemoveAll(path); err == nil {
		fmt.Printf("→ supprimé %s\n", path)
	}
}

func runBestEffort(name string, args ...string) {
	if _, err := exec.LookPath(name); err != nil {
		return
	}
	_ = exec.Command(name, args...).Run()
}
