"use client";

import Link from "next/link";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/features/auth/session/use-auth";
import { tradeLoopSteps } from "@/features/landing/trade-loop";

export function HeroPitch() {
  const { status } = useAuth();
  const isSignedIn = status === "authenticated";

  return (
    <div className="flex flex-col items-start">
      <h1 className="font-display text-7xl leading-[0.9] tracking-[0.02em] text-foreground uppercase drop-shadow-[0_4px_24px_rgb(0_0_0/0.6)] sm:text-8xl lg:text-9xl">
        Fight make{" "}
        <span className="bg-linear-to-b from-amber-300 via-accent to-red-600 bg-clip-text text-transparent">rich</span>
      </h1>
      <p className="mt-5 max-w-md text-lg leading-relaxed text-muted">
        Every fighter starts with nothing. Some never stop fighting.
      </p>

      <ol className="mt-8 w-full max-w-md space-y-3">
        {tradeLoopSteps.map((tradeLoopStep, stepIndex) => {
          const StepIcon = tradeLoopStep.icon;
          return (
            <li
              key={tradeLoopStep.title}
              className="flex items-center gap-4 rounded-xl border border-border bg-surface/60 p-3 backdrop-blur"
            >
              <span className="flex size-10 shrink-0 items-center justify-center rounded-lg bg-accent/10 text-accent">
                <StepIcon className="size-5" aria-hidden="true" />
              </span>
              <div>
                <p className="text-sm font-semibold text-foreground">
                  <span className="mr-2 font-mono text-xs text-accent-soft/70">0{stepIndex + 1}</span>
                  {tradeLoopStep.title}
                </p>
                <p className="text-xs text-muted">{tradeLoopStep.body}</p>
              </div>
            </li>
          );
        })}
      </ol>

      <div className="mt-8 flex flex-wrap gap-3">
        {isSignedIn ? (
          <Button asChild size="lg">
            <Link href="/fight">Continue fighting</Link>
          </Button>
        ) : (
          <>
            <Button asChild size="lg">
              <Link href="/signup">Start fighting free</Link>
            </Button>
            <Button asChild size="lg" variant="secondary">
              <Link href="/login">Log in</Link>
            </Button>
          </>
        )}
      </div>
    </div>
  );
}
