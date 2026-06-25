"use client";

import { useState } from "react";

// SignOutButton clears the session cookie and returns to the login page. Only
// rendered when the UI is password protected.
export function SignOutButton() {
  const [busy, setBusy] = useState(false);

  async function signOut() {
    setBusy(true);
    try {
      await fetch("/api/logout", { method: "POST" });
    } catch {
      // ignore — redirect regardless
    }
    window.location.href = "/login";
  }

  return (
    <button
      onClick={signOut}
      disabled={busy}
      aria-label="Sign out"
      className="rounded-lg border border-border bg-card px-3 py-1.5 text-sm hover:border-primary transition-colors disabled:opacity-60"
    >
      {busy ? "…" : "⏻ sign out"}
    </button>
  );
}
