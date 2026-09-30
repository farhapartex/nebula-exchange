"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { Check } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { fetchOnboardingProgress, onboardingProgressQueryKey } from "@/features/hangar/api/onboarding-api";
import { onboardingStepPresentation } from "@/features/hangar/onboarding-steps";
import { cn } from "@/utils/class-names";

export function OnboardingChecklist() {
  const progressQuery = useQuery({ queryKey: onboardingProgressQueryKey, queryFn: fetchOnboardingProgress });

  if (progressQuery.isPending) {
    return <Skeleton className="h-72 rounded-2xl" />;
  }
  if (!progressQuery.data || progressQuery.data.completed_count === progressQuery.data.steps.length) {
    return null;
  }

  const { steps, completed_count: completedCount } = progressQuery.data;
  const nextStep = steps.find((step) => !step.is_completed);

  return (
    <Card>
      <CardHeader
        title="Getting started"
        description={`${completedCount} of ${steps.length} done`}
        action={
          <div className="h-2 w-24 overflow-hidden rounded-full bg-border" aria-hidden="true">
            <div
              className="h-full rounded-full bg-accent transition-all"
              style={{ width: `${(completedCount / steps.length) * 100}%` }}
            />
          </div>
        }
      />
      <CardContent>
        <ol className="space-y-2">
          {steps.map((step) => {
            const presentation = onboardingStepPresentation[step.key];
            const StepIcon = presentation.icon;
            const isNextStep = step.key === nextStep?.key;
            return (
              <li
                key={step.key}
                className={cn(
                  "flex items-center gap-3 rounded-xl border p-3",
                  isNextStep ? "border-accent/50 bg-accent/5" : "border-border bg-background/40",
                )}
              >
                <span
                  className={cn(
                    "flex size-9 shrink-0 items-center justify-center rounded-lg",
                    step.is_completed ? "bg-up/15 text-up" : "bg-surface-raised text-muted",
                  )}
                >
                  {step.is_completed ? <Check className="size-4" aria-label="Done" /> : <StepIcon className="size-4" />}
                </span>
                <div className="min-w-0 flex-1">
                  <p
                    className={cn(
                      "text-sm font-medium",
                      step.is_completed ? "text-muted line-through" : "text-foreground",
                    )}
                  >
                    {presentation.title}
                  </p>
                  {!step.is_completed && <p className="text-xs text-muted">{presentation.description}</p>}
                </div>
                {isNextStep && (
                  <Button asChild size="sm">
                    <Link href={presentation.href}>{presentation.actionLabel}</Link>
                  </Button>
                )}
              </li>
            );
          })}
        </ol>
      </CardContent>
    </Card>
  );
}
