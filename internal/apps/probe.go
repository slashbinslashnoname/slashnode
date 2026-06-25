package apps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var probeClient = &http.Client{Timeout: 5 * time.Second}

// ProbeResult is what an app is currently doing, surfaced to the frontend.
type ProbeResult struct {
	Type   string         `json:"type"`             // none | http | rpc
	OK     bool           `json:"ok"`               // probe succeeded
	Detail string         `json:"detail,omitempty"` // human summary / error
	Result map[string]any `json:"result,omitempty"` // rpc result fields
}

// RunProbe runs the manifest's probe (if any) against the running app and
// reports what it sees: for http, reachability; for rpc, the call result (e.g.
// bitcoind getblockchaininfo → block tip).
func RunProbe(dir, id string) (*ProbeResult, error) {
	man, err := Find(dir, id)
	if err != nil {
		return nil, err
	}
	if man.Probe == nil {
		return &ProbeResult{Type: "none"}, nil
	}
	p := man.Probe

	switch p.Type {
	case "http":
		path := p.Path
		if path == "" {
			path = "/"
		}
		url := fmt.Sprintf("http://127.0.0.1:%d%s", p.Port, path)
		resp, err := probeClient.Get(url)
		if err != nil {
			return &ProbeResult{Type: "http", OK: false, Detail: err.Error()}, nil
		}
		defer resp.Body.Close()
		return &ProbeResult{
			Type:   "http",
			OK:     resp.StatusCode < 500,
			Detail: fmt.Sprintf("HTTP %d", resp.StatusCode),
		}, nil

	case "rpc":
		return rpcProbe(id, p), nil

	default:
		return &ProbeResult{Type: p.Type, OK: false, Detail: "unsupported probe type"}, nil
	}
}

// rpcProbe issues a Bitcoin-style JSON-RPC call using the app's stored
// credentials and returns the result.
func rpcProbe(id string, p *Probe) *ProbeResult {
	user := ""
	if p.UserInput != "" {
		user = LoadState().Installed[id].Inputs[p.UserInput]
	}
	pass := ""
	if p.PassSecret != "" {
		pass = loadAppSecrets(id)[p.PassSecret]
	}

	reqBody, _ := json.Marshal(map[string]any{
		"jsonrpc": "1.0",
		"id":      "slashnode",
		"method":  p.Method,
		"params":  []any{},
	})
	url := fmt.Sprintf("http://127.0.0.1:%d/", p.Port)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return &ProbeResult{Type: "rpc", OK: false, Detail: err.Error()}
	}
	req.SetBasicAuth(user, pass)
	req.Header.Set("Content-Type", "application/json")

	resp, err := probeClient.Do(req)
	if err != nil {
		return &ProbeResult{Type: "rpc", OK: false, Detail: err.Error()}
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var rpcResp struct {
		Result map[string]any `json:"result"`
		Error  any            `json:"error"`
	}
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return &ProbeResult{Type: "rpc", OK: false, Detail: fmt.Sprintf("HTTP %d", resp.StatusCode)}
	}
	if rpcResp.Error != nil {
		return &ProbeResult{Type: "rpc", OK: false, Detail: fmt.Sprintf("%v", rpcResp.Error)}
	}
	return &ProbeResult{
		Type:   "rpc",
		OK:     true,
		Detail: p.Method,
		Result: rpcResp.Result,
	}
}
