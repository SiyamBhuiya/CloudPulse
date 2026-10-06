export const percent = (n: number) => `${Math.round(n)}%`;
export const mbps = (n: number) => `${Math.round(n)} MB/s`;
export const ms = (n: number) => `${Math.round(n)}`;
export const number = (n: number) => n.toLocaleString("en-US");

export const timeAgo = (ts: number) => {
  const s = Math.max(0, Math.round((Date.now() - ts) / 1000));
  if (s < 60) return `${s}s ago`;
  if (s < 3600) return `${Math.round(s / 60)} min ago`;
  return `${Math.round(s / 3600)} h ago`;
};
export const uptime = (n: number) => `${n.toFixed(2)}%`;