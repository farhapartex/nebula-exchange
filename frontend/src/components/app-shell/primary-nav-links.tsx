"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

import { isNavigationLinkActive, primaryNavigationLinks } from "@/components/app-shell/navigation-links";
import { cn } from "@/utils/class-names";

export function PrimaryNavLinks() {
  const currentPathname = usePathname();

  return (
    <nav aria-label="Main" className="hidden items-center gap-1 lg:flex">
      {primaryNavigationLinks.map((navigationLink) => {
        const isActive = isNavigationLinkActive(navigationLink, currentPathname);
        return (
          <Link
            key={navigationLink.href}
            href={navigationLink.href}
            aria-current={isActive ? "page" : undefined}
            className={cn(
              "relative rounded-md px-3 py-2 text-sm font-medium transition-colors",
              "focus-visible:ring-2 focus-visible:ring-highlight focus-visible:outline-none",
              isActive ? "text-foreground" : "text-muted hover:bg-surface-raised hover:text-foreground",
            )}
          >
            {navigationLink.label}
            {isActive && (
              <span className="absolute inset-x-3 -bottom-[13px] h-0.5 rounded-full bg-linear-to-r from-accent to-highlight" />
            )}
          </Link>
        );
      })}
    </nav>
  );
}
