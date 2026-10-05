import * as THREE from "three";

import { createTextTexture } from "@/features/fight/stage/stage-textures";

type ActiveBurst = { points: THREE.Points; velocities: Float32Array; remainingMs: number };
type ActiveLabel = { sprite: THREE.Sprite; remainingMs: number; totalMs: number };

const burstLifetimeMs = 420;
const labelLifetimeMs = 800;

export class HitEffects {
  readonly group = new THREE.Group();
  private bursts: ActiveBurst[] = [];
  private labels: ActiveLabel[] = [];

  constructor(private readonly sparkTexture: THREE.Texture) {}

  burst(position: THREE.Vector3, isBlocked: boolean): void {
    const count = isBlocked ? 10 : 26;
    const positions = new Float32Array(count * 3);
    const velocities = new Float32Array(count * 3);
    for (let sparkIndex = 0; sparkIndex < count; sparkIndex += 1) {
      positions.set([position.x, position.y, position.z + 0.05], sparkIndex * 3);
      const angle = Math.random() * Math.PI * 2;
      const speed = (isBlocked ? 1.2 : 2.6) * (0.4 + Math.random());
      velocities.set(
        [Math.cos(angle) * speed, Math.sin(angle) * speed + 0.6, (Math.random() - 0.5) * speed * 0.5],
        sparkIndex * 3,
      );
    }
    const geometry = new THREE.BufferGeometry();
    geometry.setAttribute("position", new THREE.BufferAttribute(positions, 3));
    const points = new THREE.Points(
      geometry,
      new THREE.PointsMaterial({
        size: isBlocked ? 0.07 : 0.11,
        map: this.sparkTexture,
        color: isBlocked ? "#cbd5e1" : "#fde68a",
        transparent: true,
        depthWrite: false,
        blending: THREE.AdditiveBlending,
      }),
    );
    this.group.add(points);
    this.bursts.push({ points, velocities, remainingMs: burstLifetimeMs });
  }

  label(position: THREE.Vector3, text: string, color: string): void {
    const { texture, aspectRatio } = createTextTexture(text, color);
    const sprite = new THREE.Sprite(
      new THREE.SpriteMaterial({ map: texture, transparent: true, depthWrite: false, depthTest: false }),
    );
    const labelHeight = 0.32;
    sprite.scale.set(labelHeight * aspectRatio, labelHeight, 1);
    sprite.position.copy(position);
    this.group.add(sprite);
    this.labels.push({ sprite, remainingMs: labelLifetimeMs, totalMs: labelLifetimeMs });
  }

  update(deltaMs: number): void {
    const deltaSeconds = deltaMs / 1000;
    this.bursts = this.bursts.filter((burst) => {
      burst.remainingMs -= deltaMs;
      const positionAttribute = burst.points.geometry.getAttribute("position") as THREE.BufferAttribute;
      const positions = positionAttribute.array as Float32Array;
      for (let index = 0; index < positions.length; index += 3) {
        burst.velocities[index + 1] -= 6 * deltaSeconds;
        positions[index] += burst.velocities[index] * deltaSeconds;
        positions[index + 1] += burst.velocities[index + 1] * deltaSeconds;
        positions[index + 2] += burst.velocities[index + 2] * deltaSeconds;
      }
      positionAttribute.needsUpdate = true;
      (burst.points.material as THREE.PointsMaterial).opacity = Math.max(0, burst.remainingMs / burstLifetimeMs);
      if (burst.remainingMs <= 0) {
        this.group.remove(burst.points);
        burst.points.geometry.dispose();
        (burst.points.material as THREE.Material).dispose();
        return false;
      }
      return true;
    });
    this.labels = this.labels.filter((activeLabel) => {
      activeLabel.remainingMs -= deltaMs;
      activeLabel.sprite.position.y += 0.7 * deltaSeconds;
      activeLabel.sprite.material.opacity = Math.max(0, activeLabel.remainingMs / activeLabel.totalMs);
      if (activeLabel.remainingMs <= 0) {
        this.group.remove(activeLabel.sprite);
        activeLabel.sprite.material.map?.dispose();
        activeLabel.sprite.material.dispose();
        return false;
      }
      return true;
    });
  }
}
