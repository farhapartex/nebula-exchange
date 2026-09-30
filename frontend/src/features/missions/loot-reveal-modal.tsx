"use client";

import { Sparkles } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Modal } from "@/components/ui/modal";
import type { CatalogItem } from "@/features/catalog/api/catalog-types";
import { ItemIcon } from "@/features/catalog/components/item-icon";
import type { Mission } from "@/features/missions/api/missions-api";

type LootRevealModalProps = {
  mission: Mission | null;
  itemsByID: Map<number, CatalogItem>;
  onClose: () => void;
};

export function LootRevealModal({ mission, itemsByID, onClose }: LootRevealModalProps) {
  return (
    <Modal
      isOpen={mission !== null}
      onOpenChange={(isOpen) => !isOpen && onClose()}
      title={mission ? `${mission.zone_name} haul` : "Loot"}
      description="Everything below is now in your inventory."
      footer={<Button onClick={onClose}>Nice</Button>}
    >
      <ul className="grid gap-3 sm:grid-cols-2">
        {mission?.loot.map((roll, rollIndex) => {
          const lootItem = itemsByID.get(roll.item_id);
          return (
            <li
              key={roll.item_id}
              style={{ animationDelay: `${rollIndex * 120}ms` }}
              className="flex animate-pop-in items-center gap-3 rounded-xl border border-border bg-background/60 p-3 [animation-fill-mode:both]"
            >
              {lootItem && <ItemIcon item={lootItem} />}
              <div>
                <p className="font-mono text-lg font-semibold text-up tabular-nums">+{roll.quantity}</p>
                <p className="text-sm text-foreground">{lootItem?.name ?? `Item ${roll.item_id}`}</p>
              </div>
              {lootItem && lootItem.rarity_rank >= 5 && (
                <Sparkles className="ml-auto size-4 text-warning" aria-label="Rare" />
              )}
            </li>
          );
        })}
      </ul>
    </Modal>
  );
}
