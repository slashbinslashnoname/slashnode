"use client";

import { useState } from "react";

export function UpdateBanner({
  current,
  latest,
}: {
  current: string;
  latest: string;
}) {
  const [state, setState] = useState<"idle" | "applying" | "done" | "error">(
    "idle",
  );

  async function apply() {
    setState("applying");
    try {
      const res = await fetch("/api/update/apply", { method: "POST" });
      setState(res.ok ? "done" : "error");
    } catch {
      setState("error");
    }
  }

  return (
    <div className="w-full max-w-md rounded-xl border border-primary/40 bg-primary/10 px-4 py-3">
      <div className="flex items-center justify-between gap-4">
        <div className="text-sm">
          <span className="font-semibold text-primary">
            Mise à jour disponible
          </span>
          <div className="text-muted">
            {current} → {latest}
          </div>
        </div>
        <button
          onClick={apply}
          disabled={state === "applying" || state === "done"}
          className="rounded-lg bg-primary px-3 py-1.5 text-sm font-semibold text-white disabled:opacity-60"
        >
          {state === "idle" && "Appliquer"}
          {state === "applying" && "Mise à jour…"}
          {state === "done" && "Redémarrage…"}
          {state === "error" && "Échec — réessayer"}
        </button>
      </div>
    </div>
  );
}
