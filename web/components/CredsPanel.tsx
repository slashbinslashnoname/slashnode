"use client";

import { useEffect, useState } from "react";
import type { CredField, AppEndpoint } from "@/lib/api";

// CredsPanel shows an installed app's full configuration in a single view:
// its connection endpoints, stored parameters (rpc user, passwords…) and its
// exposed config (the endpoints other apps wire to), with reveal + copy.
export function CredsPanel({
  id,
  endpoints,
}: {
  id: string;
  endpoints?: AppEndpoint[];
}) {
  const [data, setData] = useState<{
    fields: CredField[];
    exports: Record<string, string>;
  } | null>(null);
  const [host, setHost] = useState("");

  useEffect(() => {
    setHost(location.hostname);
    fetch(`/api/apps/${id}/credentials`)
      .then((r) => r.json())
      .then((j) => setData({ fields: j.fields ?? [], exports: j.exports ?? {} }))
      .catch(() => setData({ fields: [], exports: {} }));
  }, [id]);

  if (!data) return null;

  const exportRows = Object.entries(data.exports);

  return (
    <div className="flex flex-col gap-4 rounded-lg border border-border bg-card p-3 text-sm">
      {endpoints && endpoints.length > 0 && (
        <Section title="Endpoints">
          {endpoints.map((e) => {
            const isWeb = e.scheme === "http" || e.scheme === "https";
            const value = isWeb
              ? `${e.scheme}://${host}:${e.port}${e.path ?? ""}`
              : `${host}:${e.port}`;
            return (
              <Row key={e.label + e.port} label={e.label}>
                <code className="break-all text-fg">{host ? value : "…"}</code>
                <CopyBtn value={value} />
                {isWeb && host && (
                  <a
                    href={value}
                    target="_blank"
                    rel="noreferrer"
                    className="text-muted hover:text-primary"
                    aria-label="Open"
                  >
                    ↗
                  </a>
                )}
              </Row>
            );
          })}
        </Section>
      )}

      <Section title="Parameters">
        {data.fields.length > 0 ? (
          data.fields.map((c) => <CredRow key={c.key} field={c} />)
        ) : (
          <span className="text-xs text-muted">No parameters.</span>
        )}
      </Section>

      <Section title="Config">
        {exportRows.length > 0 ? (
          exportRows.map(([k, v]) => (
            <CredRow
              key={k}
              field={{
                key: k,
                label: k,
                value: v,
                secret: /pass|secret|key|token/i.test(k),
              }}
            />
          ))
        ) : (
          <span className="text-xs text-muted">Nothing exposed.</span>
        )}
      </Section>
    </div>
  );
}

function Section({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <div className="flex flex-col gap-1">
      <div className="text-xs font-semibold text-muted">{title}</div>
      {children}
    </div>
  );
}

function Row({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div className="flex items-center justify-between gap-3 text-xs">
      <span className="text-muted">{label}</span>
      <span className="flex items-center gap-2">{children}</span>
    </div>
  );
}

function CopyBtn({ value }: { value: string }) {
  return (
    <button
      onClick={() => navigator.clipboard?.writeText(value)}
      className="text-muted hover:text-primary"
      aria-label="Copy"
    >
      ⧉
    </button>
  );
}

function CredRow({ field }: { field: CredField }) {
  const [show, setShow] = useState(false);
  const masked = field.secret && !show;
  return (
    <Row label={field.label}>
      <code className="break-all text-fg">
        {masked
          ? "•".repeat(Math.min(field.value.length || 8, 16))
          : field.value || "—"}
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
      {field.value && <CopyBtn value={field.value} />}
    </Row>
  );
}
