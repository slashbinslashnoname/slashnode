"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

export function UninstallButton({ id, name }: { id: string; name: string }) {
  const router = useRouter();
  const [confirming, setConfirming] = useState(false);
  const [busy, setBusy] = useState<"" | "keep" | "purge">("");
  const [err, setErr] = useState("");

  async function uninstall(purge: boolean) {
    setErr("");
    setBusy(purge ? "purge" : "keep");
    try {
      const res = await fetch(`/api/apps/${id}/uninstall?purge=${purge}`, {
        method: "POST",
      });
      if (!res.ok) {
        const j = await res.json().catch(() => ({}));
        setErr(j.error || "uninstall failed");
        setBusy("");
        return;
      }
      setConfirming(false);
      router.push("/store");
      router.refresh();
    } catch {
      setErr("daemon unreachable");
      setBusy("");
    }
  }

  if (!confirming) {
    return (
      <button
        onClick={() => {
          setConfirming(true);
          setErr("");
        }}
        className="cursor-pointer rounded-lg border border-border px-4 py-2 text-sm text-muted hover:border-primary hover:text-primary"
      >
        Uninstall
      </button>
    );
  }

  const working = busy !== "";

  return (
    <div className="flex flex-col gap-2 rounded-lg border border-primary/40 bg-primary/10 p-4 text-sm">
      <span>Remove {name}?</span>
      {err && <span className="text-primary">{err}</span>}
      <div className="flex flex-wrap gap-2">
        <button
          onClick={() => uninstall(false)}
          disabled={working}
          className="cursor-pointer rounded-lg border border-border px-3 py-2 hover:border-primary disabled:cursor-default disabled:opacity-60"
        >
          {busy === "keep" ? "removing…" : "remove (keep data)"}
        </button>
        <button
          onClick={() => uninstall(true)}
          disabled={working}
          className="cursor-pointer rounded-lg bg-primary px-3 py-2 font-semibold text-white disabled:cursor-default disabled:opacity-60"
        >
          {busy === "purge" ? "removing…" : "remove + delete data"}
        </button>
        <button
          onClick={() => setConfirming(false)}
          disabled={working}
          className="cursor-pointer rounded-lg px-3 py-2 text-muted hover:text-fg disabled:cursor-default disabled:opacity-60"
        >
          cancel
        </button>
      </div>
    </div>
  );
}
