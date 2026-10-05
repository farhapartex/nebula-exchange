"use client";

import { useEffect, useState, type ReactNode } from "react";
import { usePathname, useRouter } from "next/navigation";

import { Spinner } from "@/components/ui/spinner";
import { decideProtectedRouteAccess, loginPathForProtectedPage } from "@/features/auth/session/protected-route-access";
import { useAuth } from "@/features/auth/session/use-auth";

export function RequireAuthentication({ children }: { children: ReactNode }) {
  const router = useRouter();
  const currentPathname = usePathname();
  const { status } = useAuth();
  const [hasBeenAuthenticated, setHasBeenAuthenticated] = useState(false);
  if (status === "authenticated" && !hasBeenAuthenticated) {
    setHasBeenAuthenticated(true);
  }

  const access = decideProtectedRouteAccess(status, hasBeenAuthenticated);

  useEffect(() => {
    if (access === "redirect_to_login") {
      router.replace(loginPathForProtectedPage(currentPathname + window.location.search));
    }
  }, [access, currentPathname, router]);

  if (access !== "render") {
    return (
      <div className="flex flex-1 items-center justify-center py-24">
        <Spinner className="size-6 text-accent" label="Checking your session" />
      </div>
    );
  }
  return children;
}
