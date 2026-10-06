import { useEffect, useRef, useState } from "react";
import { HISTORY_SIZE, TOKEN_KEY, WS_URL } from "../utils/constants";
import type { MetricSnapshot } from "../types/metric";

const MOCK = import.meta.env.VITE_MOCK === "true";

const clamp = (v: number, lo: number, hi: number) => Math.max(lo, Math.min(hi, v));
const walk = (v: number, step: number, lo: number, hi: number) =>
  clamp(v + (Math.random() - 0.5) * step, lo, hi);

export function useWebSocket(serverId?: string) {
  const [history, setHistory] = useState<MetricSnapshot[]>([]);
  const [connected, setConnected] = useState(false);
  const retry = useRef(0);

  useEffect(() => {
    // Fake data for building the UI without a backend
    if (MOCK) {
      let cpu = 72, memory = 61, disk = 48, network = 24, requests = 18000;

      const makeSnap = (): MetricSnapshot => {
        cpu = walk(cpu, 8, 30, 95);
        memory = walk(memory, 2, 50, 75);
        disk = walk(disk, 0.2, 46, 50);
        network = walk(network, 8, 5, 58);
        requests += Math.round(Math.random() * 40);
        return {
          serverId: serverId ?? "mock",
          timestamp: Date.now(),
          cpu,
          memory,
          disk,
          network,
          requests,
          errorRate: 0.43,
          p95: walk(184, 30, 120, 260),
        };
      };

      // Start with a full graph, then add one point per second
      setHistory(Array.from({ length: HISTORY_SIZE }, makeSnap));
      setConnected(true);
      const id = window.setInterval(
        () => setHistory((h) => [...h, makeSnap()].slice(-HISTORY_SIZE)),
        1000
      );
      return () => window.clearInterval(id);
    }

    // Real WebSocket
    let ws: WebSocket | null = null;
    let timer: number | undefined;
    let closed = false;

    const connect = () => {
      const token = localStorage.getItem(TOKEN_KEY) ?? "";
      const qs = new URLSearchParams({ token });
      if (serverId) qs.set("server", serverId);

      ws = new WebSocket(`${WS_URL}?${qs}`);

      ws.onopen = () => {
        retry.current = 0;
        setConnected(true);
      };

      ws.onmessage = (e) => {
        try {
          const snap = JSON.parse(e.data) as MetricSnapshot;
          setHistory((h) => [...h, snap].slice(-HISTORY_SIZE));
        } catch {
          /* ignore malformed messages */
        }
      };

      ws.onclose = () => {
        setConnected(false);
        if (closed) return;
        // Reconnect with backoff: 1s, 2s, 4s ... max 15s
        const delay = Math.min(1000 * 2 ** retry.current++, 15000);
        timer = window.setTimeout(connect, delay);
      };
    };

    connect();

    return () => {
      closed = true;
      window.clearTimeout(timer);
      ws?.close();
    };
  }, [serverId]);

  return { history, latest: history[history.length - 1] ?? null, connected };
}