import type { ReactNode } from "react";

import { cn } from "@/utils/class-names";

type HubPanelProps = {
  title?: string;
  action?: ReactNode;
  children: ReactNode;
  className?: string;
};

export function HubPanel({ title, action, children, className }: HubPanelProps) {
  return (
    <section className={cn("rounded-2xl border border-border bg-surface/70 p-5 backdrop-blur", className)}>
      {(title || action) && (
        <div className="mb-4 flex items-center justify-between gap-3">
          {title && <h2 className="font-display text-xl tracking-[0.06em] text-foreground">{title}</h2>}
          {action}
        </div>
      )}
      {children}
    </section>
  );
}
