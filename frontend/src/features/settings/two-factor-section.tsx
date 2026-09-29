"use client";

import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { Copy, ShieldCheck, ShieldOff } from "lucide-react";
import { QRCodeSVG } from "qrcode.react";

import { ContentSection } from "@/components/layout/content-section";
import { Button } from "@/components/ui/button";
import { OneTimeCodeInput } from "@/components/ui/one-time-code-input";
import { StatusBadge } from "@/components/ui/status-badge";
import { useToast } from "@/components/ui/toast/use-toast";
import { TwoFactorPrompt } from "@/components/ui/two-factor-prompt";
import {
  beginTwoFactorSetup,
  disableTwoFactor,
  enableTwoFactor,
  type TwoFactorSetupDetails,
} from "@/features/auth/api/two-factor-api";
import { useAuth } from "@/features/auth/session/use-auth";
import { isApiError } from "@/lib/api/api-error";

export function TwoFactorSection() {
  const { user, replaceUser } = useAuth();
  const { showToast } = useToast();
  const [setupDetails, setSetupDetails] = useState<TwoFactorSetupDetails | null>(null);
  const [confirmationCode, setConfirmationCode] = useState("");
  const [isDisablePromptOpen, setIsDisablePromptOpen] = useState(false);

  const setupMutation = useMutation({ mutationFn: beginTwoFactorSetup, onSuccess: setSetupDetails });
  const enableMutation = useMutation({
    mutationFn: () => enableTwoFactor(confirmationCode),
    onSuccess: () => {
      if (user) {
        replaceUser({ ...user, two_factor_enabled: true });
      }
      setSetupDetails(null);
      setConfirmationCode("");
      showToast({ tone: "success", title: "Two-factor authentication is on" });
    },
  });

  if (!user) {
    return null;
  }

  const enableErrorMessage = isApiError(enableMutation.error)
    ? enableMutation.error.fieldErrors.code
      ? `Code ${enableMutation.error.fieldErrors.code}`
      : enableMutation.error.message
    : undefined;

  async function copySecret(secret: string) {
    await navigator.clipboard.writeText(secret);
    showToast({ tone: "info", title: "Secret copied" });
  }

  return (
    <ContentSection
      title="Two-factor authentication"
      description="Protect logins and withdrawals with a code from an authenticator app such as Google Authenticator."
    >
      <div className="mb-5 flex flex-wrap items-center justify-between gap-3">
        <StatusBadge
          label={user.two_factor_enabled ? "On" : "Off"}
          tone={user.two_factor_enabled ? "success" : "neutral"}
        />
        {user.two_factor_enabled ? (
          <Button variant="secondary" size="sm" onClick={() => setIsDisablePromptOpen(true)}>
            <ShieldOff className="size-4" />
            Turn off
          </Button>
        ) : (
          !setupDetails && (
            <Button size="sm" isLoading={setupMutation.isPending} onClick={() => setupMutation.mutate()}>
              <ShieldCheck className="size-4" />
              Set up
            </Button>
          )
        )}
      </div>

      {user.two_factor_enabled && (
        <p className="text-sm text-muted">
          You&apos;ll be asked for a code when you log in and before sensitive actions like withdrawals.
        </p>
      )}

      {!user.two_factor_enabled && setupDetails && (
        <div className="grid gap-6 sm:grid-cols-[auto_1fr]">
          <div className="justify-self-center rounded-xl bg-white p-3">
            <QRCodeSVG value={setupDetails.otpauth_url} size={168} level="M" />
          </div>
          <div className="space-y-4">
            <ol className="list-decimal space-y-1 pl-5 text-sm text-muted">
              <li>Scan the QR code with your authenticator app.</li>
              <li>Or enter this secret by hand:</li>
            </ol>
            <div className="flex items-center gap-2 rounded-lg border border-border bg-background/60 px-3 py-2">
              <code className="min-w-0 flex-1 truncate font-mono text-sm text-foreground">{setupDetails.secret}</code>
              <Button
                variant="ghost"
                size="sm"
                aria-label="Copy secret"
                onClick={() => void copySecret(setupDetails.secret)}
              >
                <Copy className="size-4" />
              </Button>
            </div>
            <OneTimeCodeInput
              label="Enter the 6-digit code to confirm"
              value={confirmationCode}
              onValueChange={setConfirmationCode}
              errorMessage={enableErrorMessage}
            />
            <div className="flex gap-3">
              <Button
                isLoading={enableMutation.isPending}
                disabled={confirmationCode.length !== 6}
                onClick={() => enableMutation.mutate()}
              >
                Turn on
              </Button>
              <Button variant="ghost" onClick={() => setSetupDetails(null)}>
                Cancel
              </Button>
            </div>
          </div>
        </div>
      )}

      <TwoFactorPrompt
        isOpen={isDisablePromptOpen}
        onOpenChange={setIsDisablePromptOpen}
        title="Turn off two-factor authentication?"
        description="Enter a current code from your authenticator app to confirm."
        confirmLabel="Turn off"
        tone="danger"
        onSubmitCode={disableTwoFactor}
        onSuccess={() => {
          replaceUser({ ...user, two_factor_enabled: false });
          showToast({ tone: "warning", title: "Two-factor authentication is off" });
        }}
      />
    </ContentSection>
  );
}
