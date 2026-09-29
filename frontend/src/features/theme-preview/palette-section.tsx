import { ContentSection } from "@/components/layout/content-section";
import { colorTokenGroups, type ColorToken } from "@/features/theme-preview/theme-tokens";
import { cn } from "@/utils/class-names";

function ColorSwatch({ token }: { token: ColorToken }) {
  return (
    <li className="flex items-start gap-3 rounded-xl border border-border bg-background/60 p-3">
      <span className={cn("size-10 shrink-0 rounded-lg ring-1 ring-border-strong", token.swatchClassName)} />
      <span className="min-w-0">
        <span className="block truncate font-mono text-sm text-foreground">{token.name}</span>
        <span className="block font-mono text-xs text-subtle">{token.hex}</span>
        <span className="mt-0.5 block text-xs leading-snug text-muted">{token.usage}</span>
      </span>
    </li>
  );
}

export function PaletteSection() {
  return (
    <ContentSection
      title="Color palette"
      description="Semantic tokens used as Tailwind classes, e.g. bg-surface, text-muted."
    >
      <div className="space-y-6">
        {colorTokenGroups.map((tokenGroup) => (
          <div key={tokenGroup.title}>
            <h3 className="mb-3 text-xs font-semibold tracking-wider text-subtle uppercase">{tokenGroup.title}</h3>
            <ul className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
              {tokenGroup.tokens.map((token) => (
                <ColorSwatch key={token.name} token={token} />
              ))}
            </ul>
          </div>
        ))}
      </div>
    </ContentSection>
  );
}
