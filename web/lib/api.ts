// Accès à l'API Go locale (slashnoded). L'URL et le token sont injectés par le
// démon Go au lancement de `next start` (variables d'environnement).

const API_URL = process.env.SLASHNODE_API_URL || "http://127.0.0.1:8081";
const API_TOKEN = process.env.SLASHNODE_API_TOKEN || "";

export type Status = {
  node_id: string;
  version: string;
  hostname: string;
  port: number;
};

export type UpdateInfo = {
  current: string;
  latest: string;
  available: boolean;
  checked_at: string;
};

async function apiGet<T>(path: string): Promise<T | null> {
  try {
    const res = await fetch(`${API_URL}${path}`, {
      headers: { Authorization: `Bearer ${API_TOKEN}` },
      cache: "no-store",
    });
    if (!res.ok) return null;
    return (await res.json()) as T;
  } catch {
    return null;
  }
}

export const getStatus = () => apiGet<Status>("/api/v1/status");
export const getUpdate = () => apiGet<UpdateInfo>("/api/v1/update");

export function apiBase() {
  return { url: API_URL, token: API_TOKEN };
}
