import Link from "next/link";
import { ArrowUpRight, type LucideIcon } from "lucide-react";

import { StatusBadge } from "@/components/ui/status-badge";

type PreviewLinkCardProps = {
  href: string;
  label: string;
  description: string;
  icon: LucideIcon;
  plannedTask?: string;
};

export function PreviewLinkCard({ href, label, description, icon: CardIcon, plannedTask }: PreviewLinkCardProps) {
  return (
    <Link
      href={href}
      className="group flex h-full flex-col rounded-2xl border border-border bg-surface/70 p-5 transition-colors hover:border-accent/50 hover:bg-surface-raised focus-visible:ring-2 focus-visible:ring-highlight focus-visible:outline-none"
    >
      <div className="mb-4 flex items-start justify-between gap-3">
        <span className="flex size-10 items-center justify-center rounded-xl border border-border-strong bg-background/60 text-accent-soft">
          <CardIcon className="size-5" />
        </span>
        <ArrowUpRight className="size-4 text-subtle transition-colors group-hover:text-highlight" aria-hidden="true" />
      </div>
      <p className="font-medium text-foreground">{label}</p>
      <p className="mt-1 flex-1 text-sm text-muted">{description}</p>
      {plannedTask && <StatusBadge className="mt-4 self-start" label={`Planned in ${plannedTask}`} tone="neutral" />}
    </Link>
  );
}
