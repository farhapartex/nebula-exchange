import { Suspense } from "react";
import type { Metadata } from "next";

import { ForgotPasswordForm } from "@/features/auth/forgot-password/forgot-password-form";

export const metadata: Metadata = {
  title: "Forgot password",
};

export default function ForgotPasswordPage() {
  return (
    <Suspense>
      <ForgotPasswordForm />
    </Suspense>
  );
}
