"use client";

import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { ArrowRight } from "lucide-react";

import { NcAmount } from "@/components/money/nc-amount";
import { Button } from "@/components/ui/button";
import { useToast } from "@/components/ui/toast/use-toast";
import type { CatalogItem, Upgrade } from "@/features/catalog/api/catalog-types";
import { ItemAttributeList } from "@/features/catalog/components/item-attributes";
import { ItemIcon } from "@/features/catalog/components/item-icon";
import { PaymentMethodPicker } from "@/features/payments/components/payment-method-picker";
import {
  describePaymentChoices,
  preferredPaymentChoice,
  type PaymentChoice,
} from "@/features/payments/payment-method-rules";
import { useCardCheckout } from "@/features/payments/use-card-checkout";
import { buyUpgrade, craftUpgrade } from "@/features/workshop/api/workshop-api";
import { RequirementList } from "@/features/workshop/requirement-list";
import { hasAllRequirements } from "@/features/workshop/requirement-rules";
import { useRefreshAfterWorkshopChange } from "@/features/workshop/use-refresh-after-workshop-change";
import { createIdempotencyKey } from "@/lib/api/api-client";

type UpgradeCardProps = {
  upgrade: Upgrade;
  itemsByID: Map<number, CatalogItem>;
  ownedByItem: Map<number, number>;
  availableNc: bigint | null;
};

export function UpgradeCard({ upgrade, itemsByID, ownedByItem, availableNc }: UpgradeCardProps) {
  const { showToast } = useToast();
  const refreshAfterWorkshopChange = useRefreshAfterWorkshopChange();
  const fromItem = itemsByID.get(upgrade.from_item_id);
  const toItem = itemsByID.get(upgrade.to_item_id);
  const ownedBaseItems = ownedByItem.get(upgrade.from_item_id) ?? 0;
  const hasBaseItem = ownedBaseItems > 0;
  const hasMaterials = hasAllRequirements(upgrade.inputs, ownedByItem);
  const hasCraftFee = availableNc === null || availableNc >= BigInt(upgrade.craft_fee);

  const buyChoices = upgrade.buy_price ? describePaymentChoices("shop", BigInt(upgrade.buy_price), availableNc) : [];
  const [chosenPayment, setChosenPayment] = useState<PaymentChoice | null>(null);
  const selectedPayment =
    chosenPayment && buyChoices.some((availability) => availability.choice === chosenPayment && availability.isAllowed)
      ? chosenPayment
      : preferredPaymentChoice(buyChoices);

  const onUpgraded = () => {
    refreshAfterWorkshopChange();
    showToast({ tone: "success", title: `Upgraded to ${toItem?.name ?? "the next tier"}` });
  };
  const craftMutation = useMutation({
    mutationFn: () => craftUpgrade(upgrade.id, createIdempotencyKey()),
    onSuccess: onUpgraded,
  });
  const buyMutation = useMutation({
    mutationFn: () => buyUpgrade(upgrade.id, createIdempotencyKey()),
    onSuccess: onUpgraded,
  });
  const { startCheckout, isStartingCheckout, checkoutError } = useCardCheckout();
  const upgradeError = craftMutation.error ?? buyMutation.error ?? checkoutError;

  return (
    <article className="flex flex-col gap-4 rounded-2xl border border-border bg-surface/80 p-5">
      <div className="flex items-center gap-3">
        {fromItem && <ItemIcon item={fromItem} />}
        <ArrowRight className="size-5 text-subtle" aria-hidden="true" />
        {toItem && <ItemIcon item={toItem} />}
        <div className="min-w-0 flex-1">
          <h3 className="text-base font-semibold text-foreground">
            {fromItem?.name} → {toItem?.name}
          </h3>
          <p className={hasBaseItem ? "text-xs text-muted" : "text-xs text-down"}>
            {hasBaseItem ? `${ownedBaseItems} free ${fromItem?.name} to upgrade` : `You need a free ${fromItem?.name}`}
          </p>
        </div>
      </div>
      {toItem && <ItemAttributeList item={toItem} />}

      <section className="rounded-xl border border-border bg-background/50 p-3">
        <p className="mb-2 flex items-center justify-between text-xs font-medium tracking-wide text-subtle uppercase">
          Craft path
          <span className="normal-case">
            fee <NcAmount amount={upgrade.craft_fee} className={hasCraftFee ? "text-foreground" : "text-down"} />
          </span>
        </p>
        <RequirementList requirements={upgrade.inputs} ownedByItem={ownedByItem} itemsByID={itemsByID} />
        <Button
          className="mt-3 w-full"
          variant="secondary"
          isLoading={craftMutation.isPending}
          disabled={!hasBaseItem || !hasMaterials || !hasCraftFee}
          onClick={() => craftMutation.mutate()}
        >
          {!hasMaterials ? "Not enough materials" : "Craft upgrade"}
        </Button>
      </section>

      <section className="rounded-xl border border-border bg-background/50 p-3">
        {upgrade.buy_price ? (
          <>
            <p className="mb-3 flex items-center justify-between text-xs font-medium tracking-wide text-subtle uppercase">
              Buy path
              <NcAmount amount={upgrade.buy_price} className="text-sm text-foreground normal-case" />
            </p>
            <PaymentMethodPicker
              isCompact
              label="Pay with"
              choices={buyChoices}
              value={selectedPayment}
              onValueChange={setChosenPayment}
              availableBalance={availableNc?.toString()}
            />
            <Button
              className="mt-3 w-full"
              isLoading={buyMutation.isPending || isStartingCheckout}
              disabled={!hasBaseItem || selectedPayment === null}
              onClick={() =>
                selectedPayment === "card"
                  ? startCheckout({ purpose: "UPGRADE_PURCHASE", upgrade_id: upgrade.id })
                  : buyMutation.mutate()
              }
            >
              {selectedPayment === "card" ? "Buy by card" : "Buy upgrade"}
            </Button>
          </>
        ) : (
          <p className="text-sm text-muted">This tier can only be crafted.</p>
        )}
      </section>

      {upgradeError && (
        <p role="alert" className="text-sm text-down">
          {upgradeError.message}
        </p>
      )}
    </article>
  );
}
