"use client";

import Link from "next/link";
import { ChevronDown, LogOut } from "lucide-react";
import { DropdownMenu } from "radix-ui";

import { profileNavigationLinks } from "@/components/app-shell/navigation-links";
import { PilotAvatar } from "@/components/app-shell/pilot-avatar";
import { placeholderPilot } from "@/components/app-shell/placeholder-pilot";

const menuItemClassName =
  "flex h-9 cursor-pointer items-center gap-2.5 rounded-md px-2.5 text-sm text-foreground outline-none select-none data-disabled:cursor-not-allowed data-disabled:opacity-40 data-highlighted:bg-border";

export function ProfileMenu() {
  return (
    <DropdownMenu.Root>
      <DropdownMenu.Trigger
        aria-label="Open profile menu"
        className="hidden items-center gap-2 rounded-lg py-1 pr-2 pl-1 transition-colors hover:bg-surface-raised focus-visible:ring-2 focus-visible:ring-highlight focus-visible:outline-none sm:flex"
      >
        <PilotAvatar username={placeholderPilot.username} />
        <ChevronDown className="size-4 text-muted" />
      </DropdownMenu.Trigger>
      <DropdownMenu.Portal>
        <DropdownMenu.Content
          align="end"
          sideOffset={8}
          className="z-50 w-60 rounded-xl border border-border-strong bg-surface-raised p-1.5 shadow-xl shadow-black/50 data-[state=open]:animate-pop-in"
        >
          <div className="px-2.5 pt-1.5 pb-2.5">
            <p className="text-sm font-medium text-foreground">{placeholderPilot.username}</p>
            <p className="truncate text-xs text-muted">{placeholderPilot.email}</p>
          </div>
          <DropdownMenu.Separator className="my-1 h-px bg-border" />
          {profileNavigationLinks.map((navigationLink) => {
            const LinkIcon = navigationLink.icon;
            return (
              <DropdownMenu.Item key={navigationLink.href} asChild className={menuItemClassName}>
                <Link href={navigationLink.href}>
                  <LinkIcon className="size-4 text-muted" />
                  {navigationLink.label}
                </Link>
              </DropdownMenu.Item>
            );
          })}
          <DropdownMenu.Separator className="my-1 h-px bg-border" />
          <DropdownMenu.Item disabled className={menuItemClassName}>
            <LogOut className="size-4 text-muted" />
            Log out
          </DropdownMenu.Item>
        </DropdownMenu.Content>
      </DropdownMenu.Portal>
    </DropdownMenu.Root>
  );
}
