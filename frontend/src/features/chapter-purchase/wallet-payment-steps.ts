import type { WalletPaymentAsset } from "@/features/chapter-purchase/wallet-payment-asset";

export type WalletPaymentStep = "PREPARING" | "APPROVING" | "PAYING" | "SUBMITTED";

export const walletPaymentStepLabels: Record<WalletPaymentStep, string> = {
  PREPARING: "Prepare the payment",
  APPROVING: "Allow the USDC amount",
  PAYING: "Send the payment",
  SUBMITTED: "Wait for the network",
};

export function walletPaymentStepsFor(asset: WalletPaymentAsset): WalletPaymentStep[] {
  return asset === "USDC" ? ["PREPARING", "APPROVING", "PAYING", "SUBMITTED"] : ["PREPARING", "PAYING", "SUBMITTED"];
}

export function walletPaymentStepState(
  steps: WalletPaymentStep[],
  step: WalletPaymentStep,
  currentStep: WalletPaymentStep,
): "DONE" | "ACTIVE" | "WAITING" {
  const stepIndex = steps.indexOf(step);
  const currentIndex = steps.indexOf(currentStep);
  if (stepIndex < currentIndex) {
    return "DONE";
  }
  return stepIndex === currentIndex ? "ACTIVE" : "WAITING";
}
