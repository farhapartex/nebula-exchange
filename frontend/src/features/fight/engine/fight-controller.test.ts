import { describe, expect, it } from "vitest";

import { FightController, type FightSnapshot } from "@/features/fight/engine/fight-controller";

const snapshot: FightSnapshot = {
  player: { health: 100, maxHealth: 100, stamina: 100, maxStamina: 100 },
  enemy: { health: 100, maxHealth: 100, stamina: 100, maxStamina: 100 },
  remainingSeconds: 60,
  outcome: null,
  isPaused: false,
  damageDealt: 0,
  damageTaken: 0,
};

describe("FightController", () => {
  it("turns held controls into movement and consumes taps once", () => {
    const controller = new FightController(snapshot);
    controller.press("right");
    controller.tap("punch");
    expect(controller.readPlayerIntent()).toEqual({ move: 1, attack: "punch", block: false, dodge: false });
    expect(controller.readPlayerIntent().attack).toBeNull();
    controller.press("left");
    expect(controller.readPlayerIntent().move).toBe(0);
  });

  it("pauses, notifies listeners and refuses to pause a finished fight", () => {
    const controller = new FightController(snapshot);
    let notifications = 0;
    controller.subscribe(() => (notifications += 1));
    controller.togglePause();
    expect(controller.isPaused).toBe(true);
    expect(notifications).toBe(1);
    controller.publish({ ...controller.getSnapshot(), isPaused: false, outcome: "won" });
    controller.togglePause();
    expect(controller.isPaused).toBe(false);
  });
});
