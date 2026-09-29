import Link from "next/link";
import { LockKeyhole, Timer } from "lucide-react";

import { describeWaitTime, formatCountdown } from "@/utils/time/format-duration";

type LoginThrottleBannerProps = {
  reason: "LOGIN_LOCKED" | "RATE_LIMITED";
  remainingSeconds: number;
};

export function LoginThrottleBanner({ reason, remainingSeconds }: LoginThrottleBannerProps) {
  const isAccountLock = reason === "LOGIN_LOCKED";
  const BannerIcon = isAccountLock ? LockKeyhole : Timer;

  return (
    <div role="alert" className="flex items-start gap-2.5 rounded-lg border border-warning/40 bg-warning/10 p-3">
      <BannerIcon className="mt-0.5 size-4 shrink-0 text-warning" />
      <div className="min-w-0 flex-1 text-sm">
        <div className="flex items-center justify-between gap-3">
          <p className="font-medium text-foreground">
            {isAccountLock ? "Too many failed attempts" : "Too many login attempts"}
          </p>
          <span className="font-mono text-sm text-warning tabular-nums" aria-hidden="true">
            {formatCountdown(remainingSeconds)}
          </span>
        </div>
        <p className="mt-0.5 text-muted">
          {isAccountLock
            ? `For your security, logins to this account are paused. Try again in ${describeWaitTime(remainingSeconds)}.`
            : `Please wait ${describeWaitTime(remainingSeconds)} before trying again.`}
        </p>
        {isAccountLock && (
          <Link
            href="/forgot"
            className="mt-2 inline-block font-medium text-accent-soft underline-offset-2 hover:underline"
          >
            Forgot your password?
          </Link>
        )}
      </div>
    </div>
  );
}
