// Package tailscale connects the node to a Tailscale tailnet so two SlashNodes
// in different physical locations can reach each other privately (WireGuard,
// NAT-piercing) and back one up onto the other off-site.
//
// Like the backup package, the tailscale binary is never installed on the host:
// tailscaled runs from the official tailscale/tailscale Docker image in a
// long-lived container that shares the host network namespace (--network=host).
// That gives the host itself a stable 100.x tailnet address, so any service the
// host already exposes (the node's sshd for the peer-backup SFTP transport, the
// web UI…) becomes reachable from every other node on the tailnet — no port
// forwarding, no public IP.
package tailscale

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/slashbinslashnoname/slashnode/internal/orchestrator"
)

// image is the engine; pinned-ish to a moving tag (re-pulled on demand).
const image = "tailscale/tailscale:latest"

// containerName / stateVolume identify the long-lived tailscaled container and
// the volume that holds its node identity (so the node keeps the same tailnet
// address across restarts and upgrades).
const (
	containerName = "slashnode-tailscale"
	stateVolume   = "slashnode-tailscale"
)

// Peer is one machine on the tailnet (the node itself, or another node).
type Peer struct {
	Host   string `json:"host"`   // short hostname (MagicDNS label)
	DNS    string `json:"dns"`    // fully-qualified MagicDNS name
	IP     string `json:"ip"`     // first IPv4 tailnet address (100.x)
	OS     string `json:"os"`     // reported OS
	Online bool   `json:"online"` // currently reachable
}

// Status is the tailnet view reported to the UI.
type Status struct {
	Available bool   `json:"available"` // docker present and not disabled
	Running   bool   `json:"running"`   // the tailscaled container is up
	Backend   string `json:"backend"`   // BackendState: Running | NeedsLogin | Stopped | NoState
	Self      Peer   `json:"self"`
	Peers     []Peer `json:"peers"`
}

// Available reports whether Tailscale can run here (docker present and not
// disabled via SLASHNODE_NO_TAILSCALE).
func Available() bool {
	if os.Getenv("SLASHNODE_NO_TAILSCALE") != "" {
		return false
	}
	return orchestrator.Available()
}

// Up (re)starts the tailscaled container and brings the node onto the tailnet.
// authKey authenticates the node the first time (a Tailscale auth key); on
// subsequent calls it may be empty — the persisted state volume re-authenticates
// the node automatically. hostname is the machine name shown on the tailnet
// (optional; Tailscale derives one otherwise).
func Up(authKey, hostname string, out io.Writer) error {
	if !Available() {
		return fmt.Errorf("docker is not available")
	}
	// Replace any previous container so env changes (new auth key/hostname) take
	// effect; the identity lives in the state volume, which is preserved.
	_ = exec.Command("docker", "rm", "-f", containerName).Run()
	args := runArgs(authKey, hostname)
	fmt.Fprintf(out, "--> starting tailscale (%s)\n", image)
	return stream(out, "docker", args...)
}

// Down stops and removes the tailscaled container. The state volume is kept so
// re-enabling later reconnects the node under its existing tailnet identity.
func Down(out io.Writer) error {
	if !orchestrator.Available() {
		return fmt.Errorf("docker is not available")
	}
	fmt.Fprintf(out, "--> stopping tailscale\n")
	return stream(out, "docker", "rm", "-f", containerName)
}

// Reconcile ensures the tailscaled container is running when Tailscale is
// enabled in the node config, and absent when it is not. Called (best-effort) on
// serve startup so the tailnet connection survives reboots without an auth key
// (the state volume re-authenticates). enabled/hostname come from the config.
func Reconcile(enabled bool, hostname string, out io.Writer) {
	if !Available() {
		return
	}
	if !enabled {
		if running() {
			_ = Down(out)
		}
		return
	}
	if running() {
		return // already up; leave the live connection alone
	}
	// No auth key here: rely on the persisted state volume. On a node that was
	// never authenticated this simply parks in NeedsLogin until the operator
	// supplies a key via the UI/CLI.
	if err := Up("", hostname, out); err != nil {
		fmt.Fprintf(out, "    (tailscale reconcile: %v)\n", err)
	}
}

