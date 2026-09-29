import { Construction } from "lucide-react";

import { navigationLinkFor, type GameSection } from "@/components/app-shell/navigation-links";
import { PageContainer } from "@/components/layout/page-container";
import { StatusBadge } from "@/components/ui/status-badge";

export function SectionPlaceholder({ section }: { section: GameSection }) {
  const navigationLink = navigationLinkFor(section);
  const SectionIcon = navigationLink.icon;

  return (
    <PageContainer className="py-8 sm:py-10">
      <header className="mb-8 flex items-start gap-4">
        <span className="flex size-12 shrink-0 items-center justify-center rounded-xl border border-border-strong bg-surface-raised text-accent-soft">
          <SectionIcon className="size-6" />
        </span>
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-foreground">{navigationLink.label}</h1>
          <p className="mt-1 max-w-2xl text-sm text-muted">{navigationLink.description}</p>
        </div>
      </header>
      <div className="flex flex-col items-center justify-center rounded-2xl border border-dashed border-border-strong bg-surface/40 px-6 py-16 text-center">
        <Construction className="mb-3 size-8 text-subtle" aria-hidden="true" />
        <p className="text-sm font-medium text-foreground">This screen is not built yet</p>
        <p className="mt-1 max-w-sm text-sm text-muted">It arrives with its task in the build plan.</p>
        <StatusBadge className="mt-4" label={`Planned in ${navigationLink.plannedTask}`} tone="accent" />
      </div>
    </PageContainer>
  );
}
