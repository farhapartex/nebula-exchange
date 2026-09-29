import type { Metadata } from "next";
import { notFound } from "next/navigation";

import { PageContainer } from "@/components/layout/page-container";
import { ButtonShowcase } from "@/features/ui-showcase/button-showcase";
import { CardShowcase } from "@/features/ui-showcase/card-showcase";
import { FeedbackShowcase } from "@/features/ui-showcase/feedback-showcase";
import { FormShowcase } from "@/features/ui-showcase/form-showcase";
import { OverlayShowcase } from "@/features/ui-showcase/overlay-showcase";
import { TabsShowcase } from "@/features/ui-showcase/tabs-showcase";

export const metadata: Metadata = {
  title: "UI components",
};

export default function UiShowcasePage() {
  if (process.env.NODE_ENV === "production") {
    notFound();
  }

  return (
    <PageContainer className="py-10 sm:py-14">
      <div className="mb-10 max-w-2xl">
        <p className="mb-3 text-xs font-semibold tracking-[0.2em] text-highlight uppercase">Developer preview</p>
        <h1 className="text-3xl font-semibold tracking-tight text-foreground sm:text-4xl">UI components</h1>
        <p className="mt-3 text-muted">
          Every shared component in one place. This page is not available in production.
        </p>
      </div>
      <div className="grid gap-6 lg:grid-cols-2">
        <ButtonShowcase />
        <FormShowcase />
        <OverlayShowcase />
        <TabsShowcase />
        <CardShowcase />
        <div className="lg:col-span-2">
          <FeedbackShowcase />
        </div>
      </div>
    </PageContainer>
  );
}
