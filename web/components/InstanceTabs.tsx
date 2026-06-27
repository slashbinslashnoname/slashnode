"use client";

import Link from "next/link";
import type { App } from "@/lib/api";

// InstanceTabs sits above the Config/Version/Domain/View tabs and lets the
// operator switch between instances of the same app (SlashSlack, SlashSlack 2, …)
// or start a new one. Each instance is its own install (id "<base>-<n>"); the
// tabs just link to /store/<instanceId>. Only shown once an app is installed.
export function InstanceTabs({ app }: { app: App }) {
  const base = app.base_id ?? app.id;
  const instances = app.instances ?? [];
  if (!app.installed || instances.length === 0) return null;

  // Next free instance number for "+ run another instance".
  const used = new Set(
    instances.map((i) => (i.id === base ? 1 : Number(i.id.slice(base.length + 1)))),
  );
  let next = 2;
  while (used.has(next)) next++;
  const nextId = `${base}-${next}`;

  return (
    <div className="mb-4 flex flex-wrap items-center gap-2 border-b border-border pb-3">
      {instances.map((i) => (
        <Link
          key={i.id}
          href={`/store/${i.id}`}
          className={`rounded-md px-3 py-1.5 text-sm font-semibold ${
            i.id === app.id
              ? "bg-primary text-white"
              : "text-muted hover:text-fg"
          }`}
        >
          {i.name}
        </Link>
      ))}
      <Link
        href={`/store/${nextId}`}
        title="Install another independent instance of this app"
        className="rounded-md border border-dashed border-border px-3 py-1.5 text-sm text-muted hover:border-primary hover:text-primary"
      >
        + run another instance
      </Link>
    </div>
  );
}
