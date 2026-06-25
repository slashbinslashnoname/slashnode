"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

export function LoginForm({ next }: { next: string }) {
  const router = useRouter();
  const [password, setPassword] = useState("");
  const [state, setState] = useState<"idle" | "checking" | "error">("idle");

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setState("checking");
    try {
      const res = await fetch("/api/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ password }),
      });
      if (!res.ok) {
        setState("error");
        return;
      }
      router.push(next || "/");
      router.refresh();
    } catch {
      setState("error");
    }
  }

  return (
    <form onSubmit={submit} className="flex w-full max-w-xs flex-col gap-3">
      <input
        type="password"
        autoFocus
        placeholder="admin password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        className="rounded-lg border border-border bg-bg px-3 py-2 outline-none focus:border-primary"
      />
      {state === "error" && (
        <p className="text-sm text-primary">Invalid password.</p>
      )}
      <button
        type="submit"
        disabled={state === "checking"}
        className="rounded-lg bg-primary px-5 py-2.5 font-semibold text-white disabled:opacity-60"
      >
        {state === "checking" ? "Checking…" : "Unlock"}
      </button>
    </form>
  );
}
