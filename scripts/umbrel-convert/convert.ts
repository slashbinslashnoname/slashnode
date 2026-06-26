/**
 * umbrel-convert — best-effort converter from getumbrel/umbrel-apps to SlashNode
 * app manifests.
 *
 * It fetches each app's `umbrel-app.yml` + `docker-compose.yml`, mechanically
 * rewrites the parts that differ between the two platforms, and writes a
 * `slashnode-app.json` into ../../converted/<id>/ for REVIEW (not into apps/).
 * A REPORT.md summarizes which apps converted cleanly and which still reference
 * Umbrel-only constructs that a human must resolve.
 *
 * What it translates:
 *   - umbrel-app.yml metadata  -> manifest fields (id/name/version/category/…)
 *   - the `app_proxy` service   -> a `web` block + publishing the web port so
 *     SlashNode's Caddy (which proxies to a host port) can reach it
 *   - `${APP_DATA_DIR}/…` binds -> named volumes (slashnode-<id>_<svc>_<key>)
 *   - the shared `slashnode` external network + container_name per service
 *   - image@sha256 digests stripped; a few known env vars remapped
 *
 * What it CANNOT translate (flagged in REPORT.md as needs-review):
 *   - Umbrel infra env vars it doesn't know (e.g. ${APP_BITCOIN_NODE_IP},
 *     ${TOR_PROXY_IP}); shared ${UMBREL_ROOT}/data/storage mounts; long-syntax
 *     volumes. These are left verbatim so a human can finish them.
 *
 * Usage:  npm install && npm start            # converts the built-in allowlist
 *         npm start -- audiobookshelf navidrome   # convert specific ids
 */
import yaml from "js-yaml";
import { mkdir, writeFile } from "node:fs/promises";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const HERE = dirname(fileURLToPath(import.meta.url));
const OUT_DIR = join(HERE, "..", "..", "converted");
const RAW = "https://raw.githubusercontent.com/getumbrel/umbrel-apps/master";

// Curated "self-hosting essentials" present in the Umbrel repo.
const ALLOWLIST = [
  "adguard-home", "audiobookshelf", "bazarr", "calibre-web", "code-server",
  "duplicati", "excalidraw", "forgejo", "freshrss", "frigate", "ghost",
  "gitea", "grafana", "grocy", "homarr", "home-assistant", "immich",
  "invidious", "jackett", "jellyfin", "jellyseerr", "librespeed", "lidarr",
  "mealie", "memos", "minio", "mosquitto", "n8n", "navidrome", "nextcloud",
  "node-red", "overseerr", "owncloud", "photoprism", "plausible", "portainer",
  "prowlarr", "qbittorrent", "radarr", "readarr", "sabnzbd", "searxng",
  "sonarr", "stirling-pdf", "syncthing", "tailscale", "tautulli",
  "transmission", "umami", "uptime-kuma", "vaultwarden", "vikunja",
  "wireguard", "wordpress", "zigbee2mqtt",
];

const CATEGORY_ICON: Record<string, string> = {
  media: "🎬", files: "📁", networking: "🌐", productivity: "📝",
  finance: "💰", automation: "🔧", social: "💬", developer: "💻",
  bitcoin: "₿", lightning: "⚡", ai: "🤖", health: "🩺", home: "🏠",
};

