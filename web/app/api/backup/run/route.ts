import { apiBase } from "@/lib/api";

// Runs a backup and streams rclone's live progress straight through to the
// browser (text/plain, chunked). `?all=true` includes chain volumes.
export async function POST(req: Request) {
  const { url, token } = apiBase();
  const all = new URL(req.url).searchParams.get("all") === "true";
  try {
    const res = await fetch(`${url}/api/v1/backup/run${all ? "?all=true" : ""}`, {
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
