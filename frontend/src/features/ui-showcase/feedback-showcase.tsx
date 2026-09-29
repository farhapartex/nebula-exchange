"use client";

import { useState } from "react";
import { Rocket } from "lucide-react";

import { ContentSection } from "@/components/layout/content-section";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/empty-state";
import { ErrorState } from "@/components/ui/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { StatusBadge } from "@/components/ui/status-badge";
import type { Tone } from "@/components/ui/tone";

const statusExamples: { label: string; tone: Tone; isPulsing?: boolean }[] = [
  { label: "Open", tone: "info" },
  { label: "Partially filled", tone: "accent" },
  { label: "Filled", tone: "success" },
  { label: "Pending review", tone: "warning", isPulsing: true },
  { label: "Rejected", tone: "danger" },
  { label: "Cancelled", tone: "neutral" },
];

export function FeedbackShowcase() {
  const [isRetrying, setIsRetrying] = useState(false);

  function simulateRetry() {
    setIsRetrying(true);
    window.setTimeout(() => setIsRetrying(false), 1200);
  }

  return (
    <ContentSection
      title="StatusBadge, Skeleton, EmptyState and ErrorState"
      description="Loading, empty and failure feedback."
    >
      <div className="space-y-6">
        <div className="flex flex-wrap gap-2">
          {statusExamples.map((statusExample) => (
            <StatusBadge key={statusExample.label} {...statusExample} />
          ))}
        </div>
        <div className="space-y-2">
          <Skeleton className="h-5 w-1/3" />
          <Skeleton className="h-4 w-full" />
          <Skeleton className="h-4 w-5/6" />
          <Skeleton className="h-24 w-full rounded-xl" />
        </div>
        <div className="grid gap-4 md:grid-cols-2">
          <EmptyState
            icon={Rocket}
            title="No missions running"
            description="Send a ship to a zone to start mining."
            action={<Button size="sm">Launch mission</Button>}
          />
          <ErrorState message="We couldn't load your orders." onRetry={simulateRetry} isRetrying={isRetrying} />
        </div>
      </div>
    </ContentSection>
  );
}
