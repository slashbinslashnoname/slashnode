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
	"path/filepath"
	"strconv"
	"strings"
	"time"
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

// ContainerStatus is the status of one service container of an app. Container is
// the real Docker container name (needed to attach a console — it differs from
// the service name for multi-service apps and extra instances).
type ContainerStatus struct {
	Service   string `json:"service"`
	Container string `json:"container"`
	State     string `json:"state"`
	Status    string `json:"status"`
	Health    string `json:"health,omitempty"`
}

// composePS mirrors the fields `docker compose ps --format json` emits (matched
// case-insensitively); Name is the container name.
type composePS struct {
	Service string
	Name    string
	State   string
	Status  string
	Health  string
}

func (c composePS) toStatus() ContainerStatus {
	return ContainerStatus{Service: c.Service, Container: c.Name, State: c.State, Status: c.Status, Health: c.Health}
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
		var c composePS
		if err := json.Unmarshal([]byte(line), &c); err == nil && c.Service != "" {
			res = append(res, c.toStatus())
			continue
		}
		var arr []composePS
		if err := json.Unmarshal([]byte(line), &arr); err == nil {
			for _, a := range arr {
				res = append(res, a.toStatus())
			}
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
	return runComposePull(w, "docker", "compose", "-p", project(appID), "-f", composeFile, "pull")
}

// UpStreamed brings the app up, streaming docker's output to w. `up` also pulls
// any images not present locally, so it goes through the same rate-limit-aware
// wrapper as PullStreamed.
func UpStreamed(appID, composeFile string, w io.Writer) error {
	return runComposePull(w, "docker", "compose", "-p", project(appID), "-f", composeFile, "up", "-d")
}

// composePullParallel caps how many images docker compose pulls at once. A burst
// of simultaneous anonymous pulls (a multi-image app like supabase fetches ~8)
// is what trips Docker Hub's per-IP rate limiter, so we serialise them.
const composePullParallel = 3

// runComposePull runs a compose command that may pull images. It bounds pull
// concurrency and, on a Docker Hub 429 (Too Many Requests), retries with backoff
// — cached layers make each retry cheaper and the limit is a short rolling
// window, so a burst that got throttled usually succeeds on a later attempt.
// For a genuinely exhausted quota, `docker login` on the host is the real fix.
func runComposePull(w io.Writer, name string, args ...string) error {
	const maxAttempts = 4
	env := []string{"COMPOSE_PARALLEL_LIMIT=" + strconv.Itoa(composePullParallel)}
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		var captured strings.Builder
		err := runStreamedEnv(io.MultiWriter(w, &captured), env, name, args...)
		if err == nil {
			return nil
		}
		lastErr = err
		if attempt >= maxAttempts || !isRateLimited(captured.String()) {
			break
		}
		delay := time.Duration(attempt*attempt) * 5 * time.Second
		fmt.Fprintf(w, "\n--> Docker Hub rate limit hit (429); retrying in %s (attempt %d/%d)…\n",
			delay, attempt+1, maxAttempts)
		time.Sleep(delay)
	}
	return lastErr
}

// isRateLimited reports whether docker's output indicates a Docker Hub 429.
func isRateLimited(out string) bool {
	l := strings.ToLower(out)
	return strings.Contains(l, "toomanyrequests") ||
		strings.Contains(l, "429 too many requests") ||
		strings.Contains(l, "too many requests")
}

// Prune reclaims disk by removing dangling images (old layers left behind by
// image/version updates). It deliberately does NOT touch containers (stopped
// apps are preserved) or volumes (app data: chains, databases… are preserved).
// Best-effort; output is streamed to w.
func Prune(w io.Writer) error {
	return runStreamed(w, "docker", "image", "prune", "-f")
}

// ExecStreamed runs a one-off command inside a running service container (used
// by per-app migrations), streaming output to w.
func ExecStreamed(appID, composeFile, service, command string, w io.Writer) error {
	return runStreamed(w, "docker", "compose", "-p", project(appID), "-f", composeFile,
		"exec", "-T", service, "sh", "-c", command)
}

