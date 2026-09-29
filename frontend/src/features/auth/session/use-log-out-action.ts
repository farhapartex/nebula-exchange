"use client";

import { useCallback, useState } from "react";
import { useRouter } from "next/navigation";

import { loginPathAfterSessionEnd } from "@/features/auth/session/session-end-reasons";
import { useAuth } from "@/features/auth/session/use-auth";

export function useLogOutAction() {
  const router = useRouter();
  const { logOut } = useAuth();
  const [isLoggingOut, setIsLoggingOut] = useState(false);

  const logOutAndLeave = useCallback(async () => {
    setIsLoggingOut(true);
    try {
      await logOut();
      router.replace(loginPathAfterSessionEnd("signed_out"));
    } finally {
      setIsLoggingOut(false);
    }
  }, [logOut, router]);

  return { logOutAndLeave, isLoggingOut };
}
