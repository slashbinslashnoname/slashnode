"use client";

import { useState } from "react";
import type { App } from "@/lib/api";
import { InstallForm } from "@/components/InstallForm";
import { CredsPanel } from "@/components/CredsPanel";
import { VersionSelector } from "@/components/VersionSelector";

// AppTabs splits an installed app's detail into two top-level tabs:
//   - "Config" : the editable form (rpc user/password, …) → reconfigure & launch.
//   - "View"   : the read-only stored Parameters + exposed Config values.
// For an app that isn't installed yet, only the install form is shown.
export function AppTabs({ app }: { app: App }) {
  const [tab, setTab] = useState<"config" | "view">(
    app.installed ? "view" : "config",
  );

  if (!app.installed) return <InstallForm app={app} />;

  return (
    <div className="flex flex-col gap-4">
      <div className="flex gap-1 text-sm">
        <Tab active={tab === "config"} onClick={() => setTab("config")}>
          Config
        </Tab>
        <Tab active={tab === "view"} onClick={() => setTab("view")}>
          View
        </Tab>
      </div>

      {tab === "config" ? (
        <div className="flex flex-col gap-4">
          {app.images && Object.keys(app.images).length > 0 && (
            <div className="flex flex-col gap-1">
              <span className="text-sm font-medium">Image version</span>
              <VersionSelector
                id={app.id}
                images={app.images}
                suggest={app.versions}
              />
            </div>
          )}
          <InstallForm app={app} />
        </div>
      ) : (
        <CredsPanel id={app.id} endpoints={app.endpoints} />
      )}
    </div>
  );
}

function Tab({
  active,
  onClick,
  children,
}: {
  active: boolean;
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <button
      onClick={onClick}
      className={`rounded-md px-3 py-1.5 font-semibold ${
        active ? "bg-primary text-white" : "text-muted hover:text-fg"
      }`}
    >
      {children}
    </button>
  );
}
