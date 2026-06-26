// Access to the local Go API (slashnoded). The URL and token are injected by
// the Go daemon when it launches `next start` (environment variables).

const API_URL = process.env.SLASHNODE_API_URL || "http://127.0.0.1:8081";
const API_TOKEN = process.env.SLASHNODE_API_TOKEN || "";

export type Status = {
  node_id: string;
  version: string;
  hostname: string;
  port: number;
  onion?: string;
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
  images?: Record<string, string>;
  category: string;
  description?: string;
  icon?: string;
  dependencies?: string[];
  inputs?: AppInput[];
  installed: boolean;
  installed_version?: string;
  update_available?: boolean;
  url?: string;
  onion_url?: string;
  onion?: string;
  subdomain?: string;
  domain?: string;
  host?: string;
  hidden?: boolean;
  web?: { port: number; path?: string };
  endpoints?: AppEndpoint[];
  probe?: { type: string };
  notes?: string;
};

export type AppEndpoint = {
  label: string;
  scheme?: string;
  port: number;
  path?: string;
};

export type ServiceStatus = {
  service: string;
  state: string;
  status: string;
  health?: string;
};

export type CredField = {
  key: string;
  label: string;
  value: string;
  secret: boolean;
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

export type Config = {
  version: string;
  node_id: string;
  hostname: string;
  data_dir: string;
  http: { bind: string; port: number; api_port: number };
  access: { mode: string; address: string; password_protected: boolean };
  tor: { enabled: boolean };
  theme: { mode: string; primary: string };
  update: { policy: string; channel: string };
  created_at: string;
};

export type SystemStats = {
  disk: { path: string; total: number; used: number; free: number; percent: number };
  mem_total: number;
  mem_used: number;
  mem_percent: number;
  load1: number;
  disk_warn: boolean;
  disk_critical: boolean;
};

export const getStatus = () => apiGet<Status>("/api/v1/status");
export const getUpdate = () => apiGet<UpdateInfo>("/api/v1/update");
export const getConfig = () => apiGet<Config>("/api/v1/config");
export const getSystem = () => apiGet<SystemStats>("/api/v1/system");

export const getApps = () => apiGet<{ apps: App[] }>("/api/v1/apps");
export const getApp = (id: string) => apiGet<App>(`/api/v1/apps/${id}`);

export function apiBase() {
  return { url: API_URL, token: API_TOKEN };
}
