import { readSessionMockState, writeSessionMockState } from "@/mocks/utils/session-mock-store";

const mockCoinBalanceStorageKey = "street-born-mock-coin-balance";
const startingMockCoins = 1250;

export function readMockCoinBalance(): number {
  return readSessionMockState<number>(mockCoinBalanceStorageKey, startingMockCoins);
}

export function spendMockCoins(coins: number): boolean {
  const balance = readMockCoinBalance();
  if (balance < coins) {
    return false;
  }
  writeSessionMockState(mockCoinBalanceStorageKey, balance - coins);
  return true;
}
