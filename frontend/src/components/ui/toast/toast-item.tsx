"use client";

import { CircleAlert, CircleCheck, Info, TriangleAlert, X, type LucideIcon } from "lucide-react";
import { Toast } from "radix-ui";

import type { ToastMessage, ToastTone } from "@/components/ui/toast/toast-context";
import { cn } from "@/utils/class-names";

const toneAppearance: Record<ToastTone, { icon: LucideIcon; iconClassName: string; borderClassName: string }> = {
  success: { icon: CircleCheck, iconClassName: "text-up", borderClassName: "border-up/40" },
  error: { icon: CircleAlert, iconClassName: "text-down", borderClassName: "border-down/40" },
  warning: { icon: TriangleAlert, iconClassName: "text-warning", borderClassName: "border-warning/40" },
  info: { icon: Info, iconClassName: "text-info", borderClassName: "border-info/40" },
};

type ToastItemProps = {
  toastMessage: ToastMessage;
  onDismiss: (toastId: string) => void;
};

export function ToastItem({ toastMessage, onDismiss }: ToastItemProps) {
  const appearance = toneAppearance[toastMessage.tone];
  const ToneIcon = appearance.icon;

  return (
    <Toast.Root
      duration={toastMessage.durationInMilliseconds}
      onOpenChange={(isOpen) => !isOpen && onDismiss(toastMessage.id)}
      className={cn(
        "flex w-full items-start gap-3 rounded-xl border bg-surface-raised p-4 shadow-xl shadow-black/50",
        "data-[state=open]:animate-slide-in data-[swipe=end]:opacity-0",
        appearance.borderClassName,
      )}
    >
      <ToneIcon className={cn("mt-0.5 size-4 shrink-0", appearance.iconClassName)} />
      <div className="min-w-0 flex-1">
        <Toast.Title className="text-sm font-medium text-foreground">{toastMessage.title}</Toast.Title>
        {toastMessage.description && (
          <Toast.Description className="mt-0.5 text-sm text-muted">{toastMessage.description}</Toast.Description>
        )}
      </div>
      <Toast.Close aria-label="Dismiss" className="rounded p-0.5 text-subtle transition-colors hover:text-foreground">
        <X className="size-4" />
      </Toast.Close>
    </Toast.Root>
  );
}
