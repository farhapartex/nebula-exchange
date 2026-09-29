"use client";

import { useId, type ReactNode } from "react";
import { Check } from "lucide-react";
import { Checkbox as CheckboxPrimitive } from "radix-ui";

import { formFieldDescriptionId } from "@/components/ui/form-field";
import { cn } from "@/utils/class-names";

type CheckboxProps = {
  isChecked: boolean;
  onCheckedChange: (isChecked: boolean) => void;
  label: ReactNode;
  errorMessage?: string;
  isDisabled?: boolean;
  name?: string;
  className?: string;
};

export function Checkbox({
  isChecked,
  onCheckedChange,
  label,
  errorMessage,
  isDisabled,
  name,
  className,
}: CheckboxProps) {
  const checkboxId = useId();

  return (
    <div className={cn("flex flex-col gap-1.5", className)}>
      <div className="flex items-start gap-3">
        <CheckboxPrimitive.Root
          id={checkboxId}
          name={name}
          checked={isChecked}
          onCheckedChange={(checkedState) => onCheckedChange(checkedState === true)}
          disabled={isDisabled}
          aria-invalid={errorMessage ? true : undefined}
          aria-describedby={errorMessage ? formFieldDescriptionId(checkboxId) : undefined}
          className={cn(
            "mt-0.5 flex size-5 shrink-0 items-center justify-center rounded-md border border-border-strong bg-background/70 transition-colors",
            "focus-visible:ring-2 focus-visible:ring-highlight focus-visible:ring-offset-2 focus-visible:ring-offset-background focus-visible:outline-none",
            "data-[state=checked]:border-accent data-[state=checked]:bg-accent",
            "disabled:cursor-not-allowed disabled:opacity-50 aria-invalid:border-down",
          )}
        >
          <CheckboxPrimitive.Indicator>
            <Check className="size-3.5 text-white" strokeWidth={3} />
          </CheckboxPrimitive.Indicator>
        </CheckboxPrimitive.Root>
        <label htmlFor={checkboxId} className="text-sm leading-snug text-muted select-none">
          {label}
        </label>
      </div>
      {errorMessage && (
        <p id={formFieldDescriptionId(checkboxId)} className="pl-8 text-xs text-down">
          {errorMessage}
        </p>
      )}
    </div>
  );
}
