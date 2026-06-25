// Package orchestrator drives the container lifecycle (docker compose up/down/…)
// for each app from a docker-compose document and parses image references back
// out of it for the version picker.
//
// Apps ship a real compose document in their manifest's `compose` field. By
// convention each service joins the shared external network ("slashnode") and
// sets a container_name equal to its service name, so that exports such as
// "rpc.host": "bitcoind" resolve across apps by DNS.
package orchestrator

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Network is the shared Docker network all app containers join.
const Network = "slashnode"

// ParseComposeImages extracts the service → image map from a docker-compose
// document (YAML or our JSON-rendered equivalent), so the version picker can
// list/override the tag of each service. Best-effort line parser: it tracks the
// service keys under the top-level `services:` map and reads each one's `image:`.
func ParseComposeImages(content string) map[string]string {
	out := map[string]string{}
	inServices := false
	servicesIndent := -1
	svcKeyIndent := -1
	current := ""
	for _, raw := range strings.Split(content, "\n") {
		line := strings.TrimRight(raw, "\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " \t"))

		if !inServices {
			if trimmed == "services:" {
				inServices = true
				servicesIndent = indent
				svcKeyIndent = -1
			}
			continue
		}
		// A key at or above the `services:` indent ends the services block.
		if indent <= servicesIndent {
			inServices = false
			current = ""
			continue
		}
		if svcKeyIndent == -1 {
			svcKeyIndent = indent
		}
		switch {
		case indent == svcKeyIndent && strings.HasSuffix(trimmed, ":"):
			current = strings.TrimSpace(strings.TrimSuffix(trimmed, ":"))
		case indent > svcKeyIndent && current != "" && strings.HasPrefix(trimmed, "image:"):
			img := strings.TrimSpace(strings.TrimPrefix(trimmed, "image:"))
			img = strings.Trim(img, `"'`)
			if i := strings.IndexAny(img, " \t#"); i >= 0 { // strip trailing comment
				img = strings.TrimSpace(img[:i])
			}
			if img != "" {
				out[current] = img
			}
		}
	}
	return out
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

// Pull fetches the latest images for the app's services.
func Pull(appID, composeFile string) error {
	return run("docker", "compose", "-p", project(appID), "-f", composeFile, "pull")
}

// PullStreamed pulls the app's images, streaming docker's output to w.
func PullStreamed(appID, composeFile string, w io.Writer) error {
	return runStreamed(w, "docker", "compose", "-p", project(appID), "-f", composeFile, "pull")
}

// UpStreamed brings the app up, streaming docker's output to w.
func UpStreamed(appID, composeFile string, w io.Writer) error {
	return runStreamed(w, "docker", "compose", "-p", project(appID), "-f", composeFile, "up", "-d")
}

// ImagesOutdated reports whether any of the app's images has a newer version in
// its registry (remote manifest digest differs from the local one). Best-effort:
// returns false on any error or for not-yet-pulled images.
func ImagesOutdated(appID, composeFile string) bool {
	out, err := output("docker", "compose", "-p", project(appID), "-f", composeFile, "config", "--images")
	if err != nil {
		return false
	}
	for _, img := range strings.Fields(out) {
		if imageOutdated(img) {
			return true
		}
	}
	return false
}

func imageOutdated(image string) bool {
	local, err := output("docker", "image", "inspect", "--format", "{{index .RepoDigests 0}}", image)
	if err != nil {
		return false // not pulled yet — nothing to compare
	}
	remote, err := output("docker", "buildx", "imagetools", "inspect", "--format", "{{.Manifest.Digest}}", image)
	if err != nil {
		return false
	}
	ld := strings.TrimSpace(local)
	if i := strings.Index(ld, "@"); i >= 0 {
		ld = ld[i+1:]
	}
	rd := strings.TrimSpace(remote)
	return ld != "" && rd != "" && ld != rd
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

// ClearLogs truncates the JSON log file of each of the app's containers.
func ClearLogs(appID, composeFile string) error {
	out, err := output("docker", "compose", "-p", project(appID), "-f", composeFile, "ps", "-q")
	if err != nil {
		return err
	}
	for _, id := range strings.Fields(out) {
		lp, err := output("docker", "inspect", "--format", "{{.LogPath}}", id)
		if err != nil {
			continue
		}
		if p := strings.TrimSpace(lp); p != "" {
			_ = exec.Command("truncate", "-s", "0", p).Run()
		}
	}
	return nil
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

// runStreamed runs a command writing its combined stdout+stderr to w live.
func runStreamed(w io.Writer, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = w
	cmd.Stderr = w
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return nil
}
