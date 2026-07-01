"use client";

import { useEffect, useRef, useState } from "react";

const inputCls =
  "rounded-lg border border-border bg-bg px-3 py-2 text-sm outline-none focus:border-primary";
const btnPrimary =
  "cursor-pointer rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-white disabled:cursor-default disabled:opacity-60";
const btnGhost =
  "cursor-pointer rounded-lg border border-border px-3 py-2 text-sm hover:border-primary disabled:cursor-default disabled:opacity-60";

type Peer = {
  host: string;
  dns: string;
  ip: string;
  os: string;
  online: boolean;
};

type State = {
  enabled: boolean;
  hostname: string;
  status: {
    available: boolean;
    running: boolean;
    backend: string;
    self: Peer;
    peers: Peer[];
  };
};

// TailscalePanel joins the node to a Tailscale tailnet so two SlashNodes in
// different locations can reach each other privately, and one can back up onto
// the other off-site. The auth key is used once and never stored or echoed back.
export function TailscalePanel() {
  const [state, setState] = useState<State | null>(null);
  const [authkey, setAuthkey] = useState("");
  const [hostname, setHostname] = useState("");
  const [busy, setBusy] = useState<"" | "up" | "down">("");
  const [log, setLog] = useState("");
  const preRef = useRef<HTMLPreElement>(null);

  async function refresh() {
    try {
      const r = await fetch("/api/tailscale", { cache: "no-store" });
      setState(await r.json());
    } catch {
      /* daemon unreachable — leave prior state */
    }
  }

  useEffect(() => {
    refresh();
  }, []);

  useEffect(() => {
    if (preRef.current) preRef.current.scrollTop = preRef.current.scrollHeight;
  }, [log]);

  async function connect() {
    setBusy("up");
    setLog("");
    try {
      const res = await fetch("/api/tailscale/up", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ authkey, hostname }),
      });
      const reader = res.body?.getReader();
      const dec = new TextDecoder();
      while (reader) {
        const { done, value } = await reader.read();
        if (done) break;
        setLog((l) => l + dec.decode(value, { stream: true }));
      }
    } catch {
      setLog((l) => l + "\n[connection lost]");
    }
    setAuthkey("");
    setBusy("");
    refresh();
  }

  async function disconnect() {
    setBusy("down");
    try {
      await fetch("/api/tailscale/down", { method: "POST" });
    } catch {
      /* ignore */
    }
    setBusy("");
    refresh();
  }

  const st = state?.status;
  const running = !!st?.running;
  const self = st?.self;

  if (state && st && !st.available) {
    return (
      <p className="text-sm text-muted">
        Tailscale needs Docker, which isn’t available on this node.
      </p>
    );
  }

  return (
    <div className="flex flex-col gap-4">
      <p className="text-xs text-muted">
        Join a <b>Tailscale</b> tailnet to privately connect two SlashNodes in
        different locations (WireGuard, NAT-piercing — no port forwarding). Once
        both nodes are on the same tailnet, point this node’s backup at the other
        using the <b>“node (Tailscale peer)”</b> destination for real off-site,
        fully self-hosted backups on hardware you own. Get a reusable{" "}
        <b>auth key</b> from your Tailscale admin console. The key is used once to
        authenticate and is never stored on the node.
      </p>

      {/* Connection state */}
      <div className="rounded-lg border border-border bg-bg p-3 text-sm">
        {!state ? (
          <span className="text-muted">Loading…</span>
        ) : running && self?.ip ? (
          <div className="flex flex-col gap-1">
            <span>
              <span className="text-primary">●</span> Connected
              {st?.backend && st.backend !== "Running" ? ` (${st.backend})` : ""}
            </span>
            <span className="text-muted">
              This node: <code>{self.dns || self.host}</code> ·{" "}
              <code>{self.ip}</code>
            </span>
          </div>
        ) : running ? (
          <span className="text-muted">
            Starting… {st?.backend === "NeedsLogin" ? "supply an auth key below to authenticate." : ""}
          </span>
        ) : (
          <span className="text-muted">○ Not connected</span>
        )}
      </div>

      {/* Peers — the candidate backup targets */}
      {running && st && st.peers.length > 0 && (
        <div className="flex flex-col gap-1">
          <span className="text-sm font-medium">Machines on your tailnet</span>
          <div className="overflow-x-auto rounded-lg border border-border">
            <table className="w-full text-sm">
              <tbody>
                {st.peers.map((p) => (
                  <tr key={p.ip || p.host} className="border-b border-border last:border-0">
                    <td className="px-3 py-2">
                      <span className={p.online ? "text-primary" : "text-muted"}>●</span>{" "}
                      {p.host}
                    </td>
                    <td className="px-3 py-2 text-muted">{p.os}</td>
                    <td className="px-3 py-2">
                      <code>{p.ip}</code>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <span className="text-xs text-muted">
            Use a peer’s <code>100.x</code> address as the host in the backup
            panel’s “node (Tailscale peer)” destination.
          </span>
        </div>
      )}

      {/* Connect / disconnect */}
      {running ? (
        <div>
          <button onClick={disconnect} disabled={!!busy} className={btnGhost}>
            {busy === "down" ? "disconnecting…" : "Disconnect from tailnet"}
          </button>
        </div>
      ) : (
        <div className="flex flex-col gap-3">
          <Field label="Auth key">
            <input
              type="password"
              value={authkey}
              placeholder="tskey-auth-…"
              onChange={(e) => setAuthkey(e.target.value)}
              className={inputCls}
            />
          </Field>
          <Field label="Machine name (optional)">
            <input
              value={hostname}
              placeholder="slashnode-home"
              onChange={(e) => setHostname(e.target.value)}
              className={inputCls}
            />
          </Field>
          <div>
            <button onClick={connect} disabled={!!busy || !authkey} className={btnPrimary}>
              {busy === "up" ? "connecting…" : "Connect to tailnet"}
            </button>
          </div>
        </div>
      )}

      {log && (
        <pre ref={preRef} className="max-h-52 overflow-auto rounded-lg bg-bg p-3 text-xs leading-relaxed">
          {log}
        </pre>
      )}
    </div>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="flex flex-col gap-1">
      <span className="text-sm font-medium">{label}</span>
      {children}
    </label>
  );
}
