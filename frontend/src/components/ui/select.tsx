"use client";

import { useId } from "react";
import { Check, ChevronDown } from "lucide-react";
import { Select as SelectPrimitive } from "radix-ui";

import { fieldControlClassName, FormField, formFieldDescriptionId } from "@/components/ui/form-field";
import { cn } from "@/utils/class-names";

export type SelectOption = {
  value: string;
  label: string;
  isDisabled?: boolean;
};

type SelectProps = {
  options: SelectOption[];
  value?: string;
  defaultValue?: string;
  onValueChange?: (selectedValue: string) => void;
  placeholder?: string;
  label?: string;
  hint?: string;
  errorMessage?: string;
  isDisabled?: boolean;
  name?: string;
  className?: string;
};

export function Select({
  options,
  value,
  defaultValue,
  onValueChange,
  placeholder = "Select an option",
  label,
  hint,
  errorMessage,
  isDisabled,
  name,
  className,
}: SelectProps) {
  const triggerId = useId();
  const hasDescription = Boolean(errorMessage ?? hint);

  return (
    <FormField inputId={triggerId} label={label} hint={hint} errorMessage={errorMessage} className={className}>
      <SelectPrimitive.Root
        value={value}
        defaultValue={defaultValue}
        onValueChange={onValueChange}
        disabled={isDisabled}
        name={name}
      >
        <SelectPrimitive.Trigger
          id={triggerId}
          aria-invalid={errorMessage ? true : undefined}
          aria-describedby={hasDescription ? formFieldDescriptionId(triggerId) : undefined}
          className={cn(fieldControlClassName, "flex items-center justify-between gap-2 data-placeholder:text-subtle")}
        >
          <SelectPrimitive.Value placeholder={placeholder} />
          <SelectPrimitive.Icon>
            <ChevronDown className="size-4 text-muted" />
          </SelectPrimitive.Icon>
        </SelectPrimitive.Trigger>
        <SelectPrimitive.Portal>
          <SelectPrimitive.Content
            position="popper"
            sideOffset={6}
            className="z-50 max-h-72 min-w-(--radix-select-trigger-width) overflow-hidden rounded-lg border border-border-strong bg-surface-raised shadow-xl shadow-black/40"
          >
            <SelectPrimitive.Viewport className="p-1">
              {options.map((option) => (
                <SelectPrimitive.Item
                  key={option.value}
                  value={option.value}
                  disabled={option.isDisabled}
                  className="relative flex h-9 cursor-pointer items-center rounded-md pr-8 pl-3 text-sm text-foreground outline-none select-none data-disabled:cursor-not-allowed data-disabled:opacity-40 data-highlighted:bg-border"
                >
                  <SelectPrimitive.ItemText>{option.label}</SelectPrimitive.ItemText>
                  <SelectPrimitive.ItemIndicator className="absolute right-2.5">
                    <Check className="size-4 text-highlight" />
                  </SelectPrimitive.ItemIndicator>
                </SelectPrimitive.Item>
              ))}
            </SelectPrimitive.Viewport>
          </SelectPrimitive.Content>
        </SelectPrimitive.Portal>
      </SelectPrimitive.Root>
    </FormField>
  );
}
