import { apiBase } from "@/lib/api";

// Proxy for the self-update check (current vs latest release). The DataProvider
// polls this from the browser to drive the dashboard's update banner; the Go API
// token is added server-side and never exposed to the client.
export async function GET() {
  const { url, token } = apiBase();
  try {
    const res = await fetch(`${url}/api/v1/update`, {
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
