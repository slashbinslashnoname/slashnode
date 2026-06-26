"use client";

import Link from "next/link";
import { useState } from "react";
import type { App } from "@/lib/api";

// StoreList renders the searchable app grid: a search box filters by app name,
// description and category as you type.
export function StoreList({ apps }: { apps: App[] }) {
  const [q, setQ] = useState("");
  const query = q.trim().toLowerCase();
  const filtered = query
    ? apps.filter((a) =>
        [a.name, a.description ?? "", a.category]
          .join(" ")
          .toLowerCase()
          .includes(query),
      )
    : apps;

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

      {filtered.length === 0 ? (
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
            {app.category} · v{app.version}
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
