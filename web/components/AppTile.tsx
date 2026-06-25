"use client";

import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import type { App, ServiceStatus, ProbeResult, CredField } from "@/lib/api";
import { useConsole } from "@/components/console/ConsoleProvider";

export function AppTile({ app }: { app: App }) {
  const router = useRouter();
  const [services, setServices] = useState<ServiceStatus[] | null>(null);
  const [docker, setDocker] = useState(true);
  const [probe, setProbe] = useState<ProbeResult | null>(null);
  const [logs, setLogs] = useState<string | null>(null);
  const [busy, setBusy] = useState("");
  const [config, setConfig] = useState<{
    fields: CredField[];
    exports: Record<string, string>;
  } | null>(null);
  const consoles = useConsole();
  const [openUrl, setOpenUrl] = useState<string | null>(null);
  const [confirming, setConfirming] = useState(false);
  const [uninstallErr, setUninstallErr] = useState("");

  // Build the "open" URL on the client from the host actually being used: the
  // app's published port (works by IP or slashnode.local), or the HTTPS
  // subdomain when browsing over TLS via Caddy.
  useEffect(() => {
    if (!app.web) return;
    if (location.protocol === "https:" && app.url) {
      setOpenUrl(app.url);
    } else {
      setOpenUrl(`http://${location.hostname}:${app.web.port}`);
    }
  }, []);

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

  async function toggleConfig() {
    if (config !== null) {
      setConfig(null);
      return;
    }
    try {
      const j = await fetch(`/api/apps/${app.id}/credentials`).then((r) => r.json());
      setConfig({ fields: j.fields ?? [], exports: j.exports ?? {} });
    } catch {
      setConfig({ fields: [], exports: {} });
    }
  }

  async function updateApp() {
    setBusy("update");
    await fetch(`/api/apps/${app.id}/update`, { method: "POST" });
    setBusy("");
    router.refresh();
  }

  async function uninstall(purge: boolean) {
    setUninstallErr("");
    const res = await fetch(`/api/apps/${app.id}/uninstall?purge=${purge}`, {
      method: "POST",
    });
    if (!res.ok) {
      const j = await res.json().catch(() => ({}));
      setUninstallErr(j.error || "uninstall failed");
      return;
    }
    setConfirming(false);
    router.refresh();
  }

  async function clearLogs() {
    await fetch(`/api/apps/${app.id}/clear-logs`, { method: "POST" });
    try {
      const j = await fetch(`/api/apps/${app.id}/logs?tail=200`).then((r) => r.json());
      setLogs(j.logs || "(no logs)");
    } catch {
      setLogs("(no logs)");
    }
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
          <div className="text-xs text-muted">
            v{app.installed_version || app.version}
            {app.update_available && (
              <span className="text-primary"> → v{app.version}</span>
            )}
          </div>
        </div>
        <span className={`ml-auto text-xs font-semibold ${badge.c}`}>
          ● {badge.t}
        </span>
      </div>

      {probe && <ProbeLine probe={probe} />}

      {app.notes && (
        <div className="rounded-lg border border-primary/30 bg-primary/5 px-3 py-2 text-xs text-muted">
          ℹ {app.notes}
        </div>
      )}

      <div className="flex flex-wrap gap-1.5">
        {app.update_available && (
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
        <Btn onClick={toggleConfig}>{config !== null ? "hide config" : "config"}</Btn>
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
        <button
          onClick={() => {
            setConfirming((c) => !c);
            setUninstallErr("");
          }}
          className="ml-auto rounded-md border border-border px-2 py-1 text-xs text-muted hover:border-primary hover:text-primary"
        >
          uninstall
        </button>
      </div>

      {confirming && (
        <div className="flex flex-col gap-2 rounded-lg border border-primary/40 bg-primary/10 p-3 text-xs">
          <span>Remove {app.name}?</span>
          {uninstallErr && <span className="text-primary">{uninstallErr}</span>}
          <div className="flex flex-wrap gap-2">
            <button
              onClick={() => uninstall(false)}
              className="rounded-md border border-border px-2 py-1 hover:border-primary"
            >
              remove (keep data)
            </button>
            <button
              onClick={() => uninstall(true)}
              className="rounded-md bg-primary px-2 py-1 font-semibold text-white"
            >
              remove + delete data
            </button>
            <button
              onClick={() => setConfirming(false)}
              className="rounded-md px-2 py-1 text-muted hover:text-fg"
            >
              cancel
            </button>
          </div>
        </div>
      )}

      {config !== null && (
        <div className="flex flex-col gap-2 rounded-lg bg-bg p-3 text-xs">
          {config.fields.length > 0 && (
            <div className="flex flex-col gap-1">
              {config.fields.map((c) => (
                <CredRow key={c.key} field={c} />
              ))}
            </div>
          )}
          {Object.keys(config.exports).length > 0 && (
            <div className="flex flex-col gap-1 border-t border-border pt-2">
              <span className="text-muted">exposes</span>
              {Object.entries(config.exports).map(([k, v]) => (
                <CredRow
                  key={k}
                  field={{
                    key: k,
                    label: k,
                    value: v,
                    secret: /pass|secret|key|token/i.test(k),
                  }}
                />
              ))}
            </div>
          )}
        </div>
      )}

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

function CredRow({ field }: { field: CredField }) {
  const [show, setShow] = useState(false);
  const masked = field.secret && !show;
  return (
    <div className="flex items-center justify-between gap-3">
      <span className="text-muted">{field.label}</span>
      <span className="flex items-center gap-2">
        <code className="text-fg">
          {masked ? "•".repeat(Math.min(field.value.length || 8, 16)) : field.value || "—"}
        </code>
        {field.secret && (
          <button
            onClick={() => setShow((s) => !s)}
            className="text-muted hover:text-primary"
            aria-label={show ? "Hide" : "Show"}
          >
            {show ? "🙈" : "👁"}
          </button>
        )}
        {field.value && (
          <button
            onClick={() => navigator.clipboard?.writeText(field.value)}
            className="text-muted hover:text-primary"
            aria-label="Copy"
          >
            ⧉
          </button>
        )}
      </span>
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
