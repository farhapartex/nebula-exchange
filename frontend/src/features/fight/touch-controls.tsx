"use client";

import type { ReactNode } from "react";
import { ChevronLeft, ChevronRight, Footprints, HandFist, Shield, Zap } from "lucide-react";

import type { FightController, HeldControl, TappedControl } from "@/features/fight/engine/fight-controller";
import { cn } from "@/utils/class-names";

const controlButtonClassName =
  "pointer-events-auto flex select-none items-center justify-center rounded-full border border-white/15 bg-black/45 text-foreground backdrop-blur active:scale-95 active:bg-accent/40 touch-none";

function HoldButton({
  controller,
  control,
  label,
  children,
  className,
}: {
  controller: FightController;
  control: HeldControl;
  label: string;
  children: ReactNode;
  className?: string;
}) {
  return (
    <button
      type="button"
      aria-label={label}
      className={cn(controlButtonClassName, className)}
      onPointerDown={() => controller.press(control)}
      onPointerUp={() => controller.release(control)}
      onPointerCancel={() => controller.release(control)}
      onPointerLeave={() => controller.release(control)}
    >
      {children}
    </button>
  );
}

function TapButton({
  controller,
  control,
  label,
  children,
  className,
}: {
  controller: FightController;
  control: TappedControl;
  label: string;
  children: ReactNode;
  className?: string;
}) {
  return (
    <button
      type="button"
      aria-label={label}
      className={cn(controlButtonClassName, className)}
      onPointerDown={() => controller.tap(control)}
    >
      {children}
    </button>
  );
}

export function TouchControls({ controller }: { controller: FightController }) {
  return (
    <div className="pointer-events-none absolute inset-x-0 bottom-0 z-10 hidden items-end justify-between p-4 [@media(pointer:coarse)]:flex">
      <div className="flex gap-3">
        <HoldButton controller={controller} control="left" label="Move left" className="size-16">
          <ChevronLeft className="size-8" />
        </HoldButton>
        <HoldButton controller={controller} control="right" label="Move right" className="size-16">
          <ChevronRight className="size-8" />
        </HoldButton>
      </div>
      <div className="grid grid-cols-2 gap-3">
        <HoldButton controller={controller} control="block" label="Block" className="size-14">
          <Shield className="size-6" />
        </HoldButton>
        <TapButton controller={controller} control="dodge" label="Dodge" className="size-14">
          <Footprints className="size-6" />
        </TapButton>
        <TapButton controller={controller} control="punch" label="Punch" className="size-16 bg-accent/50">
          <HandFist className="size-7" />
        </TapButton>
        <TapButton controller={controller} control="kick" label="Kick" className="size-16 bg-accent/50">
          <Zap className="size-7" />
        </TapButton>
      </div>
    </div>
  );
}
