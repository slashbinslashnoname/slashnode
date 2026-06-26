"use client";

import { createContext, useCallback, useContext, useRef, useState } from "react";
import { ConsoleWindow } from "./ConsoleWindow";
import { LogsWindow } from "@/components/windows/LogsWindow";
import { ConfigWindow } from "@/components/windows/ConfigWindow";

type Win =
  | { id: number; kind: "console"; container: string }
  | { id: number; kind: "logs"; appId: string; name: string }
  | { id: number; kind: "config"; appId: string; name: string };

const Ctx = createContext<{
  open: (container: string) => void;
  openLogs: (appId: string, name: string) => void;
  openConfig: (appId: string, name: string) => void;
}>({ open: () => {}, openLogs: () => {}, openConfig: () => {} });

export function useConsole() {
  return useContext(Ctx);
}

// ConsoleProvider hosts the floating windows (container consoles, app logs, app
// config). It lives in the root layout, so windows persist across page
// navigation — you find them back when you change pages — and each clamps itself
// into the visible viewport.
export function ConsoleProvider({ children }: { children: React.ReactNode }) {
  const [wins, setWins] = useState<Win[]>([]);
  const idRef = useRef(0);

  const open = useCallback((container: string) => {
    setWins((w) => [...w, { id: ++idRef.current, kind: "console", container }]);
  }, []);
  // Logs/config are single-instance per app (re-opening focuses the existing one
  // by leaving it in place rather than stacking duplicates).
  const openLogs = useCallback((appId: string, name: string) => {
    setWins((w) =>
      w.some((x) => x.kind === "logs" && x.appId === appId)
        ? w
        : [...w, { id: ++idRef.current, kind: "logs", appId, name }],
    );
  }, []);
  const openConfig = useCallback((appId: string, name: string) => {
    setWins((w) =>
      w.some((x) => x.kind === "config" && x.appId === appId)
        ? w
        : [...w, { id: ++idRef.current, kind: "config", appId, name }],
    );
  }, []);
  const close = (id: number) => setWins((w) => w.filter((x) => x.id !== id));

  return (
    <Ctx.Provider value={{ open, openLogs, openConfig }}>
      {children}
      {wins.map((w, i) => {
        if (w.kind === "console") {
          return (
            <ConsoleWindow key={w.id} container={w.container} index={i} onClose={() => close(w.id)} />
          );
        }
        if (w.kind === "logs") {
          return (
            <LogsWindow key={w.id} id={w.appId} name={w.name} index={i} onClose={() => close(w.id)} />
          );
        }
        return (
          <ConfigWindow key={w.id} id={w.appId} name={w.name} index={i} onClose={() => close(w.id)} />
        );
      })}
    </Ctx.Provider>
  );
}
