"use client";

import { useEffect, useState } from "react";
import type { SystemStats } from "@/lib/api";

function fmt(n: number): string {
  if (!n) return "0 B";
  const u = ["B", "KB", "MB", "GB", "TB"];
  let i = 0;
  let v = n;
  while (v >= 1024 && i < u.length - 1) {
    v /= 1024;
    i++;
  }
  return `${v.toFixed(v < 10 && i > 0 ? 1 : 0)} ${u[i]}`;
}

function useSystem(intervalMs = 30000) {
  const [sys, setSys] = useState<SystemStats | null>(null);
  useEffect(() => {
    let alive = true;
    const load = () =>
      fetch("/api/system", { cache: "no-store" })
        .then((r) => r.json())
        .then((j) => alive && j && typeof j.disk === "object" && setSys(j))
        .catch(() => {});
    load();
    const t = setInterval(load, intervalMs);
    return () => {
      alive = false;
      clearInterval(t);
    };
  }, [intervalMs]);
  return sys;
}

// StorageBanner shows a prominent warning only when the disk is running low; it
// renders nothing otherwise. Placed on the dashboard.
export function StorageBanner() {
  const sys = useSystem();
  if (!sys || !sys.disk_warn) return null;
  const critical = sys.disk_critical;
  return (
    <div
      className={`mb-6 rounded-xl border px-4 py-3 text-sm ${
        critical
          ? "border-primary bg-primary/15"
          : "border-amber-500/50 bg-amber-500/10"
      }`}
    >
      <span className={critical ? "font-semibold text-primary" : "font-semibold text-amber-500"}>
        {critical ? "⚠ Storage almost full" : "⚠ Storage running low"}
      </span>{" "}
      <span className="text-muted">
        {sys.disk.percent}% used · {fmt(sys.disk.free)} free. Free space by pruning
        images (Settings → Maintenance) or removing apps.
      </span>
    </div>
  );
}

function Bar({ percent }: { percent: number }) {
  const color =
    percent >= 95 ? "bg-primary" : percent >= 85 ? "bg-amber-500" : "bg-primary/60";
  return (
    <div className="h-1.5 w-full overflow-hidden rounded-full bg-border">
      <div className={`h-full ${color}`} style={{ width: `${Math.min(100, percent)}%` }} />
    </div>
  );
}

// SystemIndicators shows disk / memory / load gauges. Placed in Settings.
export function SystemIndicators() {
  const sys = useSystem(15000);
  if (!sys) return <p className="text-sm text-muted">Loading…</p>;
  return (
    <div className="flex flex-col gap-3">
      <Metric
        label="Disk"
        percent={sys.disk.percent}
        detail={`${fmt(sys.disk.used)} / ${fmt(sys.disk.total)} · ${fmt(sys.disk.free)} free`}
      />
      {sys.mem_total > 0 && (
        <Metric
          label="Memory"
          percent={sys.mem_percent}
          detail={`${fmt(sys.mem_used)} / ${fmt(sys.mem_total)}`}
        />
      )}
      {sys.load1 > 0 && (
        <div className="flex items-center justify-between text-sm">
          <span className="text-muted">Load (1m)</span>
          <code className="text-fg">{sys.load1.toFixed(2)}</code>
        </div>
      )}
    </div>
  );
}

function Metric({ label, percent, detail }: { label: string; percent: number; detail: string }) {
  return (
    <div className="flex flex-col gap-1">
      <div className="flex items-center justify-between text-sm">
        <span className="text-muted">
          {label} <span className="text-fg">{percent}%</span>
        </span>
        <span className="text-xs text-muted">{detail}</span>
      </div>
      <Bar percent={percent} />
    </div>
  );
}
