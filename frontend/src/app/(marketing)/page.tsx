import Link from "next/link";
import { ArrowRight } from "lucide-react";

import { primaryNavigationLinks, profileNavigationLinks } from "@/components/app-shell/navigation-links";
import { PageContainer } from "@/components/layout/page-container";
import { Button } from "@/components/ui/button";
import { developerPages } from "@/features/build-preview/developer-pages";
import { PreviewLinkCard } from "@/features/build-preview/preview-link-card";
import { PreviewLinkGrid } from "@/features/build-preview/preview-link-grid";

export default function BuildPreviewPage() {
  const isDevelopment = process.env.NODE_ENV !== "production";

  return (
    <PageContainer className="py-10 sm:py-14">
      <div className="mb-10 flex flex-col gap-6 sm:flex-row sm:items-end sm:justify-between">
        <div className="max-w-2xl">
          <p className="mb-3 text-xs font-semibold tracking-[0.2em] text-highlight uppercase">Build preview</p>
          <h1 className="text-3xl font-semibold tracking-tight text-foreground sm:text-4xl">Nebula Exchange</h1>
          <p className="mt-3 text-muted">
            Every screen we have so far, in one place. The real landing page replaces this hub later in the plan.
          </p>
        </div>
        <Button asChild size="lg">
          <Link href="/hangar">
            Enter the game
            <ArrowRight className="size-4" />
          </Link>
        </Button>
      </div>

      <PreviewLinkGrid title="Game" description="The main sections from the top navigation.">
        {primaryNavigationLinks.map((navigationLink) => (
          <PreviewLinkCard key={navigationLink.href} {...navigationLink} />
        ))}
      </PreviewLinkGrid>

      <PreviewLinkGrid title="Account" description="Sections reached from the profile menu.">
        {profileNavigationLinks.map((navigationLink) => (
          <PreviewLinkCard key={navigationLink.href} {...navigationLink} />
        ))}
      </PreviewLinkGrid>

      {isDevelopment && (
        <PreviewLinkGrid title="Developer" description="Design and component references. Hidden in production.">
          {developerPages.map((developerPage) => (
            <PreviewLinkCard key={developerPage.href} {...developerPage} />
          ))}
        </PreviewLinkGrid>
      )}
    </PageContainer>
  );
}
