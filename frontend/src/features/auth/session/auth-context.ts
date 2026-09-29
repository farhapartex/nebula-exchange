"use client";

import { createContext } from "react";

import type { EstablishedSession, UserProfile } from "@/features/auth/api/auth-types";

export type AuthStatus = "restoring" | "authenticated" | "anonymous";

export type AuthContextValue = {
  status: AuthStatus;
  user: UserProfile | null;
  startSession: (establishedSession: EstablishedSession) => void;
  endSession: () => void;
  logOut: () => Promise<void>;
};

export const AuthContext = createContext<AuthContextValue | null>(null);
