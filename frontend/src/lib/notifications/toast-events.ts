import type { ToastRequest } from "@/components/ui/toast/toast-context";

type ToastListener = (toastRequest: ToastRequest) => void;

const toastListeners = new Set<ToastListener>();

export function subscribeToToastEvents(listener: ToastListener): () => void {
  toastListeners.add(listener);
  return () => {
    toastListeners.delete(listener);
  };
}

export function publishToastEvent(toastRequest: ToastRequest): void {
  toastListeners.forEach((listener) => listener(toastRequest));
}
