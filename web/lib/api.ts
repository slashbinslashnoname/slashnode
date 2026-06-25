// Access to the local Go API (slashnoded). The URL and token are injected by
// the Go daemon when it launches `next start` (environment variables).

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

export type AppInput = {
  key: string;
  label: string;
  type: "text" | "email" | "password" | "number" | "textarea" | "select" | "boolean";
  required?: boolean;
  default?: unknown;
  placeholder?: string;
  help?: string;
  secret?: boolean;
  options?: string[];
  minLength?: number;
};

export type App = {
  id: string;
  name: string;
  version: string;
  category: string;
  description?: string;
  icon?: string;
  dependencies?: string[];
  inputs?: AppInput[];
  installed: boolean;
  url?: string;
  web?: { port: number; path?: string };
  probe?: { type: string };
};

export type ServiceStatus = {
  service: string;
  state: string;
  status: string;
  health?: string;
};

export type ProbeStat = { label: string; value: string };

export type ProbeResult = {
  type: string;
  ok: boolean;
  detail?: string;
  stats?: ProbeStat[];
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

export const getApps = () => apiGet<{ apps: App[] }>("/api/v1/apps");
export const getApp = (id: string) => apiGet<App>(`/api/v1/apps/${id}`);

export function apiBase() {
  return { url: API_URL, token: API_TOKEN };
}
