import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

// Password protection and access mode are driven by the daemon (config), passed
// in as environment variables when it launches the front end.
const PROTECTED = process.env.SLASHNODE_PASSWORD_PROTECTED === "true";
const SESSION = process.env.SLASHNODE_SESSION_SECRET || "";
const SECURE = process.env.SLASHNODE_ACCESS_MODE === "server"; // served over HTTPS

function setSession(res: NextResponse) {
  res.cookies.set("slashnode_session", SESSION, {
    httpOnly: true,
    path: "/",
    sameSite: "lax",
    maxAge: 604800,
    secure: SECURE,
  });
  return res;
}

export function middleware(req: NextRequest) {
  const { pathname } = req.nextUrl;
  const hasSession = !!SESSION && req.cookies.get("slashnode_session")?.value === SESSION;

  if (PROTECTED) {
    // Always allow the login page and its API.
    if (pathname.startsWith("/login") || pathname.startsWith("/api/login")) {
      return NextResponse.next();
    }
    if (hasSession) return NextResponse.next();
    const url = req.nextUrl.clone();
    url.pathname = "/login";
    url.searchParams.set("next", pathname);
    return NextResponse.redirect(url);
  }

  // Open (no-password) mode: there's no login, but we still bind a same-origin
  // session cookie and require it on the proxy API. With SameSite=Lax the cookie
  // is never sent on cross-site fetch/POST, so a malicious website the operator
  // visits (CSRF) — or any non-browser client that never loaded the UI — cannot
  // read secrets (/api/apps/*/credentials) or drive privileged actions (install,
  // config, password, console, …). Pages always load and (re)bind the cookie, so
  // the browser's own API calls carry it.
  if (pathname.startsWith("/api/") && !pathname.startsWith("/api/login")) {
    if (!hasSession) {
      return NextResponse.json({ error: "forbidden" }, { status: 403 });
    }
    return NextResponse.next();
  }
  const res = NextResponse.next();
  if (!hasSession) setSession(res);
  return res;
}

// Run on everything except Next internals and static assets.
export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico).*)"],
};
