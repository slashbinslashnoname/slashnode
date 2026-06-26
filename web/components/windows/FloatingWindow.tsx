"use client";

import { useFloating } from "./useFloating";

// FloatingWindow is the shared chrome for a draggable/resizable, viewport-clamped
// panel (logs, config, …). The console uses the same useFloating hook.
export function FloatingWindow({
  title,
  index,
  onClose,
  children,
}: {
  title: React.ReactNode;
  index: number;
  onClose: () => void;
  children: React.ReactNode;
}) {
  const { pos, size, z, bringFront, startDrag, startResize } = useFloating(index);

  return (
    <div
      className="fixed flex flex-col overflow-hidden rounded-xl border border-border bg-card shadow-2xl"
      style={{ left: pos.x, top: pos.y, width: size.w, height: size.h, zIndex: z }}
      onPointerDownCapture={bringFront}
    >
      <div
        onPointerDown={startDrag}
        className="flex cursor-move select-none items-center justify-between border-b border-border px-3 py-1.5"
      >
        <span className="font-mono text-xs text-fg">{title}</span>
        <button
          onPointerDown={(e) => e.stopPropagation()}
          onClick={onClose}
          aria-label="Close"
          className="cursor-pointer rounded px-2 text-muted hover:text-primary"
        >
          ✕
        </button>
      </div>
      <div className="min-h-0 flex-1 overflow-auto p-3">{children}</div>
      <div
        onPointerDown={startResize}
        className="absolute bottom-0 right-0 h-4 w-4 cursor-nwse-resize text-muted"
        style={{ lineHeight: "1rem" }}
        title="resize"
      >
        ◢
      </div>
    </div>
  );
}