// SlashNode templating tokens that are legitimate in a rendered manifest.
const SLASHNODE_TOKEN = /\$\{(input|secret|self|node|app)\./;

type Report = {
  id: string;
  status: "ok" | "needs-review" | "failed";
  notes: string[];
  security: string[];
};

async function fetchText(url: string): Promise<string | null> {
  const res = await fetch(url);
  if (!res.ok) return null;
  return res.text();
}

function iconFor(category: string): string {
  return CATEGORY_ICON[category?.toLowerCase?.()] ?? "📦";
}

// envEntries normalizes compose `environment` (object or array form) to pairs.
function envEntries(env: unknown): [string, string][] {
  if (!env) return [];
  if (Array.isArray(env)) {
    return env.map((e) => {
      const s = String(e);
      const i = s.indexOf("=");
      return i < 0 ? [s, ""] : [s.slice(0, i), s.slice(i + 1)];
    });
  }
  return Object.entries(env as Record<string, unknown>).map(([k, v]) => [k, String(v)]);
}

// remapTokens rewrites known Umbrel ${…} tokens; records secrets/unmapped.
function remapTokens(
  value: string,
  secrets: Set<string>,
  unmapped: Set<string>,
): string {
  return value.replace(/\$\{([A-Z0-9_]+)\}/g, (whole, name: string) => {
    if (name === "DEVICE_DOMAIN" || name === "DEVICE_HOSTNAME") {
      return "${node.exports.host}";
    }
    if (/^APP_.*(PASSWORD|SEED|SECRET|KEY|TOKEN|APIKEY)$/.test(name)) {
      secrets.add(name);
      return "${secret." + name + "}";
    }
    // Anything else is an Umbrel-injected value we can't satisfy — leave it and
    // flag it for review.
    unmapped.add(name);
    return whole;
  });
}

// convertVolumes rewrites a service's volume mounts to named volumes / binds.
function convertVolumes(
  appId: string,
  svc: string,
  vols: unknown,
  topVolumes: Record<string, unknown>,
  notes: Set<string>,
  unmapped: Set<string>,
): string[] {
  if (!Array.isArray(vols)) return [];
  const out: string[] = [];
  for (const raw of vols) {
    if (typeof raw !== "string") {
      notes.add("long-syntax volume skipped — review");
      continue;
    }
    const parts = raw.split(":");
    // host[:container[:mode]] — but Windows-style paths are irrelevant here.
    const src = parts[0];
    const dst = parts[1] ?? "";
    const mode = parts[2] ? ":" + parts[2] : "";
    const named = (key: string) => {
      const safe = key.replace(/[^a-zA-Z0-9_.-]+/g, "-").replace(/^-+|-+$/g, "") || "data";
      const volKey = `${svc}_${safe}`;
      topVolumes[volKey] = { name: `slashnode-${appId}_${svc}_${safe}` };
      out.push(`${volKey}:${dst}${mode}`);
    };
    if (src.startsWith("${APP_DATA_DIR}")) {
      named(src.replace("${APP_DATA_DIR}", "").replace(/^\/+/, ""));
    } else if (src.startsWith("${UMBREL_ROOT}")) {
      notes.add("shared ${UMBREL_ROOT} storage mapped to a private volume — review");
      named("shared-" + src.replace("${UMBREL_ROOT}", "").replace(/^\/+/, ""));
    } else if (src.startsWith("/")) {
      // Host bind (e.g. /etc/localtime, /var/run/docker.sock) — keep as-is.
      if (src.includes("docker.sock")) notes.add("mounts docker.sock (privileged) — review");
      out.push(`${src}:${dst}${mode}`);
    } else if (/^[a-zA-Z0-9_.-]+$/.test(src)) {
      // A named volume declared at the top level.
      named(src);
    } else {
      unmapped.add(src);
      out.push(raw);
    }
  }
  return out;
}

async function convert(appId: string): Promise<Report> {
  const notes = new Set<string>();
  const [metaRaw, composeRaw] = await Promise.all([
    fetchText(`${RAW}/${appId}/umbrel-app.yml`),
    fetchText(`${RAW}/${appId}/docker-compose.yml`),
  ]);
  if (!metaRaw || !composeRaw) {
    return { id: appId, status: "failed", notes: ["could not fetch umbrel-app.yml / docker-compose.yml"], security: [] };
  }
  const meta = yaml.load(metaRaw) as any;
  const compose = yaml.load(composeRaw) as any;
  const services: Record<string, any> = compose?.services ?? {};

  // Identify the web service + internal port from the app_proxy declaration.
  const proxy = services["app_proxy"];
  const proxyEnv = Object.fromEntries(envEntries(proxy?.environment));
  let webService = "";
  const appHost = proxyEnv["APP_HOST"] ?? "";
  // APP_HOST is "<id>_<service>_1" → recover <service>.
  const m = appHost.match(new RegExp(`^${appId.replace(/[^a-z0-9]/g, "[^_]*")}_(.+)_1$`));
  if (m) webService = m[1];
  const nonProxy = Object.keys(services).filter((s) => s !== "app_proxy");
  if (!webService || !services[webService]) {
    webService = nonProxy.length === 1 ? nonProxy[0] : (nonProxy.includes("web") ? "web" : nonProxy[0] ?? "");
  }
  const internalPort = Number(proxyEnv["APP_PORT"]) || Number(meta.port) || 80;
  const webHostPort = Number(meta.port) || internalPort;

  const secrets = new Set<string>();
  const unmapped = new Set<string>();
  const security = new Set<string>();
  const topVolumes: Record<string, unknown> = {};
  const outServices: Record<string, any> = {};

  // A static (non-empty) default password is a security smell to surface.
  if (meta.defaultPassword && String(meta.defaultPassword).trim() !== "") {
    security.add(`ships a default password ("${meta.defaultPassword}") — must be changed on first login`);
  }

  for (const svc of nonProxy) {
    const s = services[svc] ?? {};
    const ns: any = {};
    ns.image = String(s.image ?? "").split("@")[0]; // strip @sha256 digest
    ns.container_name = svc;
    ns.restart = "unless-stopped";
    ns.networks = ["slashnode"];
    if (s.user) ns.user = s.user;
    if (s.init) ns.init = s.init;
    if (s.command) ns.command = s.command;
    if (s.entrypoint) ns.entrypoint = s.entrypoint;
    if (s.depends_on) ns.depends_on = s.depends_on;

    // Elevated-privilege constructs: preserved (so the app still works) but
    // flagged, because they widen the attack surface beyond SlashNode's model.
    if (s.privileged) { ns.privileged = true; security.add(`${svc}: runs privileged`); }
    if (s.cap_add) { ns.cap_add = s.cap_add; security.add(`${svc}: extra capabilities ${JSON.stringify(s.cap_add)}`); }
    if (s.devices) { ns.devices = s.devices; security.add(`${svc}: maps host devices ${JSON.stringify(s.devices)}`); }
    if (s.network_mode) {
      ns.network_mode = s.network_mode;
      if (String(s.network_mode).includes("host")) {
        security.add(`${svc}: host network mode — bypasses Caddy/Tor isolation, binds host interfaces directly`);
      }
    }
    if (JSON.stringify(s.volumes ?? []).includes("docker.sock")) {
      security.add(`${svc}: mounts docker.sock — container-escape / root-equivalent access`);
    }

    const env = envEntries(s.environment);
    if (env.length) {
      ns.environment = Object.fromEntries(
        env.map(([k, v]) => [k, remapTokens(v, secrets, unmapped)]),
      );
    }

    const vols = convertVolumes(appId, svc, s.volumes, topVolumes, notes, unmapped);
    if (vols.length) ns.volumes = vols;

    if (svc === webService && !s.network_mode) {
      // Publish the web UI ONLY on loopback: SlashNode's Caddy reverse-proxies
      // <sub>.<host> → 127.0.0.1:<port> behind the admin login + HTTPS (and Tor).
      // Binding 0.0.0.0 here would expose the UI unauthenticated on the public
      // host, bypassing that — never do it. (Skipped under host network mode,
      // where `ports` is invalid — those apps are flagged for review anyway.)
      ns.ports = [`127.0.0.1:${webHostPort}:${internalPort}`];
    } else if (s.ports) {
      // Non-web services: do NOT auto-publish their ports (default-deny). Many
      // are internal; any genuinely needed inbound port must be added by hand.
      security.add(`${svc}: declares ports ${JSON.stringify(s.ports)} — left UNpublished, add by hand if required`);
    }
    outServices[svc] = ns;
  }

  const newCompose: any = {
    name: `slashnode-${appId}`,
    services: outServices,
    networks: { slashnode: { external: true } },
  };
  if (Object.keys(topVolumes).length) newCompose.volumes = topVolumes;

  const composeYaml = yaml.dump(newCompose, { lineWidth: -1, noRefs: true });

  // Build inputs for detected secrets.
  const inputs = [...secrets].map((name) => ({
    key: name,
    label: name.replace(/^APP_/, "").replace(/_/g, " ").toLowerCase(),
    type: "password",
    required: true,
    secret: true,
    minLength: 12,
    help: `Mapped from Umbrel ${"${" + name + "}"}. Stored encrypted.`,
  }));

  const manifest: any = {
    manifestVersion: 1,
    id: appId,
    name: meta.name ?? appId,
    version: String(meta.version ?? "latest"),
    category: meta.category ?? "apps",
    description: meta.tagline ?? meta.name ?? appId,
    icon: iconFor(meta.category),
    dependencies: Array.isArray(meta.dependencies) ? meta.dependencies : [],
    inputs,
    compose: composeYaml,
    web: { port: webHostPort, path: meta.path && meta.path !== "" ? meta.path : "/" },
    probe: { type: "http", port: webHostPort, path: "/" },
    notes:
      `Auto-converted from getumbrel/umbrel-apps (${appId}). Review before shipping. ` +
      (meta.website ? `Upstream: ${meta.website}. ` : "") +
      (meta.defaultUsername || meta.defaultPassword
        ? `Default login — user: "${meta.defaultUsername ?? ""}", password: "${meta.defaultPassword ?? ""}". `
        : "") +
      (security.size ? `⚠️ SECURITY: ${[...security].join("; ")}. ` : ""),
  };

  const json = JSON.stringify(manifest, null, 2) + "\n";

  // Validate: leftover non-SlashNode ${…} tokens anywhere => needs review.
  const leftover = [...json.matchAll(/\$\{([A-Za-z0-9_.]+)\}/g)]
    .map((x) => x[0])
    .filter((t) => !SLASHNODE_TOKEN.test(t));
  for (const t of leftover) unmapped.add(t.replace(/[${}]/g, ""));

  await mkdir(join(OUT_DIR, appId), { recursive: true });
  await writeFile(join(OUT_DIR, appId, "slashnode-app.json"), json);

  if (!outServices[webService]) notes.add("no web service resolved");
  const allNotes = [...notes];
  if (unmapped.size) allNotes.unshift(`unmapped tokens: ${[...unmapped].join(", ")}`);
  const status: Report["status"] =
    unmapped.size || notes.size || security.size ? "needs-review" : "ok";
  return { id: appId, status, notes: allNotes, security: [...security] };
}

async function main() {
  const ids = process.argv.slice(2).length ? process.argv.slice(2) : ALLOWLIST;
  await mkdir(OUT_DIR, { recursive: true });
  const reports: Report[] = [];
  for (const id of ids) {
    try {
      const r = await convert(id);
      reports.push(r);
      console.log(`${r.status === "ok" ? "✓" : r.status === "failed" ? "✗" : "△"} ${id}${r.notes.length ? "  — " + r.notes[0] : ""}`);
    } catch (e) {
      reports.push({ id, status: "failed", notes: [String((e as Error).message)], security: [] });
      console.log(`✗ ${id}  — ${(e as Error).message}`);
    }
  }

  const ok = reports.filter((r) => r.status === "ok");
  const review = reports.filter((r) => r.status === "needs-review");
  const failed = reports.filter((r) => r.status === "failed");
  const flagged = reports.filter((r) => r.security.length);
  const md = [
    "# Umbrel → SlashNode conversion report",
    "",
    `Converted ${reports.length} apps · ✓ ${ok.length} clean · △ ${review.length} need review · ✗ ${failed.length} failed · 🔒 ${flagged.length} with security flags.`,
    "",
    "These manifests are **auto-generated for review** and are NOT in the shipped",
    "catalog (they live here, not under `apps/`). Promote one by moving its folder",
    "into `apps/` after verifying it.",
    "",
    "Every converted app publishes its web UI on **127.0.0.1 only** (reached via",
    "Caddy + the admin login, and Tor) — never on a public interface. Non-web",
    "ports are left unpublished by default.",
    "",
    "## 🔒 Security review (per app)",
    "",
    "Apps below use elevated-privilege constructs or ship a default credential —",
    "review each before promoting. Apps not listed declared nothing notable.",
    "",
    ...(flagged.length
      ? flagged.map((r) => `### \`${r.id}\`\n` + r.security.map((s) => `- ⚠️ ${s}`).join("\n"))
      : ["_None — no app requested elevated privileges._"]),
    "",
    "## ✓ Clean",
    ...ok.map((r) => `- \`${r.id}\``),
    "",
    "## △ Needs review",
    ...review.map((r) => `- \`${r.id}\` — ${r.notes.join("; ")}`),
    "",
    "## ✗ Failed",
    ...failed.map((r) => `- \`${r.id}\` — ${r.notes.join("; ")}`),
    "",
  ].join("\n");
  await writeFile(join(OUT_DIR, "REPORT.md"), md);
  console.log(`\n${ok.length} clean, ${review.length} need review, ${failed.length} failed → converted/REPORT.md`);
}

main();
