import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

// The web UI always requires an admin login. The session cookie is a per-login,
// expiring token "<expiryUnix>.<hmac>" (issued by the daemon, same format the Go
// console verifies). We verify it here with the shared session secret.
const SESSION = process.env.SLASHNODE_SESSION_SECRET || "";

async function validToken(token: string | undefined): Promise<boolean> {
  if (!token || !SESSION) return false;
  const dot = token.lastIndexOf(".");
  if (dot <= 0) return false;
  const exp = token.slice(0, dot);
  const sig = token.slice(dot + 1);
  const n = Number(exp);
  if (!Number.isFinite(n) || n * 1000 < Date.now()) return false;
  const key = await crypto.subtle.importKey(
    "raw",
    new TextEncoder().encode(SESSION),
    { name: "HMAC", hash: "SHA-256" },
    false,
    ["sign"],
  );
  const macBuf = await crypto.subtle.sign("HMAC", key, new TextEncoder().encode(exp));
  const expected = Array.from(new Uint8Array(macBuf))
    .map((b) => b.toString(16).padStart(2, "0"))
    .join("");
  if (expected.length !== sig.length) return false;
  let diff = 0;
  for (let i = 0; i < expected.length; i++) diff |= expected.charCodeAt(i) ^ sig.charCodeAt(i);
  return diff === 0;
}

export async function middleware(req: NextRequest) {
  const { pathname } = req.nextUrl;
  // Always allow the login page and its API.
  if (pathname.startsWith("/login") || pathname.startsWith("/api/login")) {
    return NextResponse.next();
  }
  if (await validToken(req.cookies.get("slashnode_session")?.value)) {
    return NextResponse.next();
  }
  // Unauthenticated: API gets 401, pages redirect to login.
  if (pathname.startsWith("/api/")) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }
  const url = req.nextUrl.clone();
  url.pathname = "/login";
  url.searchParams.set("next", pathname);
  return NextResponse.redirect(url);
}

// Run on everything except Next internals and static assets.
export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico).*)"],
};
