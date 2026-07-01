import { apiBase } from "@/lib/api";

// Leave the tailnet (the node identity is preserved in the state volume).
export async function POST() {
  const { url, token } = apiBase();
  try {
    const res = await fetch(`${url}/api/v1/tailscale/down`, {
      method: "POST",
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
