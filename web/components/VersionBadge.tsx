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
  const [state, setState] = useState<"idle" | "updating" | "done" | "error">(
    "idle",
  );

  async function update() {
    setState("updating");
    try {
      const r = await fetch("/api/update/apply", { method: "POST" });
      if (!r.ok) {
        setState("error");
        return;
      }
      setState("done");
      // The daemon updates (binary + web + apps), reapplies and restarts;
      // reload once it's likely back up.
      setTimeout(() => location.reload(), 15000);
    } catch {
      setState("error");
    }
  }

  return (
    <div className="fixed bottom-3 right-3 z-50 flex items-center gap-2 rounded-full border border-border bg-card px-3 py-1.5 text-xs text-muted shadow-lg">
      <span>v{version}</span>
      {available && (
        <button
          onClick={update}
          disabled={state !== "idle"}
          className="rounded-full bg-primary px-2 py-0.5 font-semibold text-white disabled:opacity-60"
        >
          {state === "idle" && `update → ${latest}`}
          {state === "updating" && "updating…"}
          {state === "done" && "restarting…"}
          {state === "error" && "failed"}
        </button>
      )}
    </div>
  );
}
