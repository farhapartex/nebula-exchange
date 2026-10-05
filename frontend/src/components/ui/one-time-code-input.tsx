"use client";

import type { ComponentProps } from "react";

import { Input } from "@/components/ui/input";
import { cn } from "@/utils/class-names";

type OneTimeCodeInputProps = Omit<ComponentProps<typeof Input>, "type" | "inputMode" | "maxLength" | "onChange"> & {
  value: string;
  onValueChange: (code: string) => void;
};

export function OneTimeCodeInput({ value, onValueChange, className, ...inputProps }: OneTimeCodeInputProps) {
  return (
    <Input
      type="text"
      inputMode="numeric"
      autoComplete="one-time-code"
      maxLength={6}
      placeholder="000000"
      value={value}
      onChange={(changeEvent) => onValueChange(changeEvent.target.value.replace(/\D/g, "").slice(0, 6))}
      className={cn("text-center font-mono text-lg tracking-[0.5em] tabular-nums", className)}
      {...inputProps}
    />
  );
}
