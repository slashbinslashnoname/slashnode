"use client";

import { useEffect, useRef, useState } from "react";

const inputCls =
  "rounded-lg border border-border bg-bg px-3 py-2 text-sm outline-none focus:border-primary";
const btnPrimary =
  "cursor-pointer rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-white disabled:cursor-default disabled:opacity-60";
const btnGhost =
  "cursor-pointer rounded-lg border border-border px-3 py-2 text-sm hover:border-primary disabled:cursor-default disabled:opacity-60";

type Kind = "local" | "s3" | "sftp";

type Cfg = {
  kind?: Kind;
  prefix?: string;
  provider?: string;
  endpoint?: string;
  region?: string;
  bucket?: string;
  host?: string;
  port?: number;
  user?: string;
  has_secret?: boolean;
  all?: boolean;
  configured?: boolean;
  last_run?: string;
  last_result?: string;
};

// BackupPanel configures an rclone backup destination (local mount, S3 or SFTP)
// and runs incremental backup/restore, streaming live progress. Credentials are
// stored on the node (0600) and never sent back to the browser.
export function BackupPanel() {
  const [kind, setKind] = useState<Kind>("local");
  const [prefix, setPrefix] = useState("");
  const [provider, setProvider] = useState("");
  const [endpoint, setEndpoint] = useState("");
  const [region, setRegion] = useState("");
  const [bucket, setBucket] = useState("");
  const [accessKey, setAccessKey] = useState("");
  const [secretKey, setSecretKey] = useState("");
  const [host, setHost] = useState("");
  const [port, setPort] = useState("");
  const [user, setUser] = useState("");
  const [pass, setPass] = useState("");
  const [all, setAll] = useState(false);
  const [hasSecret, setHasSecret] = useState(false);
  const [meta, setMeta] = useState<Cfg>({});

  const [save, setSave] = useState<"idle" | "saving" | "saved" | "error">("idle");
  const [test, setTest] = useState<string>("");
  const [log, setLog] = useState("");
  const [busy, setBusy] = useState<"" | "backup" | "restore">("");
  const [confirm, setConfirm] = useState("");
  const preRef = useRef<HTMLPreElement>(null);

  useEffect(() => {
    fetch("/api/backup/config")
      .then((r) => r.json())
      .then((c: Cfg) => {
        setKind((c.kind as Kind) || "local");
        setPrefix(c.prefix ?? "");
        setProvider(c.provider ?? "");
        setEndpoint(c.endpoint ?? "");
        setRegion(c.region ?? "");
        setBucket(c.bucket ?? "");
        setHost(c.host ?? "");
        setPort(c.port ? String(c.port) : "");
        setUser(c.user ?? "");
        setAll(!!c.all);
        setHasSecret(!!c.has_secret);
        setMeta(c);
      })
      .catch(() => {});
  }, []);

  useEffect(() => {
    if (preRef.current) preRef.current.scrollTop = preRef.current.scrollHeight;
  }, [log]);

  function destination() {
    return {
      kind,
      prefix,
      provider,
      endpoint,
      region,
      bucket,
      access_key: accessKey,
      secret_key: secretKey,
      host,
      port: port ? Number(port) : 0,
      user,
      pass,
    };
  }

  async function saveConfig() {
    setSave("saving");
    try {
      const res = await fetch("/api/backup/config", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ destination: destination(), all }),
      });
      setSave(res.ok ? "saved" : "error");
      if (res.ok) {
        setSecretKey("");
        setPass("");
        setHasSecret(hasSecret || !!secretKey || !!pass || kind === "local");
      }
    } catch {
      setSave("error");
    }
  }

  async function runTest() {
    setTest("…");
    try {
      const res = await fetch("/api/backup/test", { method: "POST" });
      const j = await res.json();
      setTest(j.ok ? "✓ reachable" : `✗ ${j.error || "failed"}`);
    } catch {
      setTest("✗ failed");
    }
  }

  async function stream(path: string, kindBusy: "backup" | "restore") {
    setBusy(kindBusy);
    setLog("");
    try {
      const res = await fetch(path, { method: "POST" });
      const reader = res.body?.getReader();
      const dec = new TextDecoder();
      while (reader) {
        const { done, value } = await reader.read();
        if (done) break;
        setLog((l) => l + dec.decode(value, { stream: true }));
      }
    } catch {
      setLog((l) => l + "\n[connection lost]");
    }
    setBusy("");
  }

  return (
    <div className="flex flex-col gap-4">
      <p className="text-xs text-muted">
        Incrementally syncs your node — config, secrets, <code>.onion</code> keys
        and app data volumes — to a destination via rclone. Large re-syncable
        chains (Bitcoin, Monero…) are skipped unless “include chains” is on.
        Credentials are stored on the node (mode 0600) and never leave it.
      </p>

      <SelectRow
        label="Destination type"
        value={kind}
        onChange={(v) => setKind(v as Kind)}
        options={["local", "s3", "sftp"]}
      />

      {kind === "local" && (
        <Field label="Path on the node (USB / NFS mount)">
          <input value={prefix} placeholder="/mnt/usb/slashnode" onChange={(e) => setPrefix(e.target.value)} className={inputCls} />
        </Field>
      )}

      {kind === "s3" && (
        <div className="flex flex-col gap-3">
          <Field label="Bucket"><input value={bucket} placeholder="my-bucket" onChange={(e) => setBucket(e.target.value)} className={inputCls} /></Field>
          <Field label="Path prefix (in bucket)"><input value={prefix} placeholder="slashnode" onChange={(e) => setPrefix(e.target.value)} className={inputCls} /></Field>
          <Field label="Provider (AWS, Minio, Other…)"><input value={provider} placeholder="Other" onChange={(e) => setProvider(e.target.value)} className={inputCls} /></Field>
          <Field label="Endpoint (S3-compatible; blank for AWS)"><input value={endpoint} placeholder="https://s3.example.com" onChange={(e) => setEndpoint(e.target.value)} className={inputCls} /></Field>
          <Field label="Region"><input value={region} placeholder="us-east-1" onChange={(e) => setRegion(e.target.value)} className={inputCls} /></Field>
          <Field label="Access key"><input value={accessKey} onChange={(e) => setAccessKey(e.target.value)} className={inputCls} /></Field>
          <Field label="Secret key"><input type="password" value={secretKey} placeholder={hasSecret ? "•••••• (unchanged)" : ""} onChange={(e) => setSecretKey(e.target.value)} className={inputCls} /></Field>
        </div>
      )}

      {kind === "sftp" && (
        <div className="flex flex-col gap-3">
          <Field label="Host"><input value={host} placeholder="backup.example.com" onChange={(e) => setHost(e.target.value)} className={inputCls} /></Field>
          <Field label="Port"><input value={port} placeholder="22" onChange={(e) => setPort(e.target.value)} className={inputCls} /></Field>
          <Field label="User"><input value={user} onChange={(e) => setUser(e.target.value)} className={inputCls} /></Field>
          <Field label="Password"><input type="password" value={pass} placeholder={hasSecret ? "•••••• (unchanged)" : ""} onChange={(e) => setPass(e.target.value)} className={inputCls} /></Field>
          <Field label="Remote path"><input value={prefix} placeholder="slashnode" onChange={(e) => setPrefix(e.target.value)} className={inputCls} /></Field>
        </div>
      )}

      <label className="flex items-center gap-2 text-sm">
        <input type="checkbox" checked={all} onChange={(e) => setAll(e.target.checked)} />
        Include large chain volumes (Bitcoin, Monero, Geth, Electrs…)
      </label>

      <div className="flex flex-wrap items-center gap-3">
        <button onClick={saveConfig} disabled={save === "saving"} className={btnPrimary}>
          {save === "saving" ? "saving…" : save === "saved" ? "✓ saved" : "Save destination"}
        </button>
        <button onClick={runTest} className={btnGhost}>Test</button>
        {test && <span className="text-sm text-muted">{test}</span>}
      </div>

      {meta.last_run && (
        <p className="text-xs text-muted">Last backup: {meta.last_run} ({meta.last_result})</p>
      )}

      <div className="flex flex-wrap items-center gap-3 border-t border-border pt-4">
        <button onClick={() => stream(`/api/backup/run${all ? "?all=true" : ""}`, "backup")} disabled={!!busy} className={btnPrimary}>
          {busy === "backup" ? "backing up…" : "Back up now"}
        </button>
        <div className="flex items-center gap-2">
          <input
            value={confirm}
            placeholder='type "restore"'
            onChange={(e) => setConfirm(e.target.value)}
            className={`${inputCls} w-32`}
          />
          <button
            onClick={() => stream("/api/restore", "restore")}
            disabled={!!busy || confirm !== "restore"}
            className={`cursor-pointer rounded-lg border border-primary/50 px-3 py-2 text-sm text-primary hover:bg-primary/10 disabled:cursor-default disabled:opacity-60`}
          >
            {busy === "restore" ? "restoring…" : "Restore"}
          </button>
        </div>
      </div>
      <p className="text-xs text-muted">
        Restore is destructive — it overwrites this node’s state and volumes, then
        brings the apps back up. Best run on a freshly-initialised node.
      </p>

      {log && (
        <pre ref={preRef} className="max-h-72 overflow-auto rounded-lg bg-bg p-3 text-xs leading-relaxed">
          {log}
        </pre>
      )}
    </div>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="flex flex-col gap-1">
      <span className="text-sm font-medium">{label}</span>
      {children}
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
    <label className="flex items-center justify-between gap-3">
      <span className="text-sm font-medium">{label}</span>
      <select
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="rounded-lg border border-border bg-bg px-3 py-2 text-sm outline-none focus:border-primary"
      >
        {options.map((o) => (
          <option key={o} value={o}>{o}</option>
        ))}
      </select>
    </label>
  );
}
