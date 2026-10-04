export interface MetricSnapshot {
  serverId: string;
  timestamp: number; // unix ms
  cpu: number;       // 0-100
  memory: number;    // 0-100
  network: number;   // MB/s
  requests: number;
  errorRate: number; // percent
  p95: number;       // ms
}