import { useEffect, useState } from "react";

import Button from "../components/ui/Button";
import Modal from "../components/ui/Modal";
import AlertList from "../components/alerts/AlertList";
import AlertRuleForm from "../components/alerts/AlertRuleForm";
import { alertService } from "../services/alertService";
import { serverService } from "../services/serverService";
import { when } from "../utils/formatters";
import type { AlertEvent, AlertRule, AlertRuleInput } from "../types/alert";
import type { Server } from "../types/server";

type Tab = "rules" | "history";

export default function Alerts() {
  const [tab, setTab] = useState<Tab>("rules");
  const [rules, setRules] = useState<AlertRule[]>([]);
  const [history, setHistory] = useState<AlertEvent[]>([]);
  const [servers, setServers] = useState<Server[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);

  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<AlertRule | null>(null);

  useEffect(() => {
    Promise.all([alertService.listRules(), alertService.history(), serverService.list()])
      .then(([r, h, s]) => {
        setRules(r);
        setHistory(h);
        setServers(s);
      })
      .catch(() => setError(true))
      .finally(() => setLoading(false));
  }, []);

  function openNew() {
    setEditing(null);
    setFormOpen(true);
  }

  function openEdit(rule: AlertRule) {
    setEditing(rule);
    setFormOpen(true);
  }

  async function save(input: AlertRuleInput) {
    if (editing) {
      const updated = await alertService.updateRule({ ...editing, ...input });
      setRules((rs) => rs.map((r) => (r.id === updated.id ? updated : r)));
    } else {
      const created = await alertService.createRule(input);
      setRules((rs) => [...rs, created]);
    }
    setFormOpen(false);
  }

  async function toggle(rule: AlertRule) {
    const next = { ...rule, enabled: !rule.enabled };
    setRules((rs) => rs.map((r) => (r.id === rule.id ? next : r))); // update right away
    try {
      await alertService.updateRule(next);
    } catch {
      setRules((rs) => rs.map((r) => (r.id === rule.id ? rule : r))); // put it back
    }
  }

  async function remove(rule: AlertRule) {
    if (!window.confirm(`Delete "${rule.name}"? This cannot be undone.`)) return;
    await alertService.deleteRule(rule.id);
    setRules((rs) => rs.filter((r) => r.id !== rule.id));
  }

  return (
    <div className="mx-auto max-w-5xl">
      <header className="mb-5 flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-[28px] font-bold tracking-tight">Alerts</h1>
          <p className="text-muted">Get notified when a server crosses a limit.</p>
        </div>
        <Button onClick={openNew}>New rule</Button>
      </header>

      <div role="tablist" className="mb-5 flex gap-1 border-b border-line">
        {(["rules", "history"] as Tab[]).map((t) => (
          <button
            key={t}
            role="tab"
            aria-selected={tab === t}
            onClick={() => setTab(t)}
            className={`-mb-px border-b-2 px-3.5 py-2.5 font-semibold capitalize ${
              tab === t ? "border-cpu text-ink" : "border-transparent text-muted"
            }`}
          >
            {t}
            <span className="ml-1.5 rounded-full border border-line bg-panel px-2 text-xs">
              {t === "rules" ? rules.length : history.length}
            </span>
          </button>
        ))}
      </div>

      {loading ? (
        <p className="text-muted">Loading alerts...</p>
      ) : error ? (
        <p role="alert" className="text-bad">Could not load alerts. Refresh to try again.</p>
      ) : tab === "rules" ? (
        rules.length === 0 ? (
          <section className="rounded-xl border border-line bg-panel p-10 text-center">
            <h2 className="text-lg font-semibold">No alert rules yet</h2>
            <p className="mb-4 mt-1 text-muted">Create a rule to get notified about high CPU, memory or disk use.</p>
            <Button onClick={openNew}>New rule</Button>
          </section>
        ) : (
          <AlertList rules={rules} servers={servers} onToggle={toggle} onEdit={openEdit} onDelete={remove} />
        )
      ) : history.length === 0 ? (
        <p className="text-muted">No alerts have fired yet.</p>
      ) : (
        <section className="overflow-x-auto rounded-xl border border-line bg-panel px-5 py-1.5">
          <table className="w-full">
            <thead>
              <tr className="text-left text-[13px] text-muted">
                <th className="pb-2 pt-3.5 font-medium">Rule</th>
                <th className="pb-2 pt-3.5 font-medium">Server</th>
                <th className="pb-2 pt-3.5 font-medium">When</th>
                <th className="pb-2 pt-3.5 text-right font-medium">Status</th>
              </tr>
            </thead>
            <tbody>
              {history.map((h) => (
                <tr key={h.id} className="border-t border-line">
                  <td className="py-3">{h.ruleName}</td>
                  <td className="py-3">{h.serverName}</td>
                  <td className="py-3">{when(h.at)}</td>
                  <td className="py-3 text-right">
                    <span className="inline-flex items-center gap-1.5 rounded-full border border-line px-2.5 py-0.5 text-[13px] font-medium">
                      <i className={`h-[7px] w-[7px] rounded-full ${h.status === "firing" ? "bg-bad" : "bg-ok"}`} />
                      {h.status === "firing" ? "Firing" : "Resolved"}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </section>
      )}

      <Modal
        open={formOpen}
        onClose={() => setFormOpen(false)}
        title={editing ? "Edit alert rule" : "New alert rule"}
      >
        <AlertRuleForm
          rule={editing}
          servers={servers}
          onSubmit={save}
          onCancel={() => setFormOpen(false)}
        />
      </Modal>
    </div>
  );
}