import { apiBase } from "@/lib/api";

// Proxies an admin-password change to the Go API (token added server-side).
export async function POST(req: Request) {
  const { url, token } = apiBase();
  const body = await req.text();
  try {
    const res = await fetch(`${url}/api/v1/password`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json",
      },
      body,
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
