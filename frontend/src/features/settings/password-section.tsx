"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useForm, useWatch } from "react-hook-form";
import { z } from "zod";

import { ContentSection } from "@/components/layout/content-section";
import { Button } from "@/components/ui/button";
import { PasswordInput } from "@/components/ui/password-input";
import { useToast } from "@/components/ui/toast/use-toast";
import { PasswordStrengthMeter } from "@/features/auth/shared/password-strength-meter";
import { changePassword } from "@/features/settings/api/settings-api";
import { applyServerFieldErrors } from "@/utils/forms/apply-server-field-errors";
import { maximumPasswordLength, minimumPasswordLength } from "@/utils/security/password-strength";

const passwordChangeSchema = z
  .object({
    currentPassword: z.string().min(1, "Enter your current password"),
    newPassword: z
      .string()
      .min(minimumPasswordLength, `Use at least ${minimumPasswordLength} characters`)
      .max(maximumPasswordLength, `Use at most ${maximumPasswordLength} characters`),
    confirmPassword: z.string(),
  })
  .refine((formValues) => formValues.newPassword === formValues.confirmPassword, {
    path: ["confirmPassword"],
    message: "Passwords don't match",
  });

type PasswordChangeFormValues = z.infer<typeof passwordChangeSchema>;

const serverFieldNames = { current_password: "currentPassword", new_password: "newPassword" } as const;

export function PasswordSection() {
  const { showToast } = useToast();
  const queryClient = useQueryClient();

  const {
    register,
    control,
    handleSubmit,
    setError,
    reset,
    formState: { errors },
  } = useForm<PasswordChangeFormValues>({
    resolver: zodResolver(passwordChangeSchema),
    defaultValues: { currentPassword: "", newPassword: "", confirmPassword: "" },
    mode: "onTouched",
  });
  const typedNewPassword = useWatch({ control, name: "newPassword" });

  const changeMutation = useMutation({
    mutationFn: (formValues: PasswordChangeFormValues) =>
      changePassword(formValues.currentPassword, formValues.newPassword),
    onSuccess: (changeResult) => {
      reset();
      void queryClient.invalidateQueries({ queryKey: ["active-sessions"] });
      showToast({
        tone: "success",
        title: "Password changed",
        description:
          changeResult.signed_out_other_sessions > 0
            ? `${changeResult.signed_out_other_sessions} other session(s) were signed out.`
            : "You're still signed in on this device.",
      });
    },
    onError: (error) => applyServerFieldErrors(error, serverFieldNames, setError),
  });

  return (
    <ContentSection title="Password" description="Changing your password signs you out everywhere except this device.">
      <form onSubmit={handleSubmit((formValues) => changeMutation.mutate(formValues))} noValidate className="space-y-4">
        <PasswordInput
          label="Current password"
          autoComplete="current-password"
          errorMessage={errors.currentPassword?.message}
          {...register("currentPassword")}
        />
        <div className="grid gap-4 sm:grid-cols-2">
          <PasswordInput
            label="New password"
            autoComplete="new-password"
            errorMessage={errors.newPassword?.message}
            belowField={<PasswordStrengthMeter password={typedNewPassword} />}
            {...register("newPassword")}
          />
          <PasswordInput
            label="Confirm new password"
            autoComplete="new-password"
            errorMessage={errors.confirmPassword?.message}
            {...register("confirmPassword")}
          />
        </div>
        <Button type="submit" isLoading={changeMutation.isPending}>
          Change password
        </Button>
      </form>
    </ContentSection>
  );
}
