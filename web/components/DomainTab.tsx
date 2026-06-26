"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import type { App } from "@/lib/api";

function parseUrl(url?: string): { sub: string; base: string } {
  if (!url) return { sub: "", base: "" };
  const noScheme = url.replace(/^https?:\/\//, "");
  const dot = noScheme.indexOf(".");
  if (dot < 0) return { sub: noScheme, base: "" };
  return { sub: noScheme.slice(0, dot), base: noScheme.slice(dot + 1) };
}

// DomainTab lets the operator change the reverse-proxy subdomain an app is
// served under (https://<sub>.<host>) and shows the DNS records to configure.
export function DomainTab({ app }: { app: App }) {
  const router = useRouter();
  const { sub, base } = parseUrl(app.url);
  const [value, setValue] = useState(sub);
  const [state, setState] = useState<"idle" | "saving" | "saved" | "error">("idle");
  const [err, setErr] = useState("");

  if (!app.web) {
    return (
      <p className="text-sm text-muted">
        This app has no web UI, so it isn’t served on a subdomain.
      </p>
    );
  }

  async function save() {
    setErr("");
    setState("saving");
    try {
      const r = await fetch(
        `/api/apps/${app.id}/domain?subdomain=${encodeURIComponent(value)}`,
        { method: "POST" },
      );
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

  const label = value || app.id;
  const domain = base || "<your-domain>";
  const fullUrl = base ? `https://${label}.${base}` : "";

  return (
    <div className="flex flex-col gap-4">
      <label className="flex flex-col gap-1">
        <span className="text-sm font-medium">Subdomain</span>
        <div className="flex items-center gap-2">
          <input
            value={value}
            onChange={(e) => setValue(e.target.value)}
            placeholder={app.id}
            className="w-48 rounded-lg border border-border bg-bg px-3 py-2 text-sm outline-none focus:border-primary"
          />
          {base && <span className="text-sm text-muted">.{base}</span>}
        </div>
        <span className="text-xs text-muted">
          Lowercase letters, digits and hyphens. Empty resets to “{app.id}”.
        </span>
      </label>

      {fullUrl && (
        <div className="rounded-lg bg-bg p-3 text-sm">
          <span className="text-muted">App URL: </span>
          <code className="break-all text-fg">{fullUrl}</code>
        </div>
      )}

      <div className="flex items-center gap-3">
        <button
          onClick={save}
          disabled={state === "saving"}
          className="cursor-pointer rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-white disabled:cursor-default disabled:opacity-60"
        >
          {state === "saving" ? "saving…" : state === "saved" ? "✓ saved" : "Save subdomain"}
        </button>
        {err && <span className="text-sm text-primary">{err}</span>}
      </div>

      <div className="rounded-lg border border-border bg-card p-4 text-sm">
        <div className="mb-2 text-xs font-semibold text-muted">Name server (DNS) setup</div>
        <p className="text-muted">
          At your registrar, point a record at this node so the subdomain resolves
          and Caddy can issue an HTTPS certificate:
        </p>
        <pre className="mt-2 overflow-auto rounded-md bg-bg p-3 text-xs leading-relaxed">{`A   *.${domain}        <node-ip>   ← wildcard, covers every app
  — or, per app —
A   ${label}.${domain}   <node-ip>`}</pre>
        <p className="mt-2 text-muted">
          With a wildcard <code>*.{domain}</code> record, any subdomain works
          immediately — no extra DNS per app. Ports 80 and 443 must be reachable
          from the internet for the certificate.
        </p>
      </div>
    </div>
  );
}
