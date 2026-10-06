import { requestData } from "@/lib/api/api-client";

export type AttackStats = {
  damage: number;
  stamina_cost: number;
  range: number;
  windup_ms: number;
  active_ms: number;
  recovery_ms: number;
  knockback: number;
  hit_stun_ms: number;
};

export type FighterStats = {
  max_health: number;
  max_stamina: number;
  stamina_regen_per_second: number;
  walk_speed: number;
  block_damage_reduction: number;
  block_stamina_cost: number;
  dodge_stamina_cost: number;
  dodge_duration_ms: number;
  dodge_distance: number;
  punch: AttackStats;
  kick: AttackStats;
};

export type EnemyBrainProfile = {
  aggression: number;
  block_chance: number;
  dodge_chance: number;
  preferred_range: number;
  reaction_ms: number;
  punch_weight: number;
  kick_weight: number;
};

export type FighterLook = {
  body_color: string;
  outline_color: string;
  height: number;
  width: number;
};

export type FighterSetup = {
  id: string;
  name: string;
  title: string;
  stats: FighterStats;
  look: FighterLook;
};

export type EnemySetup = FighterSetup & {
  brain: EnemyBrainProfile;
};

export type ImageCrop = {
  x: number;
  y: number;
  width: number;
  height: number;
};

export type StageLayer = {
  id: string;
  image: string;
  crop: ImageCrop;
  position: [number, number, number];
  width: number;
  brightness: number;
  edge_fade: number;
  is_lit: boolean;
};

export type StageSilhouette = {
  id: string;
  position: [number, number, number];
  size: [number, number];
  rotation_degrees: number;
  color: string;
};

export type FireLight = {
  position: [number, number, number];
  color: string;
  intensity: number;
  distance: number;
  flicker: number;
};

export type StageCamera = {
  field_of_view: number;
  height: number;
  look_height: number;
  close_distance: number;
  far_distance: number;
  follow_speed: number;
  knockout_zoom: number;
};

export type StageSetup = {
  world_units_per_pixel: number;
  background_color: string;
  fog: { color: string; density: number };
  ambient_light: { color: string; intensity: number };
  rim_light: { color: string; intensity: number; position: [number, number, number] };
  fire_lights: FireLight[];
  floor: { color: string; depth: number };
  layers: StageLayer[];
  silhouettes: StageSilhouette[];
  smoke: { color: string; count: number; opacity: number };
  embers: { count: number; colors: string[] };
  bloom: { strength: number; radius: number; threshold: number };
  grade: { vignette: number; grain: number; tint: string };
  camera: StageCamera;
};

export type ArenaSetup = {
  name: string;
  width: number;
  floor_y: number;
  stage: StageSetup;
};

export type FightSetup = {
  level_id: string;
  time_limit_seconds: number;
  arena: ArenaSetup;
  player: FighterSetup;
  enemy: EnemySetup;
};

export const fightSetupQueryKey = (levelID: string) => ["levels", levelID, "fight"] as const;

export function fetchFightSetup(levelID: string): Promise<FightSetup> {
  return requestData<FightSetup>(`/levels/${encodeURIComponent(levelID)}/fight-setup`);
}
