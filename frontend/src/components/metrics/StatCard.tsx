interface Props {
  label: string;
  value: string;
  unit?: string;
}

export default function StatCard({ label, value, unit }: Props) {
  return (
    <div>
      <small className="block text-sm text-muted">{label}</small>
      <strong className="text-[26px] font-bold tracking-tight">
        {value}
        {unit && <em className="ml-1 text-sm font-medium not-italic text-muted">{unit}</em>}
      </strong>
    </div>
  );
}