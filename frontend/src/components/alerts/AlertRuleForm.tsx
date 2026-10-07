import { useId, useState, type FormEvent } from "react";

import Button from "../ui/Button";
import Input from "../ui/Input";
import type { Server } from "../../types/server";
import type { AlertMetric, AlertRule, AlertRuleInput, Channel } from "../../types/alert";

interface Props {
  rule: AlertRule | null; // null = creating a new rule
  servers: Server[];
  onSubmit: (input: AlertRuleInput) => Promise<void>;
  onCancel: () => void;
}

const selectClass =
  "w-full rounded-lg border border-line bg-bg px-3 py-2.5 text-ink focus-visible:outline focus-visible:outline-2 focus-visible:outline-net";

function Select({
  label,
  value,
  onChange,
  options,
}: {
  label: string;
  value: string;
  onChange: (v: string) => void;
  options: { value: string; label: string }[];
}) {
  const id = useId();
  return (
    <div className="mb-4">
      <label htmlFor={id} className="mb-1.5 block font-medium">
        {label}
      </label>
      <select id={id} value={value} onChange={(e) => onChange(e.target.value)} className={selectClass}>
        {options.map((o) => (
          <option key={o.value} value={o.value}>
            {o.label}
          </option>
        ))}
      </select>
    </div>
  );
}

export default function AlertRuleForm({ rule, servers, onSubmit, onCancel }: Props) {
  const [name, setName] = useState(rule?.name ?? "");
  const [serverId, setServerId] = useState(rule?.serverId ?? "all");
  const [metric, setMetric] = useState<AlertMetric>(rule?.metric ?? "cpu");
  const [threshold, setThreshold] = useState(String(rule?.threshold ?? 90));
  const [durationMin, setDurationMin] = useState(String(rule?.durationMin ?? 5));
  const [channel, setChannel] = useState<Channel>(rule?.channel ?? "email");
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setError(null);

    const t = Number(threshold);
    const d = Number(durationMin);
    if (!(t > 0 && t <= 100)) return setError("Threshold must be between 1 and 100.");
    if (!(d >= 1)) return setError("Duration must be at least 1 minute.");

    setSaving(true);
    try {
      await onSubmit({ name, serverId, metric, threshold: t, durationMin: d, channel });
    } catch {
      setError("Could not save the rule. Try again.");
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

      <Input
        label="Rule name"
        placeholder="High CPU on production"
        value={name}
        onChange={(e) => setName(e.target.value)}
        required
      />

      <Select
        label="Servers"
        value={serverId}
        onChange={setServerId}
        options={[
          { value: "all", label: "All servers" },
          ...servers.map((s) => ({ value: s.id, label: s.name })),
        ]}
      />

      <div className="grid grid-cols-3 gap-2.5 max-sm:grid-cols-1">
        <Select
          label="Metric"
          value={metric}
          onChange={(v) => setMetric(v as AlertMetric)}
          options={[
            { value: "cpu", label: "CPU" },
            { value: "memory", label: "Memory" },
            { value: "disk", label: "Disk" },
          ]}
        />
        <Input
          label="Above (%)"
          type="number"
          min={1}
          max={100}
          value={threshold}
          onChange={(e) => setThreshold(e.target.value)}
          required
        />
        <Input
          label="For (min)"
          type="number"
          min={1}
          value={durationMin}
          onChange={(e) => setDurationMin(e.target.value)}
          required
        />
      </div>

      <Select
        label="Notify by"
        value={channel}
        onChange={(v) => setChannel(v as Channel)}
        options={[
          { value: "email", label: "Email" },
          { value: "slack", label: "Slack" },
          { value: "webhook", label: "Webhook" },
        ]}
      />

      <div className="mt-1 flex justify-end gap-2.5">
        <button
          type="button"
          onClick={onCancel}
          className="rounded-lg border border-line px-4 py-2.5 font-semibold"
        >
          Cancel
        </button>
        <Button type="submit" disabled={saving}>
          {saving ? "Saving..." : rule ? "Save changes" : "Create rule"}
        </Button>
      </div>
    </form>
  );
}