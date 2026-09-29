"use client";

import { useCallback, useMemo, useState, type ReactNode } from "react";
import { Toast } from "radix-ui";

import { ToastContext, type ToastMessage, type ToastRequest } from "@/components/ui/toast/toast-context";
import { ToastItem } from "@/components/ui/toast/toast-item";

const defaultToastDurationInMilliseconds = 5000;
const maximumVisibleToasts = 4;

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toastMessages, setToastMessages] = useState<ToastMessage[]>([]);

  const showToast = useCallback((toastRequest: ToastRequest) => {
    const toastMessage: ToastMessage = {
      id: crypto.randomUUID(),
      title: toastRequest.title,
      description: toastRequest.description,
      tone: toastRequest.tone ?? "info",
      durationInMilliseconds: toastRequest.durationInMilliseconds ?? defaultToastDurationInMilliseconds,
    };
    setToastMessages((currentMessages) => [...currentMessages, toastMessage].slice(-maximumVisibleToasts));
  }, []);

  const dismissToast = useCallback((toastId: string) => {
    setToastMessages((currentMessages) => currentMessages.filter((toastMessage) => toastMessage.id !== toastId));
  }, []);

  const toastContextValue = useMemo(() => ({ showToast }), [showToast]);

  return (
    <ToastContext.Provider value={toastContextValue}>
      <Toast.Provider swipeDirection="right">
        {children}
        {toastMessages.map((toastMessage) => (
          <ToastItem key={toastMessage.id} toastMessage={toastMessage} onDismiss={dismissToast} />
        ))}
        <Toast.Viewport className="fixed right-0 bottom-0 z-[60] flex w-full max-w-sm flex-col gap-2 p-4 outline-none" />
      </Toast.Provider>
    </ToastContext.Provider>
  );
}
