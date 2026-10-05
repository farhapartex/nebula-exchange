import * as THREE from "three";

import { safeHexColor } from "@/utils/colors/safe-hex-color";

const volume = { minX: -7, maxX: 7, minY: 0, maxY: 4.5, minZ: -6, maxZ: 3 };

export class EmberField {
  readonly points: THREE.Points;
  private readonly velocities: Float32Array;

  constructor(count: number, colors: string[], spriteTexture: THREE.Texture) {
    const positions = new Float32Array(count * 3);
    const vertexColors = new Float32Array(count * 3);
    this.velocities = new Float32Array(count * 3);
    const palette = (colors.length > 0 ? colors : ["#f97316"]).map(
      (color) => new THREE.Color(safeHexColor(color, "#f97316")),
    );
    for (let emberIndex = 0; emberIndex < count; emberIndex += 1) {
      this.respawn(positions, emberIndex, true);
      const color = palette[emberIndex % palette.length];
      vertexColors.set([color.r, color.g, color.b], emberIndex * 3);
    }
    const geometry = new THREE.BufferGeometry();
    geometry.setAttribute("position", new THREE.BufferAttribute(positions, 3));
    geometry.setAttribute("color", new THREE.BufferAttribute(vertexColors, 3));
    this.points = new THREE.Points(
      geometry,
      new THREE.PointsMaterial({
        size: 0.06,
        map: spriteTexture,
        vertexColors: true,
        transparent: true,
        depthWrite: false,
        blending: THREE.AdditiveBlending,
      }),
    );
  }

  private respawn(positions: Float32Array, emberIndex: number, isAnywhere: boolean): void {
    const offset = emberIndex * 3;
    positions[offset] = volume.minX + Math.random() * (volume.maxX - volume.minX);
    positions[offset + 1] = isAnywhere ? Math.random() * volume.maxY : volume.minY;
    positions[offset + 2] = volume.minZ + Math.random() * (volume.maxZ - volume.minZ);
    this.velocities[offset] = (Math.random() - 0.4) * 0.25;
    this.velocities[offset + 1] = 0.35 + Math.random() * 0.6;
    this.velocities[offset + 2] = (Math.random() - 0.5) * 0.1;
  }

  update(deltaSeconds: number, timeSeconds: number): void {
    const positionAttribute = this.points.geometry.getAttribute("position") as THREE.BufferAttribute;
    const positions = positionAttribute.array as Float32Array;
    for (let emberIndex = 0; emberIndex < positions.length / 3; emberIndex += 1) {
      const offset = emberIndex * 3;
      positions[offset] += (this.velocities[offset] + Math.sin(timeSeconds * 1.7 + emberIndex) * 0.08) * deltaSeconds;
      positions[offset + 1] += this.velocities[offset + 1] * deltaSeconds;
      positions[offset + 2] += this.velocities[offset + 2] * deltaSeconds;
      if (positions[offset + 1] > volume.maxY) {
        this.respawn(positions, emberIndex, false);
      }
    }
    positionAttribute.needsUpdate = true;
  }
}
