"use client";

import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import type { App, AppInput } from "@/lib/api";

export function InstallForm({ app }: { app: App }) {
  const router = useRouter();
  const [values, setValues] = useState<Record<string, string>>(() =>
    Object.fromEntries(
      (app.inputs ?? []).map((i) => [
        i.key,
        i.default !== undefined ? String(i.default) : "",
      ]),
    ),
  );
  const [state, setState] = useState<"idle" | "saving" | "done" | "error">(
    "idle",
  );
  const [error, setError] = useState("");
  const [output, setOutput] = useState("");
  const [showOutput, setShowOutput] = useState(false);
  const preRef = useRef<HTMLPreElement>(null);

  function set(key: string, v: string) {
    setValues((prev) => ({ ...prev, [key]: v }));
  }

  // Keep the live log scrolled to the bottom as output streams in.
  useEffect(() => {
    if (preRef.current) preRef.current.scrollTop = preRef.current.scrollHeight;
  }, [output]);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setState("saving");
    setError("");
    setOutput("");
    setShowOutput(true);
    try {
      const res = await fetch(`/api/apps/${app.id}/install-stream`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ inputs: values }),
      });
      if (!res.ok || !res.body) {
        const body = await res.text().catch(() => "");
        setError(body || "install failed");
        setState("error");
        return;
      }
      const reader = res.body.getReader();
      const decoder = new TextDecoder();
      let acc = "";
      for (;;) {
        const { done, value } = await reader.read();
        if (done) break;
        acc += decoder.decode(value, { stream: true });
        setOutput(acc);
      }
      if (acc.includes("INSTALL FAILED")) {
        setError("install failed — see output below");
        setState("error");
      } else {
        setState("done");
      }
      router.refresh();
    } catch {
      setError("daemon unreachable");
      setState("error");
    }
  }

  return (
    <form onSubmit={submit} className="flex flex-col gap-4">
      {(app.inputs ?? []).map((input) => (
        <Field key={input.key} input={input} value={values[input.key] ?? ""} onChange={set} />
      ))}

      {error && <p className="text-sm text-primary">{error}</p>}

      <div className="flex items-center gap-3">
        <button
          type="submit"
          disabled={state === "saving"}
          className="rounded-lg bg-primary px-5 py-2.5 font-semibold text-white disabled:opacity-60"
        >
          {state === "idle" && (app.installed ? "Reconfigure & launch" : "Install & launch")}
          {state === "saving" && "Launching…"}
          {state === "done" && "✓ Launched"}
          {state === "error" && "Retry"}
        </button>
        {output && (
          <button
            type="button"
            onClick={() => setShowOutput((s) => !s)}
            className="text-sm text-muted hover:text-primary"
          >
            {showOutput ? "hide output" : "show output"}
          </button>
        )}
      </div>

      {showOutput && output && (
        <pre
          ref={preRef}
          className="max-h-72 overflow-auto rounded-lg bg-bg p-3 text-xs leading-relaxed"
        >
          {output}
        </pre>
      )}
    </form>
  );
}

function Field({
  input,
  value,
  onChange,
}: {
  input: AppInput;
  value: string;
  onChange: (key: string, v: string) => void;
}) {
  const [show, setShow] = useState(false);
  const base =
    "rounded-lg border border-border bg-bg px-3 py-2 outline-none focus:border-primary";

  return (
    <label className="flex flex-col gap-1">
      <span className="text-sm font-medium">
        {input.label}
        {input.required && <span className="text-primary"> *</span>}
      </span>

      {input.type === "password" ? (
        <div className="flex items-stretch gap-2">
          <input
            className={`${base} flex-1`}
            type={show ? "text" : "password"}
            value={value}
            placeholder={input.placeholder}
            minLength={input.minLength}
            required={input.required}
            onChange={(e) => onChange(input.key, e.target.value)}
          />
          <button
            type="button"
            onClick={() => setShow((s) => !s)}
            aria-label={show ? "Hide" : "Show"}
            className="rounded-lg border border-border px-3 text-sm hover:border-primary"
          >
            {show ? "🙈" : "👁"}
          </button>
        </div>
      ) : input.type === "select" ? (
        <select
          className={base}
          value={value}
          onChange={(e) => onChange(input.key, e.target.value)}
        >
          {(input.options ?? []).map((o) => (
            <option key={o} value={o}>
              {o}
            </option>
          ))}
        </select>
      ) : input.type === "boolean" ? (
        <input
          type="checkbox"
          className="h-5 w-5 accent-[var(--primary)]"
          checked={value === "true"}
          onChange={(e) => onChange(input.key, String(e.target.checked))}
        />
      ) : input.type === "textarea" ? (
        <textarea
          className={base}
          value={value}
          placeholder={input.placeholder}
          onChange={(e) => onChange(input.key, e.target.value)}
        />
      ) : (
        <input
          className={base}
          type={
            input.type === "email"
              ? "email"
              : input.type === "number"
                ? "number"
                : "text"
          }
          value={value}
          placeholder={input.placeholder}
          minLength={input.minLength}
          required={input.required}
          onChange={(e) => onChange(input.key, e.target.value)}
        />
      )}

      {input.help && <span className="text-xs text-muted">{input.help}</span>}
    </label>
  );
}
