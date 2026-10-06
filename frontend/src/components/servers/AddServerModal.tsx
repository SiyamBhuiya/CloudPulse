import { useState, type FormEvent } from "react";

import Button from "../ui/Button";
import Input from "../ui/Input";
import Modal from "../ui/Modal";
import { serverService } from "../../services/serverService";
import type { Server } from "../../types/server";

interface Props {
  open: boolean;
  onClose: () => void;
  onCreated: (server: Server) => void;
}

export default function AddServerModal({ open, onClose, onCreated }: Props) {
  const [name, setName] = useState("");
  const [region, setRegion] = useState("");
  const [agentKey, setAgentKey] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  const command = `./cloudpulse-agent --key ${agentKey}`;

  function close() {
    onClose();
    // Reset after the modal closes so the next open starts fresh
    setTimeout(() => {
      setName("");
      setRegion("");
      setAgentKey(null);
      setError(null);
    }, 200);
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      const res = await serverService.create({ name, region: region || undefined });
      setAgentKey(res.agentKey);
      onCreated(res.server);
    } catch {
      setError("Could not create the server. Try again.");
    } finally {
      setLoading(false);
    }
  }

  async function copy() {
    try {
      await navigator.clipboard.writeText(command);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch {
      /* clipboard blocked: the user can select the text manually */
    }
  }

  return (
    <Modal open={open} onClose={close} title={agentKey ? "Start the agent" : "Add a server"}>
      {!agentKey ? (
        <form onSubmit={onSubmit}>
          <p className="mb-4 mt-1 text-muted">
            Give it a name you will recognize. You get an agent key next.
          </p>

          {error && (
            <div role="alert" className="mb-4 rounded-lg border border-bad px-3 py-2 text-sm">
              {error}
            </div>
          )}

          <Input
            label="Server name"
            placeholder="api-prod-3"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
          />
          <Input
            label="Region (optional)"
            placeholder="Frankfurt"
            value={region}
            onChange={(e) => setRegion(e.target.value)}
          />

          <div className="mt-1.5 flex justify-end gap-2.5">
            <button
              type="button"
              onClick={close}
              className="rounded-lg border border-line px-4 py-2.5 font-semibold"
            >
              Cancel
            </button>
            <Button type="submit" disabled={loading}>
              {loading ? "Creating..." : "Create server"}
            </Button>
          </div>
        </form>
      ) : (
        <div>
          <p className="mb-4 mt-1 text-muted">
            Run this on the server. Metrics appear within a few seconds.
          </p>

          <div className="mb-3.5 flex items-center gap-2 rounded-xl border border-line bg-bg px-3 py-2.5">
            <code className="flex-1 overflow-x-auto whitespace-nowrap font-mono text-[13px] font-medium text-mem">
              {command}
            </code>
            <button
              onClick={copy}
              className="rounded-md border border-line px-2.5 py-1 text-[13px] font-semibold"
            >
              {copied ? "Copied" : "Copy"}
            </button>
          </div>

          <p className="mb-2 text-[13px] text-muted">
            Copy the key now. For safety, it is only shown once.
          </p>

          <div className="flex justify-end">
            <Button onClick={close}>Done</Button>
          </div>
        </div>
      )}
    </Modal>
  );
}