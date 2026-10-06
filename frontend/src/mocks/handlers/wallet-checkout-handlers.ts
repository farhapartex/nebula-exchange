import { bypass, http, passthrough } from "msw";
import { keccak256, toHex } from "viem";

import type { Plan } from "@/features/chapter-purchase/api/plan-api";
import type {
  CheckoutSessionRequest,
  CheckoutSessionState,
  CreatedWalletCheckout,
  PaymentMethod,
} from "@/features/subscriptions/api/checkout-session-api";
import type { Subscription } from "@/features/subscriptions/api/subscription-api";
import { buildApiUrl } from "@/lib/api/api-config";
import type { DataEnvelope, ListEnvelope } from "@/lib/api/api-types";
import { gameChain } from "@/lib/web3/chain-config";
import { mockDataResponse, mockErrorResponse, mockListResponse, simulateLatency } from "@/mocks/utils/mock-responses";
import { readSessionMockState, writeSessionMockState } from "@/mocks/utils/session-mock-store";

const walletCheckoutsStorageKey = "street-born-mock-wallet-checkouts";
const walletCheckoutLifetimeInMilliseconds = 30 * 60 * 1000;
const statusChecksBeforePaid = 2;
const usdcUnitsPerCent = 10000n;
const placeholderTokenAddress = "0x5FbDB2315678afecb367f032d93F642f64180aa3";
const placeholderVaultAddress = "0xe7f1725E7734CE288F8367e1Bb143E90bb3F0512";

type MockWalletCheckout = {
  statusChecks: number;
  subscription: Subscription;
};

type CheckoutSessionBody = CheckoutSessionRequest & { payment_method?: PaymentMethod };

function readWalletCheckouts(): Record<string, MockWalletCheckout> {
  return readSessionMockState<Record<string, MockWalletCheckout>>(walletCheckoutsStorageKey, {});
}

async function findPlanOption(request: Request, checkoutRequest: CheckoutSessionRequest) {
  const plansRequest = new Request(buildApiUrl("/plans?limit=100"), {
    headers: { Authorization: request.headers.get("Authorization") ?? "" },
  });
  const plansResponse = await fetch(bypass(plansRequest));
  const plans = ((await plansResponse.json()) as DataEnvelope<Plan[]>).data ?? [];
  const plan = plans.find((candidatePlan) => candidatePlan.id === checkoutRequest.plan_id);
  const option = plan?.options.find(
    (candidateOption) => candidateOption.chapter_count === checkoutRequest.chapter_count,
  );
  return plan && option ? { plan, option } : null;
}

export const walletCheckoutHandlers = [
  http.post(buildApiUrl("/checkout-sessions"), async ({ request }) => {
    const checkoutBody = (await request.clone().json()) as CheckoutSessionBody;
    if (checkoutBody.payment_method !== "WALLET") {
      return passthrough();
    }
    await simulateLatency();
    const planOption = await findPlanOption(request, checkoutBody);
    if (!planOption) {
      return mockErrorResponse(422, "VALIDATION_FAILED", "Some fields are invalid", {
        chapter_count: "is not offered by this plan",
      });
    }
    const checkoutID = crypto.randomUUID();
    const paidAt = new Date().toISOString();
    const walletCheckouts = readWalletCheckouts();
    walletCheckouts[checkoutID] = {
      statusChecks: 0,
      subscription: {
        id: checkoutID,
        plan_id: planOption.plan.id,
        plan_name: planOption.plan.name,
        plan_kind: planOption.plan.kind,
        status: "PAID",
        chapters: planOption.option.chapters.map(({ id, number, title }) => ({ id, number, title })),
        subtotal_cents: planOption.option.subtotal_cents,
        discount_percent: planOption.option.discount_percent,
        discount_cents: planOption.option.discount_cents,
        total_cents: planOption.option.total_cents,
        paid_at: paidAt,
        refunded_at: null,
      },
    };
    writeSessionMockState(walletCheckoutsStorageKey, walletCheckouts);
    return mockDataResponse<CreatedWalletCheckout>(
      {
        id: checkoutID,
        payment_reference: keccak256(toHex(checkoutID)),
        amount_units: (BigInt(planOption.option.total_cents) * usdcUnitsPerCent).toString(),
        token_address: placeholderTokenAddress,
        vault_address: placeholderVaultAddress,
        chain_id: gameChain.id,
        expires_at: new Date(Date.now() + walletCheckoutLifetimeInMilliseconds).toISOString(),
      },
      201,
    );
  }),

  http.get(buildApiUrl("/checkout-sessions/:checkoutSessionID"), async ({ params }) => {
    const checkoutID = String(params.checkoutSessionID);
    const walletCheckouts = readWalletCheckouts();
    const walletCheckout = walletCheckouts[checkoutID];
    if (!walletCheckout) {
      return passthrough();
    }
    await simulateLatency();
    walletCheckout.statusChecks += 1;
    writeSessionMockState(walletCheckoutsStorageKey, walletCheckouts);
    const isPaid = walletCheckout.statusChecks > statusChecksBeforePaid;
    return mockDataResponse<CheckoutSessionState>({ id: checkoutID, status: isPaid ? "PAID" : "OPEN" });
  }),

  http.get(buildApiUrl("/subscriptions"), async ({ request }) => {
    const realResponse = await fetch(bypass(request));
    if (!realResponse.ok) {
      return realResponse;
    }
    const realSubscriptions = (await realResponse.json()) as ListEnvelope<Subscription>;
    const paidWalletSubscriptions = Object.values(readWalletCheckouts())
      .filter((walletCheckout) => walletCheckout.statusChecks > statusChecksBeforePaid)
      .map((walletCheckout) => walletCheckout.subscription);
    return mockListResponse([...paidWalletSubscriptions, ...realSubscriptions.data], new URL(request.url), 100);
  }),
];
