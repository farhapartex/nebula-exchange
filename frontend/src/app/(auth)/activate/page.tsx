import { Suspense } from "react";
import type { Metadata } from "next";

import { ActivationFlow } from "@/features/auth/activation/activation-flow";

export const metadata: Metadata = {
  title: "Activate account",
};

export default function ActivatePage() {
  return (
    <Suspense>
      <ActivationFlow />
    </Suspense>
  );
}
