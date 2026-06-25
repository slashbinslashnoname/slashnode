"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

export function UpdateAppButton({
  id,
  from,
  to,
}: {
  id: string;
  from?: string;
  to: string;
}) {
  const router = useRouter();
  const [busy, setBusy] = useState(false);

  async function update() {
    setBusy(true);
    await fetch(`/api/apps/${id}/update`, { method: "POST" });
    setBusy(false);
    router.refresh();
  }

  return (
    <div className="mb-6 flex items-center justify-between gap-4 rounded-lg border border-primary/40 bg-primary/10 p-4 text-sm">
      <span>
        Update available{from ? `: v${from}` : ""} → <b>v{to}</b>
      </span>
      <button
        onClick={update}
        disabled={busy}
        className="rounded-lg bg-primary px-4 py-2 font-semibold text-white disabled:opacity-60"
      >
        {busy ? "Updating…" : "Update"}
      </button>
    </div>
  );
}
