export type CheckStatus = "up" | "slow" | "down";
export type DayStatus = "ok" | "degraded" | "down";

export interface HealthCheck {
  id: string;
  name: string;
  url: string;
  status: CheckStatus;
  responseMs: number | null; // null = no response
  uptime: number;            // 30-day percent
  intervalSec: number;
  expectedStatus: number;
  lastChecked: number;       // unix ms
  enabled: boolean;
  days: DayStatus[];         // 30 entries, oldest first
}

export interface HealthCheckInput {
  name: string;
  url: string;
  intervalSec: number;
  expectedStatus: number;
}