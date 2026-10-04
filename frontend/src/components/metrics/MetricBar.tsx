interface Props {
  label: string;
  value: number; // 0-100, controls bar width
  text: string;  // what to show on the right
  color: "cpu" | "mem" | "net";
}

const fill = { cpu: "bg-cpu", mem: "bg-mem", net: "bg-net" };

export default function MetricBar({ label, value, text, color }: Props) {
  return (
    <div className="mb-3.5 grid grid-cols-[78px_1fr_72px] items-center gap-3 last:mb-0">
      <span>{label}</span>
      <div className="h-2.5 overflow-hidden rounded-full bg-line/70">
        <div
          className={`h-full rounded-full transition-[width] duration-700 motion-reduce:transition-none ${fill[color]}`}
          style={{ width: `${Math.min(100, Math.max(0, value))}%` }}
        />
      </div>
      <b className="text-right font-semibold">{text}</b>
    </div>
  );
}