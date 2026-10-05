import * as THREE from "three";
import { EffectComposer } from "three/examples/jsm/postprocessing/EffectComposer.js";
import { OutputPass } from "three/examples/jsm/postprocessing/OutputPass.js";
import { RenderPass } from "three/examples/jsm/postprocessing/RenderPass.js";
import { UnrealBloomPass } from "three/examples/jsm/postprocessing/UnrealBloomPass.js";

import type { FightSetup } from "@/features/fight/api/fight-setup-api";
import { createEnemyBrain, thinkEnemy, type EnemyBrain, type RandomNumber } from "@/features/fight/engine/enemy-brain";
import type { FightController } from "@/features/fight/engine/fight-controller";
import { attackPhaseOf, createFightState, stepFight, type FightRules } from "@/features/fight/engine/fight-simulation";
import { buildSnapshot } from "@/features/fight/engine/fight-snapshot";
import type { FighterRole, FightEvent, FightState } from "@/features/fight/engine/fight-types";
import {
  cameraGoal,
  createCameraRig,
  shakeCamera,
  shakeOffset,
  stepCameraRig,
  type CameraRig,
} from "@/features/fight/stage/cinematic-camera";
import { disposeObjectTree } from "@/features/fight/stage/dispose-object-tree";
import { EmberField } from "@/features/fight/stage/ember-field";
import {
  advancePacing,
  createFightPacing,
  freezeForImpact,
  isInSlowMotion,
  startSlowMotion,
  type FightPacing,
} from "@/features/fight/stage/fight-pacing";
import { createShadowTexture, FighterBillboard } from "@/features/fight/stage/fighter-billboard";
import { FireLights } from "@/features/fight/stage/fire-lights";
import { HitEffects } from "@/features/fight/stage/hit-effects";
import { SmokeDrift } from "@/features/fight/stage/smoke-drift";
import { toWorldLength, toWorldX } from "@/features/fight/stage/stage-coordinates";
import { buildStageScenery } from "@/features/fight/stage/stage-scenery";
import { createRadialTexture } from "@/features/fight/stage/stage-textures";
import { safeHexColor } from "@/utils/colors/safe-hex-color";

export type FightStageOptions = {
  host: HTMLElement;
  setup: FightSetup;
  controller: FightController;
  random: RandomNumber;
};

const maximumFrameDeltaMs = 50;
const snapshotIntervalMs = 60;
const flashDurationMs = 140;
const knockoutSlowMotionMs = 900;
const zoomEaseSpeed = 1.8;

export class FightStage {
  private readonly renderer: THREE.WebGLRenderer;
  private readonly scene = new THREE.Scene();
  private readonly camera: THREE.PerspectiveCamera;
  private readonly composer: EffectComposer;
  private readonly bloomPass: UnrealBloomPass;
  private readonly rules: FightRules;
  private readonly rig: CameraRig;
  private readonly fireLights: FireLights;
  private readonly smoke: SmokeDrift;
  private readonly embers: EmberField;
  private readonly hitEffects: HitEffects;
  private readonly fighters: Record<FighterRole, FighterBillboard>;
  private readonly resizeObserver: ResizeObserver;
  private state: FightState;
  private brain: EnemyBrain = createEnemyBrain();
  private pacing: FightPacing = createFightPacing();
  private flashRemainingMs: Record<FighterRole, number> = { player: 0, enemy: 0 };
  private damageDealt = 0;
  private damageTaken = 0;
  private millisecondsSinceSnapshot = 0;
  private hasPublishedOutcome = false;
  private animationFrameId: number | null = null;
  private previousFrameTimeMs: number | null = null;
  private isDestroyed = false;

