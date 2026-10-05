import { PageContainer } from "@/components/layout/page-container";
import { HeroCharacter } from "@/features/landing/hero-character";
import { HeroPitch } from "@/features/landing/hero-pitch";

export default function LandingPage() {
  return (
    <section className="relative flex min-h-[calc(100svh-4rem)] items-center overflow-hidden">
      <div aria-hidden="true" className="pointer-events-none absolute inset-0">
        <div className="absolute -bottom-40 -left-20 size-[36rem] rounded-full bg-accent/15 blur-3xl" />
        <div className="absolute -top-20 right-0 size-[28rem] rounded-full bg-highlight/10 blur-3xl" />
      </div>
      <PageContainer className="relative grid items-center gap-10 py-10 lg:grid-cols-2">
        <HeroCharacter />
        <HeroPitch />
      </PageContainer>
    </section>
  );
}
