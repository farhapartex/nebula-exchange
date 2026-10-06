"use client";

import { useEffect, useRef, useState, useSyncExternalStore } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";

import { Spinner } from "@/components/ui/spinner";
import { currentPlayerQueryKey } from "@/features/auth/session/use-current-player";
import { submitFightResult } from "@/features/fight/api/fight-session-api";
import type { FightSetup } from "@/features/fight/api/fight-setup-api";
import { FightController } from "@/features/fight/engine/fight-controller";
import { createFightState } from "@/features/fight/engine/fight-simulation";
import { buildSnapshot } from "@/features/fight/engine/fight-snapshot";
import { ControlsHint, PauseOverlay, ResultOverlay } from "@/features/fight/fight-overlays";
import { FightHud } from "@/features/fight/fight-hud";
import { TouchControls } from "@/features/fight/touch-controls";
import { StageGradeOverlay } from "@/features/fight/stage-grade-overlay";
import { buildFightResultReport } from "@/features/fight/fight-result-report";
import { useFightKeyboard } from "@/features/fight/use-fight-keyboard";
import { publishToastEvent } from "@/lib/notifications/toast-events";

function createController(setup: FightSetup): FightController {
  const rules = { player: setup.player.stats, enemy: setup.enemy.stats };
  return new FightController(
    buildSnapshot(createFightState(rules, setup.arena.width, setup.time_limit_seconds), rules, false, 0, 0),
  );
}

type FightArenaProps = {
  setup: FightSetup;
  fightSessionID: string | null;
  onRestartFight?: () => Promise<boolean>;
};

export function FightArena({ setup, fightSessionID, onRestartFight }: FightArenaProps) {
  const canvasHostRef = useRef<HTMLDivElement>(null);
  const [controller] = useState(() => createController(setup));
  const [isEngineReady, setIsEngineReady] = useState(false);
  const [isRestarting, setIsRestarting] = useState(false);
  const reportedFightSessionIDRef = useRef<string | null>(null);
  const snapshot = useSyncExternalStore(controller.subscribe, controller.getSnapshot, controller.getSnapshot);
  const queryClient = useQueryClient();
  useFightKeyboard(controller);

  const submitResultMutation = useMutation({
    mutationFn: ({
      sessionID,
      report,
    }: {
      sessionID: string;
      report: NonNullable<ReturnType<typeof buildFightResultReport>>;
    }) => submitFightResult(sessionID, report),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: currentPlayerQueryKey }),
    onError: () =>
      publishToastEvent({
        tone: "error",
        title: "Your result was not saved",
        description: "This fight will not count. Try the level again.",
      }),
  });
  const { mutate: reportResult } = submitResultMutation;

  useEffect(() => {
    const report = buildFightResultReport(snapshot);
    if (!report || !fightSessionID || reportedFightSessionIDRef.current === fightSessionID) {
      return;
    }
    reportedFightSessionIDRef.current = fightSessionID;
    reportResult({ sessionID: fightSessionID, report });
  }, [fightSessionID, reportResult, snapshot]);

  async function restartFight() {
    if (!onRestartFight) {
      controller.restart();
      return;
    }
    setIsRestarting(true);
    const hasNewFight = await onRestartFight();
    setIsRestarting(false);
    if (hasNewFight) {
      controller.restart();
    }
  }

  useEffect(() => {
    let stage: { destroy: () => void } | null = null;
    let isCancelled = false;
    void import("@/features/fight/stage/fight-stage").then(async ({ FightStage }) => {
      if (isCancelled || !canvasHostRef.current) {
        return;
      }
      const fightStage = new FightStage({ host: canvasHostRef.current, setup, controller, random: Math.random });
      stage = fightStage;
      await fightStage.start();
      if (!isCancelled) {
        setIsEngineReady(true);
      }
    });
    return () => {
      isCancelled = true;
      stage?.destroy();
    };
  }, [controller, setup]);

  return (
    <div className="flex min-h-svh flex-col items-center justify-center gap-4 bg-background p-2 sm:p-6">
      <div className="relative aspect-video w-full max-w-6xl overflow-hidden rounded-xl border border-border bg-black shadow-2xl shadow-black">
        <div ref={canvasHostRef} className="absolute inset-0" />
        <StageGradeOverlay grade={setup.arena.stage.grade} />
        {!isEngineReady && (
          <div className="absolute inset-0 flex items-center justify-center">
            <Spinner />
          </div>
        )}
        <FightHud
          player={setup.player}
          enemy={setup.enemy}
          snapshot={snapshot}
          onPause={() => controller.togglePause()}
        />
        <TouchControls controller={controller} />
        {snapshot.isPaused && !snapshot.outcome && (
          <PauseOverlay onResume={() => controller.togglePause()} onRestart={() => void restartFight()} />
        )}
        {snapshot.outcome && (
          <ResultOverlay snapshot={snapshot} onRetry={() => void restartFight()} isRetrying={isRestarting} />
        )}
      </div>
      <ControlsHint />
    </div>
  );
}
