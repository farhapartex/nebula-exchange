"use client";

import { useCallback, useEffect, useMemo, useRef, useState, type ReactNode } from "react";

import type { EstablishedSession, UserProfile } from "@/features/auth/api/auth-types";
import { refreshSession } from "@/features/auth/api/session-api";
import { AuthContext, type AuthStatus } from "@/features/auth/session/auth-context";
import { refreshAccessTokenOnce, registerAccessTokenRefresher, setAccessToken } from "@/lib/api/access-token-store";

const refreshLeadTimeInMilliseconds = 60_000;
const minimumRefreshDelayInMilliseconds = 5_000;

export function AuthProvider({ children }: { children: ReactNode }) {
  const [status, setStatus] = useState<AuthStatus>("restoring");
  const [user, setUser] = useState<UserProfile | null>(null);
  const refreshTimerRef = useRef<number | null>(null);

  const clearRefreshTimer = useCallback(() => {
    if (refreshTimerRef.current !== null) {
      window.clearTimeout(refreshTimerRef.current);
      refreshTimerRef.current = null;
    }
  }, []);

  const endSession = useCallback(() => {
    clearRefreshTimer();
    setAccessToken(null);
    setUser(null);
    setStatus("anonymous");
  }, [clearRefreshTimer]);

  const startSession = useCallback(
    (establishedSession: EstablishedSession) => {
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

  useEffect(() => {
    registerAccessTokenRefresher(async () => {
      try {
        const refreshedSession = await refreshSession();
        startSession(refreshedSession);
        return refreshedSession.access_token;
      } catch {
        endSession();
        return null;
      }
    });
    void refreshAccessTokenOnce();

    return () => {
      registerAccessTokenRefresher(null);
      clearRefreshTimer();
    };
  }, [startSession, endSession, clearRefreshTimer]);

  const authContextValue = useMemo(
    () => ({ status, user, startSession, endSession }),
    [status, user, startSession, endSession],
  );

  return <AuthContext.Provider value={authContextValue}>{children}</AuthContext.Provider>;
}
