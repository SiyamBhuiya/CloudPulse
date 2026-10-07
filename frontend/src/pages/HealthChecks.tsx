import { useEffect, useId, useState, type FormEvent } from "react";

import Button from "../components/ui/Button";
import Input from "../components/ui/Input";
import Modal from "../components/ui/Modal";
import { healthService } from "../services/healthService";
import { timeAgo } from "../utils/formatters";
import type { CheckStatus, DayStatus, HealthCheck, HealthCheckInput } from "../types/healthcheck";

const badge: Record<CheckStatus, { label: string; dot: string }> = {
  up: { label: "Up", dot: "bg-ok" },
  slow: { label: "Slow", dot: "bg-cpu" },
  down: { label: "Down", dot: "bg-bad" },
};

const dayColor: Record<DayStatus, string> = {
  ok: "bg-ok",
  degraded: "bg-cpu",
  down: "bg-bad",
};

const intervalLabel = (s: number) => (s < 60 ? `${s} sec` : `${s / 60} min`);

/* ---------- Add check form (shown inside the modal) ---------- */

function AddCheckForm({
  onSubmit,
  onCancel,
}: {
  onSubmit: (input: HealthCheckInput) => Promise<void>;
  onCancel: () => void;
}) {
  const intervalId = useId();
  const [name, setName] = useState("");
  const [url, setUrl] = useState("");
  const [intervalSec, setIntervalSec] = useState("60");
  const [expected, setExpected] = useState("200");
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setError(null);

    try {
      const u = new URL(url);
      if (u.protocol !== "http:" && u.protocol !== "https:") throw new Error();
    } catch {
      return setError("Enter a full URL that starts with http:// or https://");
    }
    const code = Number(expected);
    if (!(code >= 100 && code <= 599)) return setError("Status code must be between 100 and 599.");

    setSaving(true);
    try {
      await onSubmit({ name, url, intervalSec: Number(intervalSec), expectedStatus: code });
    } catch {
      setError("Could not add the check. Try again.");
      setSaving(false);
    }
  }

  return (
    <form onSubmit={submit} className="mt-4">
      {error && (
        <div role="alert" className="mb-4 rounded-lg border border-bad px-3 py-2 text-sm">
          {error}
        </div>
      )}

      <Input label="Name" placeholder="Marketing site" value={name} onChange={(e) => setName(e.target.value)} required />
      <Input
        label="URL"
        type="url"
        placeholder="https://example.com/health"
        value={url}
        onChange={(e) => setUrl(e.target.value)}
        required
      />

      <div className="grid grid-cols-2 gap-2.5 max-sm:grid-cols-1">
        <div className="mb-4">
          <label htmlFor={intervalId} className="mb-1.5 block font-medium">
            Check every
          </label>
          <select
            id={intervalId}
            value={intervalSec}
            onChange={(e) => setIntervalSec(e.target.value)}
            className="w-full rounded-lg border border-line bg-bg px-3 py-2.5 text-ink focus-visible:outline focus-visible:outline-2 focus-visible:outline-net"
          >
            <option value="30">30 seconds</option>
            <option value="60">1 minute</option>
            <option value="300">5 minutes</option>
          </select>
        </div>
        <Input
          label="Expect status"
          type="number"
          min={100}
          max={599}
          value={expected}
          onChange={(e) => setExpected(e.target.value)}
          required
        />
      </div>

      <div className="mt-1 flex justify-end gap-2.5">
        <button type="button" onClick={onCancel} className="rounded-lg border border-line px-4 py-2.5 font-semibold">
          Cancel
        </button>
        <Button type="submit" disabled={saving}>
          {saving ? "Adding..." : "Add check"}
        </Button>
      </div>
    </form>
  );
}

/* ---------- Page ---------- */

