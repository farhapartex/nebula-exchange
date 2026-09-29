"use client";

import { placeholderPilot } from "@/components/app-shell/placeholder-pilot";
import { useAuth } from "@/features/auth/session/use-auth";

export function useDisplayedPilot() {
  const { user } = useAuth();
  return {
    username: user?.username ?? placeholderPilot.username,
    email: user?.email ?? placeholderPilot.email,
    isSignedIn: user !== null,
  };
}
