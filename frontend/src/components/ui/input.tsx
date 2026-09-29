"use client";

import { useId, type ComponentProps, type ReactNode } from "react";

import { fieldControlClassName, FormField, formFieldDescriptionId } from "@/components/ui/form-field";
import { cn } from "@/utils/class-names";

type InputProps = ComponentProps<"input"> & {
  label?: string;
  hint?: string;
  errorMessage?: string;
  trailingAdornment?: ReactNode;
  containerClassName?: string;
};

export function Input({
  id,
  label,
  hint,
  errorMessage,
  trailingAdornment,
  containerClassName,
  className,
  ...inputProps
}: InputProps) {
  const generatedId = useId();
  const inputId = id ?? generatedId;
  const hasDescription = Boolean(errorMessage ?? hint);

  return (
    <FormField inputId={inputId} label={label} hint={hint} errorMessage={errorMessage} className={containerClassName}>
      <div className="relative">
        <input
          id={inputId}
          aria-invalid={errorMessage ? true : undefined}
          aria-describedby={hasDescription ? formFieldDescriptionId(inputId) : undefined}
          className={cn(fieldControlClassName, trailingAdornment && "pr-14", className)}
          {...inputProps}
        />
        {trailingAdornment && (
          <span className="pointer-events-none absolute inset-y-0 right-3 flex items-center text-xs font-medium text-muted">
            {trailingAdornment}
          </span>
        )}
      </div>
    </FormField>
  );
}
