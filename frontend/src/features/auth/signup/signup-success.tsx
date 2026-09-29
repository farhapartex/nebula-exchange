import { MailCheck } from "lucide-react";

import { StatusBadge } from "@/components/ui/status-badge";
import type { SignedUpAccount } from "@/features/auth/api/auth-types";

export function SignupSuccess({ signedUpAccount }: { signedUpAccount: SignedUpAccount }) {
  return (
    <div className="text-center">
      <span className="mx-auto mb-4 flex size-12 items-center justify-center rounded-full bg-up/10 text-up">
        <MailCheck className="size-6" />
      </span>
      <h2 className="text-xl font-semibold text-foreground">Check your email</h2>
      <p className="mt-2 text-sm text-muted">
        We sent a 6-digit code to <span className="font-medium text-foreground">{signedUpAccount.email}</span>. It
        expires in 15 minutes.
      </p>
      <p className="mt-4 text-sm text-muted">
        Welcome aboard, <span className="font-mono text-accent-soft">{signedUpAccount.username}</span>.
      </p>
      <StatusBadge className="mt-6" label="Code entry arrives in T-013" tone="neutral" />
    </div>
  );
}
