"use client";

import { useCallback, useEffect, useMemo, useRef, useState, type ReactNode } from "react";
import { usePathname, useRouter } from "next/navigation";
import { useQueryClient } from "@tanstack/react-query";

import type { EstablishedSession, UserProfile } from "@/features/auth/api/auth-types";
import { logOut as requestLogOut, refreshSession } from "@/features/auth/api/session-api";
import { AuthContext, type AuthStatus } from "@/features/auth/session/auth-context";
import { loginPathAfterSessionEnd } from "@/features/auth/session/session-end-reasons";
import { refreshAccessTokenOnce, registerAccessTokenRefresher, setAccessToken } from "@/lib/api/access-token-store";

const refreshLeadTimeInMilliseconds = 60_000;
const minimumRefreshDelayInMilliseconds = 5_000;

export function AuthProvider({ children }: { children: ReactNode }) {
  const router = useRouter();
  const currentPathname = usePathname();
  const queryClient = useQueryClient();
  const [status, setStatus] = useState<AuthStatus>("restoring");
  const [user, setUser] = useState<UserProfile | null>(null);
  const refreshTimerRef = useRef<number | null>(null);
  const hasActiveSessionRef = useRef(false);
  const currentPathnameRef = useRef(currentPathname);

  useEffect(() => {
    currentPathnameRef.current = currentPathname;
  }, [currentPathname]);

  const clearRefreshTimer = useCallback(() => {
    if (refreshTimerRef.current !== null) {
      window.clearTimeout(refreshTimerRef.current);
      refreshTimerRef.current = null;
    }
  }, []);

  const endSession = useCallback(() => {
    clearRefreshTimer();
    hasActiveSessionRef.current = false;
    setAccessToken(null);
    setUser(null);
    setStatus("anonymous");
    queryClient.clear();
  }, [clearRefreshTimer, queryClient]);

  const startSession = useCallback(
    (establishedSession: EstablishedSession) => {
      hasActiveSessionRef.current = true;
      setAccessToken(establishedSession.access_token);
      setUser(establishedSession.user);
      setStatus("authenticated");

      clearRefreshTimer();
      const expiresInMilliseconds = new Date(establishedSession.access_token_expires_at).getTime() - Date.now();
      const refreshDelay = Math.max(
        expiresInMilliseconds - refreshLeadTimeInMilliseconds,
        minimumRefreshDelayInMilliseconds,
      );
      refreshTimerRef.current = window.setTimeout(() => {
        void refreshAccessTokenOnce();
      }, refreshDelay);
    },
    [clearRefreshTimer],
  );

  const logOut = useCallback(async () => {
    await requestLogOut().catch(() => undefined);
    endSession();
  }, [endSession]);

  useEffect(() => {
    registerAccessTokenRefresher(async () => {
      try {
        const refreshedSession = await refreshSession();
        startSession(refreshedSession);
        return refreshedSession.access_token;
      } catch {
        const hadActiveSession = hasActiveSessionRef.current;
        endSession();
        if (hadActiveSession) {
          router.replace(loginPathAfterSessionEnd("session_expired", currentPathnameRef.current));
        }
        return null;
      }
    });
    void refreshAccessTokenOnce();

    return () => {
      registerAccessTokenRefresher(null);
      clearRefreshTimer();
    };
  }, [startSession, endSession, clearRefreshTimer, router]);

  const replaceUser = useCallback((updatedUser: UserProfile) => setUser(updatedUser), []);

  const authContextValue = useMemo(
    () => ({ status, user, startSession, endSession, logOut, replaceUser }),
    [status, user, startSession, endSession, logOut, replaceUser],
  );

  return <AuthContext.Provider value={authContextValue}>{children}</AuthContext.Provider>;
}
