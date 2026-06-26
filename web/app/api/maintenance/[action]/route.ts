import { apiBase } from "@/lib/api";

// Proxies a maintenance action to the Go API (token added server-side).
const ACTIONS = new Set(["reload-caddy", "reload-tor", "prune", "restart"]);

export async function POST(
  _req: Request,
  { params }: { params: Promise<{ action: string }> },
) {
  const { action } = await params;
  if (!ACTIONS.has(action)) {
    return Response.json({ error: "unknown action" }, { status: 404 });
  }
  const { url, token } = apiBase();
  try {
    const res = await fetch(`${url}/api/v1/${action}`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
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
