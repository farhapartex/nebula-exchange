import type { CatalogItem } from "@/features/catalog/api/catalog-types";
import { describeItemAppearance } from "@/features/catalog/item-appearance";
import { cn } from "@/utils/class-names";

const iconSizes = {
  sm: { frame: "size-8 rounded-lg", glyph: "size-4" },
  md: { frame: "size-12 rounded-xl", glyph: "size-6" },
  lg: { frame: "size-20 rounded-2xl", glyph: "size-10" },
};

type ItemIconProps = {
  item: CatalogItem;
  size?: keyof typeof iconSizes;
  className?: string;
};

export function ItemIcon({ item, size = "md", className }: ItemIconProps) {
  const appearance = describeItemAppearance(item);
  const IconGlyph = appearance.icon;
  const sizing = iconSizes[size];

  return (
    <span
      role="img"
      aria-label={item.name}
      className={cn(
        "relative inline-flex shrink-0 items-center justify-center border border-white/10 bg-linear-to-br shadow-inner",
        appearance.gradientClassName,
        item.category === "legendary" && "shadow-[0_0_24px_-4px] shadow-amber-400/60",
        sizing.frame,
        className,
      )}
    >
      <IconGlyph aria-hidden className={cn(sizing.glyph, appearance.glyphClassName)} strokeWidth={1.75} />
      {item.tier !== null && size !== "sm" && (
        <span className="absolute -right-1 -bottom-1 rounded-md border border-border bg-surface px-1 font-mono text-[10px] leading-4 text-muted">
          T{item.tier}
        </span>
      )}
    </span>
  );
}
