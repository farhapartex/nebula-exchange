import { Suspense } from "react";
import type { Metadata } from "next";

import { PaymentResultView } from "@/features/payments/payment-result-view";

export const metadata: Metadata = {
  title: "Payment",
};

export default function PaymentResultPage() {
  return (
    <Suspense>
      <PaymentResultView />
    </Suspense>
  );
}
