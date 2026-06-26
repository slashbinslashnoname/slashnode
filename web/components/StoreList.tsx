"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import type { App } from "@/lib/api";

// StoreList renders the searchable app grid: a search box filters by app name,
// description and category as you type.
const catsOf = (a: App): string[] => a.category ?? [];

export function StoreList({ apps }: { apps: App[] }) {
  const [q, setQ] = useState("");
  const [showRemoved, setShowRemoved] = useState(false);
  // Selected category filters (multi-select, OR semantics). Empty = all.
  const [activeCats, setActiveCats] = useState<string[]>([]);
  const query = q.trim().toLowerCase();
  // Bitcoin Knots is intentionally not offered.
  const isKnots = /\bknots\b/.test(query);
  const removedCount = apps.filter((a) => a.hidden).length;

  // All categories present in the catalog, sorted, for the filter chips.
  const allCats = [...new Set(apps.flatMap(catsOf))].sort();
  const toggleCat = (c: string) =>
    setActiveCats((cur) => (cur.includes(c) ? cur.filter((x) => x !== c) : [...cur, c]));

  const matched = query
    ? apps.filter((a) =>
        [a.name, a.description ?? "", ...catsOf(a)]
          .join(" ")
          .toLowerCase()
          .includes(query),
      )
    : apps;
  // Apps removed from the store are hidden unless "show removed" is on; they stay
  // out of the catalog while their already-installed instances keep running.
  const shown = matched
    .filter((a) => !a.hidden || showRemoved)
    .filter((a) => activeCats.length === 0 || catsOf(a).some((c) => activeCats.includes(c)));
  // Installed apps float to the top; name order (from the API) is preserved
  // within each group since Array.sort is stable.
  const filtered = [...shown].sort(
    (a, b) => Number(!!b.installed) - Number(!!a.installed),
  );

  return (
    <div className="flex flex-col gap-6">
      <input
        type="search"
        value={q}
        onChange={(e) => setQ(e.target.value)}
        placeholder="Search apps…"
        autoComplete="off"
        className="w-full rounded-lg border border-border bg-bg px-4 py-2.5 text-sm outline-none focus:border-primary"
      />

      {allCats.length > 1 && (
        <div className="flex flex-wrap gap-2">
          <CatChip active={activeCats.length === 0} onClick={() => setActiveCats([])}>
            all
          </CatChip>
          {allCats.map((c) => (
            <CatChip key={c} active={activeCats.includes(c)} onClick={() => toggleCat(c)}>
              {c}
            </CatChip>
          ))}
        </div>
      )}

      {removedCount > 0 && (
        <button
          onClick={() => setShowRemoved((v) => !v)}
          className="-mt-2 w-fit cursor-pointer text-xs text-muted hover:text-primary"
        >
          {showRemoved ? "hide" : "show"} {removedCount} removed app{removedCount > 1 ? "s" : ""}
        </button>
      )}

      {isKnots ? (
        <p className="rounded-lg border border-primary/40 bg-primary/10 px-4 py-3 text-sm font-medium text-primary">
          We do not provide softwares that can harm Bitcoin.
        </p>
      ) : filtered.length === 0 ? (
        <p className="text-muted">No apps match “{q}”.</p>
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {filtered.map((app) => (
            <AppCard key={app.id} app={app} />
          ))}
        </div>
      )}
    </div>
  );
}

function CatChip({
  active,
  onClick,
  children,
}: {
  active: boolean;
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <button
      onClick={onClick}
      className={`cursor-pointer rounded-full border px-3 py-1 text-xs transition-colors ${
        active
          ? "border-primary bg-primary text-white"
          : "border-border text-muted hover:border-primary hover:text-fg"
      }`}
    >
      {children}
    </button>
  );
}

function AppCard({ app }: { app: App }) {
  const router = useRouter();
  const [busy, setBusy] = useState(false);

  async function toggleStore(e: React.MouseEvent) {
    e.preventDefault();
    e.stopPropagation();
    setBusy(true);
    await fetch(`/api/apps/${app.id}/${app.hidden ? "unhide" : "hide"}`, { method: "POST" });
    setBusy(false);
    router.refresh();
  }

  return (
    <Link
      href={`/store/${app.id}`}
      className={`group relative flex flex-col gap-3 rounded-xl border border-border bg-card p-5 transition-colors hover:border-primary ${
        app.hidden ? "opacity-60" : ""
      }`}
    >
      <div className="flex items-center gap-3">
        <span className="text-3xl">{app.icon ?? "📦"}</span>
        <div>
          <div className="font-semibold">{app.name}</div>
          <div className="text-xs text-muted">
            {catsOf(app).join(", ")} · v{app.version}
          </div>
        </div>
        {app.update_available ? (
          <span className="ml-auto rounded-full bg-primary px-2 py-0.5 text-xs font-semibold text-white">
            update
          </span>
        ) : app.installed ? (
          <span className="ml-auto rounded-full bg-primary/15 px-2 py-0.5 text-xs font-semibold text-primary">
            installed
          </span>
        ) : null}
      </div>
      {app.description && <p className="text-sm text-muted">{app.description}</p>}
      <button
        onClick={toggleStore}
        disabled={busy}
        title={app.hidden ? "Restore to the store" : "Remove from the store (keeps installed instances)"}
        className="w-fit cursor-pointer text-xs text-muted hover:text-primary disabled:opacity-50"
      >
        {busy ? "…" : app.hidden ? "↩ restore to store" : "✕ remove from store"}
      </button>
    </Link>
  );
}
