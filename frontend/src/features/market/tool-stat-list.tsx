import { toolStatLines } from "@/features/market/tool-stat-lines";

export function ToolStatList({ baseStats }: { baseStats: Record<string, number> }) {
  return (
    <ul className="space-y-1 text-xs text-muted">
      {toolStatLines(baseStats).map((statLine) => (
        <li key={statLine} className="flex items-center gap-2">
          <span className="size-1 shrink-0 rounded-full bg-accent" aria-hidden="true" />
          {statLine}
        </li>
      ))}
    </ul>
  );
}
