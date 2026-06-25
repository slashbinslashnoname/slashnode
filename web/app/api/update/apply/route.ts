import { apiBase } from "@/lib/api";

// Proxy l'action "Appliquer la mise à jour" du navigateur vers l'API Go locale,
// en ajoutant le token côté serveur (jamais exposé au client).
export async function POST() {
  const { url, token } = apiBase();
  try {
    const res = await fetch(`${url}/api/v1/update/apply`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
    });
    const body = await res.text();
    return new Response(body || "{}", {
      status: res.status,
      headers: { "Content-Type": "application/json" },
    });
  } catch {
    return Response.json(
      { error: "démon injoignable" },
      { status: 502 },
    );
  }
}
