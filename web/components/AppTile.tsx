"use client";

import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import type { App, ServiceStatus, ProbeResult } from "@/lib/api";
import { useConsole } from "@/components/console/ConsoleProvider";
import { CredsPanel } from "@/components/CredsPanel";
import { EndpointsPanel } from "@/components/EndpointsPanel";

export function AppTile({ app }: { app: App }) {
  const router = useRouter();
  const [services, setServices] = useState<ServiceStatus[] | null>(null);
  const [docker, setDocker] = useState(true);
  const [probe, setProbe] = useState<ProbeResult | null>(null);
  const [logs, setLogs] = useState<string | null>(null);
  const [busy, setBusy] = useState("");
  const [showConfig, setShowConfig] = useState(false);
  const [imgUpdate, setImgUpdate] = useState(false);
  const [openUrl, setOpenUrl] = useState<string | null>(null);
  const consoles = useConsole();

  useEffect(() => {
    if (!app.web) return;
    if (location.protocol === "https:" && app.url) setOpenUrl(app.url);
    else setOpenUrl(`http://${location.hostname}:${app.web.port}`);
  }, []);

  // Docker image update check (no manifest bump needed).
  useEffect(() => {
    fetch(`/api/apps/${app.id}/image-update`)
      .then((r) => r.json())
      .then((j) => setImgUpdate(!!j.available))
      .catch(() => {});
  }, [app.id]);

  const refresh = useCallback(async () => {
    try {
      const s = await fetch(`/api/apps/${app.id}/status`).then((r) => r.json());
      setServices(s.services ?? []);
      setDocker(s.docker ?? false);
    } catch {
      setServices([]);
    }
    if (app.probe) {
      try {
        setProbe(await fetch(`/api/apps/${app.id}/probe`).then((r) => r.json()));
      } catch {
        setProbe(null);
      }
    }
  }, [app.id, app.probe]);

  useEffect(() => {
    refresh();
    const t = setInterval(refresh, 5000);
    return () => clearInterval(t);
  }, [refresh]);

  async function act(action: string) {
    setBusy(action);
    await fetch(`/api/apps/${app.id}/${action}`, { method: "POST" });
    await refresh();
    setBusy("");
    router.refresh();
  }

  async function updateApp() {
    setBusy("update");
    await fetch(`/api/apps/${app.id}/update`, { method: "POST" });
    setBusy("");
    setImgUpdate(false);
    await refresh();
    router.refresh();
  }

  async function clearLogs() {
    await fetch(`/api/apps/${app.id}/clear-logs`, { method: "POST" });
    await loadLogs();
  }

  async function loadLogs() {
    try {
      const j = await fetch(`/api/apps/${app.id}/logs?tail=200`).then((r) => r.json());
      setLogs(j.logs || "(no logs)");
    } catch {
      setLogs("(failed to load logs)");
    }
  }

  function toggleLogs() {
    if (logs !== null) setLogs(null);
    else loadLogs();
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
            v{app.installed_version || app.version}
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

      {app.endpoints && app.endpoints.length > 0 && (
        <EndpointsPanel endpoints={app.endpoints} />
      )}

      <div className="flex flex-wrap gap-1.5">
        {hasUpdate && (
          <button
            onClick={updateApp}
            disabled={busy === "update"}
            className="rounded-md bg-primary px-2 py-1 text-xs font-semibold text-white disabled:opacity-60"
          >
            {busy === "update" ? "updating…" : "update"}
          </button>
        )}
        <Btn onClick={() => act("start")} busy={busy === "start"}>start</Btn>
        <Btn onClick={() => act("stop")} busy={busy === "stop"}>stop</Btn>
        <Btn onClick={() => act("restart")} busy={busy === "restart"}>restart</Btn>
        <Btn onClick={toggleLogs}>{logs !== null ? "hide logs" : "logs"}</Btn>
        <Btn onClick={() => setShowConfig((s) => !s)}>
          {showConfig ? "hide config" : "config"}
        </Btn>
        {(services ?? []).map((s) => (
          <Btn key={s.service} onClick={() => consoles.open(s.service)}>
            {`console${(services ?? []).length > 1 ? `:${s.service}` : ""}`}
          </Btn>
        ))}
        {openUrl && (
          <a
            href={openUrl}
            target="_blank"
            rel="noreferrer"
            className="rounded-md border border-border px-2 py-1 text-xs hover:border-primary"
          >
            open ↗
          </a>
        )}
        {app.onion_url && (
          <a
            href={app.onion_url}
            target="_blank"
            rel="noreferrer"
            title={app.onion_url}
            className="rounded-md border border-border px-2 py-1 text-xs hover:border-primary"
          >
            open .onion ↗
          </a>
        )}
      </div>

      {showConfig && <CredsPanel id={app.id} />}

      {logs !== null && (
        <div className="flex flex-col gap-1">
          <button
            onClick={clearLogs}
            className="self-end text-xs text-muted hover:text-primary"
          >
            clear logs
          </button>
          <pre className="max-h-60 overflow-auto rounded-lg bg-bg p-3 text-xs leading-relaxed">
            {logs}
          </pre>
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
