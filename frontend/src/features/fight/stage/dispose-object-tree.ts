import * as THREE from "three";

type TexturedMaterial = THREE.Material & { map?: THREE.Texture | null; emissiveMap?: THREE.Texture | null };

function disposeMaterial(material: TexturedMaterial): void {
  material.map?.dispose();
  material.emissiveMap?.dispose();
  material.dispose();
}

export function disposeObjectTree(root: THREE.Object3D): void {
  root.traverse((object) => {
    if (object instanceof THREE.Mesh || object instanceof THREE.Points || object instanceof THREE.Sprite) {
      object.geometry.dispose();
      const materials = Array.isArray(object.material) ? object.material : [object.material];
      for (const material of materials) {
        disposeMaterial(material as TexturedMaterial);
      }
    }
  });
}
