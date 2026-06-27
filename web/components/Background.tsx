"use client";

import dynamic from "next/dynamic";
import { useEffect, useState } from "react";
import { APOLLO_SRC, BG_EVENT, getBg, getCustom, type BgKind } from "@/lib/background";

// The Starmind scene pulls in three.js — load it as its own client chunk only
// when that background is actually selected.
const Starmind = dynamic(() => import("@/components/Starmind").then((m) => m.Starmind), {
  ssr: false,
});

// Background renders the selected decorative backdrop as a fixed full-screen
// layer behind all content. It reacts to changes from the theme picker (a
// same-window custom event) and from other tabs (the storage event).
export function Background() {
  const [kind, setKind] = useState<BgKind>("none");
  const [custom, setCustom] = useState<string | null>(null);

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

  if (kind === "none") return null;

  const image = kind === "apollo" ? APOLLO_SRC : kind === "custom" ? custom : null;

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
