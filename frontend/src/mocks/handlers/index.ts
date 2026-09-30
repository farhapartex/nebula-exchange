import { activationHandlers } from "@/mocks/handlers/activation-handlers";
import { authHandlers } from "@/mocks/handlers/auth-handlers";
import { balancesHandlers } from "@/mocks/handlers/balances-handlers";
import { catalogHandlers } from "@/mocks/handlers/catalog-handlers";
import { devShowcaseHandlers } from "@/mocks/handlers/dev-showcase-handlers";
import { healthHandlers } from "@/mocks/handlers/health-handlers";
import { inventoryHandlers } from "@/mocks/handlers/inventory-handlers";
import { ledgerHistoryHandlers } from "@/mocks/handlers/ledger-history-handlers";
import { passwordResetHandlers } from "@/mocks/handlers/password-reset-handlers";
import { paymentsHandlers } from "@/mocks/handlers/payments-handlers";
import { sessionHandlers } from "@/mocks/handlers/session-handlers";
import { settingsHandlers } from "@/mocks/handlers/settings-handlers";

export const mockRequestHandlers = [
  ...healthHandlers,
  ...authHandlers,
  ...activationHandlers,
  ...passwordResetHandlers,
  ...sessionHandlers,
  ...settingsHandlers,
  ...catalogHandlers,
  ...balancesHandlers,
  ...inventoryHandlers,
  ...ledgerHistoryHandlers,
  ...paymentsHandlers,
  ...devShowcaseHandlers,
];
