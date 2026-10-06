import type { WalletAssetQuote } from "@/features/chapter-purchase/use-wallet-asset-quotes";
import {
  ethDecimals,
  shownEthDecimals,
  shownUsdcDecimals,
  usdcDecimals,
  type WalletPaymentAsset,
} from "@/features/chapter-purchase/wallet-payment-asset";
import { cn } from "@/utils/class-names";
import { formatTokenAmount } from "@/utils/web3/format-token-amount";

type AssetOption = {
  asset: WalletPaymentAsset;
  quote: WalletAssetQuote;
  decimals: number;
  shownDecimals: number;
  note: string;
};

type WalletAssetPickerProps = {
  selectedAsset: WalletPaymentAsset;
  onSelect: (asset: WalletPaymentAsset) => void;
  ethQuote: WalletAssetQuote;
  usdcQuote: WalletAssetQuote;
  isDisabled: boolean;
};

function describeBalance(option: AssetOption): string {
  if (option.quote.isUnavailable) {
    return option.asset === "ETH" ? "The ETH price is not available right now" : "USDC is not available right now";
  }
  if (option.quote.balanceUnits === null) {
    return "Checking your balance";
  }
  const balanceText = `You have ${formatTokenAmount(option.quote.balanceUnits, option.decimals, option.shownDecimals)} ${option.asset}`;
  return option.quote.hasEnoughBalance ? balanceText : `${balanceText}, not enough`;
}

export function WalletAssetPicker({
  selectedAsset,
  onSelect,
  ethQuote,
  usdcQuote,
  isDisabled,
}: WalletAssetPickerProps) {
  const assetOptions: AssetOption[] = [
    {
      asset: "ETH",
      quote: ethQuote,
      decimals: ethDecimals,
      shownDecimals: shownEthDecimals,
      note: "Live Chainlink price. Extra ETH sent comes straight back.",
    },
    {
      asset: "USDC",
      quote: usdcQuote,
      decimals: usdcDecimals,
      shownDecimals: shownUsdcDecimals,
      note: "Exact amount. Your wallet asks you to allow it first.",
    },
  ];

  return (
    <fieldset disabled={isDisabled} className="grid gap-3 sm:grid-cols-2">
      <legend className="sr-only">Pay with ETH or USDC</legend>
      {assetOptions.map((option) => {
        const isSelected = option.asset === selectedAsset;
        const isBlocked =
          option.quote.isUnavailable || (option.quote.balanceUnits !== null && !option.quote.hasEnoughBalance);
        return (
          <label
            key={option.asset}
            className={cn(
              "flex cursor-pointer flex-col gap-1 rounded-xl border p-3.5 transition-colors has-[:focus-visible]:ring-2 has-[:focus-visible]:ring-highlight",
              isSelected ? "border-accent bg-accent/10" : "border-border bg-background/40 hover:border-border-strong",
              isBlocked && "opacity-60",
            )}
          >
            <input
              type="radio"
              name="wallet-payment-asset"
              value={option.asset}
              checked={isSelected}
              onChange={() => onSelect(option.asset)}
              className="sr-only"
            />
            <span className="flex items-baseline justify-between gap-2">
              <span className="font-medium text-foreground">{option.asset}</span>
              <span className="font-mono text-sm whitespace-nowrap text-foreground tabular-nums">
                {option.quote.priceUnits === null
                  ? "..."
                  : formatTokenAmount(option.quote.priceUnits, option.decimals, option.shownDecimals)}
              </span>
            </span>
            <span className="text-xs text-muted">{option.note}</span>
            <span className={cn("text-xs", isBlocked ? "text-down" : "text-subtle")}>{describeBalance(option)}</span>
          </label>
        );
      })}
    </fieldset>
  );
}
