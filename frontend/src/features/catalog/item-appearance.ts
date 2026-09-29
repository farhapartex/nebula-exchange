import {
  Atom,
  BatteryCharging,
  Cog,
  Cpu,
  Crown,
  Drill,
  Fuel,
  Gem,
  Hexagon,
  Layers,
  Mountain,
  Orbit,
  Rocket,
  Zap,
  type LucideIcon,
} from "lucide-react";

import type { CatalogItem, ItemCategory } from "@/features/catalog/api/catalog-types";

export type ItemAppearance = {
  icon: LucideIcon;
  gradientClassName: string;
  glyphClassName: string;
  label: string;
};

const resourceAppearanceBySlug: Record<string, Omit<ItemAppearance, "label">> = {
  "iron-ore": {
    icon: Mountain,
    gradientClassName: "from-slate-500/40 to-slate-800/60",
    glyphClassName: "text-slate-200",
  },
  "copper-ore": {
    icon: Mountain,
    gradientClassName: "from-orange-500/40 to-orange-900/60",
    glyphClassName: "text-orange-200",
  },
  crystal: { icon: Gem, gradientClassName: "from-cyan-400/40 to-cyan-900/60", glyphClassName: "text-cyan-100" },
  plasma: { icon: Zap, gradientClassName: "from-fuchsia-500/40 to-fuchsia-900/60", glyphClassName: "text-fuchsia-100" },
  "rare-earth": {
    icon: Hexagon,
    gradientClassName: "from-emerald-500/40 to-emerald-900/60",
    glyphClassName: "text-emerald-100",
  },
  "void-shard": {
    icon: Orbit,
    gradientClassName: "from-violet-500/50 to-indigo-950/70",
    glyphClassName: "text-violet-100",
  },
};

const componentIconsBySlug: Record<string, LucideIcon> = {
  "alloy-plate": Layers,
  circuit: Cpu,
  "power-core": BatteryCharging,
  "void-engine": Atom,
};

const tierGradients = [
  "from-zinc-500/40 to-zinc-800/60",
  "from-sky-500/40 to-sky-900/60",
  "from-indigo-500/40 to-indigo-900/60",
  "from-purple-500/40 to-purple-900/60",
  "from-rose-500/40 to-rose-900/60",
];

const categoryLabels: Record<ItemCategory, string> = {
  resource: "Resource",
  component: "Component",
  drill: "Drill",
  ship: "Ship",
  consumable: "Consumable",
  legendary: "Legendary",
};

function tierGradient(tier: number | null): string {
  const tierIndex = Math.min(Math.max((tier ?? 1) - 1, 0), tierGradients.length - 1);
  return tierGradients[tierIndex];
}

export function describeItemAppearance(catalogItem: CatalogItem): ItemAppearance {
  const label = categoryLabels[catalogItem.category];
  switch (catalogItem.category) {
    case "resource":
      return {
        label,
        ...(resourceAppearanceBySlug[catalogItem.slug] ?? {
          icon: Mountain,
          gradientClassName: tierGradients[0],
          glyphClassName: "text-foreground",
        }),
      };
    case "component":
      return {
        label,
        icon: componentIconsBySlug[catalogItem.slug] ?? Cog,
        gradientClassName: "from-amber-500/35 to-amber-900/60",
        glyphClassName: "text-amber-100",
      };
    case "drill":
      return {
        label,
        icon: Drill,
        gradientClassName: tierGradient(catalogItem.tier),
        glyphClassName: "text-foreground",
      };
    case "ship":
      return {
        label,
        icon: Rocket,
        gradientClassName: tierGradient(catalogItem.tier),
        glyphClassName: "text-foreground",
      };
    case "consumable":
      return {
        label,
        icon: Fuel,
        gradientClassName: "from-lime-500/35 to-lime-900/60",
        glyphClassName: "text-lime-100",
      };
    case "legendary":
      return {
        label,
        icon: Crown,
        gradientClassName: "from-yellow-400/50 via-amber-600/40 to-orange-900/70",
        glyphClassName: "text-yellow-100",
      };
  }
}

export function describeItemTier(catalogItem: CatalogItem): string | null {
  return catalogItem.tier === null ? null : `Tier ${catalogItem.tier}`;
}
