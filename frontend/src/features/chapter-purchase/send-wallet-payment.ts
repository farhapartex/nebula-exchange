import { erc20Abi } from "viem";
import { readContract, waitForTransactionReceipt, writeContract } from "wagmi/actions";

import { ethToSendFor, type WalletPaymentAsset } from "@/features/chapter-purchase/wallet-payment-asset";
import type { WalletPaymentStep } from "@/features/chapter-purchase/wallet-payment-steps";
import type { CreatedWalletCheckout } from "@/features/subscriptions/api/checkout-session-api";
import { chapterPaymentVaultAbi } from "@/lib/web3/abi/chapter-payment-vault-abi";
import { wagmiConfig } from "@/lib/web3/wagmi-config";

export type WalletPaymentRequest = {
  walletCheckout: CreatedWalletCheckout;
  asset: WalletPaymentAsset;
  payerAddress: `0x${string}`;
};

export class WalletPaymentRevertedError extends Error {
  constructor() {
    super("The payment transaction was reverted");
    this.name = "WalletPaymentRevertedError";
  }
}

export async function sendWalletPayment(
  { walletCheckout, asset, payerAddress }: WalletPaymentRequest,
  onStep: (walletPaymentStep: WalletPaymentStep) => void,
): Promise<`0x${string}`> {
  const vaultAddress = walletCheckout.vault_address;
  const chainId = walletCheckout.chain_id;
  const usdCents = BigInt(walletCheckout.usd_cents);
  const authorization = {
    paymentReference: walletCheckout.payment_reference,
    usdCents,
    deadline: BigInt(walletCheckout.deadline),
    signature: walletCheckout.signature,
  };

  let paymentHash: `0x${string}`;
  if (asset === "USDC") {
    const [paymentToken, tokenUnits] = await Promise.all([
      readContract(wagmiConfig, {
        address: vaultAddress,
        abi: chapterPaymentVaultAbi,
        functionName: "paymentToken",
        chainId,
      }),
      readContract(wagmiConfig, {
        address: vaultAddress,
        abi: chapterPaymentVaultAbi,
        functionName: "quoteTokenUnits",
        args: [usdCents],
        chainId,
      }),
    ]);
    const allowance = await readContract(wagmiConfig, {
      address: paymentToken,
      abi: erc20Abi,
      functionName: "allowance",
      args: [payerAddress, vaultAddress],
      chainId,
    });
    if (allowance < tokenUnits) {
      onStep("APPROVING");
      const approvalHash = await writeContract(wagmiConfig, {
        address: paymentToken,
        abi: erc20Abi,
        functionName: "approve",
        args: [vaultAddress, tokenUnits],
        account: payerAddress,
        chainId,
      });
      await waitForSuccessfulReceipt(approvalHash, chainId);
    }
    onStep("PAYING");
    paymentHash = await writeContract(wagmiConfig, {
      address: vaultAddress,
      abi: chapterPaymentVaultAbi,
      functionName: "payWithToken",
      args: [authorization],
      account: payerAddress,
      chainId,
    });
  } else {
    onStep("PAYING");
    const quotedWei = await readContract(wagmiConfig, {
      address: vaultAddress,
      abi: chapterPaymentVaultAbi,
      functionName: "quoteWei",
      args: [usdCents],
      chainId,
    });
    paymentHash = await writeContract(wagmiConfig, {
      address: vaultAddress,
      abi: chapterPaymentVaultAbi,
      functionName: "payWithEth",
      args: [authorization],
      value: ethToSendFor(quotedWei),
      account: payerAddress,
      chainId,
    });
  }
  onStep("SUBMITTED");
  await waitForSuccessfulReceipt(paymentHash, chainId);
  return paymentHash;
}

async function waitForSuccessfulReceipt(transactionHash: `0x${string}`, chainId: number) {
  const receipt = await waitForTransactionReceipt(wagmiConfig, { hash: transactionHash, chainId });
  if (receipt.status !== "success") {
    throw new WalletPaymentRevertedError();
  }
}
