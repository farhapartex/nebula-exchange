"use client";

import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { CircleAlert } from "lucide-react";

import { NcAmount } from "@/components/money/nc-amount";
import { Button } from "@/components/ui/button";
import { Modal } from "@/components/ui/modal";
import { QuantityStepper } from "@/components/ui/quantity-stepper";
import { useToast } from "@/components/ui/toast/use-toast";
import { balancesQueryKey } from "@/features/balances/api/balances-api";
import { useBalances } from "@/features/balances/use-balances";
import type { CatalogItem, ShopItem } from "@/features/catalog/api/catalog-types";
import { inventoryQueryKey } from "@/features/inventory/api/inventory-api";
import { PaymentMethodPicker } from "@/features/payments/components/payment-method-picker";
import {
  describePaymentChoices,
  preferredPaymentChoice,
  type PaymentChoice,
} from "@/features/payments/payment-method-rules";
import { useCardCheckout } from "@/features/payments/use-card-checkout";
import { purchaseWithBalance } from "@/features/shop/api/shop-api";
import { ShopItemContents } from "@/features/shop/shop-item-contents";
import { createIdempotencyKey } from "@/lib/api/api-client";
import { isApiError } from "@/lib/api/api-error";

const maximumPurchaseQuantity = 100;

type PurchaseModalProps = {
  shopItem: ShopItem | null;
  itemsByID: Map<number, CatalogItem>;
  onClose: () => void;
};

export function PurchaseModal({ shopItem, itemsByID, onClose }: PurchaseModalProps) {
  return (
    <Modal
      isOpen={shopItem !== null}
      onOpenChange={(isOpen) => !isOpen && onClose()}
      title={shopItem ? `Buy ${shopItem.name}` : "Buy"}
      description={shopItem?.description}
    >
      {shopItem && <PurchaseForm key={shopItem.sku} shopItem={shopItem} itemsByID={itemsByID} onClose={onClose} />}
    </Modal>
  );
}

function PurchaseForm({
  shopItem,
  itemsByID,
  onClose,
}: {
  shopItem: ShopItem;
  itemsByID: Map<number, CatalogItem>;
  onClose: () => void;
}) {
  const queryClient = useQueryClient();
  const { showToast } = useToast();
  const balancesQuery = useBalances();
  const [quantity, setQuantity] = useState(1);
  const [idempotencyKey, setIdempotencyKey] = useState(createIdempotencyKey);
  const [chosenPayment, setChosenPayment] = useState<PaymentChoice | null>(null);

  const total = BigInt(shopItem.price) * BigInt(quantity);
  const availableBalance = balancesQuery.data ? BigInt(balancesQuery.data.available) : null;
  const choices = describePaymentChoices("shop", total, availableBalance);
  const selectedPayment =
    chosenPayment && choices.some((availability) => availability.choice === chosenPayment && availability.isAllowed)
      ? chosenPayment
      : preferredPaymentChoice(choices);
  const isShortOfNc = availableBalance !== null && availableBalance < total;

  const { startCheckout, isStartingCheckout, checkoutError } = useCardCheckout();
  const balanceMutation = useMutation({
    mutationFn: () => purchaseWithBalance(shopItem.sku, quantity, idempotencyKey),
    onSuccess: (completedPurchase) => {
      void queryClient.invalidateQueries({ queryKey: balancesQueryKey });
      void queryClient.invalidateQueries({ queryKey: inventoryQueryKey });
      showToast({
        tone: "success",
        title: `Bought ${completedPurchase.quantity} × ${shopItem.name}`,
        description: "The items are in your inventory.",
      });
      onClose();
    },
    onError: () => setIdempotencyKey(createIdempotencyKey()),
  });

  const purchaseError = balanceMutation.error ?? checkoutError;
  const purchaseErrorMessage = purchaseError
    ? isApiError(purchaseError) && purchaseError.code === "INSUFFICIENT_FUNDS"
      ? "Your NC balance changed and no longer covers this purchase."
      : purchaseError.message
    : null;

  function confirmPurchase() {
    if (selectedPayment === "balance") {
      balanceMutation.mutate();
    } else if (selectedPayment === "card") {
      startCheckout({ purpose: "SHOP_PURCHASE", sku: shopItem.sku, quantity });
    }
  }

  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <QuantityStepper
          label="Quantity"
          value={quantity}
          maximum={maximumPurchaseQuantity}
          onValueChange={setQuantity}
        />
        <div className="text-right">
          <p className="text-xs text-muted">Total</p>
          <NcAmount amount={total} className="text-xl font-semibold text-foreground" />
        </div>
      </div>

      <div className="rounded-xl border border-border bg-background/50 p-3">
        <p className="mb-2 text-xs font-medium tracking-wide text-subtle uppercase">You receive</p>
        <ShopItemContents contents={shopItem.contents} itemsByID={itemsByID} multiplier={quantity} />
      </div>

      <PaymentMethodPicker
        choices={choices}
        value={selectedPayment}
        onValueChange={setChosenPayment}
        availableBalance={balancesQuery.data?.available}
      />

      {isShortOfNc && (
        <p className="flex items-start gap-2 text-xs text-muted">
          <CircleAlert className="mt-0.5 size-3.5 shrink-0 text-warning" aria-hidden="true" />
          Your balance doesn&apos;t cover this. Pay by card or lower the quantity.
        </p>
      )}
      {purchaseErrorMessage && (
        <p role="alert" className="rounded-lg border border-down/40 bg-down-soft/40 px-3 py-2 text-sm text-foreground">
          {purchaseErrorMessage}
        </p>
      )}

      <div className="flex justify-end gap-2">
        <Button variant="ghost" onClick={onClose}>
          Cancel
        </Button>
        <Button
          isLoading={balanceMutation.isPending || isStartingCheckout}
          disabled={selectedPayment === null}
          onClick={confirmPurchase}
        >
          {selectedPayment === "card" ? "Continue to card" : "Buy now"}
        </Button>
      </div>
    </div>
  );
}
