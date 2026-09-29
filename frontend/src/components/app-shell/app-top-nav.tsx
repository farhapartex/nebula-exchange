import { MobileNavDrawer } from "@/components/app-shell/mobile-nav-drawer";
import { NotificationBell } from "@/components/app-shell/notification-bell";
import { PrimaryNavLinks } from "@/components/app-shell/primary-nav-links";
import { ProfileMenu } from "@/components/app-shell/profile-menu";
import { NebulaLogo } from "@/components/brand/nebula-logo";
import { PageContainer } from "@/components/layout/page-container";
import { BalanceChip } from "@/features/balances/balance-chip";

export function AppTopNav() {
  return (
    <header className="sticky top-0 z-40 border-b border-border/70 bg-background/80 backdrop-blur-md">
      <PageContainer className="flex h-16 items-center gap-3">
        <MobileNavDrawer />
        <NebulaLogo wordmarkClassName="hidden sm:inline lg:hidden xl:inline" />
        <div className="ml-4 hidden flex-1 lg:flex">
          <PrimaryNavLinks />
        </div>
        <div className="ml-auto flex items-center gap-2">
          <BalanceChip />
          <NotificationBell />
          <ProfileMenu />
        </div>
      </PageContainer>
    </header>
  );
}
