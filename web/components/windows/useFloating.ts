"use client";

import { useCallback, useEffect, useRef, useState } from "react";

let zCounter = 50;

type XY = { x: number; y: number };
type WH = { w: number; h: number };

// useFloating drives a draggable/resizable floating window and keeps it inside
// the viewport: it clamps the position on drag, on resize of the window, and on
// browser resize — so a window is never left partly or fully off-screen (e.g.
// after navigating to a smaller page or rotating/resizing the viewport).
export function useFloating(index: number, initial: WH = { w: 560, h: 360 }) {
  const [pos, setPos] = useState<XY>(() => ({
    x: 90 + index * 28,
    y: 90 + index * 28,
  }));
  const [size, setSize] = useState<WH>(initial);
  const [z, setZ] = useState(() => ++zCounter);
  const sizeRef = useRef(size);
  sizeRef.current = size;

  const bringFront = useCallback(() => setZ(++zCounter), []);

  const clamp = useCallback((p: XY, s: WH): XY => {
    if (typeof window === "undefined") return p;
    const maxX = Math.max(0, window.innerWidth - s.w);
    const maxY = Math.max(0, window.innerHeight - s.h);
    return {
      x: Math.min(Math.max(0, p.x), maxX),
      y: Math.min(Math.max(0, p.y), maxY),
    };
  }, []);

  // Re-clamp on mount and whenever the viewport changes.
  useEffect(() => {
    const reclamp = () => setPos((p) => clamp(p, sizeRef.current));
    reclamp();
    window.addEventListener("resize", reclamp);
    return () => window.removeEventListener("resize", reclamp);
  }, [clamp]);

  function startDrag(e: React.PointerEvent) {
    e.preventDefault();
    bringFront();
    const sx = e.clientX,
      sy = e.clientY,
      o = pos;
    const move = (ev: PointerEvent) =>
      setPos(clamp({ x: o.x + ev.clientX - sx, y: o.y + ev.clientY - sy }, sizeRef.current));
    const up = () => {
      window.removeEventListener("pointermove", move);
      window.removeEventListener("pointerup", up);
    };
    window.addEventListener("pointermove", move);
    window.addEventListener("pointerup", up);
  }

  function startResize(e: React.PointerEvent) {
    e.preventDefault();
    e.stopPropagation();
    bringFront();
    const sx = e.clientX,
      sy = e.clientY,
      o = size;
    const move = (ev: PointerEvent) => {
      const w = Math.max(300, Math.min(o.w + ev.clientX - sx, window.innerWidth));
      const h = Math.max(180, Math.min(o.h + ev.clientY - sy, window.innerHeight));
      setSize({ w, h });
      setPos((p) => clamp(p, { w, h }));
    };
    const up = () => {
      window.removeEventListener("pointermove", move);
      window.removeEventListener("pointerup", up);
    };
    window.addEventListener("pointermove", move);
    window.addEventListener("pointerup", up);
  }

  return { pos, size, z, bringFront, startDrag, startResize };
}
