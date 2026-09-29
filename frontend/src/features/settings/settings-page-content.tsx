import { Settings } from "lucide-react";

import { PageContainer } from "@/components/layout/page-container";
import { PasswordSection } from "@/features/settings/password-section";
import { ProfileSection } from "@/features/settings/profile-section";
import { SessionsSection } from "@/features/settings/sessions-section";
import { TwoFactorSection } from "@/features/settings/two-factor-section";

export function SettingsPageContent() {
  return (
    <PageContainer className="max-w-3xl py-8 sm:py-10">
      <header className="mb-8 flex items-start gap-4">
        <span className="flex size-12 shrink-0 items-center justify-center rounded-xl border border-border-strong bg-surface-raised text-accent-soft">
          <Settings className="size-6" />
        </span>
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-foreground">Settings</h1>
          <p className="mt-1 text-sm text-muted">Your profile, password, security and active sessions.</p>
        </div>
      </header>
      <div className="space-y-6">
        <ProfileSection />
        <PasswordSection />
        <TwoFactorSection />
        <SessionsSection />
      </div>
    </PageContainer>
  );
}
