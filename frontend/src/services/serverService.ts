import { api } from "./api";
import type { CreateServerPayload, CreateServerResponse, Server } from "../types/server";

const MOCK = import.meta.env.VITE_MOCK === "true";
const ago = (s: number) => Date.now() - s * 1000;

// Fake data, only used when VITE_MOCK=true
let mockServers: Server[] = [
  { id: "1", name: "api-prod-1", region: "Frankfurt", status: "online", cpu: 72, memory: 61, uptime: 99.98, lastSeen: ago(2) },
  { id: "2", name: "api-prod-2", region: "Frankfurt", status: "online", cpu: 41, memory: 48, uptime: 99.99, lastSeen: ago(1) },
  { id: "3", name: "worker-1", region: "Singapore", status: "warning", cpu: 91, memory: 77, uptime: 99.71, lastSeen: ago(3) },
  { id: "4", name: "db-replica", region: "Singapore", status: "down", cpu: 0, memory: 0, uptime: 98.2, lastSeen: ago(240) },
];

const delay = <T,>(v: T) => new Promise<T>((r) => setTimeout(() => r(v), 300));

export const serverService = {
  list: (): Promise<Server[]> =>
    MOCK ? delay([...mockServers]) : api<Server[]>("/api/servers"),
    get: (id: string): Promise<Server> => {
    if (MOCK) {
      const s = mockServers.find((x) => x.id === id);
      return s ? delay(s) : Promise.reject(new Error("not found"));
    }
    return api<Server>(`/api/servers/${id}`);
  },

  create: (payload: CreateServerPayload): Promise<CreateServerResponse> => {
    if (MOCK) {
      const server: Server = {
        id: String(Date.now()),
        name: payload.name,
        region: payload.region || "Unknown",
        status: "down",
        cpu: 0,
        memory: 0,
        uptime: 100,
        lastSeen: Date.now(),
      };
      mockServers = [...mockServers, server];
      return delay({ server, agentKey: "cp_live_" + Math.random().toString(16).slice(2, 12) });
    }
    return api<CreateServerResponse>("/api/servers", {
      method: "POST",
      body: JSON.stringify(payload),
    });
  },

  remove: (id: string): Promise<void> => {
    if (MOCK) {
      mockServers = mockServers.filter((s) => s.id !== id);
      return delay(undefined);
    }
    return api<void>(`/api/servers/${id}`, { method: "DELETE" });
  },
};