"use client";

import { useState } from "react";

import { ContentSection } from "@/components/layout/content-section";
import { PaymentMethodPicker } from "@/features/payments/components/payment-method-picker";
import {
  describePaymentChoices,
  preferredPaymentChoice,
  type PaymentChoice,
  type PaymentUseCase,
} from "@/features/payments/payment-method-rules";

const pickerExamples: { useCase: PaymentUseCase; title: string; price: bigint; balance: bigint | null }[] = [
  { useCase: "entry_fee", title: "Entry fee (5 NC, empty balance)", price: 5_000_000n, balance: 0n },
  { useCase: "topup", title: "Top-up", price: 25_000_000n, balance: null },
  { useCase: "shop", title: "Shop: Scout (2 NC) with 12.50 NC", price: 2_000_000n, balance: 12_500_000n },
  { useCase: "shop", title: "Shop: Starter bundle (3.50 NC) with 1.20 NC", price: 3_500_000n, balance: 1_200_000n },
];

function PickerExample({ example }: { example: (typeof pickerExamples)[number] }) {
  const choices = describePaymentChoices(example.useCase, example.price, example.balance);
  const [selectedChoice, setSelectedChoice] = useState<PaymentChoice | null>(preferredPaymentChoice(choices));
  return (
    <div className="rounded-xl border border-border bg-background/40 p-4">
      <PaymentMethodPicker
        label={example.title}
        choices={choices}
        value={selectedChoice}
        onValueChange={setSelectedChoice}
        availableBalance={example.balance?.toString()}
      />
    </div>
  );
}

export function PaymentMethodShowcase() {
  return (
    <ContentSection
      title="PaymentMethodPicker"
      description="Options follow the PRD's payment method table. The wallet stays disabled until wallets ship."
    >
      <div className="space-y-4">
        {pickerExamples.map((example) => (
          <PickerExample key={example.title} example={example} />
        ))}
      </div>
    </ContentSection>
  );
}
