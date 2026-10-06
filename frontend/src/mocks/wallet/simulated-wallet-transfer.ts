import { delay } from "msw";

import type { CreatedWalletCheckout } from "@/features/subscriptions/api/checkout-session-api";
import type { WalletPaymentStep } from "@/features/chapter-purchase/wallet-payment-steps";

const simulatedWalletPromptInMilliseconds = 1400;

export async function simulateWalletTransfer(
  walletCheckout: CreatedWalletCheckout,
  onStep: (walletPaymentStep: WalletPaymentStep) => void,
): Promise<void> {
  onStep("APPROVING");
  await delay(simulatedWalletPromptInMilliseconds);
  onStep("PAYING");
  await delay(simulatedWalletPromptInMilliseconds);
  onStep("SUBMITTED");
  void walletCheckout;
}
