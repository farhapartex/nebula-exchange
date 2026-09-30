"use client";

import { useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";

import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { useToast } from "@/components/ui/toast/use-toast";
import type { CatalogItem } from "@/features/catalog/api/catalog-types";
import {
  abortMission,
  collectMission,
  listMissions,
  missionsQueryKeys,
  type Mission,
} from "@/features/missions/api/missions-api";
import { ActiveMissionCard } from "@/features/missions/active-mission-card";
import { LootRevealModal } from "@/features/missions/loot-reveal-modal";
import { maximumActiveMissions } from "@/features/missions/mission-rules";
import { useRefreshAfterMissionChange } from "@/features/missions/use-mission-actions";

const activeRefreshIntervalInMilliseconds = 5_000;

export function useActiveMissions() {
  return useQuery({
    queryKey: missionsQueryKeys.active,
    queryFn: async () => (await listMissions(["RUNNING", "COMPLETED"], { limit: maximumActiveMissions })).data,
    refetchInterval: (query) =>
      query.state.data?.some((mission) => mission.status === "RUNNING") ? activeRefreshIntervalInMilliseconds : false,
  });
}

export function ActiveMissions({ missions, itemsByID }: { missions: Mission[]; itemsByID: Map<number, CatalogItem> }) {
  const { showToast } = useToast();
  const refreshAfterMissionChange = useRefreshAfterMissionChange();
  const [missionPendingAbort, setMissionPendingAbort] = useState<Mission | null>(null);
  const [revealedMission, setRevealedMission] = useState<Mission | null>(null);

  const collectMutation = useMutation({
    mutationFn: (mission: Mission) => collectMission(mission.id),
    onSuccess: (collectedMission) => {
      refreshAfterMissionChange();
      setRevealedMission(collectedMission);
    },
  });
  const abortMutation = useMutation({
    mutationFn: (mission: Mission) => abortMission(mission.id),
    onSuccess: (abortedMission) => {
      refreshAfterMissionChange();
      showToast({
        tone: "info",
        title: `${abortedMission.zone_name} mission aborted`,
        description: "Your ship and drill are free again.",
      });
    },
  });

  return (
    <>
      {missions.length === 0 && (
        <p className="rounded-2xl border border-dashed border-border-strong px-4 py-6 text-center text-sm text-muted">
          No ships out. Pick a zone below to launch one.
        </p>
      )}
      <ul className="space-y-3">
        {missions.map((mission) => (
          <ActiveMissionCard
            key={mission.id}
            mission={mission}
            itemsByID={itemsByID}
            isCollecting={collectMutation.isPending && collectMutation.variables?.id === mission.id}
            onCollect={(missionToCollect) => collectMutation.mutate(missionToCollect)}
            onAbort={setMissionPendingAbort}
          />
        ))}
      </ul>
      <ConfirmDialog
        isOpen={missionPendingAbort !== null}
        onOpenChange={(isOpen) => !isOpen && setMissionPendingAbort(null)}
        title="Abort this mission?"
        description={`The ${missionPendingAbort?.fuel_spent ?? 0} fuel you burned is not refunded and you get no loot.`}
        confirmLabel="Abort mission"
        tone="danger"
        onConfirm={async () => {
          if (missionPendingAbort) {
            await abortMutation.mutateAsync(missionPendingAbort);
          }
        }}
      />
      <LootRevealModal mission={revealedMission} itemsByID={itemsByID} onClose={() => setRevealedMission(null)} />
    </>
  );
}
