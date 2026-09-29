import type { ComponentProps } from "react";
import { Slot } from "radix-ui";

import { Spinner } from "@/components/ui/spinner";
import { cn } from "@/utils/class-names";

export type ButtonVariant = "primary" | "secondary" | "ghost" | "danger" | "buy" | "sell";
export type ButtonSize = "sm" | "md" | "lg";

const variantClassNames: Record<ButtonVariant, string> = {
  primary: "bg-accent text-white hover:bg-accent-strong shadow-[0_0_20px_-6px] shadow-accent/70",
  secondary: "border border-border-strong bg-surface-raised text-foreground hover:bg-border",
  ghost: "text-muted hover:bg-surface-raised hover:text-foreground",
  danger: "bg-down text-white hover:bg-down/85",
  buy: "bg-up text-background hover:bg-up/85",
  sell: "bg-down text-white hover:bg-down/85",
};

const sizeClassNames: Record<ButtonSize, string> = {
  sm: "h-8 gap-1.5 rounded-md px-3 text-xs",
  md: "h-10 gap-2 rounded-lg px-4 text-sm",
  lg: "h-12 gap-2 rounded-lg px-6 text-base",
};

type ButtonProps = ComponentProps<"button"> & {
  variant?: ButtonVariant;
  size?: ButtonSize;
  isLoading?: boolean;
  asChild?: boolean;
};

export function Button({
  variant = "primary",
  size = "md",
  isLoading = false,
  asChild = false,
  disabled,
  className,
  children,
  type = "button",
  ...buttonProps
}: ButtonProps) {
  const sharedClassName = cn(
    "inline-flex shrink-0 items-center justify-center font-medium whitespace-nowrap transition-colors",
    "focus-visible:ring-2 focus-visible:ring-highlight focus-visible:ring-offset-2 focus-visible:ring-offset-background focus-visible:outline-none",
    "disabled:pointer-events-none disabled:opacity-50",
    variantClassNames[variant],
    sizeClassNames[size],
    className,
  );

  if (asChild) {
    return (
      <Slot.Root className={sharedClassName} {...buttonProps}>
        {children}
      </Slot.Root>
    );
  }

  return (
    <button
      type={type}
      disabled={disabled || isLoading}
      aria-busy={isLoading || undefined}
      className={sharedClassName}
      {...buttonProps}
    >
      {isLoading && <Spinner />}
      {children}
    </button>
  );
}
