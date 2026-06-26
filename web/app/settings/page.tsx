import Link from "next/link";
import { TopControls } from "@/components/TopControls";
import { SettingsForm } from "@/components/SettingsForm";
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
    </main>
  );
}
