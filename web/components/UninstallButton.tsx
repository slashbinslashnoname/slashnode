"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

export function UninstallButton({ id, name }: { id: string; name: string }) {
  const router = useRouter();
  const [confirming, setConfirming] = useState(false);
  const [err, setErr] = useState("");

  async function uninstall(purge: boolean) {
    setErr("");
    const res = await fetch(`/api/apps/${id}/uninstall?purge=${purge}`, {
      method: "POST",
    });
    if (!res.ok) {
      const j = await res.json().catch(() => ({}));
      setErr(j.error || "uninstall failed");
      return;
    }
    setConfirming(false);
    router.push("/store");
    router.refresh();
  }

  if (!confirming) {
    return (
      <button
        onClick={() => {
          setConfirming(true);
          setErr("");
        }}
        className="rounded-lg border border-border px-4 py-2 text-sm text-muted hover:border-primary hover:text-primary"
      >
        Uninstall
      </button>
    );
  }

  return (
    <div className="flex flex-col gap-2 rounded-lg border border-primary/40 bg-primary/10 p-4 text-sm">
      <span>Remove {name}?</span>
      {err && <span className="text-primary">{err}</span>}
      <div className="flex flex-wrap gap-2">
        <button
          onClick={() => uninstall(false)}
          className="rounded-lg border border-border px-3 py-2 hover:border-primary"
        >
          remove (keep data)
        </button>
        <button
          onClick={() => uninstall(true)}
          className="rounded-lg bg-primary px-3 py-2 font-semibold text-white"
        >
          remove + delete data
        </button>
        <button
          onClick={() => setConfirming(false)}
          className="rounded-lg px-3 py-2 text-muted hover:text-fg"
        >
          cancel
        </button>
      </div>
    </div>
  );
}
