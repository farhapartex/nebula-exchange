"use client";

import type { ReactNode } from "react";
import { X } from "lucide-react";
import { Dialog } from "radix-ui";

import { overlayBackdropClassName, overlayPanelClassName } from "@/components/ui/overlay-styles";
import { cn } from "@/utils/class-names";

type ModalProps = {
  isOpen: boolean;
  onOpenChange: (isOpen: boolean) => void;
  title: string;
  description?: string;
  children?: ReactNode;
  footer?: ReactNode;
  className?: string;
};

export function Modal({ isOpen, onOpenChange, title, description, children, footer, className }: ModalProps) {
  return (
    <Dialog.Root open={isOpen} onOpenChange={onOpenChange}>
      <Dialog.Portal>
        <Dialog.Overlay className={overlayBackdropClassName} />
        <Dialog.Content className={cn(overlayPanelClassName, className)}>
          <div className="mb-4 flex items-start justify-between gap-4">
            <div>
              <Dialog.Title className="text-lg font-semibold text-foreground">{title}</Dialog.Title>
              {description ? (
                <Dialog.Description className="mt-1 text-sm text-muted">{description}</Dialog.Description>
              ) : (
                <Dialog.Description className="sr-only">{title}</Dialog.Description>
              )}
            </div>
            <Dialog.Close
              aria-label="Close"
              className="rounded-md p-1 text-muted transition-colors hover:bg-border hover:text-foreground focus-visible:ring-2 focus-visible:ring-highlight focus-visible:outline-none"
            >
              <X className="size-4" />
            </Dialog.Close>
          </div>
          {children}
          {footer && <div className="mt-6 flex justify-end gap-3">{footer}</div>}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
