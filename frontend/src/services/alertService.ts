import { api } from "./api";
import type { AlertEvent, AlertRule, AlertRuleInput } from "../types/alert";

const MOCK = import.meta.env.VITE_MOCK === "true";
const delay = <T,>(v: T) => new Promise<T>((r) => setTimeout(() => r(v), 300));
const ago = (h: number) => Date.now() - h * 3_600_000;

// Fake data, only used when VITE_MOCK=true
let mockRules: AlertRule[] = [
  { id: "1", name: "High CPU on production", serverId: "1", metric: "cpu", threshold: 90, durationMin: 5, channel: "slack", enabled: true },
  { id: "2", name: "Memory running low", serverId: "all", metric: "memory", threshold: 85, durationMin: 10, channel: "email", enabled: true },
  { id: "3", name: "Disk almost full", serverId: "4", metric: "disk", threshold: 90, durationMin: 1, channel: "webhook", enabled: false },
];

const mockHistory: AlertEvent[] = [
  { id: "1", ruleName: "High CPU on production", serverName: "worker-1", at: ago(1), status: "firing" },
  { id: "2", ruleName: "High CPU on production", serverName: "api-prod-1", at: ago(4), status: "resolved" },
  { id: "3", ruleName: "Memory running low", serverName: "api-prod-2", at: ago(17), status: "resolved" },
  { id: "4", ruleName: "Disk almost full", serverName: "db-replica", at: ago(72), status: "resolved" },
];

export const alertService = {
  listRules: (): Promise<AlertRule[]> =>
    MOCK ? delay([...mockRules]) : api<AlertRule[]>("/api/alerts/rules"),

  history: (): Promise<AlertEvent[]> =>
    MOCK ? delay([...mockHistory]) : api<AlertEvent[]>("/api/alerts/history"),

  createRule: (input: AlertRuleInput): Promise<AlertRule> => {
    if (MOCK) {
      const rule: AlertRule = { ...input, id: String(Date.now()), enabled: true };
      mockRules = [...mockRules, rule];
      return delay(rule);
    }
    return api<AlertRule>("/api/alerts/rules", { method: "POST", body: JSON.stringify(input) });
  },

  updateRule: (rule: AlertRule): Promise<AlertRule> => {
    if (MOCK) {
      mockRules = mockRules.map((r) => (r.id === rule.id ? rule : r));
      return delay(rule);
    }
    return api<AlertRule>(`/api/alerts/rules/${rule.id}`, {
      method: "PUT",
      body: JSON.stringify(rule),
    });
  },

  deleteRule: (id: string): Promise<void> => {
    if (MOCK) {
      mockRules = mockRules.filter((r) => r.id !== id);
      return delay(undefined);
    }
    return api<void>(`/api/alerts/rules/${id}`, { method: "DELETE" });
  },
};