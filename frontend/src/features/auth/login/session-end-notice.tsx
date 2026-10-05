import { Clock, KeyRound, LogOut, type LucideIcon } from "lucide-react";

import type { SessionEndReason } from "@/features/auth/session/session-end-reasons";

const noticeByReason: Record<SessionEndReason, { icon: LucideIcon; title: string; message: string }> = {
  signed_out: {
    icon: LogOut,
    title: "You're logged out",
    message: "See you back in the streets soon.",
  },
  password_reset: {
    icon: KeyRound,
    title: "Password updated",
    message: "You were logged out everywhere. Log in with your new password.",
  },
  session_expired: {
    icon: Clock,
    title: "Your session expired",
    message: "For your security we signed you out. Log in again to continue where you left off.",
  },
};

export function SessionEndNotice({ reason }: { reason: SessionEndReason }) {
  const notice = noticeByReason[reason];
  const NoticeIcon = notice.icon;

  return (
    <div role="status" className="flex items-start gap-2.5 rounded-lg border border-info/40 bg-info/10 p-3">
      <NoticeIcon className="mt-0.5 size-4 shrink-0 text-info" />
      <div className="text-sm">
        <p className="font-medium text-foreground">{notice.title}</p>
        <p className="mt-0.5 text-muted">{notice.message}</p>
      </div>
    </div>
  );
}
