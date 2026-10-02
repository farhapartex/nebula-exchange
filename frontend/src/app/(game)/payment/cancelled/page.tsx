import { Suspense } from "react";
import type { Metadata } from "next";

import { PaymentCancelledView } from "@/features/payments/payment-cancelled-view";

export const metadata: Metadata = {
  title: "Payment cancelled",
};

export default function PaymentCancelledPage() {
  return (
    <Suspense>
      <PaymentCancelledView />
    </Suspense>
  );
}
