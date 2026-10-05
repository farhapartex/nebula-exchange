import { describe, expect, it } from "vitest";

import type { StageCamera } from "@/features/fight/api/fight-setup-api";
import {
  cameraGoal,
  createCameraRig,
  shakeCamera,
  shakeOffset,
  stepCameraRig,
} from "@/features/fight/stage/cinematic-camera";

const camera: StageCamera = {
  field_of_view: 38,
  height: 1.3,
  look_height: 1,
  close_distance: 5,
  far_distance: 8,
  follow_speed: 3,
  knockout_zoom: 0.7,
};

describe("cinematic camera", () => {
  it("frames the midpoint and pulls back as fighters separate", () => {
    const close = cameraGoal(-0.5, 0.5, camera, 1);
    const apart = cameraGoal(-3, 3, camera, 1);
    expect(close.target.x).toBe(0);
    expect(close.position.z).toBeLessThan(apart.position.z);
    expect(apart.position.z).toBe(8);
    expect(cameraGoal(-3, 3, camera, 0.7).position.z).toBeCloseTo(5.6);
  });

  it("eases toward the goal instead of jumping", () => {
    const rig = createCameraRig(camera);
    const goal = cameraGoal(2, 3, camera, 1);
    stepCameraRig(rig, goal, 16, camera.follow_speed);
    expect(rig.position.x).toBeGreaterThan(0);
    expect(rig.position.x).toBeLessThan(goal.position.x);
    for (let frame = 0; frame < 400; frame += 1) {
      stepCameraRig(rig, goal, 16, camera.follow_speed);
    }
    expect(rig.position.x).toBeCloseTo(goal.position.x, 2);
  });

  it("shakes and settles back to rest", () => {
    const rig = createCameraRig(camera);
    shakeCamera(rig, 0.1, 200);
    expect(Math.abs(shakeOffset(rig, 30).x) + Math.abs(shakeOffset(rig, 30).y)).toBeGreaterThan(0);
    stepCameraRig(rig, cameraGoal(0, 1, camera, 1), 250, camera.follow_speed);
    expect(shakeOffset(rig, 30)).toEqual({ x: 0, y: 0, z: 0 });
  });
});
