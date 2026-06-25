// Package orchestrator turns an app manifest's `services` block into a Docker
// Compose file (rendered as JSON, which is valid YAML) and drives the container
// lifecycle via `docker compose`.
//
// All app containers join a shared external network ("slashnode") and set a
// container_name equal to their service name, so that exports such as
// "rpc.host": "bitcoind" resolve across apps by DNS.
package orchestrator

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Network is the shared Docker network all app containers join.
const Network = "slashnode"

// Port maps a host port to a container port.
type Port struct {
	Host      int    `json:"host"`
	Container int    `json:"container"`
	Protocol  string `json:"protocol,omitempty"`
}

// Volume maps a named volume to a container path.
type Volume struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// Service is one container declared in a manifest's `services` block.
type Service struct {
	Image       string            `json:"image"`
	Command     string            `json:"command,omitempty"`
	ShmSize     string            `json:"shmSize,omitempty"`
	Ports       []Port            `json:"ports,omitempty"`
	Volumes     []Volume          `json:"volumes,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
}

// ParseServices decodes a manifest's raw `services` block.
func ParseServices(raw json.RawMessage) (map[string]Service, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("manifest has no services")
	}
	var svcs map[string]Service
	if err := json.Unmarshal(raw, &svcs); err != nil {
		return nil, fmt.Errorf("invalid services block: %w", err)
	}
	return svcs, nil
}

// BuildCompose renders a Compose file (as JSON) for the given app. env is merged
// into every service's environment (manifest defaults are kept, env wins).
func BuildCompose(appID string, services map[string]Service, env map[string]string) ([]byte, error) {
	svcMap := map[string]any{}
	volumes := map[string]any{}

	for name, s := range services {
		svc := map[string]any{
			"image":          s.Image,
			"container_name": name,
			"restart":        "unless-stopped",
			"networks":       []string{Network},
		}
		if s.Command != "" {
			svc["command"] = s.Command
		}
		if s.ShmSize != "" {
			svc["shm_size"] = s.ShmSize
		}
		if len(s.Ports) > 0 {
			ports := make([]string, 0, len(s.Ports))
			for _, p := range s.Ports {
				spec := fmt.Sprintf("%d:%d", p.Host, p.Container)
				if p.Protocol != "" {
					spec += "/" + p.Protocol
				}
				ports = append(ports, spec)
			}
			svc["ports"] = ports
		}
		if len(s.Volumes) > 0 {
			mounts := make([]string, 0, len(s.Volumes))
			for _, v := range s.Volumes {
				volName := appID + "_" + v.Name
				mounts = append(mounts, volName+":"+v.Path)
				volumes[volName] = map[string]any{}
			}
			svc["volumes"] = mounts
		}

		merged := map[string]string{}
		for k, v := range s.Environment {
			merged[k] = v
		}
		for k, v := range env {
			merged[k] = v
		}
		if len(merged) > 0 {
			svc["environment"] = merged
		}

		svcMap[name] = svc
	}

	compose := map[string]any{
		"name":     "slashnode-" + appID,
		"services": svcMap,
		"networks": map[string]any{Network: map[string]any{"external": true}},
	}
	if len(volumes) > 0 {
		compose["volumes"] = volumes
	}
	return json.MarshalIndent(compose, "", "  ")
}

// Available reports whether docker (with a reachable daemon) can be used.
// SLASHNODE_NO_DOCKER forces it off (used in tests and on dev machines).
func Available() bool {
	if os.Getenv("SLASHNODE_NO_DOCKER") != "" {
		return false
	}
	if _, err := exec.LookPath("docker"); err != nil {
		return false
	}
	return exec.Command("docker", "info").Run() == nil
}

// EnsureNetwork creates the shared network if it does not exist yet.
func EnsureNetwork() error {
	if exec.Command("docker", "network", "inspect", Network).Run() == nil {
		return nil
	}
	return run("docker", "network", "create", Network)
}

// Up brings the app up (docker compose up -d).
func Up(appID, composeFile string) error {
	return run("docker", "compose", "-p", project(appID), "-f", composeFile, "up", "-d")
}

// Down stops and removes the app (optionally its volumes).
func Down(appID, composeFile string, removeVolumes bool) error {
	args := []string{"compose", "-p", project(appID), "-f", composeFile, "down"}
	if removeVolumes {
		args = append(args, "-v")
	}
	return run("docker", args...)
}

func project(appID string) string { return "slashnode-" + appID }

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	var out strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s: %w\n%s", name, strings.Join(args, " "), err, out.String())
	}
	return nil
}
