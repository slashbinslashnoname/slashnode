import { apiBase } from "@/lib/api";

// Proxies the streaming install to the Go API, forwarding the chunked plain-text
// body straight back to the browser so install output appears live. The token is
// added server-side and never exposed to the client.
export async function POST(
  req: Request,
  { params }: { params: Promise<{ id: string }> },
) {
  const { id } = await params;
  const { url, token } = apiBase();
  const body = await req.text();
  try {
    const res = await fetch(`${url}/api/v1/apps/${id}/install/stream`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json",
      },
      body,
      // @ts-expect-error - duplex is required by Node fetch for streaming bodies
      duplex: "half",
    });
    return new Response(res.body, {
      status: res.status,
      headers: {
        "Content-Type": "text/plain; charset=utf-8",
        "Cache-Control": "no-cache",
      },
    });
  } catch {
    return new Response("daemon unreachable\n", { status: 502 });
  }
}
