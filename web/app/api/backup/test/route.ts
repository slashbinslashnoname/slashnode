import { apiBase } from "@/lib/api";

// Verifies the configured backup destination is reachable with current creds.
export async function POST() {
  const { url, token } = apiBase();
  try {
    const res = await fetch(`${url}/api/v1/backup/test`, {
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
