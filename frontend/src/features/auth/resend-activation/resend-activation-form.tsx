"use client";

import { useState } from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { MailCheck, TriangleAlert } from "lucide-react";
import { useForm } from "react-hook-form";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { requestActivationEmail } from "@/features/auth/api/activation-api";
import {
  resendActivationFieldNames,
  resendActivationSchema,
  type ResendActivationFormValues,
} from "@/features/auth/resend-activation/resend-activation-schema";
import { isApiError } from "@/lib/api/api-error";
import { applyServerFieldErrors } from "@/utils/forms/apply-server-field-errors";
import { useCooldown } from "@/utils/time/use-cooldown";

const resendCooldownInSeconds = 60;

export function ResendActivationForm() {
  const prefilledEmail = useSearchParams().get("email") ?? "";
  const [lastRequestedEmail, setLastRequestedEmail] = useState<string | null>(null);
  const { remainingSeconds, isCoolingDown, startCooldown } = useCooldown(resendCooldownInSeconds);

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors },
  } = useForm<ResendActivationFormValues>({
    resolver: zodResolver(resendActivationSchema),
    defaultValues: { email: prefilledEmail },
    mode: "onTouched",
  });

  const resendMutation = useMutation({
    mutationFn: (formValues: ResendActivationFormValues) => requestActivationEmail(formValues.email.trim()),
    onSuccess: (_result, formValues) => {
      setLastRequestedEmail(formValues.email.trim());
      startCooldown();
    },
    onError: (error) => {
      applyServerFieldErrors(error, resendActivationFieldNames, setError);
      if (isApiError(error) && error.code === "RATE_LIMITED") {
        startCooldown();
      }
    },
  });

  const showsFormError = resendMutation.isError && !errors.email;

  return (
    <form onSubmit={handleSubmit((formValues) => resendMutation.mutate(formValues))} noValidate className="space-y-5">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight text-foreground">Send a new activation link</h1>
        <p className="mt-1 text-sm text-muted">
          Enter the email you signed up with. Any older activation links stop working.
        </p>
      </div>

      {lastRequestedEmail && (
        <div role="status" className="flex items-start gap-2.5 rounded-lg border border-up/40 bg-up/10 p-3">
          <MailCheck className="mt-0.5 size-4 shrink-0 text-up" />
          <p className="text-sm text-muted">
            If <span className="font-medium text-foreground">{lastRequestedEmail}</span> belongs to an account that
            isn&apos;t activated yet, a new link is on its way. It stays valid for 24 hours.
          </p>
        </div>
      )}

      {showsFormError && (
        <div role="alert" className="flex items-start gap-2.5 rounded-lg border border-down/40 bg-down-soft/20 p-3">
          <TriangleAlert className="mt-0.5 size-4 shrink-0 text-down" />
          <p className="text-sm text-foreground">
            {isApiError(resendMutation.error)
              ? resendMutation.error.message
              : "Something went wrong. Please try again."}
          </p>
        </div>
      )}

      <Input
        label="Email"
        type="email"
        autoComplete="email"
        placeholder="pilot@nebula.test"
        errorMessage={errors.email?.message}
        {...register("email")}
      />

      <Button type="submit" size="lg" className="w-full" isLoading={resendMutation.isPending} disabled={isCoolingDown}>
        {isCoolingDown ? (
          <span>
            Send again in <span className="font-mono tabular-nums">{remainingSeconds}s</span>
          </span>
        ) : lastRequestedEmail ? (
          "Send another link"
        ) : (
          "Send activation link"
        )}
      </Button>

      <p className="text-center text-sm text-muted">
        Already activated?{" "}
        <Link href="/login" className="font-medium text-accent-soft underline-offset-2 hover:underline">
          Log in
        </Link>
      </p>
    </form>
  );
}
