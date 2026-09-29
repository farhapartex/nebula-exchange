import { MutationCache, QueryClient } from "@tanstack/react-query";

import { isApiError } from "@/lib/api/api-error";
import { publishToastEvent } from "@/lib/notifications/toast-events";
import { describeWaitTime } from "@/utils/time/format-duration";

const nonRetryableStatusCodes = new Set([400, 401, 403, 404, 409, 422, 429]);
const maximumQueryRetries = 2;

export type MutationMetadata = {
  showsThrottlingInline?: boolean;
};

function shouldRetryQuery(failureCount: number, error: unknown): boolean {
  if (isApiError(error) && nonRetryableStatusCodes.has(error.statusCode)) {
    return false;
  }
  return failureCount < maximumQueryRetries;
}

function announceThrottling(error: unknown, metadata: MutationMetadata | undefined) {
  if (!isApiError(error) || !error.isThrottled || metadata?.showsThrottlingInline) {
    return;
  }
  const retryAfterSeconds = error.retryAfterSeconds;
  publishToastEvent({
    tone: "warning",
    title: "Slow down a little",
    description: retryAfterSeconds
      ? `Too many requests. Try again in ${describeWaitTime(retryAfterSeconds)}.`
      : "Too many requests. Please wait a moment and try again.",
  });
}

export function createQueryClient(): QueryClient {
  return new QueryClient({
    mutationCache: new MutationCache({
      onError: (error, _variables, _context, mutation) => {
        announceThrottling(error, mutation.meta as MutationMetadata | undefined);
      },
    }),
    defaultOptions: {
      queries: {
        staleTime: 30_000,
        refetchOnWindowFocus: false,
        retry: shouldRetryQuery,
      },
      mutations: {
        retry: false,
      },
    },
  });
}
