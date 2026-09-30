export type PaymentChoice = "balance" | "card" | "wallet";

export type PaymentUseCase = "entry_fee" | "topup" | "shop";

export type PaymentChoiceAvailability = {
  choice: PaymentChoice;
  isAllowed: boolean;
  unavailableReason: string | null;
};

const allowedChoicesByUseCase: Record<PaymentUseCase, PaymentChoice[]> = {
  entry_fee: ["card", "wallet"],
  topup: ["card", "wallet"],
  shop: ["balance", "card", "wallet"],
};

const isWalletPaymentAvailable = false;

export function describePaymentChoices(
  useCase: PaymentUseCase,
  price: bigint,
  availableBalance: bigint | null,
): PaymentChoiceAvailability[] {
  return allowedChoicesByUseCase[useCase].map((choice) => {
    if (choice === "wallet" && !isWalletPaymentAvailable) {
      return { choice, isAllowed: false, unavailableReason: "Coming soon" };
    }
    if (choice === "balance" && availableBalance !== null && availableBalance < price) {
      return { choice, isAllowed: false, unavailableReason: "Not enough NC" };
    }
    return { choice, isAllowed: true, unavailableReason: null };
  });
}

export function preferredPaymentChoice(choices: PaymentChoiceAvailability[]): PaymentChoice | null {
  return choices.find((availability) => availability.isAllowed)?.choice ?? null;
}
