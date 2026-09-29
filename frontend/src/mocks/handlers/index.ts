import { authHandlers } from "@/mocks/handlers/auth-handlers";
import { devShowcaseHandlers } from "@/mocks/handlers/dev-showcase-handlers";
import { healthHandlers } from "@/mocks/handlers/health-handlers";

export const mockRequestHandlers = [...healthHandlers, ...authHandlers, ...devShowcaseHandlers];
