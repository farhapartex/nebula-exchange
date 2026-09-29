"use client";

import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Modal } from "@/components/ui/modal";
import { OneTimeCodeInput } from "@/components/ui/one-time-code-input";
import { isApiError } from "@/lib/api/api-error";

type TwoFactorPromptProps = {
  isOpen: boolean;
  onOpenChange: (isOpen: boolean) => void;
  title: string;
  description: string;
  confirmLabel: string;
  tone?: "default" | "danger";
  onSubmitCode: (code: string) => Promise<unknown>;
  onSuccess?: () => void;
};

export function TwoFactorPrompt({
  isOpen,
  onOpenChange,
  title,
  description,
  confirmLabel,
  tone = "default",
  onSubmitCode,
  onSuccess,
}: TwoFactorPromptProps) {
  const [code, setCode] = useState("");
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  function changeOpenState(nextIsOpen: boolean) {
    if (isSubmitting) {
      return;
    }
    setCode("");
    setErrorMessage(null);
    onOpenChange(nextIsOpen);
  }

  async function submitCode() {
    setIsSubmitting(true);
    setErrorMessage(null);
    try {
      await onSubmitCode(code);
      setCode("");
      onOpenChange(false);
      onSuccess?.();
    } catch (error) {
      const fieldMessage = isApiError(error) ? error.fieldErrors.code : undefined;
      setErrorMessage(
        fieldMessage ? `Code ${fieldMessage}` : isApiError(error) ? error.message : "Something went wrong.",
      );
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <Modal
      isOpen={isOpen}
      onOpenChange={changeOpenState}
      title={title}
      description={description}
      footer={
        <>
          <Button variant="secondary" onClick={() => changeOpenState(false)} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button
            variant={tone === "danger" ? "danger" : "primary"}
            isLoading={isSubmitting}
            disabled={code.length !== 6}
            onClick={() => void submitCode()}
          >
            {confirmLabel}
          </Button>
        </>
      }
    >
      <OneTimeCodeInput
        label="Authenticator code"
        value={code}
        onValueChange={setCode}
        errorMessage={errorMessage ?? undefined}
        autoFocus
      />
    </Modal>
  );
}
