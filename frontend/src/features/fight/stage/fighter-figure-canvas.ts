import type { FighterLook, FighterStats } from "@/features/fight/api/fight-setup-api";
import { attackPhaseOf } from "@/features/fight/engine/fight-simulation";
import type { FighterRuntime } from "@/features/fight/engine/fight-types";

export type FigurePose = {
  fighter: FighterRuntime;
  stats: FighterStats;
  look: FighterLook;
  originX: number;
  floorY: number;
  pixelsPerUnit: number;
  timeMs: number;
  isFlashing: boolean;
  showsTelegraph: boolean;
};

function attackReach(fighter: FighterRuntime, stats: FighterStats): number {
  const phase = attackPhaseOf(fighter, stats);
  if (!fighter.attack || !phase) {
    return 0;
  }
  return phase === "windup" ? -0.25 : phase === "active" ? 1 : 0.35;
}

export function drawFighterFigure(context: CanvasRenderingContext2D, pose: FigurePose): void {
  const { fighter, stats, look, originX, floorY, timeMs } = pose;
  const scale = pose.pixelsPerUnit;
  const facing = fighter.facing;
  const height = look.height * scale;
  const width = look.width * scale;
  const bodyColor = pose.isFlashing ? "#ffffff" : look.body_color;
  const lineWidth = Math.max(4, width * 0.16);
  context.lineCap = "round";
  context.lineJoin = "round";

  if (fighter.action === "knocked_out") {
    context.fillStyle = look.body_color;
    context.beginPath();
    context.roundRect(originX - height * 0.5, floorY - width * 0.45, height, width * 0.45, 10 * scale);
    context.fill();
    context.beginPath();
    context.arc(originX - facing * height * 0.55, floorY - width * 0.25, width * 0.28, 0, Math.PI * 2);
    context.fill();
    return;
  }

  const lean = fighter.action === "dodging" ? -facing * 0.35 : fighter.action === "stunned" ? -facing * 0.18 : 0;
  const hipY = floorY - height * 0.46;
  const shoulderY = floorY - height * 0.8;
  const shoulderX = originX + lean * height * 0.4;
  const headRadius = width * 0.3;
  const headX = shoulderX + facing * width * 0.06;
  const headY = shoulderY - headRadius * 1.25;
  const stride = fighter.action === "walking" ? Math.sin(timeMs / 90) * width * 0.35 : 0;
  const reach = attackReach(fighter, stats);
  context.globalAlpha = fighter.action === "dodging" ? 0.55 : 1;

  const drawLine = (fromX: number, fromY: number, toX: number, toY: number) => {
    context.beginPath();
    context.moveTo(fromX, fromY);
    context.lineTo(toX, toY);
    context.stroke();
  };

  if (pose.showsTelegraph && reach < 0) {
    context.strokeStyle = "rgba(244,63,94,0.55)";
    context.lineWidth = lineWidth + 14;
    context.beginPath();
    context.arc(headX, headY, headRadius + 8, 0, Math.PI * 2);
    context.stroke();
    drawLine(shoulderX, shoulderY, originX, hipY);
  }

  context.strokeStyle = bodyColor;
  context.lineWidth = lineWidth;
  const kickExtension = fighter.attack === "kick" && reach > 0 ? reach : 0;
  const frontFootX = originX + facing * (width * 0.35 + stride + kickExtension * stats.kick.range * scale * 0.85);
  const frontFootY = kickExtension > 0 ? hipY - height * 0.05 : floorY;
  drawLine(originX, hipY, frontFootX, frontFootY);
  drawLine(originX, hipY, originX - facing * (width * 0.35 + stride), floorY);
  drawLine(shoulderX, shoulderY, originX, hipY);

  context.fillStyle = bodyColor;
  context.beginPath();
  context.arc(headX, headY, headRadius, 0, Math.PI * 2);
  context.fill();
  context.strokeStyle = look.outline_color;
  context.lineWidth = 2 * scale;
  context.stroke();

  context.strokeStyle = bodyColor;
  context.lineWidth = lineWidth * 0.85;
  if (fighter.action === "blocking") {
    const guardX = shoulderX + facing * width * 0.55;
    drawLine(shoulderX, shoulderY, guardX, shoulderY - height * 0.12);
    drawLine(shoulderX, shoulderY + height * 0.06, guardX, shoulderY - height * 0.02);
    context.strokeStyle = look.outline_color;
    context.lineWidth = 3 * scale;
    context.beginPath();
    context.roundRect(guardX - 6 * scale, shoulderY - height * 0.2, 12 * scale, height * 0.26, 4 * scale);
    context.stroke();
    context.globalAlpha = 1;
    return;
  }

  const punchExtension = fighter.attack === "punch" ? reach : 0;
  const frontHandX = shoulderX + facing * (width * 0.45 + punchExtension * stats.punch.range * scale * 0.85);
  const frontHandY = shoulderY + (punchExtension > 0 ? 0 : height * 0.06);
  drawLine(shoulderX, shoulderY, frontHandX, frontHandY);
  drawLine(shoulderX, shoulderY, shoulderX + facing * width * 0.2, shoulderY + height * 0.16);
  context.fillStyle = look.outline_color;
  context.beginPath();
  context.arc(frontHandX, frontHandY, lineWidth * 0.7, 0, Math.PI * 2);
  context.fill();
  context.globalAlpha = 1;
}
