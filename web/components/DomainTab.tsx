"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import type { App } from "@/lib/api";

// DomainTab lets the operator change both the reverse-proxy subdomain an app is
// served under (https://<sub>.<host>) and an optional full custom domain
// (https://app.example.com), and shows the DNS records to configure for each.
export function DomainTab({ app }: { app: App }) {
  const router = useRouter();
  const base = app.host ?? "";

  const [sub, setSub] = useState(app.subdomain || app.id);
  const [domain, setDomain] = useState(app.domain ?? "");

  const [subState, setSubState] = useState<SaveState>("idle");
  const [subErr, setSubErr] = useState("");
  const [domState, setDomState] = useState<SaveState>("idle");
  const [domErr, setDomErr] = useState("");

  if (!app.web) {
    return (
      <p className="text-sm text-muted">
        This app has no web UI, so it isn’t served on a domain.
      </p>
    );
  }

  async function save(
    params: Record<string, string>,
    setState: (s: SaveState) => void,
    setErr: (s: string) => void,
  ) {
    setErr("");
    setState("saving");
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
  const subUrl = base ? `https://${subLabel}.${base}` : "";

  return (
    <div className="flex flex-col gap-6">
      {/* Subdomain under the node host */}
      <section className="flex flex-col gap-3">
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
        {subUrl && (
          <div className="rounded-lg bg-bg p-3 text-sm">
            <span className="text-muted">URL: </span>
            <code className="break-all text-fg">{subUrl}</code>
          </div>
        )}
        <div className="flex items-center gap-3">
          <button
            onClick={() => save({ subdomain: sub }, setSubState, setSubErr)}
            disabled={subState === "saving"}
            className={btnCls}
          >
            {subState === "saving" ? "saving…" : subState === "saved" ? "✓ saved" : "Save subdomain"}
          </button>
          {subErr && <span className="text-sm text-primary">{subErr}</span>}
        </div>
        {base && (
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
      </section>

      {/* Full custom domain */}
      <section className="flex flex-col gap-3 border-t border-border pt-5">
        <label className="flex flex-col gap-1">
          <span className="text-sm font-medium">Custom domain</span>
          <input
            value={domain}
            onChange={(e) => setDomain(e.target.value)}
            placeholder="app.example.com"
            className={inputCls}
          />
          <span className="text-xs text-muted">
            A full domain served in addition to the subdomain (e.g.{" "}
            <code>app.example.com</code> or <code>my-node.org</code>). Leave empty
            to remove. Caddy issues an HTTPS certificate automatically once DNS
            points here.
          </span>
        </label>
        {domain && (
          <div className="rounded-lg bg-bg p-3 text-sm">
            <span className="text-muted">URL: </span>
            <code className="break-all text-fg">https://{domain}</code>
          </div>
        )}
        <div className="flex items-center gap-3">
          <button
            onClick={() => save({ domain }, setDomState, setDomErr)}
            disabled={domState === "saving"}
            className={btnCls}
          >
            {domState === "saving"
              ? "saving…"
              : domState === "saved"
                ? "✓ saved"
                : "Save custom domain"}
          </button>
          {domErr && <span className="text-sm text-primary">{domErr}</span>}
        </div>
        {domain && (
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
      </section>
    </div>
  );
}

type SaveState = "idle" | "saving" | "saved" | "error";

const inputCls =
  "w-56 rounded-lg border border-border bg-bg px-3 py-2 text-sm outline-none focus:border-primary";
const btnCls =
  "cursor-pointer rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-white disabled:cursor-default disabled:opacity-60";

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