// GetStatus returns the tailnet view: whether tailscaled is running, the node's
// own address, and the reachable peers. Never errors — an unreachable backend
// yields a status with Running=false.
func GetStatus() Status {
	s := Status{Available: Available(), Running: running()}
	if !s.Running {
		return s
	}
	raw, err := capture("docker", "exec", containerName, "tailscale", "status", "--json")
	if err != nil {
		return s
	}
	var st tsStatus
	if json.Unmarshal([]byte(raw), &st) != nil {
		return s
	}
	s.Backend = st.BackendState
	if st.Self != nil {
		s.Self = st.Self.peer()
	}
	for _, p := range st.Peer {
		if p != nil {
			s.Peers = append(s.Peers, p.peer())
		}
	}
	return s
}

// --- tailscale status --json shape (only the fields we use) ----------------

type tsStatus struct {
	BackendState string             `json:"BackendState"`
	Self         *tsNode            `json:"Self"`
	Peer         map[string]*tsNode `json:"Peer"`
}

type tsNode struct {
	HostName     string   `json:"HostName"`
	DNSName      string   `json:"DNSName"`
	TailscaleIPs []string `json:"TailscaleIPs"`
	OS           string   `json:"OS"`
	Online       bool     `json:"Online"`
}

func (n *tsNode) peer() Peer {
	return Peer{
		Host:   n.HostName,
		DNS:    strings.TrimSuffix(n.DNSName, "."),
		IP:     firstIPv4(n.TailscaleIPs),
		OS:     n.OS,
		Online: n.Online,
	}
}

// --- helpers ---------------------------------------------------------------

// runArgs builds the `docker run` invocation for the tailscaled container. Host
// networking + NET_ADMIN + /dev/net/tun give the host a real tailnet interface;
// TS_ACCEPT_DNS=false keeps Tailscale from rewriting the host's /etc/resolv.conf
// (peers are addressed by their 100.x IP, so MagicDNS on the host is not needed).
func runArgs(authKey, hostname string) []string {
	args := []string{
		"run", "-d",
		"--name", containerName,
		"--restart", "unless-stopped",
		"--network", "host",
		"--cap-add", "NET_ADMIN",
		"--cap-add", "NET_RAW",
		"--device", "/dev/net/tun",
		"-v", stateVolume + ":/var/lib/tailscale",
		"-e", "TS_STATE_DIR=/var/lib/tailscale",
		"-e", "TS_USERSPACE=false",
		"-e", "TS_ACCEPT_DNS=false",
	}
	if authKey != "" {
		args = append(args, "-e", "TS_AUTHKEY="+authKey)
	}
	if hostname != "" {
		args = append(args, "-e", "TS_HOSTNAME="+hostname)
	}
	return append(args, image)
}

// running reports whether the tailscaled container exists and is running.
func running() bool {
	out, err := capture("docker", "inspect", "-f", "{{.State.Running}}", containerName)
	return err == nil && strings.TrimSpace(out) == "true"
}

func firstIPv4(ips []string) string {
	for _, ip := range ips {
		if !strings.Contains(ip, ":") {
			return ip
		}
	}
	if len(ips) > 0 {
		return ips[0]
	}
	return ""
}

// ValidHostname reports whether s is a valid tailnet machine name (a DNS label:
// letters, digits and hyphens, not starting/ending with a hyphen). Empty is
// valid — Tailscale derives a name.
func ValidHostname(s string) bool {
	if s == "" {
		return true
	}
	if len(s) > 63 || strings.HasPrefix(s, "-") || strings.HasSuffix(s, "-") {
		return false
	}
	for _, r := range s {
		ok := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-'
		if !ok {
			return false
		}
	}
	return true
}

func stream(out io.Writer, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = out
	cmd.Stderr = out
	return cmd.Run()
}

func capture(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var sb strings.Builder
	cmd.Stdout = &sb
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return sb.String(), nil
}
