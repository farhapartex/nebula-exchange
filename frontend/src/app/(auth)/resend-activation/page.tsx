import { Suspense } from "react";
import type { Metadata } from "next";

import { ResendActivationForm } from "@/features/auth/resend-activation/resend-activation-form";

export const metadata: Metadata = {
  title: "Resend activation link",
};

export default function ResendActivationPage() {
  return (
    <Suspense>
      <ResendActivationForm />
    </Suspense>
  );
}
