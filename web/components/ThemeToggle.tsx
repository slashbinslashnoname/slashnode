"use client";

import { useEffect, useState } from "react";

type Mode = "system" | "light" | "dark";
const MODES: Mode[] = ["system", "light", "dark"];
const ICONS: Record<Mode, string> = {
  system: "◐",
  light: "☀",
  dark: "☾",
};

function apply(mode: Mode) {
  const dark =
    mode === "dark" ||
    (mode === "system" &&
      window.matchMedia("(prefers-color-scheme: dark)").matches);
  document.documentElement.classList.toggle("dark", dark);
  localStorage.setItem("slashnode-theme", mode);
}

export function ThemeToggle() {
  const [mode, setMode] = useState<Mode>("system");

  useEffect(() => {
    const stored = (localStorage.getItem("slashnode-theme") as Mode) || "system";
    setMode(stored);
  }, []);

  function cycle() {
    const next = MODES[(MODES.indexOf(mode) + 1) % MODES.length];
    setMode(next);
    apply(next);
  }

  return (
    <button
      onClick={cycle}
      aria-label={`Theme: ${mode}`}
      title={`Theme: ${mode}`}
      className="cursor-pointer rounded-lg border border-border bg-card px-3 py-1.5 text-sm hover:border-primary transition-colors"
    >
      {ICONS[mode]}
    </button>
  );
}
