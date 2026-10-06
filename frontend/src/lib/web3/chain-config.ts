import { defineChain } from "viem";
import { anvil } from "viem/chains";

const defaultChainID = anvil.id;
const defaultChainRpcUrl = "http://localhost:8545";

export const chainRpcUrl = process.env.NEXT_PUBLIC_CHAIN_RPC_URL ?? defaultChainRpcUrl;

export const gameChain = defineChain({
  ...anvil,
  id: Number(process.env.NEXT_PUBLIC_CHAIN_ID ?? defaultChainID),
  name: "Anvil Local",
  rpcUrls: { default: { http: [chainRpcUrl] } },
});
