import type { ReactNode } from "react";

type UsageSectionProps = {
  title: string;
  emptyMessage: string;
  children: ReactNode[];
};

export function UsageSection({ title, emptyMessage, children }: UsageSectionProps) {
  return (
    <section className="rounded-2xl border border-border bg-surface/80 p-5">
      <h2 className="text-sm font-semibold text-foreground">{title}</h2>
      {children.length === 0 ? (
        <p className="mt-3 text-sm text-subtle">{emptyMessage}</p>
      ) : (
        <ul className="mt-3 space-y-3">{children}</ul>
      )}
    </section>
  );
}

type UsageRowProps = {
  label: ReactNode;
  detail?: ReactNode;
  children?: ReactNode;
};

export function UsageRow({ label, detail, children }: UsageRowProps) {
  return (
    <li className="rounded-xl border border-border bg-background/50 p-3">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <p className="text-sm font-medium text-foreground">{label}</p>
        {detail && <p className="text-xs text-muted">{detail}</p>}
      </div>
      {children && <div className="mt-2 flex flex-wrap gap-2">{children}</div>}
    </li>
  );
}
