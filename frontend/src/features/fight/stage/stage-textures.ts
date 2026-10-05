import * as THREE from "three";

import type { ImageCrop } from "@/features/fight/api/fight-setup-api";

export function loadImage(source: string): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const image = new Image();
    image.decoding = "async";
    image.onload = () => resolve(image);
    image.onerror = () => reject(new Error(`Could not load stage image ${source}`));
    image.src = source;
  });
}

function fadeEdges(context: CanvasRenderingContext2D, width: number, height: number, edgeFade: number): void {
  if (edgeFade <= 0) {
    return;
  }
  context.globalCompositeOperation = "destination-in";
  const horizontal = context.createLinearGradient(0, 0, width, 0);
  horizontal.addColorStop(0, "rgba(0,0,0,0)");
  horizontal.addColorStop(edgeFade, "rgba(0,0,0,1)");
  horizontal.addColorStop(1 - edgeFade, "rgba(0,0,0,1)");
  horizontal.addColorStop(1, "rgba(0,0,0,0)");
  context.fillStyle = horizontal;
  context.fillRect(0, 0, width, height);
  const vertical = context.createLinearGradient(0, 0, 0, height);
  vertical.addColorStop(0, "rgba(0,0,0,0)");
  vertical.addColorStop(edgeFade, "rgba(0,0,0,1)");
  vertical.addColorStop(1 - edgeFade * 0.5, "rgba(0,0,0,1)");
  vertical.addColorStop(1, "rgba(0,0,0,0)");
  context.fillStyle = vertical;
  context.fillRect(0, 0, width, height);
  context.globalCompositeOperation = "source-over";
}

export function createLayerTexture(
  image: HTMLImageElement,
  crop: ImageCrop,
  edgeFade: number,
  brightness: number,
): { texture: THREE.CanvasTexture; aspectRatio: number } {
  const sourceX = crop.x * image.naturalWidth;
  const sourceY = crop.y * image.naturalHeight;
  const sourceWidth = crop.width * image.naturalWidth;
  const sourceHeight = crop.height * image.naturalHeight;
  const canvas = document.createElement("canvas");
  canvas.width = Math.round(sourceWidth);
  canvas.height = Math.round(sourceHeight);
  const context = canvas.getContext("2d");
  if (context) {
    context.filter = `brightness(${brightness})`;
    context.drawImage(image, sourceX, sourceY, sourceWidth, sourceHeight, 0, 0, canvas.width, canvas.height);
    context.filter = "none";
    fadeEdges(context, canvas.width, canvas.height, edgeFade);
  }
  const texture = new THREE.CanvasTexture(canvas);
  texture.colorSpace = THREE.SRGBColorSpace;
  return { texture, aspectRatio: sourceWidth / sourceHeight };
}

export function createRadialTexture(size: number, innerAlpha: number, middleAlpha: number): THREE.CanvasTexture {
  const canvas = document.createElement("canvas");
  canvas.width = size;
  canvas.height = size;
  const context = canvas.getContext("2d");
  if (context) {
    const center = size / 2;
    const gradient = context.createRadialGradient(center, center, 0, center, center, center);
    gradient.addColorStop(0, `rgba(255,255,255,${innerAlpha})`);
    gradient.addColorStop(0.4, `rgba(255,255,255,${middleAlpha})`);
    gradient.addColorStop(1, "rgba(255,255,255,0)");
    context.fillStyle = gradient;
    context.fillRect(0, 0, size, size);
  }
  const texture = new THREE.CanvasTexture(canvas);
  texture.colorSpace = THREE.SRGBColorSpace;
  return texture;
}

export function createVerticalFadeTexture(height: number): THREE.CanvasTexture {
  const canvas = document.createElement("canvas");
  canvas.width = 4;
  canvas.height = height;
  const context = canvas.getContext("2d");
  if (context) {
    const gradient = context.createLinearGradient(0, 0, 0, height);
    gradient.addColorStop(0, "rgba(255,255,255,0)");
    gradient.addColorStop(0.55, "rgba(255,255,255,0.85)");
    gradient.addColorStop(1, "rgba(255,255,255,1)");
    context.fillStyle = gradient;
    context.fillRect(0, 0, 4, height);
  }
  const texture = new THREE.CanvasTexture(canvas);
  texture.colorSpace = THREE.SRGBColorSpace;
  return texture;
}

export function createTextTexture(text: string, color: string): { texture: THREE.CanvasTexture; aspectRatio: number } {
  const canvas = document.createElement("canvas");
  const fontSize = 96;
  const context = canvas.getContext("2d");
  const font = `${fontSize}px "Bebas Neue", Impact, sans-serif`;
  let measuredWidth = fontSize * text.length * 0.6;
  if (context) {
    context.font = font;
    measuredWidth = context.measureText(text).width;
  }
  canvas.width = Math.ceil(measuredWidth + 40);
  canvas.height = fontSize + 30;
  const drawingContext = canvas.getContext("2d");
  if (drawingContext) {
    drawingContext.font = font;
    drawingContext.textBaseline = "middle";
    drawingContext.lineWidth = 12;
    drawingContext.strokeStyle = "rgba(0,0,0,0.9)";
    drawingContext.strokeText(text, 20, canvas.height / 2);
    drawingContext.fillStyle = color;
    drawingContext.fillText(text, 20, canvas.height / 2);
  }
  const texture = new THREE.CanvasTexture(canvas);
  texture.colorSpace = THREE.SRGBColorSpace;
  return { texture, aspectRatio: canvas.width / canvas.height };
}
