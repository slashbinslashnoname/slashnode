import { apiBase } from "@/lib/api";

// Proxies the browser's "install & launch" action to the local Go API, adding
// the token server-side (never exposed to the client).
export async function POST(
  req: Request,
  { params }: { params: Promise<{ id: string }> },
) {
  const { id } = await params;
  const { url, token } = apiBase();
  try {
    const body = await req.text();
    const res = await fetch(`${url}/api/v1/apps/${id}/install`, {
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
