import {
  CartesianGrid,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import type { HistoryPoint, Range } from "../../types/metric";

function tick(t: number, range: Range) {
  const d = new Date(t);
  return range === "7d"
    ? d.toLocaleDateString("en-GB", { weekday: "short" })
    : d.toLocaleTimeString("en-GB", { hour: "2-digit", minute: "2-digit" });
}

export default function HistoryChart({ data, range }: { data: HistoryPoint[]; range: Range }) {
  return (
    <div className="h-60" role="img" aria-label="CPU, memory and disk history">
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={data} margin={{ top: 8, right: 4, bottom: 0, left: -20 }}>
          <CartesianGrid stroke="var(--line)" vertical={false} />
          <XAxis
            dataKey="t"
            tickFormatter={(t) => tick(t, range)}
            stroke="var(--muted)"
            tick={{ fontSize: 11 }}
            tickLine={false}
            axisLine={false}
            minTickGap={48}
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
          <Tooltip
            labelFormatter={(t) => new Date(t as number).toLocaleString("en-GB")}
            formatter={(v) => `${Number(v ?? 0).toFixed(0)}%`}
            contentStyle={{
              background: "var(--panel)",
              border: "1px solid var(--line)",
              borderRadius: 8,
              color: "var(--text)",
            }}
          />
          <Line type="monotone" dataKey="disk" name="Disk" stroke="var(--net)" strokeWidth={2} dot={false} isAnimationActive={false} />
          <Line type="monotone" dataKey="memory" name="Memory" stroke="var(--mem)" strokeWidth={2} dot={false} isAnimationActive={false} />
          <Line type="monotone" dataKey="cpu" name="CPU" stroke="var(--cpu)" strokeWidth={3} dot={false} isAnimationActive={false} />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}