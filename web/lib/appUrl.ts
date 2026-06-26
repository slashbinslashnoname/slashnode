// Canonical "open" URL for an app's web UI, so every place that links to it (the
// home tile button, the endpoints panel, …) agrees.
//
// Apps are reached at the HTTPS URL Caddy terminates TLS on — a dedicated port
// on the node host (app.url, https://<host>:<port>) — which works on a domain,
// an mDNS name or a bare IP. The app itself never serves TLS. We prefer app.url
// and only fall back to the current host + published port if the daemon didn't
// provide a URL. The onion form is always the app's hidden service at :80 (Tor
// maps onion:80 → web port).

export function webClearnetUrl(
  url: string | undefined,
  port: number | undefined,
): string {
  if (url) return url;
  if (typeof window !== "undefined" && port) {
    return `http://${location.hostname}:${port}`;
  }
  return "";
}
