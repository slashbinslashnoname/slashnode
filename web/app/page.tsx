import Link from "next/link";
import { TopControls } from "@/components/TopControls";
import { UpdateBanner } from "@/components/UpdateBanner";
import { StorageBanner } from "@/components/SystemStatus";
import { AppTile } from "@/components/AppTile";
import { BitcoinPrice } from "@/components/BitcoinPrice";
import { getApps, getStatus, getUpdate } from "@/lib/api";

export const dynamic = "force-dynamic";

export default async function Home() {
  const [data, update] = await Promise.all([getApps(), getUpdate()]);
  const installed = (data?.apps ?? [])
    .filter((a) => a.installed)
    .sort((a, b) => (b.priority ?? 0) - (a.priority ?? 0));

  return (
    <main className="mx-auto min-h-screen w-full max-w-5xl px-4 py-10">
      <TopControls />

      <header className="mb-8 flex items-center gap-4">
        <h1 className="flex-1 text-2xl font-bold tracking-widest">
          <span className="text-primary">/</span>SlashNode
        </h1>
        <BitcoinPrice />
        <Link
          href="/store"
          className="rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-white hover:opacity-90"
        >
          App Store →
        </Link>
      </header>

      {update?.available && (
        <div className="mb-6">
          <UpdateBanner current={update.current} latest={update.latest} />
        </div>
      )}

      <StorageBanner />

      <h2 className="mb-4 text-sm font-semibold uppercase tracking-wider text-muted">
        Your apps
      </h2>

      {installed.length === 0 ? (
        <div className="rounded-xl border border-border bg-card p-8 text-center">
          <p className="text-muted">No apps launched yet.</p>
          <Link href="/store" className="mt-2 inline-block font-semibold text-primary">
            Browse the App Store →
          </Link>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          {installed.map((app) => (
            <AppTile key={app.id} app={app} />
          ))}
        </div>
      )}
    </main>
  );
}
