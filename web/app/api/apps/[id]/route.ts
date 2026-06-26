import { apiBase } from "@/lib/api";

// Proxy for a single app's full detail (manifest + install state: endpoints,
// onion, web, url, …). The Go API token is added server-side.
export async function GET(
  _req: Request,
  { params }: { params: Promise<{ id: string }> },
) {
  const { id } = await params;
  const { url, token } = apiBase();
  try {
    const res = await fetch(`${url}/api/v1/apps/${id}`, {
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
