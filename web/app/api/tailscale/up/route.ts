import { apiBase } from "@/lib/api";

// Join / re-authenticate the tailnet, streaming the tailscaled startup output
// straight through to the browser (text/plain, chunked). The auth key is only
// forwarded to the local Go API; it is never persisted or echoed back.
export async function POST(req: Request) {
  const { url, token } = apiBase();
  const body = await req.text();
  try {
    const res = await fetch(`${url}/api/v1/tailscale/up`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
      body,
    });
    return new Response(res.body, {
      status: res.status,
      headers: { "Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache" },
    });
  } catch {
    return new Response("daemon unreachable", { status: 502 });
  }
}
