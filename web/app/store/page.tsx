"use client";

import Link from "next/link";
import { TopControls } from "@/components/TopControls";
import { StoreList } from "@/components/StoreList";
import { useData } from "@/components/store/DataProvider";

export default function Store() {
  // The catalog lives in the global store (polled in the layout), so coming back
  // to the store from another page is instant — no refetch.
  const { apps, ready } = useData();

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

      {ready && apps.length === 0 ? (
        <p className="text-muted">No apps found (is the daemon reachable?).</p>
      ) : (
        <StoreList apps={apps} />
      )}
    </main>
  );
}
