import { devShowcaseHandlers } from "@/mocks/handlers/dev-showcase-handlers";
import { healthHandlers } from "@/mocks/handlers/health-handlers";

export const mockRequestHandlers = [...healthHandlers, ...devShowcaseHandlers];
