import { useNavigate } from "react-router-dom";

import Button from "../components/ui/Button";
import MetricBar from "../components/metrics/MetricBar";
import StatCard from "../components/metrics/StatCard";
import LiveChart from "../components/metrics/LiveChart";
import { useWebSocket } from "../hooks/useWebSocket";
import { NETWORK_MAX_MBPS } from "../utils/constants";
import { mbps, ms, number, percent } from "../utils/formatters";

export default function Dashboard() {
  const navigate = useNavigate();
  const { history, latest, connected } = useWebSocket();

  return (
    <div className="mx-auto max-w-5xl">
      <header className="mb-6 flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-[28px] font-bold tracking-tight">Dashboard</h1>
          <p className="text-muted">Live overview of your servers</p>
        </div>
        <div className="flex items-center gap-3">
          <span className="inline-flex items-center gap-2 rounded-full border border-line bg-panel px-3 py-1.5 font-medium">
            <b className={`h-2 w-2 rounded-full ${connected ? "bg-ok" : "bg-bad"}`} />
            {connected ? "Live, updating every second" : "Reconnecting"}
          </span>
          <Button onClick={() => navigate("/servers")}>Add server</Button>
        </div>
      </header>

      {!latest ? (
        <section className="rounded-xl border border-line bg-panel p-10 text-center">
          <h2 className="text-lg font-semibold">Waiting for data</h2>
          <p className="mt-1 text-muted">
            Add a server and start its agent. Metrics appear here within a few seconds.
          </p>
        </section>
      ) : (
        <>
          <section className="mb-5 rounded-2xl border border-line bg-panel px-5 pb-3 pt-5">
            <div className="flex flex-wrap items-baseline justify-between gap-3">
              <h2 className="text-[17px] font-semibold">Last 60 seconds</h2>
              <div className="flex gap-4 text-sm text-muted">
                <span><i className="mr-1.5 inline-block h-[3px] w-3.5 bg-cpu align-middle" />CPU</span>
                <span><i className="mr-1.5 inline-block h-[3px] w-3.5 bg-mem align-middle" />Memory</span>
                <span><i className="mr-1.5 inline-block h-[3px] w-3.5 bg-net align-middle" />Network</span>
              </div>
            </div>
            <LiveChart data={history} />
          </section>

          <div className="grid gap-5 md:grid-cols-[1.1fr_1fr]">
            <section className="rounded-xl border border-line bg-panel p-5">
              <h3 className="mb-4 font-semibold">Resource usage</h3>
              <MetricBar label="CPU" color="cpu" value={latest.cpu} text={percent(latest.cpu)} />
              <MetricBar label="Memory" color="mem" value={latest.memory} text={percent(latest.memory)} />
              <MetricBar
                label="Network"
                color="net"
                value={(latest.network / NETWORK_MAX_MBPS) * 100}
                text={mbps(latest.network)}
              />
            </section>

            <section className="rounded-xl border border-line bg-panel p-5">
              <h3 className="mb-4 font-semibold">Traffic</h3>
              <div className="grid grid-cols-2 gap-4 sm:grid-cols-3">
                <StatCard label="Requests" value={number(latest.requests)} />
                <StatCard label="Error rate" value={latest.errorRate.toFixed(2)} unit="%" />
                <StatCard label="P95 latency" value={ms(latest.p95)} unit="ms" />
              </div>
            </section>
          </div>
        </>
      )}
    </div>
  );
}