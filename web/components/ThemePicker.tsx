"use client";

import { useEffect, useRef, useState } from "react";
import { Starmind } from "@/components/Starmind";
import {
  BG_EVENT,
  getBg,
  getCustom,
  readAndDownscale,
  setBg,
  setCustom,
  type BgKind,
} from "@/lib/background";

// ThemePicker sits to the right of the dark/light toggle. Clicking it opens a
// popover of small live previews of each background (none, Starmind animation,
// NASA image of the day, custom upload), letting the operator pick or upload.
export function ThemePicker() {
  const [open, setOpen] = useState(false);
  const [kind, setKind] = useState<BgKind>("none");
  const [custom, setCustomState] = useState<string | null>(null);
  const [nasa, setNasa] = useState<string | null>(null);
  const [error, setError] = useState("");
  const wrap = useRef<HTMLDivElement>(null);
  const fileRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    const sync = () => {
      setKind(getBg());
      setCustomState(getCustom());
    };
    sync();
    window.addEventListener(BG_EVENT, sync);
    return () => window.removeEventListener(BG_EVENT, sync);
  }, []);

  useEffect(() => {
    if (!open) return;
    function onDoc(e: MouseEvent) {
      if (wrap.current && !wrap.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    document.addEventListener("mousedown", onDoc);
    return () => document.removeEventListener("mousedown", onDoc);
  }, [open]);

  // Lazily fetch the NASA thumbnail once the popover opens.
  useEffect(() => {
    if (open && !nasa) {
      fetch("/api/nasa-bg")
        .then((r) => r.json())
        .then((j) => j?.url && setNasa(j.url))
        .catch(() => {});
    }
  }, [open, nasa]);

  function choose(k: BgKind) {
    setBg(k);
  }

  async function onFile(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    e.target.value = "";
    if (!file) return;
    setError("");
    try {
      const dataUrl = await readAndDownscale(file);
      setCustom(dataUrl);
      setBg("custom");
    } catch {
      setError("Could not load that image.");
    }
  }

  return (
    <div ref={wrap} className="relative">
      <button
        onClick={() => setOpen((o) => !o)}
        aria-label="Background"
        className="rounded-lg border border-border bg-card px-3 py-1.5 text-sm hover:border-primary transition-colors"
      >
        🖼 theme
      </button>

      {open && (
        <div className="absolute right-0 mt-2 w-64 rounded-xl border border-border bg-card p-3 shadow-xl">
          <div className="mb-2 text-xs font-semibold text-muted">Background</div>
          <div className="grid grid-cols-2 gap-2">
            <Tile label="None" active={kind === "none"} onClick={() => choose("none")}>
              <div className="h-full w-full bg-bg" />
            </Tile>

            <Tile
              label="Starmind"
              active={kind === "starmind"}
              onClick={() => choose("starmind")}
            >
              <Starmind mini className="h-full w-full" />
            </Tile>

            <Tile label="NASA" active={kind === "nasa"} onClick={() => choose("nasa")}>
              {nasa ? (
                <div
                  className="h-full w-full bg-cover bg-center"
                  style={{ backgroundImage: `url("${nasa}")` }}
                />
              ) : (
                <div className="flex h-full w-full items-center justify-center text-[10px] text-muted">
                  loading…
                </div>
              )}
            </Tile>

            <Tile
              label="Custom"
              active={kind === "custom"}
              onClick={() => fileRef.current?.click()}
            >
              {custom ? (
                <div
                  className="h-full w-full bg-cover bg-center"
                  style={{ backgroundImage: `url("${custom}")` }}
                />
              ) : (
                <div className="flex h-full w-full items-center justify-center text-[10px] text-muted">
                  ↑ upload
                </div>
              )}
            </Tile>
          </div>

          {custom && kind !== "custom" && (
            <button
              onClick={() => choose("custom")}
              className="mt-2 w-full rounded-md border border-border py-1 text-xs hover:border-primary"
            >
              use uploaded image
            </button>
          )}
          {error && <p className="mt-2 text-xs text-primary">{error}</p>}

          <input
            ref={fileRef}
            type="file"
            accept="image/*"
            onChange={onFile}
            className="hidden"
          />
        </div>
      )}
    </div>
  );
}

function Tile({
  label,
  active,
  onClick,
  children,
}: {
  label: string;
  active: boolean;
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <button
      onClick={onClick}
      className={`group flex flex-col gap-1 rounded-lg border p-1 text-left transition-colors ${
        active ? "border-primary" : "border-border hover:border-primary/60"
      }`}
    >
      <span className="h-14 w-full overflow-hidden rounded-md border border-border bg-bg">
        {children}
      </span>
      <span
        className={`px-0.5 text-[11px] ${active ? "text-primary" : "text-muted"}`}
      >
        {label}
      </span>
    </button>
  );
}