// CopyVolume copies the contents of the `from` docker volume into a new `to`
// volume (used by migrations that rename a volume). The destination is created
// if missing.
func CopyVolume(from, to string) error {
	if err := run("docker", "volume", "create", to); err != nil {
		return err
	}
	return run("docker", "run", "--rm", "-v", from+":/from:ro", "-v", to+":/to",
		"alpine", "sh", "-c", "cp -a /from/. /to/")
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

// Logs returns the last `tail` log lines across the app's services. If the logs
// have been cleared, only entries after the clear instant are returned.
func Logs(appID, composeFile string, tail int) (string, error) {
	args := []string{"compose", "-p", project(appID), "-f", composeFile,
		"logs", "--no-color", "--tail", strconv.Itoa(tail)}
	if since := readLogsSince(composeFile); since != "" {
		args = append(args, "--since", since)
	}
	return output("docker", args...)
}

// ClearLogs clears an app's logs by recording a "since" marker: Logs() filters
// to entries after this instant, which works for every logging driver.
//
// We deliberately do NOT truncate the container's on-disk log file. Docker holds
// the file descriptor open and keeps writing at its old offset, so an in-place
// truncate-to-zero leaves a hole of NUL bytes before the next entry. The
// json-file/local readers then hit a corrupted leading "line", which makes
// subsequent `docker compose logs` (i.e. the next refresh) return garbage or
// nothing — so the truncation both fails to reclaim disk (the file keeps growing
// from the old offset) and breaks reads. The marker alone is the reliable clear;
// disk is bounded by the logging driver's rotation, not by us.
func ClearLogs(appID, composeFile string) error {
	// Host clock is shared with the daemon; RFC3339Nano is accepted by
	// `docker compose logs --since`.
	since := time.Now().UTC().Format(time.RFC3339Nano)
	if err := writeLogsSince(composeFile, since); err != nil {
		return fmt.Errorf("record clear marker: %w", err)
	}
	return nil
}

func logsSinceFile(composeFile string) string {
	return filepath.Join(filepath.Dir(composeFile), "logs-since")
}

func readLogsSince(composeFile string) string {
	b, err := os.ReadFile(logsSinceFile(composeFile))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func writeLogsSince(composeFile, ts string) error {
	return os.WriteFile(logsSinceFile(composeFile), []byte(ts), 0o600)
}

// PrepareVolumes makes each of an app's named volumes writable by the user its
// service actually runs as. Docker creates a fresh named volume owned by root
// whenever the mount path is absent from the image; a service that runs as a
// non-root user then gets EACCES on first write (e.g. an app whose image has no
// pre-created /data dir). For every service that runs as a non-root uid and
// mounts an EMPTY read-write named volume, chown the volume to that uid:gid so
// the app can write it — removing the need for per-app `user: "0:0"` overrides.
//
// Empty-only: an already-populated volume is left untouched, so we never restamp
// ownership of live data. Best-effort: any docker hiccup is logged-and-skipped,
// leaving the prior behaviour. The image, volume name and uid all come from the
// rendered compose (trusted catalog data) / `id` output (numeric), never from
// request-time input, and are passed as exec args (no shell), so there is no
// injection sink here.
func PrepareVolumes(appID, composeFile string, out io.Writer) {
	cfg, err := output("docker", "compose", "-p", project(appID), "-f", composeFile,
		"config", "--format", "json")
	if err != nil {
		return
	}
	var doc composeConfig
	if json.Unmarshal([]byte(cfg), &doc) != nil {
		return
	}
	for _, svc := range doc.Services {
		uid, gid := "", ""
		resolved := false
		for _, v := range svc.Volumes {
			if v.Type != "volume" || v.ReadOnly {
				continue
			}
			vol := resolveVolName(v.Source, doc.Volumes)
			if vol == "" || !volumeEmpty(svc.Image, vol) {
				continue
			}
			if !resolved {
				uid, gid = resolveServiceUser(svc.Image, svc.User)
				resolved = true
			}
			if !isUID(uid) || uid == "0" {
				break // root or unresolved → the volume is already writable
			}
			if err := chownVolume(svc.Image, vol, uid, gid); err == nil {
				fmt.Fprintf(out, "--> prepared volume %s (owner %s:%s)\n", vol, uid, gid)
			}
		}
	}
}

type composeConfig struct {
	Services map[string]composeService `json:"services"`
	Volumes  map[string]composeVolume  `json:"volumes"`
}

type composeService struct {
	Image   string         `json:"image"`
	User    string         `json:"user"`
	Volumes []composeMount `json:"volumes"`
}

type composeMount struct {
	Type     string `json:"type"`
	Source   string `json:"source"`
	ReadOnly bool   `json:"read_only"`
}

type composeVolume struct {
	Name string `json:"name"`
}

// resolveVolName maps a service mount's short volume name to its real docker
// volume name via the compose top-level `volumes:` map (falling back to the
// short name).
func resolveVolName(short string, top map[string]composeVolume) string {
	if v, ok := top[short]; ok && v.Name != "" {
		return v.Name
	}
	return short
}

// resolveServiceUser returns the numeric uid:gid a service runs as: an explicit
// numeric compose `user:` wins, otherwise the image's built-in default (probed
// with `id`). Returns ("","") when it can't be determined numerically.
func resolveServiceUser(image, userField string) (string, string) {
	if userField != "" {
		u, g, _ := strings.Cut(userField, ":")
		if isUID(u) {
			if !isUID(g) {
				g = u
			}
			return u, g
		}
		return "", "" // a named user in compose — can't map safely, skip
	}
	u, err := output("docker", "run", "--rm", "--entrypoint", "id", image, "-u")
	if err != nil {
		return "", ""
	}
	g, err := output("docker", "run", "--rm", "--entrypoint", "id", image, "-g")
	if err != nil {
		return strings.TrimSpace(u), strings.TrimSpace(u)
	}
	return strings.TrimSpace(u), strings.TrimSpace(g)
}

// volumeEmpty reports whether a named volume has no entries. The probe runs as
// root so directory permissions never hide content; on any error it returns
// false so a volume we can't read is left untouched.
func volumeEmpty(image, vol string) bool {
	o, err := output("docker", "run", "--rm", "-u", "0:0",
		"-v", vol+":/v", "--entrypoint", "ls", image, "-A", "/v")
	if err != nil {
		return false
	}
	return strings.TrimSpace(o) == ""
}

func chownVolume(image, vol, uid, gid string) error {
	spec := uid
	if isUID(gid) {
		spec = uid + ":" + gid
	}
	_, err := output("docker", "run", "--rm", "-u", "0:0",
		"-v", vol+":/v", "--entrypoint", "chown", image, "-R", spec, "/v")
	return err
}

func isUID(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
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
	return runStreamedEnv(w, nil, name, args...)
}

// runStreamedEnv is runStreamed with extra environment variables appended to the
// inherited environment.
func runStreamedEnv(w io.Writer, extraEnv []string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	if len(extraEnv) > 0 {
		cmd.Env = append(os.Environ(), extraEnv...)
	}
	cmd.Stdout = w
	cmd.Stderr = w
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return nil
}
