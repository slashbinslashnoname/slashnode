import { apiBase } from "@/lib/api";

// Verifies the admin password against the Go API and, on success, stores the
// per-login session token it returns as an httpOnly cookie. The API bearer token
// is never exposed to the client.
export async function POST(req: Request) {
  const { url, token } = apiBase();

  let password = "";
  try {
    ({ password } = await req.json());
  } catch {
    return Response.json({ error: "bad request" }, { status: 400 });
  }

  let session = "";
  try {
    const res = await fetch(`${url}/api/v1/auth`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ password }),
    });
    if (!res.ok) {
      return Response.json({ error: "invalid password" }, { status: 401 });
    }
    const j = await res.json().catch(() => ({}));
    session = j.token || "";
  } catch {
    return Response.json({ error: "daemon unreachable" }, { status: 502 });
  }
  if (!session) {
    return Response.json({ error: "daemon unreachable" }, { status: 502 });
  }

  const secure = process.env.SLASHNODE_ACCESS_MODE === "server" ? "; Secure" : "";
  const response = Response.json({ status: "ok" });
  response.headers.append(
    "Set-Cookie",
    `slashnode_session=${session}; HttpOnly; Path=/; SameSite=Lax; Max-Age=604800${secure}`,
  );
  return response;
}
