import { activationHandlers } from "@/mocks/handlers/activation-handlers";
import { authHandlers } from "@/mocks/handlers/auth-handlers";
import { devShowcaseHandlers } from "@/mocks/handlers/dev-showcase-handlers";
import { healthHandlers } from "@/mocks/handlers/health-handlers";
import { sessionHandlers } from "@/mocks/handlers/session-handlers";

export const mockRequestHandlers = [
  ...healthHandlers,
  ...authHandlers,
  ...activationHandlers,
  ...sessionHandlers,
  ...devShowcaseHandlers,
];