  constructor(private readonly options: FightStageOptions) {
    const { setup } = options;
    const stage = setup.arena.stage;
    this.rules = { player: setup.player.stats, enemy: setup.enemy.stats };
    this.state = createFightState(this.rules, setup.arena.width, setup.time_limit_seconds);

    this.renderer = new THREE.WebGLRenderer({ antialias: true, powerPreference: "high-performance" });
    this.renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
    this.renderer.outputColorSpace = THREE.SRGBColorSpace;
    this.renderer.toneMapping = THREE.ACESFilmicToneMapping;
    this.renderer.toneMappingExposure = 1.1;
    this.renderer.domElement.className = "block size-full";
    options.host.appendChild(this.renderer.domElement);

    this.scene.background = new THREE.Color(safeHexColor(stage.background_color, "#0b0504"));
    this.scene.fog = new THREE.FogExp2(safeHexColor(stage.fog.color, "#170906"), stage.fog.density);
    this.camera = new THREE.PerspectiveCamera(stage.camera.field_of_view, 16 / 9, 0.1, 80);
    this.rig = createCameraRig(stage.camera);

    this.fireLights = new FireLights(stage.fire_lights);
    this.smoke = new SmokeDrift(
      stage.smoke.count,
      stage.smoke.color,
      stage.smoke.opacity,
      createRadialTexture(256, 0.9, 0.35),
    );
    this.embers = new EmberField(stage.embers.count, stage.embers.colors, createRadialTexture(64, 1, 0.5));
    this.hitEffects = new HitEffects(createRadialTexture(64, 1, 0.45));
    const shadowTexture = createShadowTexture();
    this.fighters = {
      player: new FighterBillboard(setup.player.look, setup.player.stats, stage.world_units_per_pixel, shadowTexture),
      enemy: new FighterBillboard(setup.enemy.look, setup.enemy.stats, stage.world_units_per_pixel, shadowTexture),
    };
    this.scene.add(
      this.fireLights.group,
      this.smoke.group,
      this.embers.points,
      this.hitEffects.group,
      this.fighters.player.group,
      this.fighters.enemy.group,
    );

    this.composer = new EffectComposer(this.renderer);
    this.composer.addPass(new RenderPass(this.scene, this.camera));
    this.bloomPass = new UnrealBloomPass(
      new THREE.Vector2(1, 1),
      stage.bloom.strength,
      stage.bloom.radius,
      stage.bloom.threshold,
    );
    this.composer.addPass(this.bloomPass);
    this.composer.addPass(new OutputPass());

    this.resizeObserver = new ResizeObserver(() => this.resize());
    this.resizeObserver.observe(options.host);
    this.resize();
    options.controller.onRestart(() => this.restart());
  }

  async start(): Promise<void> {
    const scenery = await buildStageScenery(this.options.setup.arena.stage);
    if (this.isDestroyed) {
      disposeObjectTree(scenery);
      return;
    }
    this.scene.add(scenery);
    this.publishSnapshot();
    this.animationFrameId = requestAnimationFrame(this.renderFrame);
  }

  destroy(): void {
    this.isDestroyed = true;
    if (this.animationFrameId !== null) {
      cancelAnimationFrame(this.animationFrameId);
    }
    this.options.controller.onRestart(null);
    this.resizeObserver.disconnect();
    this.fighters.player.dispose();
    this.fighters.enemy.dispose();
    disposeObjectTree(this.scene);
    this.composer.dispose();
    this.renderer.dispose();
    this.renderer.domElement.remove();
  }

  private resize(): void {
    const width = Math.max(this.options.host.clientWidth, 1);
    const height = Math.max(this.options.host.clientHeight, 1);
    this.renderer.setSize(width, height, false);
    this.composer.setSize(width, height);
    this.bloomPass.resolution.set(width, height);
    this.camera.aspect = width / height;
    this.camera.updateProjectionMatrix();
  }

  private restart(): void {
    this.state = createFightState(this.rules, this.options.setup.arena.width, this.options.setup.time_limit_seconds);
    this.brain = createEnemyBrain();
    this.pacing = createFightPacing();
    this.flashRemainingMs = { player: 0, enemy: 0 };
    this.damageDealt = 0;
    this.damageTaken = 0;
    this.hasPublishedOutcome = false;
    this.rig.zoomFactor = 1;
    this.publishSnapshot();
  }

  private renderFrame = (frameTimeMs: number): void => {
    if (this.isDestroyed) {
      return;
    }
    const realDeltaMs = Math.min(frameTimeMs - (this.previousFrameTimeMs ?? frameTimeMs), maximumFrameDeltaMs);
    this.previousFrameTimeMs = frameTimeMs;
    const isRunning = !this.options.controller.isPaused && !this.hasPublishedOutcome;
    const gameDeltaMs = isRunning ? advancePacing(this.pacing, realDeltaMs) : 0;

    if (gameDeltaMs > 0) {
      this.advanceFight(gameDeltaMs);
    }
    this.animateAmbience(realDeltaMs, frameTimeMs);
    this.hitEffects.update(isRunning ? Math.max(gameDeltaMs, realDeltaMs * 0.25) : realDeltaMs);
    this.updateFighters(frameTimeMs, realDeltaMs);
    this.updateCamera(realDeltaMs, frameTimeMs);
    this.composer.render(realDeltaMs / 1000);
    this.animationFrameId = requestAnimationFrame(this.renderFrame);
  };

  private advanceFight(deltaMs: number): void {
    const { setup, controller, random } = this.options;
    const enemyIntent = thinkEnemy(
      this.brain,
      {
        self: this.state.enemy,
        selfStats: setup.enemy.stats,
        opponent: this.state.player,
        opponentStats: setup.player.stats,
        profile: setup.enemy.brain,
        random,
      },
      deltaMs,
    );
    const events = stepFight(
      this.state,
      this.rules,
      { player: controller.readPlayerIntent(), enemy: enemyIntent },
      deltaMs,
    );
    for (const event of events) {
      this.reactToEvent(event);
    }
    this.millisecondsSinceSnapshot += deltaMs;
    const isOutcomeReady = this.state.outcome !== null && !isInSlowMotion(this.pacing);
    if (events.length > 0 || this.millisecondsSinceSnapshot >= snapshotIntervalMs || isOutcomeReady) {
      this.publishSnapshot();
    }
  }

