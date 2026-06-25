"use client";

import { useState } from "react";

export function VersionBadge({
  version,
  available,
  latest,
}: {
  version: string;
  available: boolean;
  latest: string;
}) {
  const [state, setState] = useState<
    "idle" | "updating" | "verifying" | "done" | "error"
  >("idle");

  // Polls the status endpoint until the reported version moves off the current
  // one (the daemon restarts mid-update, so requests fail for a while). Confirms
  // the upgrade actually took effect instead of reloading on a blind timer.
  async function waitForNewVersion(): Promise<boolean> {
    const deadline = Date.now() + 120_000;
    while (Date.now() < deadline) {
      await new Promise((r) => setTimeout(r, 3000));
      try {
        const r = await fetch("/api/status", { cache: "no-store" });
        if (!r.ok) continue; // daemon still restarting
        const s = await r.json();
        if (s.version && s.version !== version) return true;
      } catch {
        // daemon down during restart — keep waiting
      }
    }
    return false;
  }

  async function update() {
    setState("updating");
    try {
      const r = await fetch("/api/update/apply", { method: "POST" });
      if (!r.ok) {
        setState("error");
        return;
      }
      setState("verifying");
      if (await waitForNewVersion()) {
        setState("done");
        setTimeout(() => location.reload(), 1500);
      } else {
        setState("error");
      }
    } catch {
      setState("error");
    }
  }

  return (
    <div className="fixed bottom-3 right-3 z-50 flex items-center gap-2 rounded-full border border-border bg-card px-3 py-1.5 text-xs text-muted shadow-lg">
      <span>v{version}</span>
      {available && state !== "error" && (
        <button
          onClick={update}
          disabled={state !== "idle"}
          className="rounded-full bg-primary px-2 py-0.5 font-semibold text-white disabled:opacity-60"
        >
          {state === "idle" && `update → ${latest}`}
          {state === "updating" && "updating…"}
          {state === "verifying" && "verifying…"}
          {state === "done" && "updated ✓"}
        </button>
      )}
      {state === "error" && (
        <button
          onClick={update}
          className="rounded-full bg-primary px-2 py-0.5 font-semibold text-white"
        >
          retry update
        </button>
      )}
    </div>
  );
}
