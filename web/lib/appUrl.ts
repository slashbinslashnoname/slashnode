// Canonical "open" URL for an app's web UI, so every place that links to it (the
// home tile button, the endpoints panel, …) agrees on one reverse-proxy
// convention.
//
// When the SlashNode UI itself is served over HTTPS (server mode / behind the
// reverse proxy) we use the proxy subdomain https://<id>.<host> — the same
// convention Caddy and the daemon emit (app.url). On a plain local-mode origin
// that subdomain isn't resolvable over mDNS (Avahi only announces the bare
// host), so we fall back to the node host + the app's published port. The onion
// form is always the app's hidden service at :80 (Tor maps onion:80 → web port).

export function webClearnetUrl(
  url: string | undefined,
  port: number | undefined,
): string {
  if (typeof window === "undefined") return url ?? "";
  if (location.protocol === "https:" && url) return url;
  if (port) return `http://${location.hostname}:${port}`;
  return url ?? "";
}
