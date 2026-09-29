"use client";

import { useState } from "react";
import Link from "next/link";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { TriangleAlert } from "lucide-react";
import { Controller, useForm, useWatch } from "react-hook-form";

import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { PasswordInput } from "@/components/ui/password-input";
import type { SignedUpAccount } from "@/features/auth/api/auth-types";
import { signUp } from "@/features/auth/api/sign-up";
import { PasswordStrengthMeter } from "@/features/auth/signup/password-strength-meter";
import { signupFieldNames, signupSchema, type SignupFormValues } from "@/features/auth/signup/signup-schema";
import { SignupSuccess } from "@/features/auth/signup/signup-success";
import { isApiError } from "@/lib/api/api-error";
import { applyServerFieldErrors } from "@/utils/forms/apply-server-field-errors";

export function SignupForm() {
  const [signedUpAccount, setSignedUpAccount] = useState<SignedUpAccount | null>(null);
  const [formErrorMessage, setFormErrorMessage] = useState<string | null>(null);

  const {
    register,
    control,
    handleSubmit,
    setError,
    formState: { errors },
  } = useForm<SignupFormValues>({
    resolver: zodResolver(signupSchema),
    defaultValues: { email: "", username: "", password: "", acceptsTerms: false },
    mode: "onTouched",
  });

  const typedPassword = useWatch({ control, name: "password" });

  const signupMutation = useMutation({
    mutationFn: (formValues: SignupFormValues) =>
      signUp({
        email: formValues.email.trim(),
        username: formValues.username.trim(),
        password: formValues.password,
        accepts_terms: formValues.acceptsTerms,
      }),
    onSuccess: setSignedUpAccount,
    onError: (error) => {
      if (applyServerFieldErrors(error, signupFieldNames, setError)) {
        return;
      }
      setFormErrorMessage(isApiError(error) ? error.message : "Something went wrong. Please try again.");
    },
  });

  function submitSignup(formValues: SignupFormValues) {
    setFormErrorMessage(null);
    signupMutation.mutate(formValues);
  }

  if (signedUpAccount) {
    return <SignupSuccess signedUpAccount={signedUpAccount} />;
  }

  return (
    <form onSubmit={handleSubmit(submitSignup)} noValidate className="space-y-5">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight text-foreground">Create your pilot account</h1>
        <p className="mt-1 text-sm text-muted">
          Sign up, activate your account from the email link, and pay the 5 NC entry fee to launch.
        </p>
      </div>

      {formErrorMessage && (
        <div role="alert" className="flex items-start gap-2.5 rounded-lg border border-down/40 bg-down-soft/20 p-3">
          <TriangleAlert className="mt-0.5 size-4 shrink-0 text-down" />
          <p className="text-sm text-foreground">{formErrorMessage}</p>
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
      <Input
        label="Username"
        autoComplete="username"
        placeholder="stardust_runner"
        hint="3 to 20 letters, numbers or underscores. Other players see this."
        errorMessage={errors.username?.message}
        {...register("username")}
      />
      <PasswordInput
        label="Password"
        autoComplete="new-password"
        placeholder="At least 10 characters"
        errorMessage={errors.password?.message}
        belowField={<PasswordStrengthMeter password={typedPassword} />}
        {...register("password")}
      />
      <Controller
        control={control}
        name="acceptsTerms"
        render={({ field }) => (
          <Checkbox
            name={field.name}
            isChecked={field.value}
            onCheckedChange={field.onChange}
            errorMessage={errors.acceptsTerms?.message}
            label={
              <>
                I am 18 or older and I accept the{" "}
                <Link href="/terms" className="text-accent-soft underline-offset-2 hover:underline">
                  Terms
                </Link>{" "}
                and{" "}
                <Link href="/privacy" className="text-accent-soft underline-offset-2 hover:underline">
                  Privacy Policy
                </Link>
                .
              </>
            }
          />
        )}
      />

      <Button type="submit" size="lg" className="w-full" isLoading={signupMutation.isPending}>
        {signupMutation.isPending ? "Creating account" : "Create account"}
      </Button>

      <p className="text-center text-sm text-muted">
        Already have an account?{" "}
        <Link href="/login" className="font-medium text-accent-soft underline-offset-2 hover:underline">
          Log in
        </Link>
      </p>
    </form>
  );
}
