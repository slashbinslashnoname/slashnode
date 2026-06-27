"use client";

import Link from "next/link";
import { useState } from "react";
import type { App } from "@/lib/api";

// StoreList renders the searchable app grid: a search box filters by app name,
// description and category as you type.
const catsOf = (a: App): string[] => a.category ?? [];

export function StoreList({ apps }: { apps: App[] }) {
  const [q, setQ] = useState("");
  // Selected category filters (multi-select, OR semantics). Empty = all.
  const [activeCats, setActiveCats] = useState<string[]>([]);
  const query = q.trim().toLowerCase();
  // Bitcoin Knots is intentionally not offered.
  const isKnots = /\bknots\b/.test(query);

  // Developer-hidden apps are never offered in the store.
  const visible = apps.filter((a) => !a.hidden);
  // All categories present in the catalog, sorted, for the filter chips.
  const allCats = [...new Set(visible.flatMap(catsOf))].sort();
  const toggleCat = (c: string) =>
    setActiveCats((cur) => (cur.includes(c) ? cur.filter((x) => x !== c) : [...cur, c]));

  const matched = query
    ? visible.filter((a) =>
        [a.name, a.description ?? "", ...catsOf(a)]
          .join(" ")
          .toLowerCase()
          .includes(query),
      )
    : visible;
  const shown = matched.filter(
    (a) => activeCats.length === 0 || catsOf(a).some((c) => activeCats.includes(c)),
  );
  // Ordering: explicit priority first (first-party apps rank above ported ones),
  // then installed, then the API's name order (preserved by the stable sort).
  const filtered = [...shown].sort(
    (a, b) =>
      (b.priority ?? 0) - (a.priority ?? 0) ||
      Number(!!b.installed) - Number(!!a.installed),
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
  return (
    <Link
      href={`/store/${app.id}`}
      className="group flex flex-col gap-3 rounded-xl border border-border bg-card p-5 transition-colors hover:border-primary"
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
    </Link>
  );
}
