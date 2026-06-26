"use client";

import { useEffect, useState } from "react";
import type { Status } from "@/lib/api";

// NodeLinks is an icon button that opens a modal listing every way the node can
// be reached: the address you're currently on, the local mDNS URL and — when
// Tor is configured — the .onion address.
export function NodeLinks() {
  const [open, setOpen] = useState(false);
  const [status, setStatus] = useState<Status | null>(null);
  const [origin, setOrigin] = useState("");

  useEffect(() => setOrigin(location.origin), []);

  function toggle() {
    const next = !open;
    setOpen(next);
    if (next && !status) {
      fetch("/api/status")
        .then((r) => r.json())
        .then(setStatus)
        .catch(() => {});
    }
  }

  const links: { label: string; url: string; hint?: string }[] = [];
  if (origin) links.push({ label: "This address", url: origin });
  if (status?.hostname) {
    const mdns = `http://${status.hostname}:${status.port}`;
    if (mdns !== origin) links.push({ label: "Local network (mDNS)", url: mdns });
  }
  if (status?.onion) {
    links.push({
      label: "Tor",
      url: `http://${status.onion}`,
      hint: "open in Tor Browser",
    });
  }

  return (
    <>
      <button
        onClick={toggle}
        aria-label="Node URLs"
        title="Node URLs"
        className="cursor-pointer rounded-lg border border-border bg-card px-3 py-1.5 text-sm hover:border-primary transition-colors"
      >
        🔗
      </button>

      {open && (
        <div
          className="fixed inset-0 z-[60] flex items-start justify-center bg-black/50 p-4 pt-24"
          onClick={() => setOpen(false)}
        >
          <div
            className="w-full max-w-md rounded-xl border border-border bg-card p-5 shadow-xl"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="mb-3 flex items-center justify-between">
              <h2 className="text-sm font-semibold">Node URLs</h2>
              <button
                onClick={() => setOpen(false)}
                aria-label="Close"
                className="cursor-pointer text-muted hover:text-primary"
              >
                ✕
              </button>
            </div>

            {links.length === 0 ? (
              <p className="text-sm text-muted">Loading…</p>
            ) : (
              <ul className="flex flex-col gap-2">
                {links.map((l) => (
                  <li key={l.label} className="flex flex-col gap-1">
                    <div className="flex items-center justify-between gap-2 text-xs">
                      <span className="text-muted">
                        {l.label}
                        {l.hint && (
                          <span className="ml-1 text-[10px]">({l.hint})</span>
                        )}
                      </span>
                      <span className="flex items-center gap-2">
                        <button
                          onClick={() => navigator.clipboard?.writeText(l.url)}
                          className="cursor-pointer text-muted hover:text-primary"
                          aria-label="Copy"
                        >
                          ⧉
                        </button>
                        <a
                          href={l.url}
                          target="_blank"
                          rel="noreferrer"
                          className="cursor-pointer text-muted hover:text-primary"
                          aria-label="Open"
                        >
                          ↗
                        </a>
                      </span>
                    </div>
                    <code className="break-all rounded-md bg-bg px-2 py-1 text-xs text-fg">
                      {l.url}
                    </code>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </div>
      )}
    </>
  );
}
