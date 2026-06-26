"use client";

import { FloatingWindow } from "./FloatingWindow";
import { CredsPanel } from "@/components/CredsPanel";

// ConfigWindow is a floating panel showing an app's stored parameters and
// exposed config (credentials), mirroring the console/logs windows.
export function ConfigWindow({
  id,
  name,
  index,
  onClose,
}: {
  id: string;
  name: string;
  index: number;
  onClose: () => void;
}) {
  return (
    <FloatingWindow
      index={index}
      onClose={onClose}
      title={
        <span>
          <span className="text-primary">⚙</span> {name} · config
        </span>
      }
    >
      <CredsPanel id={id} />
    </FloatingWindow>
  );
}
