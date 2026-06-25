import Link from "next/link";
import { Skull } from "@/components/Skull";
import { ThemeToggle } from "@/components/ThemeToggle";
import { UpdateBanner } from "@/components/UpdateBanner";
import { getStatus, getUpdate } from "@/lib/api";

export const dynamic = "force-dynamic";

export default async function Home() {
  const [status, update] = await Promise.all([getStatus(), getUpdate()]);

  const nodeId = status?.node_id ?? "—";
  const version = status?.version ?? "—";

  return (
    <main className="flex min-h-screen flex-col items-center justify-center gap-6 px-4">
      <ThemeToggle />

      <Skull />

      <div className="text-center">
        <h1 className="text-2xl font-bold tracking-widest">
          <span className="text-primary">/</span>SlashNode
        </h1>
        <p className="text-muted">your node, your rules</p>
      </div>

      {update?.available && (
        <UpdateBanner current={update.current} latest={update.latest} />
      )}

      <div className="w-full max-w-md rounded-xl border border-border bg-card p-5">
        <Row k="node" v={nodeId} />
        <Row k="version" v={version} />
        <Row
          k="status"
          v={
            <span>
              <span className="text-primary">●</span>{" "}
              {status ? "online" : "daemon unreachable"}
            </span>
          }
        />
      </div>

      <Link
        href="/store"
        className="rounded-lg bg-primary px-5 py-2.5 font-semibold text-white hover:opacity-90 transition-opacity"
      >
        Browse the App Store →
      </Link>
    </main>
  );
}

function Row({ k, v }: { k: string; v: React.ReactNode }) {
  return (
    <div className="flex justify-between gap-8 py-1">
      <span className="text-muted">{k}</span>
      <span className="font-semibold">{v}</span>
    </div>
  );
}
