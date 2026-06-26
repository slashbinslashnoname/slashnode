"use client";

import { useEffect, useState } from "react";
import { FloatingWindow } from "./FloatingWindow";
import { CredsPanel } from "@/components/CredsPanel";
import { EndpointsPanel } from "@/components/EndpointsPanel";
import type { App } from "@/lib/api";

// ConfigWindow is a floating panel showing an app's connection endpoints, its
// stored parameters and exposed config (credentials), mirroring the
// console/logs windows.
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
  const [app, setApp] = useState<App | null>(null);

  useEffect(() => {
    fetch(`/api/apps/${id}`, { cache: "no-store" })
      .then((r) => r.json())
      .then((j) => j && j.id && setApp(j))
      .catch(() => {});
  }, [id]);

  const hasEndpoints = !!app && ((app.endpoints?.length ?? 0) > 0 || !!app.web);

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
      <div className="flex flex-col gap-3">
        {hasEndpoints && (
          <EndpointsPanel
            endpoints={app!.endpoints ?? []}
            onion={app!.onion}
            web={app!.web}
            url={app!.url}
          />
        )}
        <CredsPanel id={id} />
      </div>
    </FloatingWindow>
  );
}
