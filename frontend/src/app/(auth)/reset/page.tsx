import { Suspense } from "react";
import type { Metadata } from "next";

import { ResetPasswordFlow } from "@/features/auth/reset-password/reset-password-flow";

export const metadata: Metadata = {
  title: "Reset password",
};

export default function ResetPasswordPage() {
  return (
    <Suspense>
      <ResetPasswordFlow />
    </Suspense>
  );
}
