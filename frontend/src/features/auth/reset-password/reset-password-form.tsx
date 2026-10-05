"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { TriangleAlert } from "lucide-react";
import { useForm, useWatch } from "react-hook-form";

import { Button } from "@/components/ui/button";
import { PasswordInput } from "@/components/ui/password-input";
import { resetPassword } from "@/features/auth/api/password-reset-api";
import {
  resetPasswordFieldNames,
  resetPasswordSchema,
  type ResetPasswordFormValues,
} from "@/features/auth/reset-password/reset-password-schema";
import { PasswordStrengthMeter } from "@/features/auth/shared/password-strength-meter";
import { isApiError } from "@/lib/api/api-error";
import { applyServerFieldErrors } from "@/utils/forms/apply-server-field-errors";

type ResetPasswordFormProps = {
  resetToken: string;
  emailHint: string;
  onPasswordReset: () => void;
  onLinkExpired: () => void;
};

export function ResetPasswordForm({ resetToken, emailHint, onPasswordReset, onLinkExpired }: ResetPasswordFormProps) {
  const {
    register,
    control,
    handleSubmit,
    setError,
    formState: { errors },
  } = useForm<ResetPasswordFormValues>({
    resolver: zodResolver(resetPasswordSchema),
    defaultValues: { password: "", confirmPassword: "" },
    mode: "onTouched",
  });
  const typedPassword = useWatch({ control, name: "password" });

  const resetMutation = useMutation({
    mutationFn: (formValues: ResetPasswordFormValues) => resetPassword(resetToken, formValues.password),
    meta: { showsThrottlingInline: true },
    onSuccess: onPasswordReset,
    onError: (error) => {
      if (isApiError(error) && error.statusCode === 404) {
        onLinkExpired();
        return;
      }
      applyServerFieldErrors(error, resetPasswordFieldNames, setError);
    },
  });

  const showsFormError = resetMutation.isError && Object.keys(errors).length === 0;

  return (
    <form onSubmit={handleSubmit((formValues) => resetMutation.mutate(formValues))} noValidate className="space-y-5">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight text-foreground">Choose a new password</h1>
        <p className="mt-1 text-sm text-muted">
          For <span className="font-mono text-foreground">{emailHint}</span>. You&apos;ll be logged out on every device.
        </p>
      </div>

      {showsFormError && (
        <div role="alert" className="flex items-start gap-2.5 rounded-lg border border-down/40 bg-down-soft/20 p-3">
          <TriangleAlert className="mt-0.5 size-4 shrink-0 text-down" />
          <p className="text-sm text-foreground">
            {isApiError(resetMutation.error) ? resetMutation.error.message : "Something went wrong. Please try again."}
          </p>
        </div>
      )}

      <PasswordInput
        label="New password"
        autoComplete="new-password"
        placeholder="At least 10 characters"
        errorMessage={errors.password?.message}
        belowField={<PasswordStrengthMeter password={typedPassword} />}
        {...register("password")}
      />
      <PasswordInput
        label="Confirm new password"
        autoComplete="new-password"
        placeholder="Type it again"
        errorMessage={errors.confirmPassword?.message}
        {...register("confirmPassword")}
      />

      <Button type="submit" size="lg" className="w-full" isLoading={resetMutation.isPending}>
        {resetMutation.isPending ? "Updating password" : "Update password"}
      </Button>
    </form>
  );
}
