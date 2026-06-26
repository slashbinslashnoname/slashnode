"use client";

import { useEffect, useState } from "react";
import type { AppEndpoint } from "@/lib/api";
import { webClearnetUrl } from "@/lib/appUrl";

// EndpointsPanel renders an app's connection URLs/addresses: the clearnet
// address (node host the browser is using + port) and, when the app is exposed
// over Tor, the matching `.onion` address for each endpoint. An optional Web UI
// row surfaces the app's web interface (clearnet + onion) the same way. http(s)
// endpoints get an open link; others (tcp, p2p, game servers…) are copyable.
export function EndpointsPanel({
  endpoints,
  onion,
  web,
  url,
}: {
  endpoints: AppEndpoint[];
  onion?: string;
  web?: { port: number; path?: string };
  url?: string; // the app's reverse-proxy URL (https://<id>.<host>)
}) {
  const [host, setHost] = useState("");

  useEffect(() => {
    setHost(location.hostname);
  }, []);

  if (!endpoints.length && !web) return null;

  // The web UI follows the reverse-proxy convention (same as the tile's "open"
  // button). Its onion is the hidden service at :80 (Tor maps onion:80 → web
  // port), not the published web port.
  const webClear = web && host ? webClearnetUrl(url, web.port) : "";
  const webOnion = web && onion ? `http://${onion}` : "";

  return (
    <div className="flex flex-col gap-1.5 rounded-lg border border-border bg-card p-3 text-sm">
      <div className="mb-1 text-xs font-semibold text-muted">Endpoints</div>

      {web && (
        <Row label="Web UI">
          <Addr value={webClear || "…"} open={!!host} />
          {webOnion && <Addr value={webOnion} open onionTag />}
        </Row>
      )}

      {endpoints.map((e, i) => {
        // Endpoints are connection addresses (RPC, P2P, Electrum, S3…), reached
        // with a client — not browser pages — so they show as a bare host:port
        // with no scheme prefix and no "open" link (the Web UI row above is the
        // only browser-navigable URL).
        const suffix = e.path && e.path !== "/" ? e.path : "";
        const clear = `${host}:${e.port}${suffix}`;
        const onionAddr = onion ? `${onion}:${e.port}${suffix}` : "";
        return (
          <Row key={`${e.label}-${e.port}-${i}`} label={e.label}>
            <Addr value={host ? clear : "…"} />
            {onionAddr && <Addr value={onionAddr} onionTag />}
          </Row>
        );
      })}
    </div>
  );
}

function Row({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div className="flex items-start justify-between gap-3 text-xs">
      <span className="shrink-0 text-muted">{label}</span>
      <span className="flex flex-col items-end gap-1">{children}</span>
    </div>
  );
}

function Addr({
  value,
  open,
  onionTag,
}: {
  value: string;
  open?: boolean;
  onionTag?: boolean;
}) {
  return (
    <span className="flex items-center gap-2">
      {onionTag && <span className="text-[10px] text-muted">.onion</span>}
      <code className="break-all text-fg">{value}</code>
      <button
        onClick={() => navigator.clipboard?.writeText(value)}
        className="text-muted hover:text-primary"
        aria-label="Copy"
      >
        ⧉
      </button>
      {open && value !== "…" && (
        <a
          href={value}
          target="_blank"
          rel="noreferrer"
          className="text-muted hover:text-primary"
          aria-label="Open"
        >
          ↗
        </a>
      )}
    </span>
  );
}
