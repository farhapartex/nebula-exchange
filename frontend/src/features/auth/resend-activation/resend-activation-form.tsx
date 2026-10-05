"use client";

import Link from "next/link";

import { requestActivationEmail } from "@/features/auth/api/activation-api";
import { EmailLinkRequestForm } from "@/features/auth/shared/email-link-request-form";

export function ResendActivationForm() {
  return (
    <EmailLinkRequestForm
      title="Send a new activation link"
      description="Enter the email you signed up with. Any older activation links stop working."
      submitLabel="Send activation link"
      resubmitLabel="Send another link"
      requestLink={requestActivationEmail}
      describeSentLink={(requestedEmail) => (
        <>
          If <span className="font-medium text-foreground">{requestedEmail}</span> belongs to an account that isn&apos;t
          activated yet, a new link is on its way. It stays valid for 24 hours.
        </>
      )}
      footer={
        <p className="text-center text-sm text-muted">
          Already activated?{" "}
          <Link href="/login" className="font-medium text-accent-soft underline-offset-2 hover:underline">
            Log in
          </Link>
        </p>
      }
    />
  );
}
