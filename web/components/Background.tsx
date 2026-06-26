"use client";

import { useEffect, useState } from "react";
import { Starmind } from "@/components/Starmind";
import { BG_EVENT, getBg, getCustom, type BgKind } from "@/lib/background";

// Background renders the selected decorative backdrop as a fixed full-screen
// layer behind all content. It reacts to changes from the theme picker (a
// same-window custom event) and from other tabs (the storage event).
export function Background() {
  const [kind, setKind] = useState<BgKind>("none");
  const [custom, setCustom] = useState<string | null>(null);
  const [nasa, setNasa] = useState<string | null>(null);

  useEffect(() => {
    const sync = () => {
      setKind(getBg());
      setCustom(getCustom());
    };
    sync();
    window.addEventListener(BG_EVENT, sync);
    window.addEventListener("storage", sync);
    return () => {
      window.removeEventListener(BG_EVENT, sync);
      window.removeEventListener("storage", sync);
    };
  }, []);

  useEffect(() => {
    if (kind === "nasa" && !nasa) {
      fetch("/api/nasa-bg")
        .then((r) => r.json())
        .then((j) => j?.url && setNasa(j.url))
        .catch(() => {});
    }
  }, [kind, nasa]);

  if (kind === "none") return null;

  const image = kind === "nasa" ? nasa : kind === "custom" ? custom : null;

  return (
    <div className="fixed inset-0 -z-10 overflow-hidden" aria-hidden>
      {kind === "starmind" && <Starmind className="h-full w-full" />}
      {image && (
        <>
          <div
            className="absolute inset-0 bg-cover bg-center"
            style={{ backgroundImage: `url("${image}")` }}
          />
          {/* Scrim so foreground text/cards stay legible over the photo. */}
          <div className="absolute inset-0 bg-bg/60" />
        </>
      )}
    </div>
  );
}
