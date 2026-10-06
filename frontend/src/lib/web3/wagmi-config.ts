import { createConfig, http } from "wagmi";
import { injected } from "wagmi/connectors";

import { chainRpcUrl, gameChain } from "@/lib/web3/chain-config";

export const wagmiConfig = createConfig({
  chains: [gameChain],
  connectors: [injected()],
  transports: { [gameChain.id]: http(chainRpcUrl) },
  ssr: true,
});
