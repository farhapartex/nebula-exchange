"use client";

import { useEffect } from "react";

import type { FightController, HeldControl, TappedControl } from "@/features/fight/engine/fight-controller";

const heldControlByKey: Record<string, HeldControl> = {
  a: "left",
  arrowleft: "left",
  d: "right",
  arrowright: "right",
  l: "block",
  s: "block",
  arrowdown: "block",
};

const tappedControlByKey: Record<string, TappedControl> = {
  j: "punch",
  k: "kick",
  " ": "dodge",
  shift: "dodge",
};

export function useFightKeyboard(controller: FightController): void {
  useEffect(() => {
    function handleKeyDown(keyEvent: KeyboardEvent) {
      const key = keyEvent.key.toLowerCase();
      if (key === "escape" || key === "p") {
        controller.togglePause();
        return;
      }
      const heldControl = heldControlByKey[key];
      const tappedControl = tappedControlByKey[key];
      if (heldControl || tappedControl) {
        keyEvent.preventDefault();
      }
      if (heldControl) {
        controller.press(heldControl);
      } else if (tappedControl && !keyEvent.repeat) {
        controller.tap(tappedControl);
      }
    }
    function handleKeyUp(keyEvent: KeyboardEvent) {
      const heldControl = heldControlByKey[keyEvent.key.toLowerCase()];
      if (heldControl) {
        controller.release(heldControl);
      }
    }
    function handleBlur() {
      controller.releaseEverything();
    }
    window.addEventListener("keydown", handleKeyDown);
    window.addEventListener("keyup", handleKeyUp);
    window.addEventListener("blur", handleBlur);
    return () => {
      window.removeEventListener("keydown", handleKeyDown);
      window.removeEventListener("keyup", handleKeyUp);
      window.removeEventListener("blur", handleBlur);
    };
  }, [controller]);
}
