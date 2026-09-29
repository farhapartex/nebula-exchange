"use client";

import { useEffect, useState, type ReactNode } from "react";

import { areApiMocksEnabled } from "@/lib/api/api-config";

let mockServiceWorkerStartup: Promise<void> | null = null;

function startMockServiceWorkerOnce(): Promise<void> {
  mockServiceWorkerStartup ??= import("@/mocks/browser").then(({ mockServiceWorker }) =>
    mockServiceWorker.start({ onUnhandledRequest: "bypass", quiet: true }).then(() => undefined),
  );
  return mockServiceWorkerStartup;
}

export function MockServiceWorkerGate({ children }: { children: ReactNode }) {
  const [isReady, setIsReady] = useState(!areApiMocksEnabled);

  useEffect(() => {
    if (!areApiMocksEnabled) {
      return;
    }
    let isMounted = true;
    startMockServiceWorkerOnce().finally(() => {
      if (isMounted) {
        setIsReady(true);
      }
    });
    return () => {
      isMounted = false;
    };
  }, []);

  if (!isReady) {
    return null;
  }
  return children;
}
