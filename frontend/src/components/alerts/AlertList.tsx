import type { Server } from "../../types/server";
import type { AlertMetric, AlertRule, Channel } from "../../types/alert";

interface Props {
  rules: AlertRule[];
  servers: Server[];
  onToggle: (rule: AlertRule) => void;
  onEdit: (rule: AlertRule) => void;
  onDelete: (rule: AlertRule) => void;
}

const metricLabel: Record<AlertMetric, string> = { cpu: "CPU", memory: "Memory", disk: "Disk" };
const channelLabel: Record<Channel, string> = { email: "Email", slack: "Slack", webhook: "Webhook" };

export default function AlertList({ rules, servers, onToggle, onEdit, onDelete }: Props) {
  const serverName = (id: string) =>
    id === "all" ? "All servers" : servers.find((s) => s.id === id)?.name ?? "Unknown server";

  return (
    <ul>
      {rules.map((r) => (
        <li
          key={r.id}
          className={`mb-3 grid items-center gap-x-4 gap-y-2 rounded-xl border border-line bg-panel px-5 py-4 md:grid-cols-[1fr_auto] ${
            r.enabled ? "" : "opacity-65"
          }`}
        >
          <div>
            <h3 className="text-base font-semibold">{r.name}</h3>
            <p className="text-sm text-muted">
              When <b className="font-semibold text-ink">{metricLabel[r.metric]}</b> is above{" "}
              <b className="font-semibold text-ink">{r.threshold}%</b> for{" "}
              <b className="font-semibold text-ink">{r.durationMin} min</b>
            </p>
            <div className="mt-2 flex flex-wrap gap-2">
              <span className="rounded-full border border-line px-2.5 py-0.5 text-[13px] text-muted">
                {serverName(r.serverId)}
              </span>
              <span className="rounded-full border border-line px-2.5 py-0.5 text-[13px] text-muted">
                {channelLabel[r.channel]}
              </span>
            </div>
          </div>

          <div className="flex items-center gap-3.5">
            <button
              role="switch"
              aria-checked={r.enabled}
              aria-label={`Enable ${r.name}`}
              onClick={() => onToggle(r)}
              className={`relative h-[22px] w-10 rounded-full transition-colors ${
                r.enabled ? "bg-ok" : "bg-line"
              }`}
            >
              <span
                className={`absolute left-[3px] top-[3px] h-4 w-4 rounded-full transition-transform motion-reduce:transition-none ${
                  r.enabled ? "translate-x-[18px] bg-[#10222b]" : "bg-ink"
                }`}
              />
            </button>
            <button onClick={() => onEdit(r)} className="rounded-md px-1.5 py-1 font-semibold text-muted">
              Edit
            </button>
            <button onClick={() => onDelete(r)} className="rounded-md px-1.5 py-1 font-semibold text-bad">
              Delete
            </button>
          </div>
        </li>
      ))}
    </ul>
  );
}