"use client";

import { useState } from "react";

import { ContentSection } from "@/components/layout/content-section";
import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { Input } from "@/components/ui/input";
import { Modal } from "@/components/ui/modal";
import type { ToastTone } from "@/components/ui/toast/toast-context";
import { useToast } from "@/components/ui/toast/use-toast";

const toastExamples: { tone: ToastTone; title: string; description: string }[] = [
  { tone: "success", title: "Order filled", description: "Bought 40 IRON at 0.0100 NC." },
  { tone: "error", title: "Insufficient funds", description: "You need 0.42 NC more for this order." },
  { tone: "warning", title: "Market halted", description: "IRON/NC reopens in 5 minutes." },
  { tone: "info", title: "Mission complete", description: "Your Scout is back from the Asteroid Belt." },
];

function wait(milliseconds: number) {
  return new Promise<void>((resolve) => window.setTimeout(resolve, milliseconds));
}

export function OverlayShowcase() {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isConfirmOpen, setIsConfirmOpen] = useState(false);
  const [isDangerConfirmOpen, setIsDangerConfirmOpen] = useState(false);
  const { showToast } = useToast();

  async function confirmPurchase() {
    await wait(1200);
    showToast({ tone: "success", title: "Purchase complete", description: "20 Fuel Cells added to inventory." });
  }

  return (
    <ContentSection title="Modal, ConfirmDialog and Toast" description="Overlays trap focus and close with Escape.">
      <div className="flex flex-wrap gap-3">
        <Button variant="secondary" onClick={() => setIsModalOpen(true)}>
          Open modal
        </Button>
        <Button variant="secondary" onClick={() => setIsConfirmOpen(true)}>
          Open confirm
        </Button>
        <Button variant="danger" onClick={() => setIsDangerConfirmOpen(true)}>
          Open danger confirm
        </Button>
      </div>
      <div className="mt-4 flex flex-wrap gap-3">
        {toastExamples.map((toastExample) => (
          <Button key={toastExample.tone} variant="ghost" size="sm" onClick={() => showToast(toastExample)}>
            Toast: {toastExample.tone}
          </Button>
        ))}
      </div>

      <Modal
        isOpen={isModalOpen}
        onOpenChange={setIsModalOpen}
        title="Rename ship"
        description="Give your Scout a name other players can see."
        footer={
          <>
            <Button variant="secondary" onClick={() => setIsModalOpen(false)}>
              Cancel
            </Button>
            <Button onClick={() => setIsModalOpen(false)}>Save</Button>
          </>
        }
      >
        <Input label="Ship name" placeholder="Stardust Runner" />
      </Modal>

      <ConfirmDialog
        isOpen={isConfirmOpen}
        onOpenChange={setIsConfirmOpen}
        title="Confirm purchase"
        description="Check the amounts before paying."
        confirmLabel="Pay 4.00 NC"
        onConfirm={confirmPurchase}
      >
        <dl className="space-y-2 rounded-lg border border-border bg-background/60 p-4 font-mono text-sm tabular-nums">
          <div className="flex justify-between">
            <dt className="text-muted">20 × Fuel Cell</dt>
            <dd className="text-foreground">4.000000 NC</dd>
          </div>
          <div className="flex justify-between border-t border-border pt-2">
            <dt className="text-muted">Total</dt>
            <dd className="font-semibold text-foreground">4.000000 NC</dd>
          </div>
        </dl>
      </ConfirmDialog>

      <ConfirmDialog
        isOpen={isDangerConfirmOpen}
        onOpenChange={setIsDangerConfirmOpen}
        title="Abort mission?"
        description="Fuel is not refunded and you get no loot."
        confirmLabel="Abort mission"
        tone="danger"
        onConfirm={() => wait(800)}
      />
    </ContentSection>
  );
}
