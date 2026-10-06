export type ServerStatus = "online" | "warning" | "down";

export interface Server {
  id: string;
  name: string;
  region: string;
  status: ServerStatus;
  cpu: number;    // 0-100
  memory: number; // 0-100
  uptime: number; // percent
  lastSeen: number; // unix ms
}

export interface CreateServerPayload {
  name: string;
  region?: string;
}

export interface CreateServerResponse {
  server: Server;
  agentKey: string; // shown to the user once
}