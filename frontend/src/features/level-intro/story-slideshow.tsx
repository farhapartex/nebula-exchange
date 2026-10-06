"use client";

import { useCallback, useEffect, useState, type CSSProperties } from "react";
import Link from "next/link";
import { ChevronLeft, ChevronRight, Play, X } from "lucide-react";

import { PageContainer } from "@/components/layout/page-container";
import { Button } from "@/components/ui/button";
import { CallToActionBackdrop } from "@/features/level-intro/call-to-action-backdrop";
import type { LevelStory } from "@/features/level-intro/level-story";
import { isFinalSlide, nextSlideIndex, previousSlideIndex } from "@/features/level-intro/slide-navigation";
import { resolveSlidePalette } from "@/features/level-intro/slide-palette";
import { SlideSceneImage } from "@/features/level-intro/slide-scene-image";
import { cn } from "@/utils/class-names";

type StorySlideshowProps = {
  levelStory: LevelStory;
  onPlay: () => void;
  isStartingFight: boolean;
  startFightError: string | null;
};

export function StorySlideshow({ levelStory, onPlay, isStartingFight, startFightError }: StorySlideshowProps) {
  const slideCount = levelStory.slides.length + 1;
  const [slideIndex, setSlideIndex] = useState(0);
  const isShowingCallToAction = isFinalSlide(slideIndex, slideCount);
  const currentSlide = isShowingCallToAction ? null : levelStory.slides[slideIndex];
  const callToAction = levelStory.callToAction;
  const palette = resolveSlidePalette(currentSlide?.palette ?? callToAction.palette);
  const sceneImage = currentSlide ? currentSlide.image : callToAction.image;

  const goForward = useCallback(() => setSlideIndex((index) => nextSlideIndex(index, slideCount)), [slideCount]);
  const goBack = useCallback(() => setSlideIndex((index) => previousSlideIndex(index)), []);

  useEffect(() => {
    function handleKey(keyEvent: KeyboardEvent) {
      if (keyEvent.key === "ArrowRight" || keyEvent.key === " ") {
        keyEvent.preventDefault();
        goForward();
      } else if (keyEvent.key === "ArrowLeft") {
        goBack();
      }
    }
    window.addEventListener("keydown", handleKey);
    return () => window.removeEventListener("keydown", handleKey);
  }, [goBack, goForward]);

  const accentStyle = { color: palette.accent } satisfies CSSProperties;

  return (
    <div
      className="relative flex min-h-svh flex-col overflow-hidden transition-colors duration-1000"
      style={{ backgroundColor: palette.background }}
    >
      <div aria-hidden="true" className="pointer-events-none absolute inset-0">
        <div
          className={cn(
            "absolute -bottom-56 size-[52rem] -translate-x-1/2 rounded-full blur-3xl transition-all duration-1000",
            isShowingCallToAction ? "left-1/2" : "left-1/4",
          )}
          style={{ backgroundColor: palette.glow }}
        />
      </div>
      {sceneImage &&
        (isShowingCallToAction ? (
          <CallToActionBackdrop key={sceneImage} source={sceneImage} />
        ) : (
          <SlideSceneImage key={sceneImage} source={sceneImage} />
        ))}

      <header className="relative">
        <PageContainer className="flex items-center gap-4 pt-5">
          <Link
            href="/fight"
            aria-label="Leave the level"
            className="flex size-9 shrink-0 items-center justify-center rounded-lg text-muted transition-colors hover:bg-surface-raised hover:text-foreground"
          >
            <X className="size-5" />
          </Link>
          <ol className="flex flex-1 gap-1.5" aria-label="Story progress">
            {Array.from({ length: slideCount }, (_, progressIndex) => (
              <li key={progressIndex} className="h-1 flex-1 overflow-hidden rounded-full bg-foreground/10">
                <span
                  className="block h-full rounded-full transition-all duration-500"
                  style={{ width: progressIndex <= slideIndex ? "100%" : "0%", backgroundColor: palette.accent }}
                />
              </li>
            ))}
          </ol>
        </PageContainer>
      </header>

      <main className={cn("relative flex flex-1", isShowingCallToAction ? "items-start pt-[6vh]" : "items-center")}>
        <PageContainer className="py-10">
          {currentSlide ? (
            <div key={currentSlide.id} className="max-w-xl motion-safe:animate-slide-in lg:ml-[6%]">
              {currentSlide.eyebrow && (
                <p className="mb-4 text-xs font-semibold tracking-[0.22em] uppercase" style={accentStyle}>
                  {currentSlide.eyebrow}
                </p>
              )}
              <h1 className="font-display text-5xl leading-[0.95] tracking-[0.02em] text-foreground drop-shadow-[0_4px_24px_rgb(0_0_0/0.7)] sm:text-7xl">
                {currentSlide.heading}
              </h1>
              <span
                aria-hidden="true"
                className="mt-6 block h-1 w-16 rounded-full"
                style={{ backgroundColor: palette.accent }}
              />
              <p className="mt-6 text-lg leading-relaxed text-foreground/80 sm:text-xl">{currentSlide.body}</p>
            </div>
          ) : (
            <div
              key="call-to-action"
              className="mx-auto flex max-w-3xl flex-col items-center text-center motion-safe:animate-pop-in"
            >
              <h1 className="font-display text-5xl leading-[0.92] tracking-[0.02em] text-foreground drop-shadow-[0_4px_24px_rgb(0_0_0/0.85)] sm:text-7xl">
                {callToAction.heading}
              </h1>
              <p className="mt-4 text-lg text-foreground/85 drop-shadow-[0_2px_12px_rgb(0_0_0/0.9)]">
                {callToAction.body}
              </p>
              <Button
                size="lg"
                onClick={onPlay}
                isLoading={isStartingFight}
                className="mt-8 h-16 px-14 font-display text-3xl tracking-[0.12em]"
                style={{ backgroundColor: palette.accent, boxShadow: `0 0 48px -6px ${palette.accent}` }}
              >
                {!isStartingFight && <Play className="size-6 fill-current" aria-hidden="true" />}
                {callToAction.button_label}
              </Button>
              {startFightError && (
                <p
                  role="alert"
                  className="mt-4 rounded-lg border border-down/40 bg-background/80 px-4 py-2 text-sm text-foreground"
                >
                  {startFightError}
                </p>
              )}
            </div>
          )}
        </PageContainer>
      </main>

      {!isShowingCallToAction && (
        <footer className="relative">
          <PageContainer className="flex items-center justify-between pb-8">
            <Button variant="ghost" onClick={goBack} disabled={slideIndex === 0}>
              <ChevronLeft className="size-4" aria-hidden="true" />
              Back
            </Button>
            <span className="font-mono text-xs text-subtle tabular-nums">
              {slideIndex + 1} / {levelStory.slides.length}
            </span>
            <Button onClick={goForward} style={{ backgroundColor: palette.accent, color: palette.background }}>
              Next
              <ChevronRight className="size-4" aria-hidden="true" />
            </Button>
          </PageContainer>
        </footer>
      )}
    </div>
  );
}
