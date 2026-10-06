import { Link } from "react-router-dom";
import MetricBar from "../metrics/MetricBar";
import { percent, timeAgo, uptime } from "../../utils/formatters";
import type { Server, ServerStatus } from "../../types/server";

const badge: Record<ServerStatus, { label: string; dot: string }> = {
  online: { label: "Online", dot: "bg-ok" },
  warning: { label: "High load", dot: "bg-cpu" },
  down: { label: "Down", dot: "bg-bad" },
};

export default function ServerCard({ server }: { server: Server }) {
  const b = badge[server.status];
  const down = server.status === "down";

  return (
    <Link
      to={`/servers/${server.id}`}
      className={`block rounded-xl border border-line bg-panel p-5 transition hover:border-muted ${
        down ? "opacity-85" : ""
      }`}
    >
      <div className="mb-4 flex items-start justify-between gap-3">
        <div>
          <h3 className="text-[17px] font-semibold">{server.name}</h3>
          <small className="text-muted">{server.region}</small>
        </div>
        <span className="inline-flex items-center gap-1.5 whitespace-nowrap rounded-full border border-line px-2.5 py-0.5 text-[13px] font-medium">
          <i className={`h-[7px] w-[7px] rounded-full ${b.dot}`} />
          {b.label}
        </span>
      </div>

      {down ? (
        <p className="my-1.5 text-sm text-muted">
          No data. The agent stopped reporting {timeAgo(server.lastSeen)}.
        </p>
      ) : (
        <>
          <MetricBar label="CPU" color="cpu" value={server.cpu} text={percent(server.cpu)} />
          <MetricBar label="Memory" color="mem" value={server.memory} text={percent(server.memory)} />
        </>
      )}

      <div className="mt-3.5 flex justify-between border-t border-line pt-3 text-[13px] text-muted">
        <span>Uptime {uptime(server.uptime)}</span>
        <span>Seen {timeAgo(server.lastSeen)}</span>
      </div>
    </Link>
  );
}