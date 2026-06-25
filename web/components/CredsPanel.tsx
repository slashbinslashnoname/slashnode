"use client";

import { useEffect, useState } from "react";
import type { CredField } from "@/lib/api";

// CredsPanel shows an installed app's stored inputs/secrets (rpc user,
// passwords…) and its exposed endpoints, with reveal + copy.
export function CredsPanel({ id }: { id: string }) {
  const [data, setData] = useState<{
    fields: CredField[];
    exports: Record<string, string>;
  } | null>(null);

  useEffect(() => {
    fetch(`/api/apps/${id}/credentials`)
      .then((r) => r.json())
      .then((j) => setData({ fields: j.fields ?? [], exports: j.exports ?? {} }))
      .catch(() => setData({ fields: [], exports: {} }));
  }, [id]);

  if (!data) return null;
  if (data.fields.length === 0 && Object.keys(data.exports).length === 0) {
    return null;
  }

  return (
    <div className="flex flex-col gap-2 rounded-lg border border-border bg-card p-4 text-sm">
      {data.fields.length > 0 && (
        <div className="flex flex-col gap-1">
          <span className="text-xs uppercase tracking-wider text-muted">config</span>
          {data.fields.map((c) => (
            <CredRow key={c.key} field={c} />
          ))}
        </div>
      )}
      {Object.keys(data.exports).length > 0 && (
        <div className="flex flex-col gap-1 border-t border-border pt-2">
          <span className="text-xs uppercase tracking-wider text-muted">exposes</span>
          {Object.entries(data.exports).map(([k, v]) => (
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
  );
}

function CredRow({ field }: { field: CredField }) {
  const [show, setShow] = useState(false);
  const masked = field.secret && !show;
  return (
    <div className="flex items-center justify-between gap-3 text-xs">
      <span className="text-muted">{field.label}</span>
      <span className="flex items-center gap-2">
        <code className="text-fg">
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
