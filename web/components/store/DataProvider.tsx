"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
} from "react";
import type { App, ProbeResult, ServiceStatus, UpdateInfo } from "@/lib/api";

// DataProvider is the app-wide data store. It lives in the root layout, so it
// stays mounted across page navigations: the apps list, their live status, the
// BTC price and the update check are polled here ONCE and shared, instead of
// each page/tile refetching on every navigation (which caused the reload flash).

type AppStatus = { services: ServiceStatus[]; docker: boolean; onion?: string };

type DataCtx = {
  apps: App[];
  ready: boolean; // true once the apps list has loaded at least once
  btc: string | null;
  update: UpdateInfo | null;
  status: Record<string, AppStatus>;
  probe: Record<string, ProbeResult>;
  imgUpdate: Record<string, boolean>;
  refresh: () => void;
};

const Ctx = createContext<DataCtx | null>(null);

export function useData(): DataCtx {
  const c = useContext(Ctx);
  if (!c) throw new Error("useData must be used within <DataProvider>");
  return c;
}

export function DataProvider({ children }: { children: React.ReactNode }) {
  const [apps, setApps] = useState<App[]>([]);
  const [ready, setReady] = useState(false);
  const [btc, setBtc] = useState<string | null>(null);
  const [update, setUpdate] = useState<UpdateInfo | null>(null);
  const [status, setStatus] = useState<Record<string, AppStatus>>({});
  const [probe, setProbe] = useState<Record<string, ProbeResult>>({});
  const [imgUpdate, setImgUpdate] = useState<Record<string, boolean>>({});

  const loadApps = useCallback(() => {
    fetch("/api/apps", { cache: "no-store" })
      .then((r) => r.json())
      .then((j) => Array.isArray(j?.apps) && setApps(j.apps))
      .catch(() => {})
      .finally(() => setReady(true));
  }, []);

  // Apps catalog + install state.
  useEffect(() => {
    loadApps();
    const t = setInterval(loadApps, 10_000);
    return () => clearInterval(t);
  }, [loadApps]);

  // Update check (also drives the dashboard's update banner).
  useEffect(() => {
    const load = () =>
      fetch("/api/update", { cache: "no-store" })
        .then((r) => r.json())
        .then((j) => j && setUpdate(j))
        .catch(() => {});
    load();
    const t = setInterval(load, 60_000);
    return () => clearInterval(t);
  }, []);

  // Live BTC/USD spot price.
  useEffect(() => {
    let alive = true;
    const load = () =>
      fetch("https://api.coinbase.com/v2/prices/BTC-USD/spot")
        .then((r) => r.json())
        .then((j) => {
          const a = Number(j?.data?.amount);
          if (alive && a) setBtc(a.toLocaleString("en-US", { maximumFractionDigits: 0 }));
        })
        .catch(() => {});
    load();
    const t = setInterval(load, 60_000);
    return () => {
      alive = false;
      clearInterval(t);
    };
  }, []);

  // Per-installed-app live status (running state + onion), polled fast.
  const installedKey = apps.filter((a) => a.installed).map((a) => a.id).join(",");
  useEffect(() => {
    const ids = installedKey ? installedKey.split(",") : [];
    if (!ids.length) return;
    let alive = true;
    const poll = () => {
      for (const id of ids) {
        fetch(`/api/apps/${id}/status`, { cache: "no-store" })
          .then((r) => r.json())
          .then((s) =>
            alive &&
            setStatus((p) => ({
              ...p,
              [id]: { services: s.services ?? [], docker: s.docker ?? false, onion: s.onion },
            })),
          )
          .catch(() => {});
      }
    };
    poll();
    const t = setInterval(poll, 5_000);
    return () => {
      alive = false;
      clearInterval(t);
    };
  }, [installedKey]);

  // Per-app probe (apps that declare one), polled slower.
  const probeKey = apps.filter((a) => a.installed && a.probe).map((a) => a.id).join(",");
  useEffect(() => {
    const ids = probeKey ? probeKey.split(",") : [];
    if (!ids.length) return;
    let alive = true;
    const poll = () => {
      for (const id of ids) {
        fetch(`/api/apps/${id}/probe`)
          .then((r) => r.json())
          .then((pr) => alive && setProbe((p) => ({ ...p, [id]: pr })))
          .catch(() => {});
      }
    };
    poll();
    const t = setInterval(poll, 15_000);
    return () => {
      alive = false;
      clearInterval(t);
    };
  }, [probeKey]);

  // Docker image-update check per installed app (cheap to keep fresh-ish).
  useEffect(() => {
    const ids = installedKey ? installedKey.split(",") : [];
    if (!ids.length) return;
    let alive = true;
    const poll = () => {
      for (const id of ids) {
        fetch(`/api/apps/${id}/image-update`)
          .then((r) => r.json())
          .then((j) => alive && setImgUpdate((p) => ({ ...p, [id]: !!j.available })))
          .catch(() => {});
      }
    };
    poll();
    const t = setInterval(poll, 300_000);
    return () => {
      alive = false;
      clearInterval(t);
    };
  }, [installedKey]);

  return (
    <Ctx.Provider value={{ apps, ready, btc, update, status, probe, imgUpdate, refresh: loadApps }}>
      {children}
    </Ctx.Provider>
  );
}
