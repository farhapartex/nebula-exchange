"use client";

import { useId, useState, type ComponentProps, type ReactNode } from "react";
import { Eye, EyeOff } from "lucide-react";

import { fieldControlClassName, FormField, formFieldDescriptionId } from "@/components/ui/form-field";
import { cn } from "@/utils/class-names";

type PasswordInputProps = Omit<ComponentProps<"input">, "type"> & {
  label?: string;
  hint?: string;
  errorMessage?: string;
  belowField?: ReactNode;
};

export function PasswordInput({
  id,
  label,
  hint,
  errorMessage,
  belowField,
  className,
  ...inputProps
}: PasswordInputProps) {
  const generatedId = useId();
  const inputId = id ?? generatedId;
  const [isVisible, setIsVisible] = useState(false);
  const hasDescription = Boolean(errorMessage ?? hint);

  return (
    <FormField inputId={inputId} label={label} hint={hint} errorMessage={errorMessage}>
      <div className="relative">
        <input
          id={inputId}
          type={isVisible ? "text" : "password"}
          aria-invalid={errorMessage ? true : undefined}
          aria-describedby={hasDescription ? formFieldDescriptionId(inputId) : undefined}
          className={cn(fieldControlClassName, "pr-11", className)}
          {...inputProps}
        />
        <button
          type="button"
          onClick={() => setIsVisible((wasVisible) => !wasVisible)}
          aria-label={isVisible ? "Hide password" : "Show password"}
          aria-pressed={isVisible}
          className="absolute inset-y-0 right-0 flex w-10 items-center justify-center rounded-r-lg text-subtle transition-colors hover:text-foreground focus-visible:text-foreground focus-visible:outline-none"
        >
          {isVisible ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
        </button>
      </div>
      {belowField}
    </FormField>
  );
}
