"use client";

import { useEffect, useState, type ReactNode } from "react";
import { usePathname, useRouter } from "next/navigation";

import { ErrorState } from "@/components/ui/error-state";
import { Spinner } from "@/components/ui/spinner";
import {
  decideCurrentUserView,
  decideProtectedRouteAccess,
  loginPathForProtectedPage,
} from "@/features/auth/session/protected-route-access";
import { loginPathAfterSessionEnd } from "@/features/auth/session/session-end-reasons";
import { useAuth } from "@/features/auth/session/use-auth";
import { useVerifiedCurrentPlayer } from "@/features/auth/session/use-current-player";

function SessionCheckSpinner() {
  return (
    <div className="flex flex-1 items-center justify-center py-24">
      <Spinner className="size-6 text-accent" label="Checking your session" />
    </div>
  );
}

export function RequireAuthentication({ children }: { children: ReactNode }) {
  const router = useRouter();
  const currentPathname = usePathname();
  const { status, endSession } = useAuth();
  const [hasBeenAuthenticated, setHasBeenAuthenticated] = useState(false);
  if (status === "authenticated" && !hasBeenAuthenticated) {
    setHasBeenAuthenticated(true);
  }

  const access = decideProtectedRouteAccess(status, hasBeenAuthenticated);
  const currentPlayerQuery = useVerifiedCurrentPlayer(access === "render");
  const currentPlayerView = decideCurrentUserView(currentPlayerQuery.status, currentPlayerQuery.error);

  useEffect(() => {
    if (access === "redirect_to_login") {
      router.replace(loginPathForProtectedPage(currentPathname + window.location.search));
    }
  }, [access, currentPathname, router]);

  useEffect(() => {
    if (access === "render" && currentPlayerView === "session_rejected") {
      endSession();
      router.replace(loginPathAfterSessionEnd("session_expired", currentPathname + window.location.search));
    }
  }, [access, currentPathname, currentPlayerView, endSession, router]);

  if (access !== "render" || currentPlayerView === "loading" || currentPlayerView === "session_rejected") {
    return <SessionCheckSpinner />;
  }
  if (currentPlayerView === "failed") {
    return (
      <div className="mx-auto w-full max-w-md px-4 py-24">
        <ErrorState
          title="We could not load your profile"
          message="Check your connection and try again."
          onRetry={() => void currentPlayerQuery.refetch()}
          isRetrying={currentPlayerQuery.isFetching}
        />
      </div>
    );
  }
  return children;
}
