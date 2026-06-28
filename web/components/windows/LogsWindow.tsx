"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { FloatingWindow } from "./FloatingWindow";

// LogsWindow is a floating panel showing an app's container logs, with refresh
// and clear, mirroring the console window's behaviour.
export function LogsWindow({
  id,
  name,
  index,
  onClose,
}: {
  id: string;
  name: string;
  index: number;
  onClose: () => void;
}) {
  const [logs, setLogs] = useState("loading…");
  const [query, setQuery] = useState("");
  const [refreshing, setRefreshing] = useState(false);
  const [clearing, setClearing] = useState(false);
  const preRef = useRef<HTMLPreElement>(null);

  // Filter the displayed lines by the search query (case-insensitive). The full
  // log text stays in state, so clearing the query restores everything.
  const shown = query
    ? logs
        .split("\n")
        .filter((l) => l.toLowerCase().includes(query.toLowerCase()))
        .join("\n") || "(no matching lines)"
    : logs;

  const load = useCallback(async () => {
    try {
      const j = await fetch(`/api/apps/${id}/logs?tail=300`, {
        cache: "no-store",
      }).then((r) => r.json());
      setLogs(j.logs || "(no logs)");
    } catch {
      setLogs("(failed to load logs)");
    }
  }, [id]);

  useEffect(() => {
    load();
    const t = setInterval(load, 5000);
    return () => clearInterval(t);
  }, [load]);

  useEffect(() => {
    if (preRef.current) preRef.current.scrollTop = preRef.current.scrollHeight;
  }, [logs]);

  // Manual refresh: give a brief visible "refreshing…" state so the button
  // clearly does something even when the 5s auto-poll already keeps it current.
  async function refresh() {
    setRefreshing(true);
    await load();
    setRefreshing(false);
  }

  async function clear() {
    setClearing(true);
    // Blank the view immediately for instant feedback. The server records a
    // "since" marker, so the reload returns only entries logged after now — for
    // a chatty app a few fresh lines stream straight back, which is correct.
    setLogs("");
    setQuery("");
    const r = await fetch(`/api/apps/${id}/clear-logs`, { method: "POST" });
    if (!r.ok) {
      const j = await r.json().catch(() => ({}));
      setLogs(`[clear failed: ${j.error || "error"}]`);
      setClearing(false);
      return;
    }
    await load();
    setClearing(false);
  }

  return (
    <FloatingWindow
      index={index}
      onClose={onClose}
      title={
        <span>
          <span className="text-primary">▤</span> {name} · logs
        </span>
      }
    >
      <div className="flex h-full flex-col gap-2">
        <div className="flex items-center gap-2">
          <button
            onClick={refresh}
            disabled={refreshing}
            className="cursor-pointer text-xs text-muted hover:text-primary disabled:opacity-60"
          >
            {refreshing ? "refreshing…" : "refresh"}
          </button>
          <button
            onClick={clear}
            disabled={clearing}
            className="cursor-pointer text-xs text-muted hover:text-primary disabled:opacity-60"
          >
            {clearing ? "clearing…" : "clear logs"}
          </button>
          <input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="search…"
            className="ml-auto w-32 rounded border border-border bg-bg px-2 py-0.5 text-xs outline-none focus:border-primary"
          />
          {query && (
            <button
              onClick={() => setQuery("")}
              className="cursor-pointer text-xs text-muted hover:text-primary"
              aria-label="clear search"
            >
              ✕
            </button>
          )}
        </div>
        <pre
          ref={preRef}
          className="min-h-0 flex-1 overflow-auto rounded-lg bg-bg p-3 text-xs leading-relaxed"
        >
          {shown}
        </pre>
      </div>
    </FloatingWindow>
  );
}
