"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import type { App, ProbeResult } from "@/lib/api";
import { useConsole } from "@/components/console/ConsoleProvider";
import { useData } from "@/components/store/DataProvider";
import { webClearnetUrl } from "@/lib/appUrl";
import { appVersion } from "@/lib/version";

export function AppTile({ app }: { app: App }) {
  const router = useRouter();
  // Live status/probe/update-check come from the global store (polled in the
  // layout), so this tile shows data instantly across navigations.
  const { status, probe: probeMap, imgUpdate: imgUpdateMap, refresh } = useData();
  const [busy, setBusy] = useState("");
  const [openUrl, setOpenUrl] = useState<string | null>(null);
  const consoles = useConsole();

  const st = status[app.id];
  const services = st?.services ?? null;
  const docker = st?.docker ?? true;
  const onion = st?.onion ?? app.onion ?? null;
  const probe = probeMap[app.id] ?? null;
  const imgUpdate = !!imgUpdateMap[app.id];

  // webClearnetUrl reads window.location, so defer to the client.
  useEffect(() => {
    if (app.web) setOpenUrl(webClearnetUrl(app.url, app.web.port));
  }, [app.web, app.url]);

  const onionUrl = onion && app.web ? `http://${onion}` : null;

  async function act(action: string) {
    setBusy(action);
    await fetch(`/api/apps/${app.id}/${action}`, { method: "POST" });
    setBusy("");
    refresh();
    router.refresh();
  }

  async function updateApp() {
    setBusy("update");
    await fetch(`/api/apps/${app.id}/update`, { method: "POST" });
    setBusy("");
    refresh();
    router.refresh();
  }

  const running = (services ?? []).some((s) => s.state === "running");
  const badge = !docker
    ? { t: "docker off", c: "text-muted" }
    : running
      ? { t: "running", c: "text-primary" }
      : services && services.length > 0
        ? { t: "stopped", c: "text-muted" }
        : { t: "not started", c: "text-muted" };
  const hasUpdate = app.update_available || imgUpdate;

  return (
    <div className="flex flex-col gap-3 rounded-xl border border-border bg-card p-5">
      <div className="flex items-center gap-3">
        <span className="text-3xl">{app.icon ?? "📦"}</span>
        <div className="min-w-0">
          <div className="font-semibold">{app.name}</div>
          <div className="text-xs text-muted">
            v{appVersion(app)}
            {app.update_available && (
              <span className="text-primary"> → v{app.version}</span>
            )}
            {!app.update_available && imgUpdate && (
              <span className="text-primary"> · image update</span>
            )}
          </div>
        </div>
        <span className={`ml-auto text-xs font-semibold ${badge.c}`}>
          ● {badge.t}
        </span>
      </div>

      {probe && <ProbeLine probe={probe} />}

      <div className="flex flex-wrap gap-1.5">
        {hasUpdate && (
          <button
            onClick={updateApp}
            disabled={busy === "update"}
            className="cursor-pointer rounded-md bg-primary px-2 py-1 text-xs font-semibold text-white disabled:opacity-60"
          >
            {busy === "update" ? "updating…" : "update"}
          </button>
        )}
        <Btn onClick={() => act("start")} busy={busy === "start"}>start</Btn>
        <Btn onClick={() => act("stop")} busy={busy === "stop"}>stop</Btn>
        <Btn onClick={() => act("restart")} busy={busy === "restart"}>restart</Btn>
        <Btn onClick={() => consoles.openLogs(app.id, app.name)}>logs</Btn>
        <Btn onClick={() => consoles.openConfig(app.id, app.name)}>config</Btn>
        {(services ?? []).map((s) => (
          <Btn key={s.service} onClick={() => consoles.open(s.container || s.service)}>
            {`console${(services ?? []).length > 1 ? `:${s.service}` : ""}`}
          </Btn>
        ))}
      </div>

      {(openUrl || onionUrl) && (
        <div className="flex flex-wrap gap-2">
          {openUrl && (
            <a
              href={openUrl}
              target="_blank"
              rel="noreferrer"
              className="cursor-pointer rounded-md border border-border px-2 py-1 text-xs hover:border-primary"
            >
              open ↗
            </a>
          )}
          {onionUrl && (
            <a
              href={onionUrl}
              target="_blank"
              rel="noreferrer"
              title={onionUrl}
              className="cursor-pointer rounded-md border border-border px-2 py-1 text-xs hover:border-primary"
            >
              open .onion ↗
            </a>
          )}
        </div>
      )}
    </div>
  );
}

function ProbeLine({ probe }: { probe: ProbeResult }) {
  if (probe.type === "none") return null;
  const stats = probe.stats ?? [];
  return (
    <div className="rounded-lg bg-bg px-3 py-2 text-xs">
      {stats.length > 0 ? (
        <div className="flex flex-wrap gap-x-4 gap-y-1">
          {stats.map((s) => (
            <span key={s.label} className="text-muted">
              {s.label}: <span className="text-fg">{s.value}</span>
            </span>
          ))}
        </div>
      ) : probe.ok ? (
        <span className="text-primary">● reachable</span>
      ) : (
        <span className="text-muted" title={probe.detail}>
          ○ not synced or unavailable yet
        </span>
      )}
    </div>
  );
}

function Btn({
  children,
  onClick,
  busy,
}: {
  children: React.ReactNode;
  onClick: () => void;
  busy?: boolean;
}) {
  return (
    <button
      onClick={onClick}
      disabled={busy}
      className="rounded-md border border-border px-2 py-1 text-xs hover:border-primary disabled:opacity-50"
    >
      {busy ? "…" : children}
    </button>
  );
}