  private publishSnapshot(): void {
    this.millisecondsSinceSnapshot = 0;
    const isOutcomeReady = this.state.outcome !== null && !isInSlowMotion(this.pacing);
    const snapshot = buildSnapshot(
      this.state,
      this.rules,
      this.options.controller.isPaused,
      this.damageDealt,
      this.damageTaken,
    );
    this.hasPublishedOutcome = isOutcomeReady;
    this.options.controller.publish(isOutcomeReady ? snapshot : { ...snapshot, outcome: null });
  }

  private impactPoint(role: FighterRole): THREE.Vector3 {
    const { setup } = this.options;
    const stage = setup.arena.stage;
    const look = role === "player" ? setup.player.look : setup.enemy.look;
    return new THREE.Vector3(
      toWorldX(this.state[role].x, setup.arena.width, stage.world_units_per_pixel),
      toWorldLength(look.height * 0.72, stage.world_units_per_pixel),
      0.2,
    );
  }

  private reactToEvent(event: FightEvent): void {
    if (event.kind === "hit") {
      const point = this.impactPoint(event.defender);
      this.hitEffects.burst(point, event.wasBlocked);
      if (event.attacker === "player") {
        this.damageDealt += event.damage;
      } else {
        this.damageTaken += event.damage;
      }
      if (event.wasBlocked) {
        this.hitEffects.label(point, "BLOCK", "#cbd5e1");
        shakeCamera(this.rig, 0.025, 110);
        freezeForImpact(this.pacing, 35);
        return;
      }
      const isKick = event.attack === "kick";
      this.hitEffects.label(point, `-${event.damage}`, event.defender === "player" ? "#fca5a5" : "#fde68a");
      shakeCamera(this.rig, isKick ? 0.09 : 0.055, isKick ? 220 : 150);
      freezeForImpact(this.pacing, isKick ? 90 : 60);
      this.flashRemainingMs[event.defender] = flashDurationMs;
    } else if (event.kind === "dodged") {
      this.hitEffects.label(this.impactPoint(event.defender), "DODGE", "#5eead4");
    } else if (event.kind === "knockout") {
      this.hitEffects.label(this.impactPoint(event.loser).add(new THREE.Vector3(0, 0.4, 0)), "K.O.", "#f97316");
      shakeCamera(this.rig, 0.14, 380);
      startSlowMotion(this.pacing, knockoutSlowMotionMs);
    } else if (event.kind === "time_up") {
      this.hitEffects.label(new THREE.Vector3(this.rig.target.x, 2.2, 0.4), "TIME", "#fdba74");
    }
  }

  private animateAmbience(deltaMs: number, frameTimeMs: number): void {
    const deltaSeconds = deltaMs / 1000;
    this.fireLights.update(frameTimeMs / 1000);
    this.smoke.update(deltaSeconds);
    this.embers.update(deltaSeconds, frameTimeMs / 1000);
  }

  private updateFighters(frameTimeMs: number, deltaMs: number): void {
    const { setup } = this.options;
    const unitsPerPixel = setup.arena.stage.world_units_per_pixel;
    for (const role of ["player", "enemy"] as const) {
      const fighter = this.state[role];
      const stats = role === "player" ? setup.player.stats : setup.enemy.stats;
      this.flashRemainingMs[role] = Math.max(0, this.flashRemainingMs[role] - deltaMs);
      this.fighters[role].update(
        fighter,
        toWorldX(fighter.x, setup.arena.width, unitsPerPixel),
        frameTimeMs,
        this.flashRemainingMs[role] > 0,
        role === "enemy" && attackPhaseOf(fighter, stats) === "windup",
      );
    }
  }

  private updateCamera(deltaMs: number, frameTimeMs: number): void {
    const { setup } = this.options;
    const stage = setup.arena.stage;
    const targetZoom = this.state.outcome === "won" || this.state.outcome === "lost" ? stage.camera.knockout_zoom : 1;
    this.rig.zoomFactor += (targetZoom - this.rig.zoomFactor) * (1 - Math.exp((-zoomEaseSpeed * deltaMs) / 1000));
    const playerWorldX = toWorldX(this.state.player.x, setup.arena.width, stage.world_units_per_pixel);
    const enemyWorldX = toWorldX(this.state.enemy.x, setup.arena.width, stage.world_units_per_pixel);
    stepCameraRig(
      this.rig,
      cameraGoal(playerWorldX, enemyWorldX, stage.camera, this.rig.zoomFactor),
      deltaMs,
      stage.camera.follow_speed,
    );
    const shake = shakeOffset(this.rig, frameTimeMs);
    this.camera.position.set(this.rig.position.x + shake.x, this.rig.position.y + shake.y, this.rig.position.z);
    this.camera.lookAt(this.rig.target.x + shake.x * 0.5, this.rig.target.y + shake.y * 0.5, this.rig.target.z);
  }
}
