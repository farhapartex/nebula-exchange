"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { useState } from "react";
import { useForm, useWatch } from "react-hook-form";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { PasswordInput } from "@/components/ui/password-input";
import { logIn } from "@/features/auth/api/session-api";
import { LoginErrorBanner } from "@/features/auth/login/login-error-banner";
import { LoginThrottleBanner } from "@/features/auth/login/login-throttle-banner";
import { loginFieldNames, loginSchema, type LoginFormValues } from "@/features/auth/login/login-schema";
import { SessionEndNotice } from "@/features/auth/login/session-end-notice";
import { parseSessionEndReason } from "@/features/auth/session/session-end-reasons";
import { useAuth } from "@/features/auth/session/use-auth";
import { isApiError } from "@/lib/api/api-error";
import { applyServerFieldErrors } from "@/utils/forms/apply-server-field-errors";
import { safeInternalPath } from "@/utils/navigation/safe-internal-path";
import { useCooldown } from "@/utils/time/use-cooldown";

const defaultPathAfterLogin = "/hangar";
const fallbackThrottleSeconds = 60;

type LoginThrottle = {
  reason: "LOGIN_LOCKED" | "RATE_LIMITED";
  email: string;
};

export function LoginForm() {
  const router = useRouter();
  const searchParameters = useSearchParams();
  const { startSession } = useAuth();

  const [activeThrottle, setActiveThrottle] = useState<LoginThrottle | null>(null);
  const throttleCooldown = useCooldown(fallbackThrottleSeconds);

  const {
    register,
    control,
    handleSubmit,
    setError,
    formState: { errors },
  } = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: "", password: "" },
    mode: "onTouched",
  });

  const typedEmail = useWatch({ control, name: "email" });

  const loginMutation = useMutation({
    mutationFn: (formValues: LoginFormValues) =>
      logIn({ email: formValues.email.trim(), password: formValues.password }),
    meta: { showsThrottlingInline: true },
    onSuccess: (establishedSession) => {
      startSession(establishedSession);
      router.replace(safeInternalPath(searchParameters.get("next")) ?? defaultPathAfterLogin);
    },
    onError: (error, formValues) => {
      if (isApiError(error) && (error.code === "LOGIN_LOCKED" || error.code === "RATE_LIMITED")) {
        setActiveThrottle({ reason: error.code, email: formValues.email.trim().toLowerCase() });
        throttleCooldown.startCooldown(error.retryAfterSeconds ?? fallbackThrottleSeconds);
        return;
      }
      applyServerFieldErrors(error, loginFieldNames, setError);
    },
  });

  const isThrottleActive = activeThrottle !== null && throttleCooldown.isCoolingDown;
  const isTypedEmailThrottled =
    isThrottleActive &&
    (activeThrottle.reason === "RATE_LIMITED" || activeThrottle.email === typedEmail.trim().toLowerCase());
  const shouldShowErrorBanner =
    loginMutation.isError && !isApiErrorThrottled(loginMutation.error) && Object.keys(errors).length === 0;
  const sessionEndReason = parseSessionEndReason(searchParameters.get("reason"));
  const shouldShowSessionEndNotice = sessionEndReason !== null && !loginMutation.isError;

  return (
    <form onSubmit={handleSubmit((formValues) => loginMutation.mutate(formValues))} noValidate className="space-y-5">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight text-foreground">Welcome back, pilot</h1>
        <p className="mt-1 text-sm text-muted">Log in to return to your hangar.</p>
      </div>

      {shouldShowSessionEndNotice && <SessionEndNotice reason={sessionEndReason} />}

      {isTypedEmailThrottled && (
        <LoginThrottleBanner reason={activeThrottle.reason} remainingSeconds={throttleCooldown.remainingSeconds} />
      )}

      {shouldShowErrorBanner && (
        <LoginErrorBanner error={loginMutation.error} attemptedEmail={loginMutation.variables?.email.trim()} />
      )}

      <Input
        label="Email"
        type="email"
        autoComplete="email"
        placeholder="pilot@nebula.test"
        errorMessage={errors.email?.message}
        {...register("email")}
      />
      <div className="space-y-1.5">
        <PasswordInput
          label="Password"
          autoComplete="current-password"
          placeholder="Your password"
          errorMessage={errors.password?.message}
          {...register("password")}
        />
        <div className="text-right">
          <Link href="/forgot" className="text-xs text-accent-soft underline-offset-2 hover:underline">
            Forgot password?
          </Link>
        </div>
      </div>

      <Button
        type="submit"
        size="lg"
        className="w-full"
        isLoading={loginMutation.isPending}
        disabled={isTypedEmailThrottled}
      >
        {loginMutation.isPending ? "Logging in" : "Log in"}
      </Button>

      <p className="text-center text-sm text-muted">
        New to the nebula?{" "}
        <Link href="/signup" className="font-medium text-accent-soft underline-offset-2 hover:underline">
          Create an account
        </Link>
      </p>
    </form>
  );
}

function isApiErrorThrottled(error: unknown): boolean {
  return isApiError(error) && error.isThrottled;
}
