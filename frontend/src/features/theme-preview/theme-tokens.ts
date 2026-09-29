export type ColorToken = {
  name: string;
  hex: string;
  swatchClassName: string;
  usage: string;
};

export type ColorTokenGroup = {
  title: string;
  tokens: ColorToken[];
};

export const colorTokenGroups: ColorTokenGroup[] = [
  {
    title: "Surfaces",
    tokens: [
      { name: "background", hex: "#05070F", swatchClassName: "bg-background", usage: "Page background" },
      { name: "surface", hex: "#0B1020", swatchClassName: "bg-surface", usage: "Cards and panels" },
      { name: "surface-raised", hex: "#121A31", swatchClassName: "bg-surface-raised", usage: "Menus, modals, hover" },
      { name: "border", hex: "#1E2944", swatchClassName: "bg-border", usage: "Dividers and outlines" },
      { name: "border-strong", hex: "#2C3A5E", swatchClassName: "bg-border-strong", usage: "Inputs and focus rings" },
    ],
  },
  {
    title: "Text",
    tokens: [
      { name: "foreground", hex: "#E7ECF7", swatchClassName: "bg-foreground", usage: "Primary text" },
      { name: "muted", hex: "#8D98B3", swatchClassName: "bg-muted", usage: "Secondary text and labels" },
      { name: "subtle", hex: "#5D6883", swatchClassName: "bg-subtle", usage: "Hints and disabled text" },
    ],
  },
  {
    title: "Brand",
    tokens: [
      { name: "accent", hex: "#8B5CF6", swatchClassName: "bg-accent", usage: "Primary actions and links" },
      { name: "accent-soft", hex: "#C4B5FD", swatchClassName: "bg-accent-soft", usage: "Accent text on dark" },
      { name: "highlight", hex: "#22D3EE", swatchClassName: "bg-highlight", usage: "Live data and focus glow" },
    ],
  },
  {
    title: "Signals",
    tokens: [
      { name: "up", hex: "#22C55E", swatchClassName: "bg-up", usage: "Price up, buy, success" },
      { name: "down", hex: "#F43F5E", swatchClassName: "bg-down", usage: "Price down, sell, errors" },
      { name: "warning", hex: "#F59E0B", swatchClassName: "bg-warning", usage: "Holds, halts, pending" },
      { name: "info", hex: "#38BDF8", swatchClassName: "bg-info", usage: "Neutral notices" },
    ],
  },
];

export type MarketSignalExample = {
  symbol: string;
  lastPrice: string;
  changePercent: string;
  direction: "up" | "down" | "flat";
};

export const marketSignalExamples: MarketSignalExample[] = [
  { symbol: "IRON/NC", lastPrice: "0.0100", changePercent: "+4.21%", direction: "up" },
  { symbol: "CRYSTAL/NC", lastPrice: "0.0842", changePercent: "−2.10%", direction: "down" },
  { symbol: "FUEL/NC", lastPrice: "0.2000", changePercent: "0.00%", direction: "flat" },
];
