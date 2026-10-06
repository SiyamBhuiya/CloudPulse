import { useEffect, useState } from "react";

import Button from "../components/ui/Button";
import ServerCard from "../components/servers/ServerCard";
import AddServerModal from "../components/servers/AddServerModal";
import { serverService } from "../services/serverService";
import type { Server } from "../types/server";

export default function Servers() {
  const [servers, setServers] = useState<Server[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);

  useEffect(() => {
    serverService
      .list()
      .then(setServers)
      .catch(() => setError(true))
      .finally(() => setLoading(false));
  }, []);

  const online = servers.filter((s) => s.status !== "down").length;

  return (
    <div className="mx-auto max-w-5xl">
      <header className="mb-6 flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-[28px] font-bold tracking-tight">Servers</h1>
          <p className="text-muted">
            {servers.length} {servers.length === 1 ? "server" : "servers"} · {online} online
          </p>
        </div>
        <Button onClick={() => setModalOpen(true)}>Add server</Button>
      </header>

      {loading ? (
        <p className="text-muted">Loading servers...</p>
      ) : error ? (
        <p role="alert" className="text-bad">
          Could not load servers. Refresh to try again.
        </p>
      ) : servers.length === 0 ? (
        <section className="rounded-xl border border-line bg-panel p-10 text-center">
          <h2 className="text-lg font-semibold">No servers yet</h2>
          <p className="mb-4 mt-1 text-muted">
            Add your first server to start seeing live metrics.
          </p>
          <Button onClick={() => setModalOpen(true)}>Add server</Button>
        </section>
      ) : (
        <div className="grid grid-cols-[repeat(auto-fill,minmax(300px,1fr))] gap-4">
          {servers.map((s) => (
            <ServerCard key={s.id} server={s} />
          ))}
        </div>
      )}

      <AddServerModal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        onCreated={(s) => setServers((prev) => [...prev, s])}
      />
    </div>
  );
}