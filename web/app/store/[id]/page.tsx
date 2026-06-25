import Link from "next/link";
import { notFound } from "next/navigation";
import { ThemeToggle } from "@/components/ThemeToggle";
import { InstallForm } from "@/components/InstallForm";
import { getApp } from "@/lib/api";

export const dynamic = "force-dynamic";

export default async function AppDetail({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const app = await getApp(id);
  if (!app) notFound();

  return (
    <main className="mx-auto min-h-screen w-full max-w-2xl px-4 py-10">
      <ThemeToggle />

      <Link href="/store" className="text-sm text-muted hover:text-primary">
        ← App Store
      </Link>

      <header className="mt-4 mb-6 flex items-center gap-4">
        <span className="text-5xl">{app.icon ?? "📦"}</span>
        <div>
          <h1 className="text-2xl font-bold">{app.name}</h1>
          <div className="text-sm text-muted">
            {app.category} · v{app.version}
            {app.installed && (
              <span className="ml-2 rounded-full bg-primary/15 px-2 py-0.5 text-xs font-semibold text-primary">
                installed
              </span>
            )}
          </div>
        </div>
      </header>

      {app.description && (
        <p className="mb-6 text-muted">{app.description}</p>
      )}

      {app.dependencies && app.dependencies.length > 0 && (
        <div className="mb-6 rounded-lg border border-border bg-card p-4 text-sm">
          <span className="text-muted">Requires: </span>
          {app.dependencies.join(", ")}
        </div>
      )}

      <InstallForm app={app} />
    </main>
  );
}
