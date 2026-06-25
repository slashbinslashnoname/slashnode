package apps

import (
	"bytes"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var probeClient = &http.Client{Timeout: 5 * time.Second}

// Stat is one metadata-driven display value (label + formatted value). What to
// show is declared by the manifest's probe.display, never hardcoded here.
type Stat struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// ProbeResult is what an app is currently doing, surfaced to the frontend.
type ProbeResult struct {
	Type   string `json:"type"`             // none | http | rpc | electrum | lnd
	OK     bool   `json:"ok"`               // probe succeeded
	Detail string `json:"detail,omitempty"` // human summary / error
	Stats  []Stat `json:"stats,omitempty"`  // display values from probe.display
}

// RunProbe runs the manifest's probe and extracts the declared display fields.
func RunProbe(dir, id string) (*ProbeResult, error) {
	man, err := Find(dir, id)
	if err != nil {
		return nil, err
	}
	if man.Probe == nil {
		return &ProbeResult{Type: "none"}, nil
	}
	p := man.Probe

	var (
		data   map[string]any
		ok     bool
		detail string
	)
	switch p.Type {
	case "http":
		ok, detail = httpProbe(p)
	case "tcp":
		ok, detail = tcpProbe(p)
	case "rpc":
		data, ok, detail = rpcProbe(id, p)
	case "electrum":
		data, ok, detail = electrumProbe(p)
	case "lnd":
		data, ok, detail = lndProbe(id, p)
	default:
		return &ProbeResult{Type: p.Type, OK: false, Detail: "unsupported probe type"}, nil
	}

	res := &ProbeResult{Type: p.Type, OK: ok, Detail: detail}
	if ok {
		res.Stats = buildStats(p.Display, data)
	}
	return res, nil
}

// buildStats maps the manifest's display declarations onto the probe data.
func buildStats(display []ProbeStat, data map[string]any) []Stat {
	var stats []Stat
	for _, d := range display {
		key := d.Field
		if key == "" {
			key = "_scalar"
		}
		v, present := data[key]
		if !present {
			continue
		}
		stats = append(stats, Stat{Label: d.Label, Value: formatVal(v, d.Hex)})
	}
	return stats
}

func tcpProbe(p *Probe) (bool, string) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", p.Port), 4*time.Second)
	if err != nil {
		return false, err.Error()
	}
	_ = conn.Close()
	return true, "reachable"
}

func httpProbe(p *Probe) (bool, string) {
	path := p.Path
	if path == "" {
		path = "/"
	}
	resp, err := probeClient.Get(fmt.Sprintf("http://127.0.0.1:%d%s", p.Port, path))
	if err != nil {
		return false, err.Error()
	}
	defer resp.Body.Close()
	return resp.StatusCode < 500, fmt.Sprintf("HTTP %d", resp.StatusCode)
}

// rpcProbe issues a JSON-RPC call and returns its result as a map. A scalar
// result (e.g. geth eth_blockNumber) is exposed under the "_scalar" key.
func rpcProbe(id string, p *Probe) (map[string]any, bool, string) {
	user, pass := "", ""
	if p.UserInput != "" {
		user = LoadState().Installed[id].Inputs[p.UserInput]
	}
	if p.PassSecret != "" {
		pass = loadAppSecrets(id)[p.PassSecret]
	}

	reqBody, _ := json.Marshal(map[string]any{
		"jsonrpc": "1.0", "id": "slashnode", "method": p.Method, "params": []any{},
	})
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://127.0.0.1:%d/", p.Port), bytes.NewReader(reqBody))
	if err != nil {
		return nil, false, err.Error()
	}
	if user != "" || pass != "" {
		req.SetBasicAuth(user, pass)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := probeClient.Do(req)
	if err != nil {
		return nil, false, err.Error()
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var rpcResp struct {
		Result json.RawMessage `json:"result"`
		Error  any             `json:"error"`
	}
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return nil, false, fmt.Sprintf("HTTP %d", resp.StatusCode)
	}
	if rpcResp.Error != nil {
		return nil, false, fmt.Sprintf("%v", rpcResp.Error)
	}

	var obj map[string]any
	if json.Unmarshal(rpcResp.Result, &obj) == nil {
		return obj, true, p.Method
	}
	var scalar any
	if json.Unmarshal(rpcResp.Result, &scalar) == nil {
		return map[string]any{"_scalar": scalar}, true, p.Method
	}
	return map[string]any{}, true, p.Method
}

// electrumProbe asks an Electrum server (electrs) for the current tip.
func electrumProbe(p *Probe) (map[string]any, bool, string) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", p.Port), 4*time.Second)
	if err != nil {
		return nil, false, err.Error()
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(4 * time.Second))

	if _, err := conn.Write([]byte(`{"id":1,"method":"blockchain.headers.subscribe","params":[]}` + "\n")); err != nil {
		return nil, false, err.Error()
	}
	line, err := readLine(conn)
	if err != nil {
		return nil, false, err.Error()
	}
	var er struct {
		Result map[string]any `json:"result"`
	}
	if json.Unmarshal([]byte(line), &er) != nil {
		return nil, false, "bad response"
	}
	return er.Result, true, "headers.subscribe"
}

// lndProbe calls LND REST /v1/getinfo using the admin macaroon from the app's
// Docker volume.
func lndProbe(id string, p *Probe) (map[string]any, bool, string) {
	macaroon, err := readLNDMacaroon(id)
	if err != nil {
		return nil, false, err.Error()
	}
	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("https://127.0.0.1:%d/v1/getinfo", p.Port), nil)
	req.Header.Set("Grpc-Metadata-macaroon", macaroon)
	resp, err := client.Do(req)
	if err != nil {
		return nil, false, err.Error()
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var obj map[string]any
	if json.Unmarshal(body, &obj) != nil {
		return nil, false, fmt.Sprintf("HTTP %d", resp.StatusCode)
	}
	return obj, true, "getinfo"
}

func readLNDMacaroon(id string) (string, error) {
	base := fmt.Sprintf("/var/lib/docker/volumes/slashnode-%s_%s_%s-data/_data", id, id, id)
	b, err := os.ReadFile(base + "/data/chain/bitcoin/mainnet/admin.macaroon")
	if err != nil {
		return "", fmt.Errorf("macaroon not readable yet: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func formatVal(v any, isHex bool) string {
	if isHex {
		if n, ok := toInt64(v, true); ok {
			return strconv.FormatInt(n, 10)
		}
	}
	switch x := v.(type) {
	case float64:
		if x == float64(int64(x)) {
			return strconv.FormatInt(int64(x), 10)
		}
		return strconv.FormatFloat(x, 'f', -1, 64)
	case bool:
		if x {
			return "yes"
		}
		return "no"
	default:
		return fmt.Sprint(v)
	}
}

func toInt64(v any, isHex bool) (int64, bool) {
	switch x := v.(type) {
	case float64:
		return int64(x), true
	case string:
		s := strings.TrimSpace(x)
		if isHex {
			s = strings.TrimPrefix(s, "0x")
			n, err := strconv.ParseInt(s, 16, 64)
			return n, err == nil
		}
		n, err := strconv.ParseInt(s, 10, 64)
		return n, err == nil
	}
	return 0, false
}

func readLine(conn net.Conn) (string, error) {
	var b strings.Builder
	buf := make([]byte, 1)
	for {
		n, err := conn.Read(buf)
		if n > 0 {
			if buf[0] == '\n' {
				return b.String(), nil
			}
			b.WriteByte(buf[0])
		}
		if err != nil {
			if b.Len() > 0 {
				return b.String(), nil
			}
			return "", err
		}
	}
}
