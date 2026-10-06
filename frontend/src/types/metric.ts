export interface MetricSnapshot {
  serverId: string;
  timestamp: number; // unix ms
  cpu: number;       // 0-100
  memory: number;    // 0-100
  disk: number;      // 0-100
  network: number;   // MB/s
  requests: number;
  errorRate: number; // percent
  p95: number;       // ms
}

export type Range = "1h" | "6h" | "24h" | "7d";

export interface HistoryPoint {
  t: number; // unix ms
  cpu: number;
  memory: number;
  disk: number;
}

export interface ProcessInfo {
  name: string;
  cpu: number;      // percent
  memoryMb: number;
}

export interface ServerEvent {
  id: string;
  kind: "alert" | "ok" | "error";
  text: string;
  at: number; // unix ms
  resolved?: boolean;
}