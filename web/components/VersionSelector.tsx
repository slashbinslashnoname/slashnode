"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

// VersionSelector lets the operator change the image tag of any docker image an
// app runs (e.g. bump bitcoind to v31), one service at a time, then re-applies
// it. Nothing is auto-updated — the change happens only on "change".
// `images` maps each service to its current image ref ("repo:tag"); `suggest`
// offers optional tag suggestions (manifest `versions`).
export function VersionSelector({
  id,
  images,
  suggest,
}: {
  id: string;
  images: Record<string, string>;
  suggest?: string[];
}) {
  const services = Object.keys(images);
  if (services.length === 0) return null;
  return (
    <div className="flex flex-col gap-2">
      {services.map((svc) => (
        <ServiceTag
          key={svc}
          id={id}
          service={svc}
          image={images[svc]}
          suggest={suggest}
          showService={services.length > 1}
        />
      ))}
    </div>
  );
}

function splitTag(image: string): { repo: string; tag: string } {
  const slash = image.lastIndexOf("/");
  const last = slash >= 0 ? image.slice(slash + 1) : image;
  const colon = last.indexOf(":");
  if (colon < 0) return { repo: image, tag: "latest" };
  return {
    repo: image.slice(0, image.length - (last.length - colon)),
    tag: last.slice(colon + 1),
  };
}

function ServiceTag({
  id,
  service,
  image,
  suggest,
  showService,
}: {
  id: string;
  service: string;
  image: string;
  suggest?: string[];
  showService: boolean;
}) {
  const router = useRouter();
  const { repo, tag: current } = splitTag(image);
  const [tag, setTag] = useState(current);
  const [state, setState] = useState<"idle" | "applying" | "error">("idle");
  const listId = `tags-${id}-${service}`;

  async function apply() {
    if (!tag || tag === current) return;
    setState("applying");
    try {
      const r = await fetch(
        `/api/apps/${id}/set-version?service=${encodeURIComponent(
          service,
        )}&tag=${encodeURIComponent(tag)}`,
        { method: "POST" },
      );
      setState(r.ok ? "idle" : "error");
      if (r.ok) router.refresh();
    } catch {
      setState("error");
    }
  }

  return (
    <div className="flex flex-wrap items-center gap-2 text-xs">
      <code className="text-muted">
        {showService ? `${service}: ` : ""}
        {repo}:
      </code>
      <input
        value={tag}
        list={suggest && suggest.length ? listId : undefined}
        onChange={(e) => setTag(e.target.value)}
        disabled={state === "applying"}
        className="w-28 rounded-md border border-border bg-bg px-2 py-1 outline-none focus:border-primary"
      />
      {suggest && suggest.length > 0 && (
        <datalist id={listId}>
          {suggest.map((v) => (
            <option key={v} value={v} />
          ))}
        </datalist>
      )}
      <button
        onClick={apply}
        disabled={state === "applying" || !tag || tag === current}
        className="rounded-md bg-primary px-2 py-1 font-semibold text-white disabled:opacity-50"
      >
        {state === "applying" ? "applying…" : state === "error" ? "retry" : "change"}
      </button>
    </div>
  );
}
