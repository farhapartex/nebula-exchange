import { BaseError, ContractFunctionRevertedError, UserRejectedRequestError } from "viem";

import type { ToastRequest } from "@/components/ui/toast/toast-context";

const contractErrorMessages: Record<string, string> = {
  StaleEthUsdPrice: "The ETH price from Chainlink is out of date. Pay with USDC, or try again later.",
  InvalidEthUsdPrice: "The ETH price is not available right now. Pay with USDC, or try again later.",
  PaymentAuthorizationExpired: "This checkout has expired. Press Pay again to start a new one.",
  PaymentAlreadyMade: "This checkout is already paid. Your chapters unlock in a moment.",
  InvalidPaymentAuthorization: "This payment was not approved for this wallet. Reload the page and try again.",
  InsufficientEthSent: "The ETH price moved while you were paying. Press Pay again.",
  EnforcedPause: "Payments are paused for a moment. Try again later.",
  ERC20InsufficientBalance: "Your wallet does not have enough USDC.",
  ERC20InsufficientAllowance: "The USDC amount was not allowed. Press Pay again and approve it.",
};

export function revertedContractErrorName(error: unknown): string | null {
  if (!(error instanceof BaseError)) {
    return null;
  }
  const revertedError = error.walk((cause) => cause instanceof ContractFunctionRevertedError);
  return revertedError instanceof ContractFunctionRevertedError ? (revertedError.data?.errorName ?? null) : null;
}

function isWalletRejection(error: unknown): boolean {
  if (error instanceof BaseError) {
    return error.walk((cause) => cause instanceof UserRejectedRequestError) instanceof UserRejectedRequestError;
  }
  return error instanceof Error && error.name === UserRejectedRequestError.name;
}

export function walletPaymentErrorToast(error: unknown): ToastRequest {
  if (isWalletRejection(error)) {
    return { tone: "info", title: "Payment cancelled", description: "You cancelled the request in your wallet." };
  }
  const contractErrorName = revertedContractErrorName(error);
  const contractMessage = contractErrorName ? contractErrorMessages[contractErrorName] : undefined;
  return {
    tone: "error",
    title: "Wallet payment failed",
    description: contractMessage ?? "Nothing was taken. Try again in a moment.",
  };
}
