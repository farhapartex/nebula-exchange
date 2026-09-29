import { Component, Palette, type LucideIcon } from "lucide-react";

export type DeveloperPage = {
  href: `/${string}`;
  label: string;
  description: string;
  icon: LucideIcon;
};

export const developerPages: DeveloperPage[] = [
  {
    href: "/dev/theme",
    label: "Theme preview",
    description: "Color tokens, typography and market signal styles.",
    icon: Palette,
  },
  {
    href: "/dev/ui",
    label: "UI components",
    description: "Every shared component, money formatting and the mocked API client.",
    icon: Component,
  },
];
