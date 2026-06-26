package cli

import (
	"fmt"
	"os"

	"github.com/slashbinslashnoname/slashnode/internal/apps"
)

// Prune implements `slashnoded prune` (run by the daily timer and on bootstrap/
// updates): removes dangling docker images to reclaim disk. Stopped containers
// and volumes are preserved.
func Prune(args []string) error {
	if !apps.DockerAvailable() {
		fmt.Println("docker not available — nothing to prune")
		return nil
	}
	fmt.Println("→ pruning dangling docker images…")
	apps.PruneImages(os.Stdout)
	fmt.Println(colorize("✓ prune done", ansiRed))
	return nil
}
