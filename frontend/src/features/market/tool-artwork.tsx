import Image from "next/image";
import { Lock, Shield, Swords } from "lucide-react";

import type { ToolTypeSummary } from "@/features/market/api/market-types";
import { rarityArtworkClassNames } from "@/features/market/tool-rarity";
import { toolSvgSource } from "@/features/market/tool-svg-source";
import { cn } from "@/utils/class-names";

type ToolArtworkProps = {
  toolType: ToolTypeSummary;
  isLocked?: boolean;
  className?: string;
};

export function ToolArtwork({ toolType, isLocked = false, className }: ToolArtworkProps) {
  const CategoryIcon = toolType.category === "WEAPON" ? Swords : Shield;
  return (
    <div
      className={cn(
        "relative flex aspect-[4/3] items-center justify-center overflow-hidden rounded-xl border border-border bg-linear-to-br to-background",
        rarityArtworkClassNames[toolType.rarity],
        className,
      )}
    >
      {toolType.image_svg ? (
        <Image src={toolSvgSource(toolType.image_svg)} alt="" fill unoptimized className="object-contain p-4" />
      ) : (
        <CategoryIcon className="size-12 opacity-80" aria-hidden="true" />
      )}
      {isLocked && (
        <span className="absolute inset-0 flex items-center justify-center bg-background/60 backdrop-blur-[2px]">
          <Lock className="size-6 text-muted" aria-hidden="true" />
        </span>
      )}
    </div>
  );
}
