import Link from "next/link";
import { MailWarning, ShieldBan, TriangleAlert, type LucideIcon } from "lucide-react";

import { isApiError } from "@/lib/api/api-error";
import { cn } from "@/utils/class-names";

type LoginErrorPresentation = {
  icon: LucideIcon;
  title: string;
  message: string;
  toneClassName: string;
  showsResendLink?: boolean;
};

function presentLoginError(error: unknown): LoginErrorPresentation {
  if (isApiError(error) && error.code === "ACCOUNT_NOT_ACTIVATED") {
    return {
      icon: MailWarning,
      title: "Your account isn't activated yet",
      message: "Open the activation link we emailed you. It stays valid for 24 hours.",
      toneClassName: "border-warning/40 bg-warning/10 text-warning",
      showsResendLink: true,
    };
  }
  if (isApiError(error) && error.code === "FORBIDDEN") {
    return {
      icon: ShieldBan,
      title: "Access denied",
      message: error.message,
      toneClassName: "border-down/40 bg-down-soft/20 text-down",
    };
  }
  return {
    icon: TriangleAlert,
    title: isApiError(error) && error.statusCode === 401 ? "Login failed" : "Something went wrong",
    message: isApiError(error) ? error.message : "Please try again.",
    toneClassName: "border-down/40 bg-down-soft/20 text-down",
  };
}

export function LoginErrorBanner({ error }: { error: unknown }) {
  const presentation = presentLoginError(error);
  const ErrorIcon = presentation.icon;

  return (
    <div role="alert" className={cn("flex items-start gap-2.5 rounded-lg border p-3", presentation.toneClassName)}>
      <ErrorIcon className="mt-0.5 size-4 shrink-0" />
      <div className="text-sm">
        <p className="font-medium text-foreground">{presentation.title}</p>
        <p className="mt-0.5 text-muted">{presentation.message}</p>
        {presentation.showsResendLink && (
          <Link
            href="/resend-activation"
            className="mt-2 inline-block font-medium text-accent-soft underline-offset-2 hover:underline"
          >
            Send a new activation link
          </Link>
        )}
      </div>
    </div>
  );
}
