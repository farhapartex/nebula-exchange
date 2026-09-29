"use client";

import { useState } from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { CircleCheck, LinkIcon } from "lucide-react";

import { Button } from "@/components/ui/button";
import { ErrorState } from "@/components/ui/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { fetchPasswordResetPreview } from "@/features/auth/api/password-reset-api";
import { ResetPasswordForm } from "@/features/auth/reset-password/reset-password-form";
import { loginPathAfterSessionEnd } from "@/features/auth/session/session-end-reasons";
import { useAuth } from "@/features/auth/session/use-auth";
import { isApiError } from "@/lib/api/api-error";
import { useRedirectCountdown } from "@/utils/navigation/use-redirect-countdown";

const redirectDelayInSeconds = 3;

export function ResetPasswordFlow() {
  const resetToken = useSearchParams().get("token")?.trim() ?? "";
  const { endSession } = useAuth();
  const [hasResetPassword, setHasResetPassword] = useState(false);
  const [hasLinkExpiredDuringReset, setHasLinkExpiredDuringReset] = useState(false);

  const previewQuery = useQuery({
    queryKey: ["password-reset-preview", resetToken],
    queryFn: ({ signal }) => fetchPasswordResetPreview(resetToken, signal),
    enabled: resetToken !== "",
    retry: false,
    staleTime: Infinity,
  });

  const remainingSeconds = useRedirectCountdown(
    loginPathAfterSessionEnd("password_reset"),
    redirectDelayInSeconds,
    hasResetPassword,
  );

  const isInvalidLink =
    resetToken === "" ||
    hasLinkExpiredDuringReset ||
    (isApiError(previewQuery.error) && previewQuery.error.statusCode === 404);

  if (hasResetPassword) {
    return <PasswordResetSuccess remainingSeconds={remainingSeconds} />;
  }
  if (isInvalidLink) {
    return <InvalidResetLink />;
  }
  if (previewQuery.isError) {
    return (
      <ErrorState
        message={isApiError(previewQuery.error) ? previewQuery.error.message : "Please try again in a moment."}
        onRetry={() => void previewQuery.refetch()}
        isRetrying={previewQuery.isFetching}
      />
    );
  }
  if (!previewQuery.data) {
    return (
      <div className="space-y-5" aria-busy="true">
        <Skeleton className="h-8 w-2/3" />
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-10 w-full" />
        <Skeleton className="h-10 w-full" />
        <Skeleton className="h-12 w-full" />
      </div>
    );
  }

  return (
    <ResetPasswordForm
      resetToken={resetToken}
      emailHint={previewQuery.data.email_hint}
      onPasswordReset={() => {
        endSession();
        setHasResetPassword(true);
      }}
      onLinkExpired={() => setHasLinkExpiredDuringReset(true)}
    />
  );
}

function PasswordResetSuccess({ remainingSeconds }: { remainingSeconds: number }) {
  return (
    <div className="space-y-4 text-center" aria-live="polite">
      <span className="mx-auto flex size-12 items-center justify-center rounded-full bg-up/10 text-up">
        <CircleCheck className="size-6" />
      </span>
      <h1 className="text-xl font-semibold text-foreground">Password updated</h1>
      <p className="text-sm text-muted">You were logged out on every device. Log in with your new password.</p>
      <p className="text-sm text-muted">
        Taking you to login in <span className="font-mono text-foreground tabular-nums">{remainingSeconds}</span>…
      </p>
      <Button asChild variant="secondary" size="sm">
        <Link href={loginPathAfterSessionEnd("password_reset")}>Go to login now</Link>
      </Button>
    </div>
  );
}

function InvalidResetLink() {
  return (
    <div className="space-y-4 text-center">
      <span className="mx-auto flex size-12 items-center justify-center rounded-full bg-warning/10 text-warning">
        <LinkIcon className="size-6" />
      </span>
      <h1 className="text-xl font-semibold text-foreground">This reset link doesn&apos;t work</h1>
      <p className="text-sm text-muted">
        It may have expired, already been used, or been replaced by a newer link. Reset links work once and expire after
        30 minutes.
      </p>
      <Button asChild>
        <Link href="/forgot">Request a new link</Link>
      </Button>
    </div>
  );
}
