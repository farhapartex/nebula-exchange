"use client";

import Link from "next/link";
import { ChevronDown, LogOut, ReceiptText } from "lucide-react";
import { DropdownMenu } from "radix-ui";

import { ProfileMenuWallet } from "@/components/app-shell/profile-menu-wallet";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/features/auth/session/use-auth";
import { useLogOutAction } from "@/features/auth/session/use-log-out-action";

export function ProfileMenu() {
  const { user } = useAuth();
  const { logOutAndLeave } = useLogOutAction();

  if (!user) {
    return (
      <Button asChild size="sm" variant="secondary">
        <Link href="/login?next=/fight">Log in</Link>
      </Button>
    );
  }

  const initials = user.username.slice(0, 2).toUpperCase();

  return (
    <DropdownMenu.Root>
      <DropdownMenu.Trigger className="flex items-center gap-1.5 rounded-lg p-1 transition-colors hover:bg-surface-raised focus-visible:ring-2 focus-visible:ring-highlight focus-visible:outline-none">
        <span className="flex size-8 items-center justify-center rounded-full bg-linear-to-br from-amber-400 to-accent font-display text-base text-background">
          {initials}
        </span>
        <ChevronDown className="size-4 text-muted" aria-hidden="true" />
      </DropdownMenu.Trigger>
      <DropdownMenu.Portal>
        <DropdownMenu.Content
          align="end"
          sideOffset={8}
          className="z-50 min-w-48 rounded-xl border border-border-strong bg-surface-raised p-1.5 shadow-xl shadow-black/50 data-[state=open]:animate-pop-in"
        >
          <div className="px-2.5 py-2">
            <p className="text-sm font-medium text-foreground">{user.username}</p>
            <p className="truncate text-xs text-muted">{user.email}</p>
          </div>
          <ProfileMenuWallet />
          <DropdownMenu.Separator className="my-1 h-px bg-border" />
          <DropdownMenu.Item
            asChild
            className="flex cursor-pointer items-center gap-2 rounded-lg px-2.5 py-2 text-sm text-muted outline-none data-[highlighted]:bg-border data-[highlighted]:text-foreground"
          >
            <Link href="/subscription">
              <ReceiptText className="size-4" aria-hidden="true" />
              Subscription
            </Link>
          </DropdownMenu.Item>
          <DropdownMenu.Item
            onSelect={() => void logOutAndLeave()}
            className="flex cursor-pointer items-center gap-2 rounded-lg px-2.5 py-2 text-sm text-muted outline-none data-[highlighted]:bg-border data-[highlighted]:text-foreground"
          >
            <LogOut className="size-4" aria-hidden="true" />
            Log out
          </DropdownMenu.Item>
        </DropdownMenu.Content>
      </DropdownMenu.Portal>
    </DropdownMenu.Root>
  );
}
