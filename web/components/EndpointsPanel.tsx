"use client";

import { useEffect, useState } from "react";
import type { AppEndpoint } from "@/lib/api";

// EndpointsPanel renders an app's connection URLs/addresses, built from the node
// host (the address the browser is using) + each endpoint's port. http/https
// endpoints get an open link; others (tcp, p2p, game servers) are shown as a
// copyable host:port address.
export function EndpointsPanel({ endpoints }: { endpoints: AppEndpoint[] }) {
  const [host, setHost] = useState("");

  useEffect(() => {
    setHost(location.hostname);
  }, []);

  if (!endpoints.length) return null;

  return (
    <div className="flex flex-col gap-1.5 rounded-lg border border-border bg-card p-3 text-sm">
      <div className="mb-1 text-xs font-semibold text-muted">Endpoints</div>
      {endpoints.map((e) => {
        const isWeb = e.scheme === "http" || e.scheme === "https";
        const value = isWeb
          ? `${e.scheme}://${host}:${e.port}${e.path ?? ""}`
          : `${host}:${e.port}`;
        return (
          <div
            key={e.label + e.port}
            className="flex items-center justify-between gap-3 text-xs"
          >
            <span className="text-muted">{e.label}</span>
            <span className="flex items-center gap-2">
              <code className="break-all text-fg">{host ? value : "…"}</code>
              <button
                onClick={() => navigator.clipboard?.writeText(value)}
                className="text-muted hover:text-primary"
                aria-label="Copy"
              >
                ⧉
              </button>
              {isWeb && host && (
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
          </div>
        );
      })}
    </div>
  );
}
