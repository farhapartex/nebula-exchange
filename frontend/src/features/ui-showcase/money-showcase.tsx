"use client";

import { useState } from "react";

import { ContentSection } from "@/components/layout/content-section";
import { PriceChange } from "@/components/market/price-change";
import { NcAmount } from "@/components/money/nc-amount";
import { Input } from "@/components/ui/input";
import { formatMicroUnits, parseNcToMicroUnits } from "@/utils/money/micro-units";
import { calculatePriceChange, fractionDigitsForTickSize } from "@/utils/money/price";

const amountExamples = [
  { label: "Balance", amount: "1284500000" },
  { label: "Crafting fee", amount: "20000" },
  { label: "Sub-cent remainder", amount: "4999" },
  { label: "Beyond JS number range", amount: "9223372036854775807" },
];

const priceChangeExamples = [
  { symbol: "IRON/NC", current: "10421", reference: "10000" },
  { symbol: "CRYSTAL/NC", current: "8242", reference: "8420" },
  { symbol: "FUEL/NC", current: "200000", reference: "200000" },
];

const ironTickSizeInMicroUnits = "100";

export function MoneyShowcase() {
  const [typedAmount, setTypedAmount] = useState("0.0105");
  const parsedMicroUnits = parseNcToMicroUnits(typedAmount);

  return (
    <ContentSection
      title="NcAmount and PriceChange"
      description="Exact BigInt math on micro-units. Hover or focus an amount to see all 6 decimals."
    >
      <div className="space-y-6">
        <dl className="divide-y divide-border rounded-xl border border-border">
          {amountExamples.map((amountExample) => (
            <div key={amountExample.label} className="flex items-center justify-between gap-4 px-4 py-2.5 text-sm">
              <dt className="text-muted">{amountExample.label}</dt>
              <dd className="text-foreground">
                <NcAmount amount={amountExample.amount} />
              </dd>
            </div>
          ))}
          <div className="flex items-center justify-between gap-4 px-4 py-2.5 text-sm">
            <dt className="text-muted">Price at IRON tick size</dt>
            <dd className="text-foreground">
              <NcAmount amount="10500" fractionDigits={fractionDigitsForTickSize(ironTickSizeInMicroUnits)} />
            </dd>
          </div>
        </dl>

        <ul className="flex flex-wrap gap-4">
          {priceChangeExamples.map((priceChangeExample) => (
            <li key={priceChangeExample.symbol} className="flex items-center gap-2 text-sm">
              <span className="font-mono text-foreground">{priceChangeExample.symbol}</span>
              <PriceChange
                variant="pill"
                changeInBasisPoints={
                  calculatePriceChange(priceChangeExample.current, priceChangeExample.reference).changeInBasisPoints
                }
              />
            </li>
          ))}
        </ul>

        <div className="grid items-end gap-4 sm:grid-cols-2">
          <Input
            label="Type an NC amount"
            value={typedAmount}
            onChange={(changeEvent) => setTypedAmount(changeEvent.target.value)}
            trailingAdornment="NC"
            inputMode="decimal"
            className="font-mono tabular-nums"
            errorMessage={parsedMicroUnits === null ? "Up to 6 decimals, digits only" : undefined}
          />
          <div className="rounded-lg border border-border bg-background/60 px-3 py-2 font-mono text-xs text-muted tabular-nums">
            {parsedMicroUnits === null ? (
              "—"
            ) : (
              <>
                <p>micro-units: {parsedMicroUnits.toString()}</p>
                <p>formatted: {formatMicroUnits(parsedMicroUnits)} NC</p>
              </>
            )}
          </div>
        </div>
      </div>
    </ContentSection>
  );
}
