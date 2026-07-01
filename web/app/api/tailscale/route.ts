import { apiBase } from "@/lib/api";

// Read the tailnet status (self address, peers) plus the persisted enabled flag.
export async function GET() {
  const { url, token } = apiBase();
  try {
    const res = await fetch(`${url}/api/v1/tailscale`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: "no-store",
    });
    return new Response((await res.text()) || "{}", {
      status: res.status,
      headers: { "Content-Type": "application/json" },
    });
  } catch {
    return Response.json({ error: "daemon unreachable" }, { status: 502 });
  }
}
