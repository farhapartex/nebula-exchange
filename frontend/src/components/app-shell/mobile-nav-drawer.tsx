"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { LogOut, Menu, X } from "lucide-react";
import { Dialog } from "radix-ui";

import {
  isNavigationLinkActive,
  primaryNavigationLinks,
  profileNavigationLinks,
  type NavigationLink,
} from "@/components/app-shell/navigation-links";
import { PilotAvatar } from "@/components/app-shell/pilot-avatar";
import { useDisplayedPilot } from "@/components/app-shell/use-displayed-pilot";
import { NebulaLogo } from "@/components/brand/nebula-logo";
import { cn } from "@/utils/class-names";

function DrawerLinkGroup({
  title,
  navigationLinks,
  currentPathname,
  onNavigate,
}: {
  title: string;
  navigationLinks: NavigationLink[];
  currentPathname: string;
  onNavigate: () => void;
}) {
  return (
    <div>
      <p className="mb-2 px-3 text-xs font-semibold tracking-wider text-subtle uppercase">{title}</p>
      <ul className="space-y-0.5">
        {navigationLinks.map((navigationLink) => {
          const LinkIcon = navigationLink.icon;
          const isActive = isNavigationLinkActive(navigationLink, currentPathname);
          return (
            <li key={navigationLink.href}>
              <Link
                href={navigationLink.href}
                onClick={onNavigate}
                aria-current={isActive ? "page" : undefined}
                className={cn(
                  "flex h-11 items-center gap-3 rounded-lg px-3 text-sm font-medium transition-colors",
                  isActive
                    ? "bg-accent/15 text-foreground ring-1 ring-accent/40"
                    : "text-muted hover:bg-surface-raised hover:text-foreground",
                )}
              >
                <LinkIcon className={cn("size-4", isActive ? "text-accent-soft" : "text-subtle")} />
                {navigationLink.label}
              </Link>
            </li>
          );
        })}
      </ul>
    </div>
  );
}

export function MobileNavDrawer() {
  const [isOpen, setIsOpen] = useState(false);
  const currentPathname = usePathname();
  const displayedPilot = useDisplayedPilot();
  const closeDrawer = () => setIsOpen(false);

  return (
    <Dialog.Root open={isOpen} onOpenChange={setIsOpen}>
      <Dialog.Trigger
        aria-label="Open navigation"
        className="flex size-9 items-center justify-center rounded-lg text-muted transition-colors hover:bg-surface-raised hover:text-foreground focus-visible:ring-2 focus-visible:ring-highlight focus-visible:outline-none lg:hidden"
      >
        <Menu className="size-5" />
      </Dialog.Trigger>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-50 bg-background/80 backdrop-blur-sm data-[state=open]:animate-fade-in lg:hidden" />
        <Dialog.Content className="fixed inset-y-0 left-0 z-50 flex w-[min(20rem,85vw)] flex-col border-r border-border-strong bg-surface shadow-2xl shadow-black/60 focus:outline-none data-[state=open]:animate-drawer-in lg:hidden">
          <div className="flex h-16 items-center justify-between border-b border-border px-4">
            <NebulaLogo />
            <Dialog.Close
              aria-label="Close navigation"
              className="flex size-9 items-center justify-center rounded-lg text-muted hover:bg-surface-raised hover:text-foreground"
            >
              <X className="size-5" />
            </Dialog.Close>
          </div>
          <Dialog.Title className="sr-only">Navigation</Dialog.Title>
          <Dialog.Description className="sr-only">Main sections and your account links</Dialog.Description>
          <div className="flex-1 space-y-6 overflow-y-auto p-4">
            <DrawerLinkGroup
              title="Play"
              navigationLinks={primaryNavigationLinks}
              currentPathname={currentPathname}
              onNavigate={closeDrawer}
            />
            <DrawerLinkGroup
              title="Account"
              navigationLinks={profileNavigationLinks}
              currentPathname={currentPathname}
              onNavigate={closeDrawer}
            />
          </div>
          <div className="flex items-center gap-3 border-t border-border p-4">
            <PilotAvatar username={displayedPilot.username} />
            <div className="min-w-0 flex-1">
              <p className="truncate text-sm font-medium text-foreground">{displayedPilot.username}</p>
              <p className="truncate text-xs text-muted">{displayedPilot.email}</p>
            </div>
            <button
              type="button"
              disabled
              aria-label="Log out"
              className="flex size-9 items-center justify-center rounded-lg text-muted disabled:opacity-40"
            >
              <LogOut className="size-4" />
            </button>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
