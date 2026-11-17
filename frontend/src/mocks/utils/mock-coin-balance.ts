import { readSessionMockState, writeSessionMockState } from "@/mocks/utils/session-mock-store";

const mockCoinBalanceStorageKey = "street-born-mock-coin-balance";
const welcomeGiftMockCoins = 100;

export function readMockCoinBalance(): number {
  return readSessionMockState<number>(mockCoinBalanceStorageKey, welcomeGiftMockCoins);
}

export function spendMockCoins(coins: number): boolean {
  const balance = readMockCoinBalance();
  if (balance < coins) {
    return false;
  }
  writeSessionMockState(mockCoinBalanceStorageKey, balance - coins);
  return true;
}
