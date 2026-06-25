package cli

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/slashbinslashnoname/slashnode/internal/paths"
)

// Uninstall implements `slashnoded uninstall`. Removes the systemd unit, the
// Avahi service and the binary. With --purge, also deletes config + data
// (confirmation required unless --yes).
func Uninstall(args []string) error {
	fs := flag.NewFlagSet("uninstall", flag.ContinueOnError)
	purge := fs.Bool("purge", false, "also delete /etc/slashnode and /var/lib/slashnode")
	yes := fs.Bool("yes", false, "do not ask for confirmation")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *purge && !*yes {
		fmt.Print("⚠ --purge deletes config + secrets + data. Confirm? [y/N] ")
		var resp string
		fmt.Scanln(&resp)
		if resp != "o" && resp != "O" && resp != "y" && resp != "Y" {
			return fmt.Errorf("cancelled")
		}
	}

	// Stop + disable the services (best effort, Linux).
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
		removeAll(paths.TorDataDir()) // generated hidden-service keys
	}

	fmt.Println(colorize("✓ SlashNode uninstalled.", ansiRed))
	if !*purge {
		fmt.Println(colorize("  (config + data kept; --purge to delete everything)", ansiDim))
	}
	return nil
}

func remove(path string) {
	if err := os.Remove(path); err == nil {
		fmt.Printf("→ removed %s\n", path)
	}
}

func removeAll(path string) {
	if err := os.RemoveAll(path); err == nil {
		fmt.Printf("→ removed %s\n", path)
	}
}

func runBestEffort(name string, args ...string) {
	if _, err := exec.LookPath(name); err != nil {
		return
	}
	_ = exec.Command(name, args...).Run()
}
