import { useEffect, useRef, useState } from "react";
import { HISTORY_SIZE, TOKEN_KEY, WS_URL } from "../utils/constants";
import type { MetricSnapshot } from "../types/metric";

export function useWebSocket(serverId?: string) {
  const [history, setHistory] = useState<MetricSnapshot[]>([]);
  const [connected, setConnected] = useState(false);
  const retry = useRef(0);

  useEffect(() => {
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