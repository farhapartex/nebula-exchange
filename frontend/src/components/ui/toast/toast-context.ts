"use client";

import { createContext } from "react";

export type ToastTone = "success" | "error" | "warning" | "info";

export type ToastRequest = {
  title: string;
  description?: string;
  tone?: ToastTone;
  durationInMilliseconds?: number;
};

export type ToastMessage = Required<Omit<ToastRequest, "description">> & {
  id: string;
  description?: string;
};

export type ToastContextValue = {
  showToast: (toastRequest: ToastRequest) => void;
};

export const ToastContext = createContext<ToastContextValue | null>(null);
