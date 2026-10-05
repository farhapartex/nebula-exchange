import * as THREE from "three";

import { safeHexColor } from "@/utils/colors/safe-hex-color";

type SmokePuff = { mesh: THREE.Mesh; driftSpeed: number; spinSpeed: number };

export class SmokeDrift {
  readonly group = new THREE.Group();
  private readonly puffs: SmokePuff[] = [];

  constructor(count: number, color: string, opacity: number, smokeTexture: THREE.Texture) {
    const smokeColor = new THREE.Color(safeHexColor(color, "#2a1a14"));
    for (let puffIndex = 0; puffIndex < count; puffIndex += 1) {
      const depth = -5 + (puffIndex / Math.max(count - 1, 1)) * 7.5;
      const size = 4 + Math.random() * 4;
      const mesh = new THREE.Mesh(
        new THREE.PlaneGeometry(size, size * 0.6),
        new THREE.MeshBasicMaterial({
          map: smokeTexture,
          color: smokeColor,
          transparent: true,
          opacity: depth > 1 ? opacity * 0.45 : opacity,
          depthWrite: false,
        }),
      );
      mesh.position.set(-6 + Math.random() * 12, 1.2 + Math.random() * 2.6, depth);
      mesh.rotation.z = Math.random() * Math.PI;
      this.group.add(mesh);
      this.puffs.push({ mesh, driftSpeed: 0.08 + Math.random() * 0.15, spinSpeed: (Math.random() - 0.5) * 0.06 });
    }
  }

  update(deltaSeconds: number): void {
    for (const puff of this.puffs) {
      puff.mesh.position.x += puff.driftSpeed * deltaSeconds;
      puff.mesh.rotation.z += puff.spinSpeed * deltaSeconds;
      if (puff.mesh.position.x > 8) {
        puff.mesh.position.x = -8;
      }
    }
  }
}
