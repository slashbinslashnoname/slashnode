import { apiBase } from "@/lib/api";

// Read/write the backup destination config. Credentials are added server-side
// to the Go API call; the GET response never includes stored secrets (the Go
// handler redacts them to a `has_secret` boolean).
export async function GET() {
  const { url, token } = apiBase();
  try {
    const res = await fetch(`${url}/api/v1/backup/config`, {
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

export async function POST(req: Request) {
  const { url, token } = apiBase();
  const body = await req.text();
  try {
    const res = await fetch(`${url}/api/v1/backup/config`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
      body,
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
