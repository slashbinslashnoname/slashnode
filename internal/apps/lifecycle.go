package apps

import (
	"github.com/slashbinslashnoname/slashnode/internal/orchestrator"
	"github.com/slashbinslashnoname/slashnode/internal/paths"
)

// DockerAvailable reports whether a Docker daemon is reachable.
func DockerAvailable() bool { return orchestrator.Available() }

// Status returns the container status of an installed app. When Docker is not
// available it returns nil (the caller treats it as "unknown").
func Status(id string) ([]orchestrator.ContainerStatus, error) {
	if !orchestrator.Available() {
		return nil, nil
	}
	return orchestrator.Status(id, paths.AppComposeFile(id))
}

// Start, Stop and Restart drive the app's containers (no-op without Docker).
func Start(id string) error {
	if !orchestrator.Available() {
		return nil
	}
	return orchestrator.Start(id, paths.AppComposeFile(id))
}

func Stop(id string) error {
	if !orchestrator.Available() {
		return nil
	}
	return orchestrator.Stop(id, paths.AppComposeFile(id))
}

func Restart(id string) error {
	if !orchestrator.Available() {
		return nil
	}
	return orchestrator.Restart(id, paths.AppComposeFile(id))
}

// Logs returns the last `tail` log lines of an installed app.
func Logs(id string, tail int) (string, error) {
	if !orchestrator.Available() {
		return "", nil
	}
	return orchestrator.Logs(id, paths.AppComposeFile(id), tail)
}

// ClearLogs truncates the app's container logs.
func ClearLogs(id string) error {
	if !orchestrator.Available() {
		return nil
	}
	return orchestrator.ClearLogs(id, paths.AppComposeFile(id))
}
