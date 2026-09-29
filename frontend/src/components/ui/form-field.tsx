import type { ReactNode } from "react";

import { cn } from "@/utils/class-names";

type FormFieldProps = {
  inputId: string;
  label?: string;
  hint?: string;
  errorMessage?: string;
  className?: string;
  children: ReactNode;
};

export function formFieldDescriptionId(inputId: string) {
  return `${inputId}-description`;
}

export function FormField({ inputId, label, hint, errorMessage, className, children }: FormFieldProps) {
  const descriptionText = errorMessage ?? hint;

  return (
    <div className={cn("flex flex-col gap-1.5", className)}>
      {label && (
        <label htmlFor={inputId} className="text-sm font-medium text-foreground">
          {label}
        </label>
      )}
      {children}
      {descriptionText && (
        <p id={formFieldDescriptionId(inputId)} className={cn("text-xs", errorMessage ? "text-down" : "text-subtle")}>
          {descriptionText}
        </p>
      )}
    </div>
  );
}

export const fieldControlClassName = cn(
  "h-10 w-full rounded-lg border border-border-strong bg-background/70 px-3 text-sm text-foreground",
  "transition-colors placeholder:text-subtle",
  "focus:border-highlight focus:ring-2 focus:ring-highlight/30 focus:outline-none",
  "disabled:cursor-not-allowed disabled:opacity-50",
  "aria-invalid:border-down aria-invalid:focus:ring-down/30",
);
