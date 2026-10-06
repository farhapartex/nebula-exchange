import { describe, expect, it } from "vitest";
import { parseSiweMessage } from "viem/siwe";

import { buildWalletLinkMessage } from "@/features/wallet/build-wallet-link-message";

describe("buildWalletLinkMessage", () => {
  it("builds a sign-in message bound to this site, chain and nonce", () => {
    const message = buildWalletLinkMessage({
      address: "0x8ba1f109551bD432803012645Ac136ddd64DBA72",
      chainID: 31337,
      nonce: "a1b2c3d4e5f6a7b8",
      pageOrigin: "http://localhost:3000",
      issuedAt: new Date("2026-10-06T10:00:00Z"),
    });
    const parsedMessage = parseSiweMessage(message);
    expect(parsedMessage.domain).toBe("localhost:3000");
    expect(parsedMessage.chainId).toBe(31337);
    expect(parsedMessage.nonce).toBe("a1b2c3d4e5f6a7b8");
    expect(parsedMessage.statement).toContain("does not send a transaction");
  });
});
