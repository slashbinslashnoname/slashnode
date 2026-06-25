import Link from "next/link";
import { TopControls } from "@/components/TopControls";
import { getApps, type App } from "@/lib/api";

export const dynamic = "force-dynamic";

export default async function Store() {
  const data = await getApps();
  const apps = data?.apps ?? [];

  return (
    <main className="mx-auto min-h-screen w-full max-w-5xl px-4 py-10">
      <TopControls />

      <header className="mb-8 flex items-end justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-widest">
            <span className="text-primary">/</span>App Store
          </h1>
          <p className="text-muted">browse apps to launch on your node</p>
        </div>
        <Link href="/" className="text-sm text-muted hover:text-primary">
          ← dashboard
        </Link>
      </header>

      {apps.length === 0 ? (
        <p className="text-muted">No apps found (is the daemon reachable?).</p>
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {apps.map((app) => (
            <AppCard key={app.id} app={app} />
          ))}
        </div>
      )}
    </main>
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
      {app.description && (
        <p className="text-sm text-muted">{app.description}</p>
      )}
    </Link>
  );
}
