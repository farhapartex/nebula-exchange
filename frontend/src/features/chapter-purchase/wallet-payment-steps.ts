export type WalletPaymentStep = "PREPARING" | "APPROVING" | "PAYING" | "SUBMITTED";

export const walletPaymentStepOrder: WalletPaymentStep[] = ["PREPARING", "APPROVING", "PAYING", "SUBMITTED"];

export const walletPaymentStepLabels: Record<WalletPaymentStep, string> = {
  PREPARING: "Prepare the payment",
  APPROVING: "Allow the USDC amount",
  PAYING: "Send the payment",
  SUBMITTED: "Wait for the network",
};

export function walletPaymentStepState(
  step: WalletPaymentStep,
  currentStep: WalletPaymentStep,
): "DONE" | "ACTIVE" | "WAITING" {
  const stepIndex = walletPaymentStepOrder.indexOf(step);
  const currentIndex = walletPaymentStepOrder.indexOf(currentStep);
  if (stepIndex < currentIndex) {
    return "DONE";
  }
  return stepIndex === currentIndex ? "ACTIVE" : "WAITING";
}
