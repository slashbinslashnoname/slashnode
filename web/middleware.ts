import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

// Password protection is driven by the daemon (config.access.password_protected),
// passed in as an environment variable when it launches the front end.
const PROTECTED = process.env.SLASHNODE_PASSWORD_PROTECTED === "true";
const SESSION = process.env.SLASHNODE_SESSION_SECRET || "";

export function middleware(req: NextRequest) {
  if (!PROTECTED) return NextResponse.next();

  const { pathname } = req.nextUrl;
  // Always allow the login page and its API.
  if (pathname.startsWith("/login") || pathname.startsWith("/api/login")) {
    return NextResponse.next();
  }

  const cookie = req.cookies.get("slashnode_session")?.value;
  if (cookie && SESSION && cookie === SESSION) {
    return NextResponse.next();
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
