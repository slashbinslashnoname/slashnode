import { apiBase } from "@/lib/api";

// Verifies the admin password against the Go API and, on success, sets an
// httpOnly session cookie. The token is never exposed to the client.
export async function POST(req: Request) {
  const { url, token } = apiBase();
  const session = process.env.SLASHNODE_SESSION_SECRET || "";

  let password = "";
  try {
    ({ password } = await req.json());
  } catch {
    return Response.json({ error: "bad request" }, { status: 400 });
  }

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
  } catch {
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
