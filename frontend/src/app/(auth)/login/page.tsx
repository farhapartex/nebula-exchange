import { Suspense } from "react";
import type { Metadata } from "next";

import { LoginForm } from "@/features/auth/login/login-form";

export const metadata: Metadata = {
  title: "Log in",
};

export default function LoginPage() {
  return (
    <Suspense>
      <LoginForm />
    </Suspense>
  );
}
