import { Wallet } from "lucide-react";

import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { StatusBadge } from "@/components/ui/status-badge";

export function WalletLinkPlaceholder() {
  return (
    <Card>
      <CardHeader
        title="Wallet and withdrawals"
        description="Link a wallet on Base to pay with USDC and withdraw NC or items."
        action={<StatusBadge label="Coming soon" tone="accent" />}
      />
      <CardContent className="flex items-center gap-3 text-sm text-muted">
        <Wallet className="size-5 shrink-0 text-subtle" aria-hidden="true" />
        Wallet linking, USDC payments and withdrawals arrive with the on-chain release.
      </CardContent>
    </Card>
  );
}
