"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { useForm } from "react-hook-form";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { PasswordInput } from "@/components/ui/password-input";
import { logIn } from "@/features/auth/api/session-api";
import { LoginErrorBanner } from "@/features/auth/login/login-error-banner";
import { loginFieldNames, loginSchema, type LoginFormValues } from "@/features/auth/login/login-schema";
import { useAuth } from "@/features/auth/session/use-auth";
import { applyServerFieldErrors } from "@/utils/forms/apply-server-field-errors";
import { safeInternalPath } from "@/utils/navigation/safe-internal-path";

const defaultPathAfterLogin = "/hangar";

export function LoginForm() {
  const router = useRouter();
  const searchParameters = useSearchParams();
  const { startSession } = useAuth();

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors },
  } = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: "", password: "" },
    mode: "onTouched",
  });

  const loginMutation = useMutation({
    mutationFn: (formValues: LoginFormValues) =>
      logIn({ email: formValues.email.trim(), password: formValues.password }),
    onSuccess: (establishedSession) => {
      startSession(establishedSession);
      router.replace(safeInternalPath(searchParameters.get("next")) ?? defaultPathAfterLogin);
    },
    onError: (error) => {
      applyServerFieldErrors(error, loginFieldNames, setError);
    },
  });

  const shouldShowErrorBanner = loginMutation.isError && Object.keys(errors).length === 0;

  return (
    <form onSubmit={handleSubmit((formValues) => loginMutation.mutate(formValues))} noValidate className="space-y-5">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight text-foreground">Welcome back, pilot</h1>
        <p className="mt-1 text-sm text-muted">Log in to return to your hangar.</p>
      </div>

      {shouldShowErrorBanner && <LoginErrorBanner error={loginMutation.error} />}

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

      <Button type="submit" size="lg" className="w-full" isLoading={loginMutation.isPending}>
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
