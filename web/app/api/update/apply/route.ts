import { apiBase } from "@/lib/api";

// Proxies the browser's "apply update" action to the local Go API, adding the
// token server-side (never exposed to the client).
export async function POST() {
  const { url, token } = apiBase();
  try {
    const res = await fetch(`${url}/api/v1/update/apply`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
    });
    const body = await res.text();
    return new Response(body || "{}", {
      status: res.status,
      headers: { "Content-Type": "application/json" },
    });
  } catch {
    return Response.json(
      { error: "daemon unreachable" },
      { status: 502 },
    );
  }
}
