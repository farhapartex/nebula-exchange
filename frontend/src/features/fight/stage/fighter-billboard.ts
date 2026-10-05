import * as THREE from "three";

import type { FighterLook, FighterStats } from "@/features/fight/api/fight-setup-api";
import type { FighterRuntime } from "@/features/fight/engine/fight-types";
import { drawFighterFigure } from "@/features/fight/stage/fighter-figure-canvas";
import { createRadialTexture } from "@/features/fight/stage/stage-textures";

const boxWidthInPixels = 560;
const boxHeightInPixels = 260;
const canvasPixelsPerUnit = 2;

export class FighterBillboard {
  readonly group = new THREE.Group();
  private readonly canvas = document.createElement("canvas");
  private readonly context: CanvasRenderingContext2D | null;
  private readonly texture: THREE.CanvasTexture;
  private readonly material: THREE.MeshStandardMaterial;

  constructor(
    private readonly look: FighterLook,
    private readonly stats: FighterStats,
    private readonly worldUnitsPerPixel: number,
    shadowTexture: THREE.Texture,
  ) {
    this.canvas.width = boxWidthInPixels * canvasPixelsPerUnit;
    this.canvas.height = boxHeightInPixels * canvasPixelsPerUnit;
    this.context = this.canvas.getContext("2d");
    this.texture = new THREE.CanvasTexture(this.canvas);
    this.texture.colorSpace = THREE.SRGBColorSpace;
    this.material = new THREE.MeshStandardMaterial({
      map: this.texture,
      emissiveMap: this.texture,
      emissive: new THREE.Color("#ffffff"),
      emissiveIntensity: 0.12,
      transparent: true,
      alphaTest: 0.02,
      roughness: 0.9,
      metalness: 0,
    });
    const figure = new THREE.Mesh(
      new THREE.PlaneGeometry(boxWidthInPixels * worldUnitsPerPixel, boxHeightInPixels * worldUnitsPerPixel),
      this.material,
    );
    figure.position.y = (boxHeightInPixels * worldUnitsPerPixel) / 2;
    const shadow = new THREE.Mesh(
      new THREE.PlaneGeometry(look.width * worldUnitsPerPixel * 3.2, look.width * worldUnitsPerPixel * 1.1),
      new THREE.MeshBasicMaterial({
        map: shadowTexture,
        color: "#000000",
        transparent: true,
        opacity: 0.75,
        depthWrite: false,
      }),
    );
    shadow.rotation.x = -Math.PI / 2;
    shadow.position.y = 0.01;
    this.group.add(shadow, figure);
  }

  update(fighter: FighterRuntime, worldX: number, timeMs: number, isFlashing: boolean, showsTelegraph: boolean): void {
    this.group.position.x = worldX;
    if (!this.context) {
      return;
    }
    this.context.clearRect(0, 0, this.canvas.width, this.canvas.height);
    drawFighterFigure(this.context, {
      fighter,
      stats: this.stats,
      look: this.look,
      originX: this.canvas.width / 2,
      floorY: this.canvas.height - 6,
      pixelsPerUnit: canvasPixelsPerUnit,
      timeMs,
      isFlashing,
      showsTelegraph,
    });
    this.texture.needsUpdate = true;
  }

  dispose(): void {
    this.texture.dispose();
    this.material.dispose();
  }
}

export function createShadowTexture(): THREE.CanvasTexture {
  return createRadialTexture(128, 1, 0.6);
}
