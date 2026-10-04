type Status = "online" | "warning" | "down";

const colors: Record<Status, string> = {
  online: "bg-ok",
  warning: "bg-cpu",
  down: "bg-bad",
};

export default function StatusBadge({ status }: { status: Status }) {
  return (
    <span
      className={`mr-2 inline-block h-2 w-2 rounded-full ${colors[status]}`}
      role="img"
      aria-label={status}
    />
  );
}