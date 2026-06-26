"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import type { App } from "@/lib/api";

// DomainTab lets the operator serve an app at EITHER a subdomain under the node
// host (https://<sub>.<host>) OR a full custom domain (https://app.example.com)
// — one or the other, never both. The choice is a single mode toggle.
export function DomainTab({ app }: { app: App }) {
  const router = useRouter();
  const base = app.host ?? "";

  const [mode, setMode] = useState<"subdomain" | "domain">(app.domain ? "domain" : "subdomain");
  const [sub, setSub] = useState(app.subdomain || app.id);
  const [domain, setDomain] = useState(app.domain ?? "");
  const [state, setState] = useState<"idle" | "saving" | "saved" | "error">("idle");
  const [err, setErr] = useState("");

  if (!app.web) {
    return (
      <p className="text-sm text-muted">
        This app has no web UI, so it isn’t served on a domain.
      </p>
    );
  }

  async function save() {
    setErr("");
    setState("saving");
    // Saving one mode clears the other so exactly one address is ever active:
    // subdomain mode clears the custom domain; domain mode sets it.
    const params: Record<string, string> =
      mode === "subdomain"
        ? { subdomain: sub, domain: "" }
        : { domain };
    try {
      const qs = new URLSearchParams(params).toString();
      const r = await fetch(`/api/apps/${app.id}/domain?${qs}`, { method: "POST" });
      if (!r.ok) {
        const j = await r.json().catch(() => ({}));
        setErr(j.error || "failed");
        setState("error");
        return;
      }
      setState("saved");
      router.refresh();
    } catch {
      setState("error");
    }
  }

  const subLabel = sub || app.id;
  const activeUrl =
    mode === "subdomain"
      ? base
        ? `https://${subLabel}.${base}`
        : ""
      : domain
        ? `https://${domain}`
        : "";

  return (
    <div className="flex flex-col gap-4">
      {/* Mode toggle: choose exactly one addressing scheme. */}
      <div className="inline-flex w-fit rounded-lg border border-border p-0.5 text-sm">
        <ModeBtn active={mode === "subdomain"} onClick={() => setMode("subdomain")}>
          Subdomain
        </ModeBtn>
        <ModeBtn active={mode === "domain"} onClick={() => setMode("domain")}>
          Custom domain
        </ModeBtn>
      </div>

      {mode === "subdomain" ? (
        <label className="flex flex-col gap-1">
          <span className="text-sm font-medium">Subdomain</span>
          <div className="flex items-center gap-2">
            <input
              value={sub}
              onChange={(e) => setSub(e.target.value)}
              placeholder={app.id}
              className={inputCls}
            />
            {base && <span className="text-sm text-muted">.{base}</span>}
          </div>
          <span className="text-xs text-muted">
            Lowercase letters, digits and hyphens. Empty resets to “{app.id}”.
          </span>
        </label>
      ) : (
        <label className="flex flex-col gap-1">
          <span className="text-sm font-medium">Custom domain</span>
          <input
            value={domain}
            onChange={(e) => setDomain(e.target.value)}
            placeholder="app.example.com"
            className={inputCls}
          />
          <span className="text-xs text-muted">
            A full domain (e.g. <code>app.example.com</code>). Caddy issues an HTTPS
            certificate automatically once its DNS points at this node.
          </span>
        </label>
      )}

      {activeUrl && (
        <div className="rounded-lg bg-bg p-3 text-sm">
          <span className="text-muted">Served at: </span>
          <code className="break-all text-fg">{activeUrl}</code>
        </div>
      )}

      <div className="flex items-center gap-3">
        <button
          onClick={save}
          disabled={state === "saving" || (mode === "domain" && !domain)}
          className="cursor-pointer rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-white disabled:cursor-default disabled:opacity-60"
        >
          {state === "saving" ? "saving…" : state === "saved" ? "✓ saved" : "Save"}
        </button>
        {err && <span className="text-sm text-primary">{err}</span>}
      </div>

      {/* DNS guidance for the active mode only. */}
      {mode === "subdomain" && base && (
        <DnsHelp
          title="DNS for the subdomain"
          records={`A   *.${base}        <node-ip>   ← wildcard, covers every app
  — or, per app —
A   ${subLabel}.${base}   <node-ip>`}
          note={
            <>
              With a wildcard <code>*.{base}</code> record any subdomain works
              immediately — no extra DNS per app.
            </>
          }
        />
      )}
      {mode === "domain" && domain && (
        <DnsHelp
          title="DNS for the custom domain"
          records={`A   ${domain}   <node-ip>`}
          note={
            <>
              Point <code>{domain}</code> at this node. Ports 80 and 443 must be
              reachable from the internet for the certificate.
            </>
          }
        />
      )}
    </div>
  );
}

const inputCls =
  "w-56 rounded-lg border border-border bg-bg px-3 py-2 text-sm outline-none focus:border-primary";

function ModeBtn({
  active,
  onClick,
  children,
}: {
  active: boolean;
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <button
      onClick={onClick}
      className={`cursor-pointer rounded-md px-3 py-1.5 font-medium transition-colors ${
        active ? "bg-primary text-white" : "text-muted hover:text-fg"
      }`}
    >
      {children}
    </button>
  );
}

function DnsHelp({
  title,
  records,
  note,
}: {
  title: string;
  records: string;
  note: React.ReactNode;
}) {
  return (
    <div className="rounded-lg border border-border bg-card p-4 text-sm">
      <div className="mb-2 text-xs font-semibold text-muted">{title}</div>
      <pre className="overflow-auto rounded-md bg-bg p-3 text-xs leading-relaxed">{records}</pre>
      <p className="mt-2 text-muted">{note}</p>
    </div>
  );
}
