"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import type { App } from "@/lib/api";
import { InstallForm } from "@/components/InstallForm";
import { CredsPanel } from "@/components/CredsPanel";
import { EndpointsPanel } from "@/components/EndpointsPanel";
import { VersionSelector } from "@/components/VersionSelector";
import { DomainTab } from "@/components/DomainTab";

// AppTabs splits an installed app's detail into three tabs:
//   - "Config"  : the editable form (rpc user/password, …) → reconfigure & launch.
//   - "Version" : per-service image tag picker + re-pull the current tag.
//   - "View"    : connection endpoints (clearnet + .onion) + stored credentials.
// For an app that isn't installed yet, only the install form is shown.
export function AppTabs({ app }: { app: App }) {
  const [tab, setTab] = useState<"config" | "version" | "domain" | "view">("config");

  if (!app.installed) return <InstallForm app={app} />;

  const hasImages = !!app.images && Object.keys(app.images).length > 0;

  return (
    <div className="flex flex-col gap-4">
      <div className="flex gap-1 text-sm">
        <Tab active={tab === "config"} onClick={() => setTab("config")}>
          Config
        </Tab>
        <Tab active={tab === "version"} onClick={() => setTab("version")}>
          Version
        </Tab>
        {app.web && (
          <Tab active={tab === "domain"} onClick={() => setTab("domain")}>
            Domain
          </Tab>
        )}
        <Tab active={tab === "view"} onClick={() => setTab("view")}>
          View
        </Tab>
      </div>

      {tab === "config" && <InstallForm app={app} />}

      {tab === "version" && (
        <div className="flex flex-col gap-4">
          {hasImages && (
            <div className="flex flex-col gap-1">
              <span className="text-sm font-medium">Image version</span>
              <VersionSelector id={app.id} images={app.images!} />
            </div>
          )}
          <RepullButton id={app.id} />
        </div>
      )}

      {tab === "domain" && app.web && <DomainTab app={app} />}

      {tab === "view" && (
        <div className="flex flex-col gap-4">
          <EndpointsPanel
            endpoints={app.endpoints ?? []}
            onion={app.onion}
            web={app.web}
            url={app.url}
          />
          <CredsPanel id={app.id} />
        </div>
      )}
    </div>
  );
}

// RepullButton re-pulls the images for the current tags and recreates the
// containers — useful to pick up a fresh build of a moving tag like :latest
// without changing the version.
function RepullButton({ id }: { id: string }) {
  const router = useRouter();
  const [state, setState] = useState<"idle" | "busy" | "done" | "error">("idle");

  async function repull() {
    setState("busy");
    try {
      const r = await fetch(`/api/apps/${id}/update`, { method: "POST" });
      setState(r.ok ? "done" : "error");
      if (r.ok) router.refresh();
    } catch {
      setState("error");
    }
  }

  return (
    <div className="flex flex-col gap-1">
      <button
        onClick={repull}
        disabled={state === "busy"}
        className="cursor-pointer self-start rounded-lg border border-border px-3 py-2 text-sm hover:border-primary disabled:cursor-default disabled:opacity-60"
      >
        {state === "busy"
          ? "re-pulling…"
          : state === "done"
            ? "✓ re-pulled"
            : state === "error"
              ? "failed — retry"
              : "Re-pull current version"}
      </button>
      <span className="text-xs text-muted">
        Pulls the freshest build for the current tag (e.g. :latest) and recreates
        the containers — without changing the version.
      </span>
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
      className={`cursor-pointer rounded-md px-3 py-1.5 font-semibold ${
        active ? "bg-primary text-white" : "text-muted hover:text-fg"
      }`}
    >
      {children}
    </button>
  );
}
