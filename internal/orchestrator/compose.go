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
	"strconv"
	"strings"
)

// Network is the shared Docker network all app containers join.
const Network = "slashnode"

// Port maps a host port to a container port. HostIP optionally binds the host
// side to a specific address (e.g. 127.0.0.1 to keep an RPC port local-only).
type Port struct {
	HostIP    string `json:"hostIP,omitempty"`
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
// configMounts maps a service name to extra bind-mount specs
// ("hostpath:containerpath:ro") for rendered config files.
func BuildCompose(appID string, services map[string]Service, env map[string]string, configMounts map[string][]string) ([]byte, error) {
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
				if p.HostIP != "" {
					spec = fmt.Sprintf("%s:%d:%d", p.HostIP, p.Host, p.Container)
				}
				if p.Protocol != "" {
					spec += "/" + p.Protocol
				}
				ports = append(ports, spec)
			}
			svc["ports"] = ports
		}
		mounts := make([]string, 0, len(s.Volumes))
		for _, v := range s.Volumes {
			volName := appID + "_" + v.Name
			mounts = append(mounts, volName+":"+v.Path)
			volumes[volName] = map[string]any{}
		}
		mounts = append(mounts, configMounts[name]...)
		if len(mounts) > 0 {
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

// ContainerStatus is the status of one service container of an app.
type ContainerStatus struct {
	Service string `json:"service"`
	State   string `json:"state"`
	Status  string `json:"status"`
	Health  string `json:"health,omitempty"`
}

// Status returns the per-service container status (docker compose ps).
func Status(appID, composeFile string) ([]ContainerStatus, error) {
	out, err := output("docker", "compose", "-p", project(appID), "-f", composeFile, "ps", "-a", "--format", "json")
	if err != nil {
		return nil, err
	}
	var res []ContainerStatus
	// Compose may emit NDJSON (one object per line) or a JSON array.
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var c ContainerStatus
		if err := json.Unmarshal([]byte(line), &c); err == nil && c.Service != "" {
			res = append(res, c)
			continue
		}
		var arr []ContainerStatus
		if err := json.Unmarshal([]byte(line), &arr); err == nil {
			res = append(res, arr...)
		}
	}
	return res, nil
}

// Start, Stop and Restart drive the lifecycle of an already-created app.
func Start(appID, composeFile string) error {
	return run("docker", "compose", "-p", project(appID), "-f", composeFile, "start")
}
func Stop(appID, composeFile string) error {
	return run("docker", "compose", "-p", project(appID), "-f", composeFile, "stop")
}
func Restart(appID, composeFile string) error {
	return run("docker", "compose", "-p", project(appID), "-f", composeFile, "restart")
}

// Logs returns the last `tail` log lines across the app's services.
func Logs(appID, composeFile string, tail int) (string, error) {
	return output("docker", "compose", "-p", project(appID), "-f", composeFile,
		"logs", "--no-color", "--tail", strconv.Itoa(tail))
}

func project(appID string) string { return "slashnode-" + appID }

func output(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var out, errb strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s %s: %w\n%s", name, strings.Join(args, " "), err, errb.String())
	}
	return out.String(), nil
}

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
