import type { ComponentProps, ReactNode } from "react";

import { cn } from "@/utils/class-names";

export function Card({ className, ...cardProps }: ComponentProps<"div">) {
  return <div className={cn("rounded-2xl border border-border bg-surface/80", className)} {...cardProps} />;
}

type CardHeaderProps = {
  title: string;
  description?: string;
  action?: ReactNode;
  className?: string;
};

export function CardHeader({ title, description, action, className }: CardHeaderProps) {
  return (
    <div className={cn("flex items-start justify-between gap-4 px-5 pt-5", className)}>
      <div className="min-w-0">
        <h3 className="text-base font-semibold text-foreground">{title}</h3>
        {description && <p className="mt-1 text-sm text-muted">{description}</p>}
      </div>
      {action && <div className="shrink-0">{action}</div>}
    </div>
  );
}

export function CardContent({ className, ...contentProps }: ComponentProps<"div">) {
  return <div className={cn("p-5", className)} {...contentProps} />;
}

export function CardFooter({ className, ...footerProps }: ComponentProps<"div">) {
  return (
    <div
      className={cn("flex items-center justify-end gap-3 border-t border-border px-5 py-4", className)}
      {...footerProps}
    />
  );
}
