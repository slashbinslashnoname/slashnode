import { apiBase } from "@/lib/api";

// Runs a restore from the configured destination and streams live progress to
// the browser. Destructive — the UI gates this behind a typed confirmation.
export async function POST() {
  const { url, token } = apiBase();
  try {
    const res = await fetch(`${url}/api/v1/restore`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
    });
    return new Response(res.body, {
      status: res.status,
      headers: { "Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache" },
    });
  } catch {
    return new Response("daemon unreachable", { status: 502 });
  }
}
