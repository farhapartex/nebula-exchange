"use client";

import { useState, type ReactNode } from "react";
import { AlertDialog } from "radix-ui";

import { Button } from "@/components/ui/button";
import { overlayBackdropClassName, overlayPanelClassName } from "@/components/ui/overlay-styles";

type ConfirmDialogProps = {
  isOpen: boolean;
  onOpenChange: (isOpen: boolean) => void;
  title: string;
  description?: string;
  children?: ReactNode;
  confirmLabel?: string;
  cancelLabel?: string;
  tone?: "default" | "danger";
  onConfirm: () => void | Promise<void>;
};

export function ConfirmDialog({
  isOpen,
  onOpenChange,
  title,
  description,
  children,
  confirmLabel = "Confirm",
  cancelLabel = "Cancel",
  tone = "default",
  onConfirm,
}: ConfirmDialogProps) {
  const [isConfirming, setIsConfirming] = useState(false);

  async function handleConfirm() {
    setIsConfirming(true);
    try {
      await onConfirm();
      onOpenChange(false);
    } finally {
      setIsConfirming(false);
    }
  }

  return (
    <AlertDialog.Root open={isOpen} onOpenChange={(nextIsOpen) => !isConfirming && onOpenChange(nextIsOpen)}>
      <AlertDialog.Portal>
        <AlertDialog.Overlay className={overlayBackdropClassName} />
        <AlertDialog.Content className={overlayPanelClassName}>
          <AlertDialog.Title className="text-lg font-semibold text-foreground">{title}</AlertDialog.Title>
          <AlertDialog.Description className={description ? "mt-1 text-sm text-muted" : "sr-only"}>
            {description ?? title}
          </AlertDialog.Description>
          {children && <div className="mt-4">{children}</div>}
          <div className="mt-6 flex justify-end gap-3">
            <AlertDialog.Cancel asChild>
              <Button variant="secondary" disabled={isConfirming}>
                {cancelLabel}
              </Button>
            </AlertDialog.Cancel>
            <Button variant={tone === "danger" ? "danger" : "primary"} isLoading={isConfirming} onClick={handleConfirm}>
              {confirmLabel}
            </Button>
          </div>
        </AlertDialog.Content>
      </AlertDialog.Portal>
    </AlertDialog.Root>
  );
}
