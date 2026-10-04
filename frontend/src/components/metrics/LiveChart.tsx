import {
  CartesianGrid,
  Line,
  LineChart,
  ResponsiveContainer,
  XAxis,
  YAxis,
} from "recharts";
import { NETWORK_MAX_MBPS } from "../../utils/constants";
import type { MetricSnapshot } from "../../types/metric";

export default function LiveChart({ data }: { data: MetricSnapshot[] }) {
  const rows = data.map((d, i) => ({
    t: i - data.length + 1, // seconds ago (0 = now)
    cpu: d.cpu,
    memory: d.memory,
    network: Math.min(100, (d.network / NETWORK_MAX_MBPS) * 100),
  }));

  return (
    <div className="h-64" role="img" aria-label="Live CPU, memory and network graph">
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={rows} margin={{ top: 8, right: 4, bottom: 0, left: -20 }}>
          <CartesianGrid stroke="var(--line)" vertical={false} />
          <XAxis
            dataKey="t"
            tickFormatter={(t) => (t === 0 ? "now" : `${-t}s`)}
            stroke="var(--muted)"
            tick={{ fontSize: 11 }}
            tickLine={false}
            axisLine={false}
            minTickGap={40}
          />
          <YAxis
            domain={[0, 100]}
            ticks={[0, 50, 100]}
            tickFormatter={(v) => `${v}%`}
            stroke="var(--muted)"
            tick={{ fontSize: 11 }}
            tickLine={false}
            axisLine={false}
          />
          <Line type="monotone" dataKey="network" stroke="var(--net)" strokeWidth={2} dot={false} isAnimationActive={false} />
          <Line type="monotone" dataKey="memory" stroke="var(--mem)" strokeWidth={2} dot={false} isAnimationActive={false} />
          <Line type="monotone" dataKey="cpu" stroke="var(--cpu)" strokeWidth={3} dot={false} isAnimationActive={false} />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}