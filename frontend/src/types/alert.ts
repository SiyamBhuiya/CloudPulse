export type AlertMetric = "cpu" | "memory" | "disk";
export type Channel = "email" | "slack" | "webhook";

export interface AlertRule {
  id: string;
  name: string;
  serverId: string; // a server id, or "all"
  metric: AlertMetric;
  threshold: number; // percent
  durationMin: number;
  channel: Channel;
  enabled: boolean;
}

export type AlertRuleInput = Omit<AlertRule, "id" | "enabled">;

export interface AlertEvent {
  id: string;
  ruleName: string;
  serverName: string;
  at: number; // unix ms
  status: "firing" | "resolved";
}