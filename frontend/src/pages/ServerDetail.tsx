import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";

import Button from "../components/ui/Button";
import Modal from "../components/ui/Modal";
import HistoryChart from "../components/metrics/HistoryChart";
import { useWebSocket } from "../hooks/useWebSocket";
import { metricService } from "../services/metricService";
import { serverService } from "../services/serverService";
import { NETWORK_MAX_MBPS } from "../utils/constants";
import { mbps, memSize, percent, timeAgo, uptime, when } from "../utils/formatters";
import type { Server } from "../types/server";
import type { HistoryPoint, ProcessInfo, Range, ServerEvent } from "../types/metric";

const RANGES: Range[] = ["1h", "6h", "24h", "7d"];

const badge = {
  online: { label: "Online", dot: "bg-ok" },
  warning: { label: "High load", dot: "bg-cpu" },
  down: { label: "Down", dot: "bg-bad" },
};
const eventDot = { alert: "bg-cpu", ok: "bg-ok", error: "bg-bad" };
const tileFill = { cpu: "bg-cpu", mem: "bg-mem", net: "bg-net" };

function Tile({
  label,
  value,
  pct,
  color,
}: {
  label: string;
  value: string;
  pct: number;
  color: keyof typeof tileFill;
}) {
  return (
    <div className="rounded-xl border border-line bg-panel px-4 py-4">
      <small className="block text-sm text-muted">{label}</small>
      <strong className="text-[26px] font-bold tracking-tight">{value}</strong>
      <div className="mt-2.5 h-1.5 overflow-hidden rounded-full bg-line">
        <div
          className={`h-full rounded-full transition-[width] duration-700 motion-reduce:transition-none ${tileFill[color]}`}
          style={{ width: `${Math.min(100, Math.max(0, pct))}%` }}
        />
      </div>
    </div>
  );
}

