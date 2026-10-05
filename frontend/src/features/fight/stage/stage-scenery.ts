import * as THREE from "three";

import type { StageLayer, StageSetup, StageSilhouette } from "@/features/fight/api/fight-setup-api";
import { createLayerTexture, createVerticalFadeTexture, loadImage } from "@/features/fight/stage/stage-textures";
import { safeHexColor } from "@/utils/colors/safe-hex-color";

const floorWidth = 40;
const groundMistHeight = 1.6;
const groundMistOffsetFromNearestLayer = 0.4;

function createFloor(stage: StageSetup): THREE.Mesh {
  const floor = new THREE.Mesh(
    new THREE.PlaneGeometry(floorWidth, stage.floor.depth),
    new THREE.MeshStandardMaterial({
      color: safeHexColor(stage.floor.color, "#1d0f09"),
      roughness: 0.95,
      metalness: 0,
    }),
  );
  floor.rotation.x = -Math.PI / 2;
  floor.position.z = -stage.floor.depth / 2 + 4;
  return floor;
}

function createGroundMist(stage: StageSetup): THREE.Mesh | null {
  if (stage.layers.length === 0) {
    return null;
  }
  const nearestLayerDepth = Math.max(...stage.layers.map((layer) => layer.position[2]));
  const mist = new THREE.Mesh(
    new THREE.PlaneGeometry(floorWidth, groundMistHeight),
    new THREE.MeshBasicMaterial({
      map: createVerticalFadeTexture(128),
      color: safeHexColor(stage.floor.color, "#1d0f09"),
      transparent: true,
      depthWrite: false,
    }),
  );
  mist.position.set(0, groundMistHeight / 2 - 0.05, nearestLayerDepth + groundMistOffsetFromNearestLayer);
  return mist;
}

function createSilhouette(silhouette: StageSilhouette): THREE.Mesh {
  const mesh = new THREE.Mesh(
    new THREE.PlaneGeometry(silhouette.size[0], silhouette.size[1]),
    new THREE.MeshBasicMaterial({ color: safeHexColor(silhouette.color, "#050302"), fog: false }),
  );
  mesh.position.set(...silhouette.position);
  mesh.rotation.z = THREE.MathUtils.degToRad(silhouette.rotation_degrees);
  return mesh;
}

async function createLayer(layer: StageLayer): Promise<THREE.Mesh | null> {
  try {
    const image = await loadImage(layer.image);
    const { texture, aspectRatio } = createLayerTexture(image, layer.crop, layer.edge_fade, layer.brightness);
    const material = layer.is_lit
      ? new THREE.MeshStandardMaterial({
          map: texture,
          transparent: true,
          roughness: 1,
          metalness: 0,
          emissiveMap: texture,
          emissive: new THREE.Color("#ffffff"),
          emissiveIntensity: 0.45,
        })
      : new THREE.MeshBasicMaterial({ map: texture, transparent: true });
    const mesh = new THREE.Mesh(new THREE.PlaneGeometry(layer.width, layer.width / aspectRatio), material);
    mesh.position.set(...layer.position);
    mesh.renderOrder = -1;
    return mesh;
  } catch {
    return null;
  }
}

export async function buildStageScenery(stage: StageSetup): Promise<THREE.Group> {
  const group = new THREE.Group();
  group.add(new THREE.AmbientLight(safeHexColor(stage.ambient_light.color, "#3a1a10"), stage.ambient_light.intensity));
  const rimLight = new THREE.DirectionalLight(
    safeHexColor(stage.rim_light.color, "#2dd4bf"),
    stage.rim_light.intensity,
  );
  rimLight.position.set(...stage.rim_light.position);
  group.add(rimLight, createFloor(stage));
  const layers = await Promise.all(stage.layers.map(createLayer));
  for (const layer of layers) {
    if (layer) {
      group.add(layer);
    }
  }
  const groundMist = createGroundMist(stage);
  if (groundMist) {
    group.add(groundMist);
  }
  for (const silhouette of stage.silhouettes) {
    group.add(createSilhouette(silhouette));
  }
  return group;
}