export default function HealthChecks() {
  const [checks, setChecks] = useState<HealthCheck[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);

  useEffect(() => {
    healthService
      .list()
      .then(setChecks)
      .catch(() => setError(true))
      .finally(() => setLoading(false));
  }, []);

  async function add(input: HealthCheckInput) {
    const created = await healthService.create(input);
    setChecks((cs) => [...cs, created]);
    setModalOpen(false);
  }

  async function toggle(c: HealthCheck) {
    setChecks((cs) => cs.map((x) => (x.id === c.id ? { ...x, enabled: !c.enabled } : x))); // update right away
    try {
      await healthService.setEnabled(c.id, !c.enabled);
    } catch {
      setChecks((cs) => cs.map((x) => (x.id === c.id ? c : x))); // put it back
    }
  }

  async function remove(c: HealthCheck) {
    if (!window.confirm(`Delete "${c.name}"? Its uptime history will be removed.`)) return;
    await healthService.remove(c.id);
    setChecks((cs) => cs.filter((x) => x.id !== c.id));
  }

  // Summary numbers (paused checks are left out)
  const active = checks.filter((c) => c.enabled);
  const up = active.filter((c) => c.status !== "down").length;
  const timed = active.filter((c) => c.responseMs !== null);
  const avgMs = timed.length
    ? Math.round(timed.reduce((sum, c) => sum + (c.responseMs ?? 0), 0) / timed.length)
    : null;
  const avgUptime = checks.length
    ? checks.reduce((sum, c) => sum + c.uptime, 0) / checks.length
    : null;

  return (
    <div className="mx-auto max-w-5xl">
      <header className="mb-5 flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-[28px] font-bold tracking-tight">Health checks</h1>
          <p className="text-muted">We request your URLs on a schedule and track uptime.</p>
        </div>
        <Button onClick={() => setModalOpen(true)}>Add check</Button>
      </header>

      {loading ? (
        <p className="text-muted">Loading health checks...</p>
      ) : error ? (
        <p role="alert" className="text-bad">Could not load health checks. Refresh to try again.</p>
      ) : checks.length === 0 ? (
        <section className="rounded-xl border border-line bg-panel p-10 text-center">
          <h2 className="text-lg font-semibold">No health checks yet</h2>
          <p className="mb-4 mt-1 text-muted">Add a URL and we will tell you when it goes down.</p>
          <Button onClick={() => setModalOpen(true)}>Add check</Button>
        </section>
      ) : (
        <>
          <div className="mb-5 grid gap-3.5 sm:grid-cols-3">
            {[
              ["Checks up", String(up), `of ${active.length}`],
              ["Avg response", avgMs === null ? "--" : String(avgMs), "ms"],
              ["30-day uptime", avgUptime === null ? "--" : avgUptime.toFixed(2), "%"],
            ].map(([label, value, unit]) => (
              <div key={label} className="rounded-xl border border-line bg-panel px-4 py-4">
                <small className="block text-sm text-muted">{label}</small>
                <strong className="text-[26px] font-bold tracking-tight">
                  {value}
                  <em className="ml-1 text-sm font-medium not-italic text-muted">{unit}</em>
                </strong>
              </div>
            ))}
          </div>

          <ul>
            {checks.map((c) => {
              const b = c.enabled ? badge[c.status] : { label: "Paused", dot: "bg-muted" };
              return (
                <li
                  key={c.id}
                  className={`mb-3 rounded-xl border border-line bg-panel px-5 py-4 ${c.enabled ? "" : "opacity-65"}`}
                >
                  <div className="flex flex-wrap items-start justify-between gap-3">
                    <div>
                      <h3 className="text-base font-semibold">{c.name}</h3>
                      <p className="break-all text-sm text-muted">{c.url}</p>
                    </div>
                    <span className="inline-flex items-center gap-1.5 whitespace-nowrap rounded-full border border-line px-2.5 py-0.5 text-[13px] font-medium">
                      <i className={`h-[7px] w-[7px] rounded-full ${b.dot}`} />
                      {b.label}
                    </span>
                  </div>

                  <div
                    role="img"
                    aria-label="Uptime over the last 30 days"
                    className="mb-2 mt-3.5 grid grid-cols-[repeat(30,1fr)] gap-[3px]"
                  >
                    {c.days.map((d, i) => (
                      <i key={i} className={`h-[26px] rounded-[3px] ${dayColor[d]}`} />
                    ))}
                  </div>
                  <div className="mb-3 flex justify-between text-xs text-muted">
                    <span>30 days ago</span>
                    <span>Today</span>
                  </div>

                  <div className="flex flex-wrap items-center justify-between gap-3 border-t border-line pt-3 text-sm">
                    <div className="flex flex-wrap gap-x-5 gap-y-1 text-muted">
                      <span>
                        Response{" "}
                        <b className="font-semibold text-ink">
                          {c.responseMs === null ? "No response" : `${c.responseMs} ms`}
                        </b>
                      </span>
                      <span>
                        Uptime <b className="font-semibold text-ink">{c.uptime.toFixed(2)}%</b>
                      </span>
                      <span>
                        Every <b className="font-semibold text-ink">{intervalLabel(c.intervalSec)}</b>
                      </span>
                      <span>Checked {timeAgo(c.lastChecked)}</span>
                    </div>

                    <div className="flex items-center gap-3">
                      <button
                        role="switch"
                        aria-checked={c.enabled}
                        aria-label={`Run ${c.name}`}
                        onClick={() => toggle(c)}
                        className={`relative h-[22px] w-10 rounded-full transition-colors ${c.enabled ? "bg-ok" : "bg-line"}`}
                      >
                        <span
                          className={`absolute left-[3px] top-[3px] h-4 w-4 rounded-full transition-transform motion-reduce:transition-none ${
                            c.enabled ? "translate-x-[18px] bg-[#10222b]" : "bg-ink"
                          }`}
                        />
                      </button>
                      <button onClick={() => remove(c)} className="rounded-md px-1.5 py-1 font-semibold text-bad">
                        Delete
                      </button>
                    </div>
                  </div>
                </li>
              );
            })}
          </ul>
        </>
      )}

      <Modal open={modalOpen} onClose={() => setModalOpen(false)} title="Add a health check">
        <AddCheckForm onSubmit={add} onCancel={() => setModalOpen(false)} />
      </Modal>
    </div>
  );
}