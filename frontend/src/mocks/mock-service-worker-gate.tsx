"use client";

import { useEffect, useState, type ReactNode } from "react";

import { areApiMocksEnabled } from "@/lib/api/api-config";

async function startMockServiceWorker() {
  const { mockServiceWorker } = await import("@/mocks/browser");
  await mockServiceWorker.start({ onUnhandledRequest: "bypass", quiet: true });
}

export function MockServiceWorkerGate({ children }: { children: ReactNode }) {
  const [isReady, setIsReady] = useState(!areApiMocksEnabled);

  useEffect(() => {
    if (!areApiMocksEnabled) {
      return;
    }
    startMockServiceWorker().finally(() => setIsReady(true));
  }, []);

  if (!isReady) {
    return null;
  }
  return children;
}
