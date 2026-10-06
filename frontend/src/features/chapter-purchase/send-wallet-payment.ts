import type { WalletPaymentStep } from "@/features/chapter-purchase/wallet-payment-steps";
import type { CreatedWalletCheckout } from "@/features/subscriptions/api/checkout-session-api";
import { simulateWalletTransfer } from "@/mocks/wallet/simulated-wallet-transfer";

export function sendWalletPayment(
  walletCheckout: CreatedWalletCheckout,
  onStep: (walletPaymentStep: WalletPaymentStep) => void,
): Promise<void> {
  return simulateWalletTransfer(walletCheckout, onStep);
}
