import { apiBase } from "@/lib/api";

// Generic proxy for per-app operations. GET: status | logs | probe.
// POST: install | uninstall | start | stop | restart. The Go API token is
// added server-side and never exposed to the browser.
const GET_ACTIONS = new Set([
  "status",
  "logs",
  "probe",
  "credentials",
  "image-update",
  "image-tags",
]);
const POST_ACTIONS = new Set([
  "install",
  "uninstall",
  "start",
  "stop",
  "restart",
  "clear-logs",
  "update",
  "update-latest",
  "set-version",
  "domain",
  "hide",
  "unhide",
]);

async function proxy(
  method: "GET" | "POST",
  id: string,
  action: string,
  search: string,
  body?: string,
) {
  const { url, token } = apiBase();
  try {
    const res = await fetch(`${url}/api/v1/apps/${id}/${action}${search}`, {
      method,
      headers: {
        Authorization: `Bearer ${token}`,
        ...(body ? { "Content-Type": "application/json" } : {}),
      },
      body,
      cache: "no-store",
    });
    const out = await res.text();
    return new Response(out || "{}", {
      status: res.status,
      headers: { "Content-Type": "application/json" },
    });
  } catch {
    return Response.json({ error: "daemon unreachable" }, { status: 502 });
  }
}

export async function GET(
  req: Request,
  { params }: { params: Promise<{ id: string; action: string }> },
) {
  const { id, action } = await params;
  if (!GET_ACTIONS.has(action)) {
    return Response.json({ error: "unknown action" }, { status: 404 });
  }
  const search = new URL(req.url).search;
  return proxy("GET", id, action, search);
}

export async function POST(
  req: Request,
  { params }: { params: Promise<{ id: string; action: string }> },
) {
  const { id, action } = await params;
  if (!POST_ACTIONS.has(action)) {
    return Response.json({ error: "unknown action" }, { status: 404 });
  }
  const search = new URL(req.url).search;
  const body = action === "install" ? await req.text() : undefined;
  return proxy("POST", id, action, search, body);
}
