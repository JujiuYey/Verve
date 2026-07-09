export function Metric({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="flex min-h-16 flex-col justify-center gap-1 rounded-md border bg-background px-3">
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className="text-lg font-semibold leading-none">{value}</div>
    </div>
  );
}
