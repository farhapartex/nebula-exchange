import { setupWorker } from "msw/browser";

import { mockRequestHandlers } from "@/mocks/handlers";

export const mockServiceWorker = setupWorker(...mockRequestHandlers);
