import { QueryClient } from "@tanstack/react-query";

import { isApiError } from "@/lib/api/api-error";

const nonRetryableStatusCodes = new Set([400, 401, 403, 404, 409, 422]);
const maximumQueryRetries = 2;

function shouldRetryQuery(failureCount: number, error: unknown): boolean {
  if (isApiError(error) && nonRetryableStatusCodes.has(error.statusCode)) {
    return false;
  }
  return failureCount < maximumQueryRetries;
}

export function createQueryClient(): QueryClient {
  return new QueryClient({
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
