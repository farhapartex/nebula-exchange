"use client";

import { useState } from "react";

import { ContentSection } from "@/components/layout/content-section";
import { Button } from "@/components/ui/button";
import { StatusBadge } from "@/components/ui/status-badge";
import { useAuth } from "@/features/auth/session/use-auth";
import { refreshAccessTokenOnce } from "@/lib/api/access-token-store";
import { requestData } from "@/lib/api/api-client";
import { areApiMocksEnabled } from "@/lib/api/api-config";

export function SessionShowcase() {
  const { status, user } = useAuth();
  const [isExpiring, setIsExpiring] = useState(false);

  async function simulateSessionExpiry() {
    setIsExpiring(true);
    await requestData("/dev/mock-session/expire", { method: "POST", skipSessionRefresh: true });
    await refreshAccessTokenOnce();
    setIsExpiring(false);
  }

  return (
    <ContentSection
      title="Session"
      description="Mock only: log in as pilot@nebula.test, then expire the session to see the redirect."
    >
      <div className="flex flex-wrap items-center gap-3 text-sm">
        <StatusBadge label={status} tone={status === "authenticated" ? "success" : "neutral"} />
        {user && <span className="font-mono text-muted">{user.username}</span>}
        <Button
          variant="secondary"
          size="sm"
          isLoading={isExpiring}
          disabled={!areApiMocksEnabled || status !== "authenticated"}
          onClick={() => void simulateSessionExpiry()}
        >
          Simulate session expiry
        </Button>
      </div>
    </ContentSection>
  );
}
