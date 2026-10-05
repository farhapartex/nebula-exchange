import type { StageCamera } from "@/features/fight/api/fight-setup-api";

export type Vector = { x: number; y: number; z: number };

export type CameraRig = {
  position: Vector;
  target: Vector;
  shakeRemainingMs: number;
  shakeDurationMs: number;
  shakeStrength: number;
  zoomFactor: number;
};

const maximumFramingSeparation = 6;

export function createCameraRig(camera: StageCamera): CameraRig {
  return {
    position: { x: 0, y: camera.height, z: camera.far_distance },
    target: { x: 0, y: camera.look_height, z: 0 },
    shakeRemainingMs: 0,
    shakeDurationMs: 1,
    shakeStrength: 0,
    zoomFactor: 1,
  };
}

export function cameraGoal(
  playerWorldX: number,
  enemyWorldX: number,
  camera: StageCamera,
  zoomFactor: number,
): { position: Vector; target: Vector } {
  const midpointX = (playerWorldX + enemyWorldX) / 2;
  const separationShare = Math.min(Math.abs(enemyWorldX - playerWorldX) / maximumFramingSeparation, 1);
  const distance =
    (camera.close_distance + (camera.far_distance - camera.close_distance) * separationShare) * zoomFactor;
  return {
    position: { x: midpointX, y: camera.height, z: distance },
    target: { x: midpointX, y: camera.look_height, z: 0 },
  };
}

function approach(current: number, goal: number, smoothing: number): number {
  return current + (goal - current) * smoothing;
}

export function stepCameraRig(
  rig: CameraRig,
  goal: { position: Vector; target: Vector },
  deltaMs: number,
  followSpeed: number,
): void {
  const smoothing = 1 - Math.exp((-followSpeed * deltaMs) / 1000);
  for (const axis of ["x", "y", "z"] as const) {
    rig.position[axis] = approach(rig.position[axis], goal.position[axis], smoothing);
    rig.target[axis] = approach(rig.target[axis], goal.target[axis], smoothing);
  }
  rig.shakeRemainingMs = Math.max(0, rig.shakeRemainingMs - deltaMs);
}

export function shakeCamera(rig: CameraRig, strength: number, durationMs: number): void {
  if (strength >= rig.shakeStrength * (rig.shakeRemainingMs / rig.shakeDurationMs)) {
    rig.shakeStrength = strength;
    rig.shakeDurationMs = durationMs;
    rig.shakeRemainingMs = durationMs;
  }
}

export function shakeOffset(rig: CameraRig, timeMs: number): Vector {
  if (rig.shakeRemainingMs <= 0) {
    return { x: 0, y: 0, z: 0 };
  }
  const fade = rig.shakeRemainingMs / rig.shakeDurationMs;
  const amplitude = rig.shakeStrength * fade;
  return {
    x: Math.sin(timeMs * 0.09) * amplitude,
    y: Math.cos(timeMs * 0.13) * amplitude * 0.7,
    z: 0,
  };
}
