import { createSiweMessage } from "viem/siwe";

const walletLinkStatement =
  "Link this wallet to your Street Born account. This does not send a transaction or cost gas.";
const siweMessageVersion = "1";

type WalletLinkMessageInput = {
  address: `0x${string}`;
  chainID: number;
  nonce: string;
  pageOrigin: string;
  issuedAt: Date;
};

export function buildWalletLinkMessage({
  address,
  chainID,
  nonce,
  pageOrigin,
  issuedAt,
}: WalletLinkMessageInput): string {
  const pageUrl = new URL(pageOrigin);
  return createSiweMessage({
    address,
    chainId: chainID,
    domain: pageUrl.host,
    nonce,
    uri: pageUrl.origin,
    version: siweMessageVersion,
    statement: walletLinkStatement,
    issuedAt,
  });
}
