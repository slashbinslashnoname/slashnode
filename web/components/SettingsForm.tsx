"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import type { Config, Status } from "@/lib/api";

export function SettingsForm({
  config,
  status,
}: {
  config: Config;
  status: Status | null;
}) {
  const router = useRouter();

  // Editable copy of the settings.
  const [hostname, setHostname] = useState(config.hostname);
  const [mode, setMode] = useState(config.access.mode);
  const [address, setAddress] = useState(config.access.address ?? "");
  const [pwProtected, setPwProtected] = useState(config.access.password_protected);
  const [tor, setTor] = useState(config.tor.enabled);
  const [policy, setPolicy] = useState(config.update.policy);
  const [channel, setChannel] = useState(config.update.channel);

  const [cfgState, setCfgState] = useState<"idle" | "saving" | "saved" | "error">("idle");

  async function saveConfig() {
    setCfgState("saving");
    try {
      const res = await fetch("/api/config", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          hostname,
          access: { mode, address, password_protected: pwProtected },
          tor: { enabled: tor },
          update: { policy, channel },
        }),
      });
      setCfgState(res.ok ? "saved" : "error");
      if (res.ok) router.refresh();
    } catch {
      setCfgState("error");
    }
  }

  // Password change.
  const [pw, setPw] = useState("");
  const [pwState, setPwState] = useState<"idle" | "saving" | "saved" | "error">("idle");
  const [pwErr, setPwErr] = useState("");

  async function savePassword() {
    setPwErr("");
    setPwState("saving");
    try {
      const res = await fetch("/api/password", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ password: pw }),
      });
      if (!res.ok) {
        const j = await res.json().catch(() => ({}));
        setPwErr(j.error || "failed");
        setPwState("error");
        return;
      }
      setPw("");
      setPwState("saved");
    } catch {
      setPwState("error");
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <Section title="Node">
        <Info label="Version" value={status?.version ?? config.version} />
        <Info label="Node ID" value={config.node_id} />
        <Info label="Hostname" value={config.hostname} />
        {status?.onion && <Info label="Tor (.onion)" value={status.onion} />}
      </Section>

      <Section title="Admin password">
        <div className="flex flex-wrap items-end gap-3">
          <label className="flex flex-1 flex-col gap-1">
            <span className="text-sm font-medium">New password</span>
            <input
              type="password"
              value={pw}
              minLength={8}
              placeholder="at least 8 characters"
              onChange={(e) => setPw(e.target.value)}
              className={inputCls}
            />
          </label>
          <button
            onClick={savePassword}
            disabled={pwState === "saving" || pw.length < 8}
            className={btnPrimary}
          >
            {pwState === "saving" ? "saving…" : pwState === "saved" ? "✓ changed" : "Change password"}
          </button>
        </div>
        {pwErr && <p className="text-sm text-primary">{pwErr}</p>}
      </Section>

      <Section title="Access & security">
        <Toggle
          label="Password-protect the web UI"
          hint="Requires a restart to take effect on the front end."
          checked={pwProtected}
          onChange={setPwProtected}
        />
        <SelectRow label="Access mode" value={mode} onChange={setMode} options={["local", "server"]} />
        {mode === "server" && (
          <label className="flex flex-col gap-1">
            <span className="text-sm font-medium">Public address</span>
            <input
              value={address}
              placeholder="node.example.com"
              onChange={(e) => setAddress(e.target.value)}
              className={inputCls}
            />
          </label>
        )}
      </Section>

      <Section title="Tor">
        <Toggle
          label="Expose the UI and apps as .onion hidden services"
          checked={tor}
          onChange={setTor}
        />
      </Section>

      <Section title="Updates">
        <SelectRow label="Policy" value={policy} onChange={setPolicy} options={["notify", "auto"]} />
        <SelectRow label="Channel" value={channel} onChange={setChannel} options={["stable", "beta"]} />
      </Section>

      <div className="flex items-center gap-3">
        <button onClick={saveConfig} disabled={cfgState === "saving"} className={btnPrimary}>
          {cfgState === "saving" ? "saving…" : cfgState === "saved" ? "✓ saved" : "Save settings"}
        </button>
        {cfgState === "error" && <span className="text-sm text-primary">save failed</span>}
        <span className="text-xs text-muted">
          Port, hostname and password-protection changes need a restart.
        </span>
      </div>

      <Section title="Maintenance">
        <div className="flex flex-wrap gap-2">
          <Action label="Reload Caddy" action="reload-caddy" />
          <Action label="Reload Tor" action="reload-tor" />
          <Action label="Prune images" action="prune" />
          <Action label="Restart daemon" action="restart" danger />
        </div>
      </Section>
    </div>
  );
}

const inputCls =
  "rounded-lg border border-border bg-bg px-3 py-2 text-sm outline-none focus:border-primary";
const btnPrimary =
  "cursor-pointer rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-white disabled:cursor-default disabled:opacity-60";

function Action({
  label,
  action,
  danger,
}: {
  label: string;
  action: string;
  danger?: boolean;
}) {
  const [state, setState] = useState<"idle" | "busy" | "done" | "error">("idle");
  async function run() {
    setState("busy");
    try {
      const res = await fetch(`/api/maintenance/${action}`, { method: "POST" });
      setState(res.ok ? "done" : "error");
    } catch {
      setState("error");
    }
    if (action !== "restart") setTimeout(() => setState("idle"), 2500);
  }
  return (
    <button
      onClick={run}
      disabled={state === "busy"}
      className={`cursor-pointer rounded-lg border px-3 py-2 text-sm disabled:cursor-default disabled:opacity-60 ${
        danger
          ? "border-primary/50 text-primary hover:bg-primary/10"
          : "border-border hover:border-primary"
      }`}
    >
      {state === "busy"
        ? "…"
        : state === "done"
          ? "✓ done"
          : state === "error"
            ? "failed"
            : label}
    </button>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="flex flex-col gap-3 rounded-xl border border-border bg-card p-5">
      <h2 className="text-sm font-semibold uppercase tracking-wider text-muted">{title}</h2>
      {children}
    </section>
  );
}

function Info({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-3 text-sm">
      <span className="text-muted">{label}</span>
      <code className="break-all text-fg">{value}</code>
    </div>
  );
}

function Toggle({
  label,
  hint,
  checked,
  onChange,
}: {
  label: string;
  hint?: string;
  checked: boolean;
  onChange: (v: boolean) => void;
}) {
  return (
    <label className="flex cursor-pointer items-start gap-3 text-sm">
      <input
        type="checkbox"
        checked={checked}
        onChange={(e) => onChange(e.target.checked)}
        className="mt-0.5 h-5 w-5 accent-[var(--primary)]"
      />
      <span>
        {label}
        {hint && <span className="block text-xs text-muted">{hint}</span>}
      </span>
    </label>
  );
}

function SelectRow({
  label,
  value,
  onChange,
  options,
}: {
  label: string;
  value: string;
  onChange: (v: string) => void;
  options: string[];
}) {
  return (
    <label className="flex items-center justify-between gap-3 text-sm">
      <span className="font-medium">{label}</span>
      <select value={value} onChange={(e) => onChange(e.target.value)} className={inputCls}>
        {options.map((o) => (
          <option key={o} value={o}>
            {o}
          </option>
        ))}
      </select>
    </label>
  );
}
