import * as THREE from "three";

import type { FireLight } from "@/features/fight/api/fight-setup-api";
import { safeHexColor } from "@/utils/colors/safe-hex-color";

type FlickeringLight = { light: THREE.PointLight; baseIntensity: number; flicker: number; seed: number };

export function flickerLevel(timeSeconds: number, seed: number): number {
  return (
    (Math.sin(timeSeconds * 7.3 + seed) +
      Math.sin(timeSeconds * 13.1 + seed * 2.1) * 0.5 +
      Math.sin(timeSeconds * 23.7 + seed * 3.7) * 0.25) /
    1.75
  );
}

export class FireLights {
  readonly group = new THREE.Group();
  private readonly lights: FlickeringLight[];

  constructor(fireLights: FireLight[]) {
    this.lights = fireLights.map((fireLight, lightIndex) => {
      const light = new THREE.PointLight(
        safeHexColor(fireLight.color, "#ff7a1a"),
        fireLight.intensity,
        fireLight.distance,
        1.4,
      );
      light.position.set(...fireLight.position);
      this.group.add(light);
      return { light, baseIntensity: fireLight.intensity, flicker: fireLight.flicker, seed: lightIndex * 11.7 };
    });
  }

  update(timeSeconds: number): void {
    for (const flickeringLight of this.lights) {
      flickeringLight.light.intensity =
        flickeringLight.baseIntensity * (1 + flickerLevel(timeSeconds, flickeringLight.seed) * flickeringLight.flicker);
    }
  }
}
