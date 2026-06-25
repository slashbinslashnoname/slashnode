import { apiBase } from "@/lib/api";

// Proxies the current node status (version, hostname, onion…) to the browser,
// adding the Go API token server-side. Used to poll for the new version after
// an update so the UI can confirm the upgrade actually took effect.
export async function GET() {
  const { url, token } = apiBase();
  try {
    const res = await fetch(`${url}/api/v1/status`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: "no-store",
    });
    const body = await res.text();
    return new Response(body || "{}", {
      status: res.status,
      headers: { "Content-Type": "application/json" },
    });
  } catch {
    return Response.json({ error: "daemon unreachable" }, { status: 502 });
  }
}
