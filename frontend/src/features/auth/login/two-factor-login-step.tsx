"use client";

import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { ArrowLeft, ShieldCheck } from "lucide-react";

import { Button } from "@/components/ui/button";
import { OneTimeCodeInput } from "@/components/ui/one-time-code-input";
import type { EstablishedSession, TwoFactorChallenge } from "@/features/auth/api/auth-types";
import { completeTwoFactorLogin } from "@/features/auth/api/session-api";
import { isApiError } from "@/lib/api/api-error";

type TwoFactorLoginStepProps = {
  challenge: TwoFactorChallenge;
  onSessionEstablished: (establishedSession: EstablishedSession) => void;
  onRestart: (reasonMessage?: string) => void;
};

export function TwoFactorLoginStep({ challenge, onSessionEstablished, onRestart }: TwoFactorLoginStepProps) {
  const [code, setCode] = useState("");

  const twoFactorMutation = useMutation({
    mutationFn: () => completeTwoFactorLogin(challenge.challenge_token, code),
    meta: { showsThrottlingInline: true },
    onSuccess: onSessionEstablished,
    onError: (error) => {
      if (isApiError(error) && (error.statusCode === 401 || error.code === "LOGIN_LOCKED")) {
        onRestart(error.message);
      }
      setCode("");
    },
  });

  const errorMessage =
    isApiError(twoFactorMutation.error) && twoFactorMutation.error.fieldErrors.code
      ? `Code ${twoFactorMutation.error.fieldErrors.code}`
      : undefined;

  return (
    <form
      onSubmit={(submitEvent) => {
        submitEvent.preventDefault();
        twoFactorMutation.mutate();
      }}
      className="space-y-5"
    >
      <div className="text-center">
        <span className="mx-auto mb-4 flex size-12 items-center justify-center rounded-full bg-accent/15 text-accent-soft">
          <ShieldCheck className="size-6" />
        </span>
        <h1 className="text-xl font-semibold text-foreground">Two-factor check</h1>
        <p className="mt-1 text-sm text-muted">Enter the 6-digit code from your authenticator app.</p>
      </div>

      <OneTimeCodeInput
        label="Authenticator code"
        value={code}
        onValueChange={setCode}
        errorMessage={errorMessage}
        autoFocus
      />

      <Button
        type="submit"
        size="lg"
        className="w-full"
        isLoading={twoFactorMutation.isPending}
        disabled={code.length !== 6}
      >
        Verify and log in
      </Button>

      <button
        type="button"
        onClick={() => onRestart()}
        className="mx-auto flex items-center gap-1.5 text-sm text-muted transition-colors hover:text-foreground"
      >
        <ArrowLeft className="size-4" />
        Use a different account
      </button>
    </form>
  );
}
