"use client";

import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import type { App, ServiceStatus, ProbeResult } from "@/lib/api";
import { Console } from "@/components/Console";

export function AppTile({ app }: { app: App }) {
  const router = useRouter();
  const [services, setServices] = useState<ServiceStatus[] | null>(null);
  const [docker, setDocker] = useState(true);
  const [probe, setProbe] = useState<ProbeResult | null>(null);
  const [logs, setLogs] = useState<string | null>(null);
  const [busy, setBusy] = useState("");
  const [consoleFor, setConsoleFor] = useState<string | null>(null);

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

  async function toggleLogs() {
    if (logs !== null) {
      setLogs(null);
      return;
    }
    try {
      const j = await fetch(`/api/apps/${app.id}/logs?tail=200`).then((r) => r.json());
      setLogs(j.logs || "(no logs)");
    } catch {
      setLogs("(failed to load logs)");
    }
  }

  const running = (services ?? []).some((s) => s.state === "running");
  const badge = !docker
    ? { t: "docker off", c: "text-muted" }
    : running
      ? { t: "running", c: "text-primary" }
      : services && services.length > 0
        ? { t: "stopped", c: "text-muted" }
        : { t: "not started", c: "text-muted" };

  return (
    <div className="flex flex-col gap-3 rounded-xl border border-border bg-card p-5">
      <div className="flex items-center gap-3">
        <span className="text-3xl">{app.icon ?? "📦"}</span>
        <div className="min-w-0">
          <div className="font-semibold">{app.name}</div>
          <div className="text-xs text-muted">v{app.version}</div>
        </div>
        <span className={`ml-auto text-xs font-semibold ${badge.c}`}>
          ● {badge.t}
        </span>
      </div>

      {probe && <ProbeLine probe={probe} />}

      <div className="flex flex-wrap gap-2">
        <Btn onClick={() => act("start")} busy={busy === "start"}>start</Btn>
        <Btn onClick={() => act("stop")} busy={busy === "stop"}>stop</Btn>
        <Btn onClick={() => act("restart")} busy={busy === "restart"}>restart</Btn>
        <Btn onClick={toggleLogs}>{logs !== null ? "hide logs" : "logs"}</Btn>
        {(services ?? []).map((s) => (
          <Btn key={s.service} onClick={() => setConsoleFor(s.service)}>
            {`console${(services ?? []).length > 1 ? `:${s.service}` : ""}`}
          </Btn>
        ))}
        {app.url && (
          <a
            href={app.url}
            target="_blank"
            rel="noreferrer"
            className="rounded-lg border border-border px-3 py-1.5 text-sm hover:border-primary"
          >
            open ↗
          </a>
        )}
      </div>

      {consoleFor && (
        <Console container={consoleFor} onClose={() => setConsoleFor(null)} />
      )}

      {logs !== null && (
        <pre className="max-h-60 overflow-auto rounded-lg bg-bg p-3 text-xs leading-relaxed">
          {logs}
        </pre>
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
      ) : (
        <span className={probe.ok ? "text-primary" : "text-muted"}>
          {probe.ok ? "● reachable" : "○ unreachable"}
          {probe.detail ? ` — ${probe.detail}` : ""}
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
      className="rounded-lg border border-border px-3 py-1.5 text-sm hover:border-primary disabled:opacity-50"
    >
      {busy ? "…" : children}
    </button>
  );
}
