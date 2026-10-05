"use client";

import { useEffect, useRef } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useMutation, useQuery } from "@tanstack/react-query";

import { Button } from "@/components/ui/button";
import { ErrorState } from "@/components/ui/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { AccountSummary } from "@/features/auth/activation/account-summary";
import { ActivationStatusHeader, type ActivationStage } from "@/features/auth/activation/activation-status-header";
import { activateAccount, fetchActivationPreview } from "@/features/auth/api/activation-api";
import { isApiError } from "@/lib/api/api-error";
import { useRedirectCountdown } from "@/utils/navigation/use-redirect-countdown";

const redirectDelayInSeconds = 3;
const minimumLoaderDurationInMilliseconds = 900;

function isInvalidTokenError(error: unknown): boolean {
  return isApiError(error) && (error.statusCode === 404 || error.statusCode === 422);
}

async function activateWithVisibleLoader(activationToken: string) {
  const [activatedAccount] = await Promise.all([
    activateAccount(activationToken),
    new Promise((resolve) => window.setTimeout(resolve, minimumLoaderDurationInMilliseconds)),
  ]);
  return activatedAccount;
}

export function ActivationFlow() {
  const router = useRouter();
  const activationToken = useSearchParams().get("token")?.trim() ?? "";
  const hasStartedActivation = useRef(false);

  const previewQuery = useQuery({
    queryKey: ["activation-preview", activationToken],
    queryFn: ({ signal }) => fetchActivationPreview(activationToken, signal),
    enabled: activationToken !== "",
    retry: false,
    staleTime: Infinity,
  });

  const activationMutation = useMutation({ mutationFn: activateWithVisibleLoader });

  const isAlreadyActivated = previewQuery.data?.is_activated === true;
  const hasInvalidToken =
    activationToken === "" || isInvalidTokenError(previewQuery.error) || isInvalidTokenError(activationMutation.error);

  useEffect(() => {
    if (hasInvalidToken) {
      router.replace("/");
    }
  }, [hasInvalidToken, router]);

  useEffect(() => {
    if (!previewQuery.data || isAlreadyActivated || hasStartedActivation.current) {
      return;
    }
    hasStartedActivation.current = true;
    activationMutation.mutate(activationToken);
  }, [activationMutation, activationToken, isAlreadyActivated, previewQuery.data]);

  const stage: ActivationStage = isAlreadyActivated
    ? "already_activated"
    : activationMutation.isSuccess
      ? "activated"
      : "activating";
  const isFinished = stage !== "activating";
  const remainingSeconds = useRedirectCountdown("/login", redirectDelayInSeconds, isFinished);

  if (hasInvalidToken) {
    return <ActivationSkeleton />;
  }

  const unexpectedError = previewQuery.error ?? activationMutation.error;
  if (unexpectedError) {
    return (
      <ErrorState
        title="We couldn't activate your account"
        message={isApiError(unexpectedError) ? unexpectedError.message : "Please try again in a moment."}
        onRetry={() => {
          hasStartedActivation.current = false;
          activationMutation.reset();
          void previewQuery.refetch();
        }}
        isRetrying={previewQuery.isFetching}
      />
    );
  }

  if (!previewQuery.data) {
    return <ActivationSkeleton />;
  }

  return (
    <div className="space-y-6">
      <ActivationStatusHeader stage={stage} />
      <AccountSummary activationPreview={previewQuery.data} />
      {isFinished && (
        <div className="space-y-3 text-center">
          <p className="text-sm text-muted">
            Taking you to login in <span className="font-mono text-foreground tabular-nums">{remainingSeconds}</span>…
          </p>
          <Button asChild variant="secondary" size="sm">
            <Link href="/login">Go to login now</Link>
          </Button>
        </div>
      )}
    </div>
  );
}

function ActivationSkeleton() {
  return (
    <div className="space-y-6" aria-busy="true">
      <div className="flex flex-col items-center gap-3">
        <Skeleton className="size-12 rounded-full" />
        <Skeleton className="h-6 w-48" />
        <Skeleton className="h-4 w-64" />
      </div>
      <Skeleton className="h-40 w-full rounded-xl" />
    </div>
  );
}
