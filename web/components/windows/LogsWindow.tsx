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
  const preRef = useRef<HTMLPreElement>(null);

  const load = useCallback(async () => {
    try {
      const j = await fetch(`/api/apps/${id}/logs?tail=300`).then((r) => r.json());
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

  async function clear() {
    const r = await fetch(`/api/apps/${id}/clear-logs`, { method: "POST" });
    if (!r.ok) {
      const j = await r.json().catch(() => ({}));
      setLogs((l) => l + `\n[clear failed: ${j.error || "error"}]`);
      return;
    }
    await load();
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
        <div className="flex gap-2">
          <button onClick={load} className="cursor-pointer text-xs text-muted hover:text-primary">
            refresh
          </button>
          <button onClick={clear} className="cursor-pointer text-xs text-muted hover:text-primary">
            clear logs
          </button>
        </div>
        <pre
          ref={preRef}
          className="min-h-0 flex-1 overflow-auto rounded-lg bg-bg p-3 text-xs leading-relaxed"
        >
          {logs}
        </pre>
      </div>
    </FloatingWindow>
  );
}
