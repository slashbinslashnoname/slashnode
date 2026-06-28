import { apiBase } from "@/lib/api";

// Proxy for the app catalog list (manifest + install state for every app). The
// DataProvider polls this from the browser, so it needs a same-origin route;
// the Go API token is added server-side and never exposed to the client.
export async function GET() {
  const { url, token } = apiBase();
  try {
    const res = await fetch(`${url}/api/v1/apps`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: "no-store",
    });
    const out = await res.text();
    return new Response(out || "{}", {
      status: res.status,
      headers: { "Content-Type": "application/json" },
    });
  } catch {
    return Response.json({ error: "daemon unreachable" }, { status: 502 });
  }
}
