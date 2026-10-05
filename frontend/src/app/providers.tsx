"use client";

import { useState, type ReactNode } from "react";
import { QueryClientProvider } from "@tanstack/react-query";

import { ToastProvider } from "@/components/ui/toast/toast-provider";
import { TooltipProvider } from "@/components/ui/tooltip";
import { AuthProvider } from "@/features/auth/session/auth-provider";
import { createQueryClient } from "@/lib/query/query-client";
import { MockServiceWorkerGate } from "@/mocks/mock-service-worker-gate";

export function Providers({ children }: { children: ReactNode }) {
  const [queryClient] = useState(createQueryClient);

  return (
    <MockServiceWorkerGate>
      <QueryClientProvider client={queryClient}>
        <TooltipProvider delayDuration={200}>
          <AuthProvider>
            <ToastProvider>{children}</ToastProvider>
          </AuthProvider>
        </TooltipProvider>
      </QueryClientProvider>
    </MockServiceWorkerGate>
  );
}
