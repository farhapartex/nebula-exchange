"use client";

import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { Clock, Fuel, PackageOpen } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Modal } from "@/components/ui/modal";
import { Select } from "@/components/ui/select";
import { useToast } from "@/components/ui/toast/use-toast";
import type { CatalogItem, Zone } from "@/features/catalog/api/catalog-types";
import type { InventoryHolding } from "@/features/inventory/api/inventory-api";
import { startMission } from "@/features/missions/api/missions-api";
import { LootRangeList } from "@/features/missions/loot-range-list";
import {
  drillProfileOf,
  expectedLootRanges,
  fuelCellItemID,
  maximumLootUnits,
  missionDurationSeconds,
  shipProfileOf,
  zoneAllowsShip,
} from "@/features/missions/mission-rules";
import { useRefreshAfterMissionChange } from "@/features/missions/use-mission-actions";
import { createIdempotencyKey } from "@/lib/api/api-client";
import { cn } from "@/utils/class-names";
import { formatDurationShort } from "@/utils/time/format-duration";

type MissionLauncherProps = {
  zone: Zone | null;
  holdings: InventoryHolding[];
  itemsByID: Map<number, CatalogItem>;
  onClose: () => void;
};

export function MissionLauncher({ zone, holdings, itemsByID, onClose }: MissionLauncherProps) {
  return (
    <Modal
      isOpen={zone !== null}
      onOpenChange={(isOpen) => !isOpen && onClose()}
      title={zone ? `Launch to ${zone.name}` : "Launch mission"}
      description="Your ship and drill stay locked until you collect the loot."
    >
      {zone && <LauncherForm key={zone.id} zone={zone} holdings={holdings} itemsByID={itemsByID} onClose={onClose} />}
    </Modal>
  );
}

function LauncherForm({
  zone,
  holdings,
  itemsByID,
  onClose,
}: {
  zone: Zone;
  holdings: InventoryHolding[];
  itemsByID: Map<number, CatalogItem>;
  onClose: () => void;
}) {
  const { showToast } = useToast();
  const refreshAfterMissionChange = useRefreshAfterMissionChange();
  const [idempotencyKey] = useState(createIdempotencyKey);

  const availableItems = holdings
    .filter((holding) => holding.available > 0)
    .flatMap((holding) => {
      const catalogItem = itemsByID.get(holding.item_id);
      return catalogItem ? [{ catalogItem, available: holding.available }] : [];
    });
  const usableShips = availableItems.filter(
    ({ catalogItem }) => shipProfileOf(catalogItem) && zoneAllowsShip(zone, catalogItem.id),
  );
  const usableDrills = availableItems
    .filter(({ catalogItem }) => (drillProfileOf(catalogItem)?.tier ?? 0) >= zone.minimum_drill_tier)
    .sort(
      (first, second) =>
        (drillProfileOf(second.catalogItem)?.yieldBasisPoints ?? 0) -
        (drillProfileOf(first.catalogItem)?.yieldBasisPoints ?? 0),
    );
  const ownedFuel = holdings.find((holding) => holding.item_id === fuelCellItemID)?.available ?? 0;

  const [shipItemID, setShipItemID] = useState(usableShips[0]?.catalogItem.id ?? 0);
  const [drillItemID, setDrillItemID] = useState(usableDrills[0]?.catalogItem.id ?? 0);
  const selectedShip = itemsByID.get(shipItemID);
  const selectedDrill = itemsByID.get(drillItemID);
  const shipProfile = selectedShip ? shipProfileOf(selectedShip) : null;
  const drillProfile = selectedDrill ? drillProfileOf(selectedDrill) : null;
  const lootRanges = expectedLootRanges(zone.loot, drillProfile);
  const isCargoLimited = shipProfile !== null && maximumLootUnits(lootRanges) > shipProfile.cargoCapacity;
  const hasEnoughFuel = ownedFuel >= zone.fuel_cost;

  const launchMutation = useMutation({
    mutationFn: () =>
      startMission({ zone_id: zone.id, ship_item_id: shipItemID, drill_item_id: drillItemID }, idempotencyKey),
    onSuccess: (mission) => {
      refreshAfterMissionChange();
      showToast({
        tone: "success",
        title: `${mission.zone_name} mission launched`,
        description: "Collect the loot when the timer ends.",
      });
      onClose();
    },
  });

  return (
    <div className="space-y-5">
      <div className="grid gap-4 sm:grid-cols-2">
        <Select
          label="Ship"
          value={shipItemID ? String(shipItemID) : undefined}
          onValueChange={(selectedValue) => setShipItemID(Number(selectedValue))}
          options={usableShips.map(({ catalogItem, available }) => ({
            value: String(catalogItem.id),
            label: `${catalogItem.name} (${available} free)`,
          }))}
          placeholder="No ship can fly here"
          isDisabled={usableShips.length === 0}
        />
        <Select
          label="Drill"
          value={drillItemID ? String(drillItemID) : undefined}
          onValueChange={(selectedValue) => setDrillItemID(Number(selectedValue))}
          options={usableDrills.map(({ catalogItem, available }) => ({
            value: String(catalogItem.id),
            label: `${catalogItem.name} (${available} free)`,
          }))}
          placeholder={`Needs a Tier ${zone.minimum_drill_tier} drill`}
          isDisabled={usableDrills.length === 0}
        />
      </div>

      <dl className="grid grid-cols-2 gap-3 text-sm">
        <div className="rounded-xl border border-border bg-background/50 p-3">
          <dt className="flex items-center gap-1.5 text-xs text-muted">
            <Clock className="size-3.5" aria-hidden="true" /> Flight time
          </dt>
          <dd className="mt-1 font-medium text-foreground">
            {shipProfile ? formatDurationShort(missionDurationSeconds(zone, shipProfile)) : "—"}
          </dd>
        </div>
        <div
          className={cn(
            "rounded-xl border p-3",
            hasEnoughFuel ? "border-border bg-background/50" : "border-down/50 bg-down-soft/30",
          )}
        >
          <dt className="flex items-center gap-1.5 text-xs text-muted">
            <Fuel className="size-3.5" aria-hidden="true" /> Fuel burned
          </dt>
          <dd className="mt-1 font-medium text-foreground">
            {zone.fuel_cost} <span className="text-xs text-muted">of {ownedFuel} owned</span>
          </dd>
        </div>
      </dl>

      <div className="rounded-xl border border-border bg-background/50 p-3">
        <p className="mb-2 flex items-center gap-1.5 text-xs font-medium tracking-wide text-subtle uppercase">
          <PackageOpen className="size-3.5" aria-hidden="true" /> Expected loot
        </p>
        <LootRangeList ranges={lootRanges} itemsByID={itemsByID} />
        {isCargoLimited && shipProfile && (
          <p className="mt-2 text-xs text-warning">
            Your {selectedShip?.name} holds {shipProfile.cargoCapacity} units, so big rolls are trimmed to fit.
          </p>
        )}
      </div>

      {launchMutation.error && (
        <p role="alert" className="rounded-lg border border-down/40 bg-down-soft/40 px-3 py-2 text-sm text-foreground">
          {launchMutation.error.message}
        </p>
      )}

      <div className="flex justify-end gap-2">
        <Button variant="ghost" onClick={onClose}>
          Cancel
        </Button>
        <Button
          isLoading={launchMutation.isPending}
          disabled={!shipItemID || !drillItemID || !hasEnoughFuel}
          onClick={() => launchMutation.mutate()}
        >
          Launch · burn {zone.fuel_cost} fuel
        </Button>
      </div>
    </div>
  );
}
