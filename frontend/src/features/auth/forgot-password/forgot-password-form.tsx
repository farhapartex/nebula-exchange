"use client";

import Link from "next/link";

import { requestPasswordReset } from "@/features/auth/api/password-reset-api";
import { EmailLinkRequestForm } from "@/features/auth/shared/email-link-request-form";

export function ForgotPasswordForm() {
  return (
    <EmailLinkRequestForm
      title="Forgot your password?"
      description="Enter the email you use to log in and we'll send you a link to choose a new password."
      submitLabel="Send reset link"
      resubmitLabel="Send another link"
      requestLink={requestPasswordReset}
      describeSentLink={(requestedEmail) => (
        <>
          If <span className="font-medium text-foreground">{requestedEmail}</span> has an account, a reset link is on
          its way. It works once and expires in 30 minutes.
        </>
      )}
      footer={
        <p className="text-center text-sm text-muted">
          Remembered it?{" "}
          <Link href="/login" className="font-medium text-accent-soft underline-offset-2 hover:underline">
            Back to login
          </Link>
        </p>
      }
    />
  );
}
