import type { ReactNode } from "react";
import type { LucideIcon } from "lucide-react";

type SectionHeaderProps = {
  title: string;
  description?: string;
  icon: LucideIcon;
  action?: ReactNode;
};

export function SectionHeader({ title, description, icon: SectionIcon, action }: SectionHeaderProps) {
  return (
    <header className="mb-8 flex flex-wrap items-start gap-4">
      <span className="flex size-12 shrink-0 items-center justify-center rounded-xl border border-border-strong bg-surface-raised text-accent-soft">
        <SectionIcon className="size-6" />
      </span>
      <div className="min-w-0 flex-1">
        <h1 className="text-2xl font-semibold tracking-tight text-foreground">{title}</h1>
        {description && <p className="mt-1 max-w-2xl text-sm text-muted">{description}</p>}
      </div>
      {action && <div className="shrink-0">{action}</div>}
    </header>
  );
}
