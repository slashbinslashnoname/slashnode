"use client";

import { useEffect, useRef, useState } from "react";
import {
  APOLLO_SRC,
  BG_EVENT,
  getBg,
  getCustom,
  readAndDownscale,
  setBg,
  setCustom,
  type BgKind,
} from "@/lib/background";

// ThemePicker sits to the right of the dark/light toggle. Clicking it opens a
// popover of small previews of each background (none, Starmind animation, an
// Apollo photo, custom upload), letting the operator pick or upload.
export function ThemePicker() {
  const [open, setOpen] = useState(false);
  const [kind, setKind] = useState<BgKind>("none");
  const [custom, setCustomState] = useState<string | null>(null);
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
        title="Background"
        className="cursor-pointer rounded-lg border border-border bg-card px-3 py-1.5 text-sm hover:border-primary transition-colors"
      >
        🖼
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
              <StarmindPreview />
            </Tile>

            <Tile
              label="Apollo"
              active={kind === "apollo"}
              onClick={() => choose("apollo")}
            >
              <div
                className="h-full w-full bg-cover bg-center"
                style={{ backgroundImage: `url("${APOLLO_SRC}")` }}
              />
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
              className="mt-2 w-full cursor-pointer rounded-md border border-border py-1 text-xs hover:border-primary"
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

// Static, dependency-free depiction of the Starmind background so opening the
// popover doesn't spin up a live canvas/animation (which made it feel slow).
function StarmindPreview() {
  const sats = [
    [14, 10],
    [50, 12],
    [10, 30],
    [54, 30],
    [22, 33],
    [42, 8],
  ];
  return (
    <svg
      viewBox="0 0 64 40"
      preserveAspectRatio="xMidYMid slice"
      className="h-full w-full"
    >
      {sats.map(([x, y], i) => (
        <line
          key={i}
          x1="32"
          y1="20"
          x2={x}
          y2={y}
          stroke="var(--primary)"
          strokeOpacity="0.18"
          strokeWidth="0.5"
        />
      ))}
      {sats.map(([x, y], i) => (
        <circle key={i} cx={x} cy={y} r="1.4" fill="var(--primary)" fillOpacity="0.85" />
      ))}
      <circle cx="32" cy="20" r="2" fill="var(--primary)" />
    </svg>
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
      className={`group flex cursor-pointer flex-col gap-1 rounded-lg border p-1 text-left transition-colors ${
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
