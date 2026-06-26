"use client";

import { useEffect, useRef } from "react";

// Starmind renders a constellation of satellites (red dots) orbiting a center
// and transmitting messages (pulses traveling along links) between each other.
// Pure canvas + requestAnimationFrame; sizes itself to its parent. Used both as
// a full-screen background and as a small preview tile (mini).
export function Starmind({
  className,
  mini = false,
}: {
  className?: string;
  mini?: boolean;
}) {
  const ref = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    // Non-null assertions: in useEffect the canvas is always mounted and a 2d
    // context is universally available. Typing them non-null keeps the nested
    // resize()/frame() closures from re-widening them to null.
    const cv: HTMLCanvasElement = ref.current!;
    const c: CanvasRenderingContext2D = cv.getContext("2d")!;

    const primary =
      getComputedStyle(document.documentElement)
        .getPropertyValue("--primary")
        .trim() || "#e5484d";
    const reduced = window.matchMedia(
      "(prefers-reduced-motion: reduce)",
    ).matches;

    let w = 0;
    let h = 0;
    const dpr = Math.min(window.devicePixelRatio || 1, 2);
    function resize() {
      const r = cv.getBoundingClientRect();
      w = r.width;
      h = r.height;
      cv.width = Math.max(1, Math.round(w * dpr));
      cv.height = Math.max(1, Math.round(h * dpr));
      c.setTransform(dpr, 0, 0, dpr, 0, 0);
    }
    resize();
    const ro = new ResizeObserver(resize);
    ro.observe(cv);

    const count = mini ? 6 : 16;
    const sats = Array.from({ length: count }, (_, i) => ({
      // Spread satellites across a few orbital rings.
      orbit: 0.16 + ((i % 4) + 1) * 0.16 + Math.random() * 0.05,
      angle: Math.random() * Math.PI * 2,
      speed: (Math.random() > 0.5 ? 1 : -1) * (0.06 + Math.random() * 0.12),
      size: mini ? 1.4 : 1.6 + Math.random() * 1.4,
    }));

    type Msg = { from: number; to: number; t: number; speed: number };
    let msgs: Msg[] = [];

    function pos(i: number) {
      const cx = w / 2;
      const cy = h / 2;
      const radius = Math.min(w, h) * sats[i].orbit;
      return {
        x: cx + Math.cos(sats[i].angle) * radius,
        y: cy + Math.sin(sats[i].angle) * radius * 0.78, // slight ellipse
      };
    }

    function rgba(alpha: number) {
      // primary may be hex; convert to rgba for translucent strokes.
      const hex = primary.replace("#", "");
      if (hex.length === 6) {
        const n = parseInt(hex, 16);
        return `rgba(${(n >> 16) & 255}, ${(n >> 8) & 255}, ${n & 255}, ${alpha})`;
      }
      return primary;
    }

    let raf = 0;
    let last = 0;
    let acc = 0;
    function frame(ts: number) {
      const dt = last ? Math.min((ts - last) / 1000, 0.05) : 0;
      last = ts;
      c.clearRect(0, 0, w, h);

      // Advance orbits.
      for (const s of sats) s.angle += s.speed * dt;

      // Spawn messages between random satellites.
      acc += dt;
      const interval = mini ? 0.9 : 0.55;
      if (!reduced && acc > interval && msgs.length < (mini ? 3 : 9)) {
        acc = 0;
        const from = (Math.random() * count) | 0;
        let to = (Math.random() * count) | 0;
        if (to === from) to = (to + 1) % count;
        msgs.push({ from, to, t: 0, speed: 0.4 + Math.random() * 0.5 });
      }

      // Draw links + traveling pulses.
      msgs = msgs.filter((m) => m.t < 1);
      for (const m of msgs) {
        m.t += m.speed * dt;
        const a = pos(m.from);
        const b = pos(m.to);
        c.strokeStyle = rgba(0.12);
        c.lineWidth = 1;
        c.beginPath();
        c.moveTo(a.x, a.y);
        c.lineTo(b.x, b.y);
        c.stroke();
        const px = a.x + (b.x - a.x) * m.t;
        const py = a.y + (b.y - a.y) * m.t;
        c.fillStyle = rgba(0.9);
        c.beginPath();
        c.arc(px, py, mini ? 1.3 : 2, 0, Math.PI * 2);
        c.fill();
      }

      // Draw satellites.
      for (let i = 0; i < count; i++) {
        const p = pos(i);
        c.fillStyle = rgba(0.85);
        c.beginPath();
        c.arc(p.x, p.y, sats[i].size, 0, Math.PI * 2);
        c.fill();
        c.fillStyle = rgba(0.18);
        c.beginPath();
        c.arc(p.x, p.y, sats[i].size * 2.4, 0, Math.PI * 2);
        c.fill();
      }

      // Center node.
      c.fillStyle = rgba(0.9);
      c.beginPath();
      c.arc(w / 2, h / 2, mini ? 1.8 : 3, 0, Math.PI * 2);
      c.fill();

      raf = requestAnimationFrame(frame);
    }
    raf = requestAnimationFrame(frame);

    return () => {
      cancelAnimationFrame(raf);
      ro.disconnect();
    };
  }, [mini]);

  return <canvas ref={ref} className={className} />;
}
