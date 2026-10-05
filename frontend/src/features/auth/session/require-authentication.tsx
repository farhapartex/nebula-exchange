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
import { useCurrentUser } from "@/features/auth/session/use-current-user";

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
  const { status, endSession, replaceUser } = useAuth();
  const [hasBeenAuthenticated, setHasBeenAuthenticated] = useState(false);
  if (status === "authenticated" && !hasBeenAuthenticated) {
    setHasBeenAuthenticated(true);
  }

  const access = decideProtectedRouteAccess(status, hasBeenAuthenticated);
  const currentUserQuery = useCurrentUser(access === "render");
  const currentUserView = decideCurrentUserView(currentUserQuery.status, currentUserQuery.error);
  const currentUser = currentUserQuery.data;

  useEffect(() => {
    if (access === "redirect_to_login") {
      router.replace(loginPathForProtectedPage(currentPathname + window.location.search));
    }
  }, [access, currentPathname, router]);

  useEffect(() => {
    if (currentUser) {
      replaceUser(currentUser);
    }
  }, [currentUser, replaceUser]);

  useEffect(() => {
    if (access === "render" && currentUserView === "session_rejected") {
      endSession();
      router.replace(loginPathAfterSessionEnd("session_expired", currentPathname + window.location.search));
    }
  }, [access, currentPathname, currentUserView, endSession, router]);

  if (access !== "render" || currentUserView === "loading" || currentUserView === "session_rejected") {
    return <SessionCheckSpinner />;
  }
  if (currentUserView === "failed") {
    return (
      <div className="mx-auto w-full max-w-md px-4 py-24">
        <ErrorState
          title="We could not load your profile"
          message="Check your connection and try again."
          onRetry={() => void currentUserQuery.refetch()}
          isRetrying={currentUserQuery.isFetching}
        />
      </div>
    );
  }
  return children;
}
