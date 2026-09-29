"use client";

import { RotateCw, TriangleAlert } from "lucide-react";

import { Button } from "@/components/ui/button";
import { cn } from "@/utils/class-names";

type ErrorStateProps = {
  title?: string;
  message: string;
  onRetry?: () => void;
  isRetrying?: boolean;
  className?: string;
};

export function ErrorState({
  title = "Something went wrong",
  message,
  onRetry,
  isRetrying = false,
  className,
}: ErrorStateProps) {
  return (
    <div
      role="alert"
      className={cn(
        "flex flex-col items-center justify-center rounded-xl border border-down/40 bg-down-soft/20 px-6 py-10 text-center",
        className,
      )}
    >
      <span className="mb-3 flex size-11 items-center justify-center rounded-full bg-down-soft/60 text-down">
        <TriangleAlert className="size-5" />
      </span>
      <p className="text-sm font-medium text-foreground">{title}</p>
      <p className="mt-1 max-w-sm text-sm text-muted">{message}</p>
      {onRetry && (
        <Button variant="secondary" size="sm" className="mt-4" isLoading={isRetrying} onClick={onRetry}>
          {!isRetrying && <RotateCw className="size-3.5" />}
          Try again
        </Button>
      )}
    </div>
  );
}
