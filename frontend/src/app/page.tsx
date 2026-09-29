import { PageContainer } from "@/components/layout/page-container";
import { MarketSignalSection } from "@/features/theme-preview/market-signal-section";
import { PaletteSection } from "@/features/theme-preview/palette-section";
import { TypographySection } from "@/features/theme-preview/typography-section";

export default function HomePage() {
  return (
    <PageContainer className="py-10 sm:py-14">
      <div className="mb-10 max-w-2xl">
        <p className="mb-3 text-xs font-semibold tracking-[0.2em] text-highlight uppercase">Theme preview</p>
        <h1 className="text-3xl font-semibold tracking-tight text-foreground sm:text-4xl">Nebula Exchange</h1>
        <p className="mt-3 text-muted">
          A placeholder page to review the dark space theme before we build real screens. The landing page replaces it
          later.
        </p>
      </div>
      <div className="grid gap-6 lg:grid-cols-5">
        <div className="lg:col-span-3">
          <PaletteSection />
        </div>
        <div className="space-y-6 lg:col-span-2">
          <TypographySection />
          <MarketSignalSection />
        </div>
      </div>
    </PageContainer>
  );
}
