import { Hammer, HandFist, Shield, Sword, type LucideIcon } from "lucide-react";

import type { ToolKey } from "@/features/fight-hub/api/fight-hub-types";

export const toolIcons: Record<ToolKey, LucideIcon> = {
  bare_fists: HandFist,
  iron_pipe: Hammer,
  street_blade: Sword,
  scrap_shield: Shield,
};
