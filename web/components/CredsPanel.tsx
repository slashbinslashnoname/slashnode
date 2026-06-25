"use client";

import { useEffect, useState } from "react";
import type { CredField } from "@/lib/api";

// CredsPanel shows an installed app's stored parameters (rpc user, passwords…)
// and its exposed endpoints, in two tabs, with reveal + copy.
export function CredsPanel({ id }: { id: string }) {
  const [data, setData] = useState<{
    fields: CredField[];
    exports: Record<string, string>;
  } | null>(null);
  const [tab, setTab] = useState<"params" | "config">("params");

  useEffect(() => {
    fetch(`/api/apps/${id}/credentials`)
      .then((r) => r.json())
      .then((j) => setData({ fields: j.fields ?? [], exports: j.exports ?? {} }))
      .catch(() => setData({ fields: [], exports: {} }));
  }, [id]);

  if (!data) return null;

  const exportRows = Object.entries(data.exports);

  return (
    <div className="flex flex-col gap-2 rounded-lg border border-border bg-card p-3 text-sm">
      <div className="flex gap-1 text-xs">
        <Tab active={tab === "params"} onClick={() => setTab("params")}>
          Parameters
        </Tab>
        <Tab active={tab === "config"} onClick={() => setTab("config")}>
          Config
        </Tab>
      </div>

      {tab === "params" ? (
        data.fields.length > 0 ? (
          <div className="flex flex-col gap-1">
            {data.fields.map((c) => (
              <CredRow key={c.key} field={c} />
            ))}
          </div>
        ) : (
          <span className="text-xs text-muted">No parameters.</span>
        )
      ) : exportRows.length > 0 ? (
        <div className="flex flex-col gap-1">
          {exportRows.map(([k, v]) => (
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
      ) : (
        <span className="text-xs text-muted">Nothing exposed.</span>
      )}
    </div>
  );
}

function Tab({
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
      className={`rounded-md px-2 py-1 ${
        active ? "bg-primary text-white" : "text-muted hover:text-fg"
      }`}
    >
      {children}
    </button>
  );
}

function CredRow({ field }: { field: CredField }) {
  const [show, setShow] = useState(false);
  const masked = field.secret && !show;
  return (
    <div className="flex items-center justify-between gap-3 text-xs">
      <span className="text-muted">{field.label}</span>
      <span className="flex items-center gap-2">
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
