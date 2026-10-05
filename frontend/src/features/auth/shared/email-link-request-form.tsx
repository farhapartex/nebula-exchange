"use client";

import { useState, type ReactNode } from "react";
import { useSearchParams } from "next/navigation";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { MailCheck, TriangleAlert } from "lucide-react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { isApiError } from "@/lib/api/api-error";
import { applyServerFieldErrors } from "@/utils/forms/apply-server-field-errors";
import { useCooldown } from "@/utils/time/use-cooldown";

const resendCooldownInSeconds = 60;

const emailLinkRequestSchema = z.object({
  email: z.email("Enter a valid email address"),
});

type EmailLinkRequestValues = z.infer<typeof emailLinkRequestSchema>;

const emailLinkRequestFieldNames = ["email"] as const;

type EmailLinkRequestFormProps = {
  title: string;
  description: string;
  submitLabel: string;
  resubmitLabel: string;
  describeSentLink: (requestedEmail: string) => ReactNode;
  requestLink: (emailAddress: string) => Promise<unknown>;
  footer: ReactNode;
};

export function EmailLinkRequestForm({
  title,
  description,
  submitLabel,
  resubmitLabel,
  describeSentLink,
  requestLink,
  footer,
}: EmailLinkRequestFormProps) {
  const prefilledEmail = useSearchParams().get("email") ?? "";
  const [lastRequestedEmail, setLastRequestedEmail] = useState<string | null>(null);
  const { remainingSeconds, isCoolingDown, startCooldown } = useCooldown(resendCooldownInSeconds);

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors },
  } = useForm<EmailLinkRequestValues>({
    resolver: zodResolver(emailLinkRequestSchema),
    defaultValues: { email: prefilledEmail },
    mode: "onTouched",
  });

  const linkRequestMutation = useMutation({
    mutationFn: (formValues: EmailLinkRequestValues) => requestLink(formValues.email.trim()),
    meta: { showsThrottlingInline: true },
    onSuccess: (_result, formValues) => {
      setLastRequestedEmail(formValues.email.trim());
      startCooldown();
    },
    onError: (error) => {
      applyServerFieldErrors(error, emailLinkRequestFieldNames, setError);
      if (isApiError(error) && error.isThrottled) {
        startCooldown(error.retryAfterSeconds ?? undefined);
      }
    },
  });

  const showsFormError = linkRequestMutation.isError && !errors.email;

  return (
    <form
      onSubmit={handleSubmit((formValues) => linkRequestMutation.mutate(formValues))}
      noValidate
      className="space-y-5"
    >
      <div>
        <h1 className="text-2xl font-semibold tracking-tight text-foreground">{title}</h1>
        <p className="mt-1 text-sm text-muted">{description}</p>
      </div>

      {lastRequestedEmail && (
        <div role="status" className="flex items-start gap-2.5 rounded-lg border border-up/40 bg-up/10 p-3">
          <MailCheck className="mt-0.5 size-4 shrink-0 text-up" />
          <p className="text-sm text-muted">{describeSentLink(lastRequestedEmail)}</p>
        </div>
      )}

      {showsFormError && (
        <div role="alert" className="flex items-start gap-2.5 rounded-lg border border-down/40 bg-down-soft/20 p-3">
          <TriangleAlert className="mt-0.5 size-4 shrink-0 text-down" />
          <p className="text-sm text-foreground">
            {isApiError(linkRequestMutation.error)
              ? linkRequestMutation.error.message
              : "Something went wrong. Please try again."}
          </p>
        </div>
      )}

      <Input
        label="Email"
        type="email"
        autoComplete="email"
        placeholder="you@example.com"
        errorMessage={errors.email?.message}
        {...register("email")}
      />

      <Button
        type="submit"
        size="lg"
        className="w-full"
        isLoading={linkRequestMutation.isPending}
        disabled={isCoolingDown}
      >
        {isCoolingDown ? (
          <span>
            Send again in <span className="font-mono tabular-nums">{remainingSeconds}s</span>
          </span>
        ) : lastRequestedEmail ? (
          resubmitLabel
        ) : (
          submitLabel
        )}
      </Button>

      {footer}
    </form>
  );
}
