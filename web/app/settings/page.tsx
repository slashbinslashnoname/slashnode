import Link from "next/link";
import { TopControls } from "@/components/TopControls";
import { SettingsForm } from "@/components/SettingsForm";
import { BackupPanel } from "@/components/BackupPanel";
import { TailscalePanel } from "@/components/TailscalePanel";
import { getConfig, getStatus } from "@/lib/api";

export const dynamic = "force-dynamic";

export default async function Settings() {
  const [config, status] = await Promise.all([getConfig(), getStatus()]);

  return (
    <main className="mx-auto min-h-screen w-full max-w-2xl px-4 py-10">
      <TopControls />

      <header className="mb-8 flex items-end justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-widest">
            <span className="text-primary">/</span>Settings
          </h1>
          <p className="text-muted">configure your node</p>
        </div>
        <Link href="/" className="text-sm text-muted hover:text-primary">
          ← dashboard
        </Link>
      </header>

      {config ? (
        <SettingsForm config={config} status={status} />
      ) : (
        <p className="text-muted">Could not load settings (is the daemon reachable?).</p>
      )}

      <section className="mt-8">
        <h2 className="mb-3 text-sm font-semibold uppercase tracking-wider text-muted">
          Tailscale
        </h2>
        <div className="rounded-xl border border-border bg-card p-5">
          <TailscalePanel />
        </div>
      </section>

      <section className="mt-8">
        <h2 className="mb-3 text-sm font-semibold uppercase tracking-wider text-muted">
          Backup &amp; restore
        </h2>
        <div className="rounded-xl border border-border bg-card p-5">
          <BackupPanel />
        </div>
      </section>
    </main>
  );
}
