"use client";

import { useContext } from "react";

import { AuthContext } from "@/features/auth/session/auth-context";

export function useAuth() {
  const authContext = useContext(AuthContext);
  if (!authContext) {
    throw new Error("useAuth must be used inside AuthProvider");
  }
  return authContext;
}
