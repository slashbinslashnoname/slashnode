import { apiBase } from "@/lib/api";

// Proxies host health indicators (disk/memory/load) from the Go API.
export async function GET() {
  const { url, token } = apiBase();
  try {
    const res = await fetch(`${url}/api/v1/system`, {
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
