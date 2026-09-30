import { Construction } from "lucide-react";

import { navigationLinkFor, type GameSection } from "@/components/app-shell/navigation-links";
import { PageContainer } from "@/components/layout/page-container";
import { SectionHeader } from "@/components/layout/section-header";
import { StatusBadge } from "@/components/ui/status-badge";

export function SectionPlaceholder({ section }: { section: GameSection }) {
  const navigationLink = navigationLinkFor(section);

  return (
    <PageContainer className="py-8 sm:py-10">
      <SectionHeader title={navigationLink.label} description={navigationLink.description} icon={navigationLink.icon} />
      <div className="flex flex-col items-center justify-center rounded-2xl border border-dashed border-border-strong bg-surface/40 px-6 py-16 text-center">
        <Construction className="mb-3 size-8 text-subtle" aria-hidden="true" />
        <p className="text-sm font-medium text-foreground">This screen is not built yet</p>
        <p className="mt-1 max-w-sm text-sm text-muted">It arrives with its task in the build plan.</p>
        <StatusBadge className="mt-4" label={`Planned in ${navigationLink.plannedTask}`} tone="accent" />
      </div>
    </PageContainer>
  );
}
