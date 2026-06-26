"use client";

import { useState } from "react";

export function UpdateBanner({
  current,
  latest,
}: {
  current: string;
  latest: string;
}) {
  const [state, setState] = useState<
    "idle" | "applying" | "verifying" | "done" | "error"
  >("idle");

  // After the daemon accepts the update it restarts, so requests fail for a
  // while. Poll /status until the reported version moves off the current one,
  // confirming the upgrade actually took effect instead of spinning forever.
  async function waitForNewVersion(): Promise<boolean> {
    const deadline = Date.now() + 120_000;
    while (Date.now() < deadline) {
      await new Promise((r) => setTimeout(r, 3000));
      try {
        const r = await fetch("/api/status", { cache: "no-store" });
        if (!r.ok) continue; // daemon still restarting
        const s = await r.json();
        if (s.version && s.version !== current) return true;
      } catch {
        // daemon down during restart — keep waiting
      }
    }
    return false;
  }

  async function apply() {
    setState("applying");
    try {
      const res = await fetch("/api/update/apply", { method: "POST" });
      if (!res.ok) {
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
    <div className="w-full max-w-md rounded-xl border border-primary/40 bg-primary/10 px-4 py-3">
      <div className="flex items-center justify-between gap-4">
        <div className="text-sm">
          <span className="font-semibold text-primary">Update available</span>
          <div className="text-muted">
            {current} → {latest}
          </div>
        </div>
        <button
          onClick={apply}
          disabled={state === "applying" || state === "verifying" || state === "done"}
          className="rounded-lg bg-primary px-3 py-1.5 text-sm font-semibold text-white disabled:opacity-60"
        >
          {state === "idle" && "Apply"}
          {state === "applying" && "Updating…"}
          {state === "verifying" && "Restarting…"}
          {state === "done" && "Updated ✓"}
          {state === "error" && "Failed — retry"}
        </button>
      </div>
    </div>
  );
}
