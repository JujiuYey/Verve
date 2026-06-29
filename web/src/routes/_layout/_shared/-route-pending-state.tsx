export function RoutePendingState({ message }: { message: string }) {
  return (
    <div
      className="absolute inset-0 z-50 flex min-h-full w-full items-center justify-center bg-background"
      role="status"
      aria-live="polite"
    >
      <div className="flex flex-col items-center justify-center px-6 text-center">
        <div className="route-pending-orbit" aria-hidden="true">
          <span className="route-pending-dot route-pending-dot-1" />
          <span className="route-pending-dot route-pending-dot-2" />
          <span className="route-pending-dot route-pending-dot-3" />
        </div>

        <div className="mt-6 space-y-1.5">
          <div className="text-base text-foreground">{message}</div>
          <div className="text-xs text-muted-foreground">正在加载</div>
        </div>
      </div>
    </div>
  );
}
