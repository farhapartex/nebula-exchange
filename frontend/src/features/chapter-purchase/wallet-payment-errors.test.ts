import { describe, expect, it } from "vitest";
import { ContractFunctionRevertedError, UserRejectedRequestError, type Abi } from "viem";

import { walletPaymentErrorToast } from "@/features/chapter-purchase/wallet-payment-errors";

const vaultErrorAbi = [
  { type: "error", name: "StaleEthUsdPrice", inputs: [{ name: "updatedAt", type: "uint256" }] },
] as const satisfies Abi;

describe("walletPaymentErrorToast", () => {
  it("explains a contract revert in plain words", () => {
    const revertedError = new ContractFunctionRevertedError({
      abi: vaultErrorAbi,
      functionName: "payWithEth",
      data: "0x71d167920000000000000000000000000000000000000000000000000000000000000001",
    });
    const toast = walletPaymentErrorToast(revertedError);
    expect(toast.tone).toBe("error");
    expect(toast.description).toBe("The ETH price from Chainlink is out of date. Pay with USDC, or try again later.");
  });

  it("treats a cancelled wallet prompt as a cancellation, not an error", () => {
    const toast = walletPaymentErrorToast(new UserRejectedRequestError(new Error("User rejected")));
    expect(toast.tone).toBe("info");
    expect(toast.title).toBe("Payment cancelled");
  });

  it("falls back to a safe message for anything else", () => {
    expect(walletPaymentErrorToast(new Error("network down")).description).toBe(
      "Nothing was taken. Try again in a moment.",
    );
  });
});