export default function ServerDetail() {
  const { id = "" } = useParams();
  const navigate = useNavigate();
  const { latest } = useWebSocket(id);

  const [server, setServer] = useState<Server | null>(null);
  const [notFound, setNotFound] = useState(false);
  const [range, setRange] = useState<Range>("6h");
  const [history, setHistory] = useState<HistoryPoint[]>([]);
  const [processes, setProcesses] = useState<ProcessInfo[]>([]);
  const [events, setEvents] = useState<ServerEvent[]>([]);
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [deleting, setDeleting] = useState(false);

  useEffect(() => {
    serverService.get(id).then(setServer).catch(() => setNotFound(true));
    metricService.processes(id).then(setProcesses).catch(() => {});
    metricService.events(id).then(setEvents).catch(() => {});
  }, [id]);

  useEffect(() => {
    metricService.history(id, range).then(setHistory).catch(() => setHistory([]));
  }, [id, range]);

  async function onDelete() {
    setDeleting(true);
    try {
      await serverService.remove(id);
      navigate("/servers", { replace: true });
    } catch {
      setDeleting(false);
    }
  }

  if (notFound) {
    return (
      <div className="mx-auto max-w-5xl">
        <h1 className="text-[28px] font-bold">Server not found</h1>
        <Link to="/servers" className="text-muted underline">Back to all servers</Link>
      </div>
    );
  }
  if (!server) return <p className="text-muted">Loading server...</p>;

  const b = badge[server.status];

  return (
    <div className="mx-auto max-w-5xl">
      <Link to="/servers" className="text-sm text-muted">&larr; All servers</Link>

      <header className="mb-6 mt-1.5 flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="flex flex-wrap items-center gap-3 text-[28px] font-bold tracking-tight">
            {server.name}
            <span className="inline-flex items-center gap-1.5 rounded-full border border-line px-2.5 py-0.5 text-[13px] font-medium">
              <i className={`h-[7px] w-[7px] rounded-full ${b.dot}`} />
              {b.label}
            </span>
          </h1>
          <p className="text-muted">
            {server.region} · Uptime {uptime(server.uptime)} · Seen {timeAgo(server.lastSeen)}
          </p>
        </div>
        <button
          onClick={() => setConfirmOpen(true)}
          className="rounded-lg border border-bad px-3.5 py-2 font-semibold text-bad"
        >
          Delete server
        </button>
      </header>

      <div className="mb-5 grid grid-cols-2 gap-3.5 lg:grid-cols-4">
        <Tile label="CPU" color="cpu" value={latest ? percent(latest.cpu) : "--"} pct={latest?.cpu ?? 0} />
        <Tile label="Memory" color="mem" value={latest ? percent(latest.memory) : "--"} pct={latest?.memory ?? 0} />
        <Tile label="Disk" color="net" value={latest ? percent(latest.disk) : "--"} pct={latest?.disk ?? 0} />
        <Tile
          label="Network"
          color="net"
          value={latest ? mbps(latest.network) : "--"}
          pct={latest ? (latest.network / NETWORK_MAX_MBPS) * 100 : 0}
        />
      </div>

      <section className="mb-5 rounded-2xl border border-line bg-panel p-5">
        <div className="mb-1.5 flex flex-wrap items-center justify-between gap-3">
          <h2 className="text-[17px] font-semibold">History</h2>
          <div role="group" aria-label="Time range" className="flex gap-1 rounded-lg border border-line bg-bg p-[3px]">
            {RANGES.map((r) => (
              <button
                key={r}
                aria-pressed={range === r}
                onClick={() => setRange(r)}
                className={`rounded-md px-3 py-1 text-sm font-semibold ${
                  range === r ? "bg-panel text-ink ring-1 ring-line" : "text-muted"
                }`}
              >
                {r}
              </button>
            ))}
          </div>
        </div>
        <div className="flex gap-4 text-sm text-muted">
          <span><i className="mr-1.5 inline-block h-[3px] w-3.5 bg-cpu align-middle" />CPU</span>
          <span><i className="mr-1.5 inline-block h-[3px] w-3.5 bg-mem align-middle" />Memory</span>
          <span><i className="mr-1.5 inline-block h-[3px] w-3.5 bg-net align-middle" />Disk</span>
        </div>
        <HistoryChart data={history} range={range} />
      </section>

      <div className="grid gap-5 md:grid-cols-[1.3fr_1fr]">
        <section className="rounded-xl border border-line bg-panel p-5">
          <h3 className="mb-3.5 font-semibold">Top processes</h3>
          <table className="w-full text-sm">
            <thead>
              <tr className="text-left text-[13px] text-muted">
                <th className="pb-2 font-medium">Process</th>
                <th className="pb-2 text-right font-medium">CPU</th>
                <th className="pb-2 text-right font-medium">Memory</th>
              </tr>
            </thead>
            <tbody>
              {processes.map((p) => (
                <tr key={p.name} className="border-t border-line">
                  <td className="py-2.5">{p.name}</td>
                  <td className="py-2.5 text-right">{percent(p.cpu)}</td>
                  <td className="py-2.5 text-right">{memSize(p.memoryMb)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </section>

        <section className="rounded-xl border border-line bg-panel p-5">
          <h3 className="mb-3.5 font-semibold">Recent events</h3>
          {events.length === 0 ? (
            <p className="text-sm text-muted">No events yet.</p>
          ) : (
            <ul className="flex flex-col gap-3 text-sm">
              {events.map((e) => (
                <li key={e.id} className="flex gap-2.5">
                  <i className={`mt-[7px] h-2 w-2 flex-none rounded-full ${eventDot[e.kind]}`} />
                  <div>
                    {e.text}
                    <small className="block text-[13px] text-muted">
                      {when(e.at)}
                      {e.resolved ? " · resolved" : ""}
                    </small>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </section>
      </div>

      <Modal open={confirmOpen} onClose={() => setConfirmOpen(false)} title={`Delete ${server.name}?`}>
        <p className="mb-5 mt-1 text-muted">
          This removes the server and all of its history. The agent will stop being accepted. This
          cannot be undone.
        </p>
        <div className="flex justify-end gap-2.5">
          <button
            onClick={() => setConfirmOpen(false)}
            className="rounded-lg border border-line px-4 py-2.5 font-semibold"
          >
            Cancel
          </button>
          <Button onClick={onDelete} disabled={deleting} className="!bg-bad">
            {deleting ? "Deleting..." : "Delete server"}
          </Button>
        </div>
      </Modal>
    </div>
  );
}