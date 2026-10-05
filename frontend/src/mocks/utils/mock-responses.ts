import { delay, HttpResponse } from "msw";

import type { ApiErrorCode, DataEnvelope, ErrorEnvelope, ListEnvelope } from "@/lib/api/api-types";

export const mockNetworkLatencyInMilliseconds = 400;

export async function simulateLatency(milliseconds = mockNetworkLatencyInMilliseconds) {
  await delay(milliseconds);
}

export function mockDataResponse<Payload>(payload: Payload, status = 200): Response {
  return HttpResponse.json<DataEnvelope<Payload>>({ data: payload }, { status });
}

export function mockErrorResponse(
  status: number,
  code: ApiErrorCode,
  message: string,
  details?: ErrorEnvelope["error"]["details"],
): Response {
  return HttpResponse.json<ErrorEnvelope>({ error: { code, message, details } }, { status });
}

export function mockListResponse<Item>(allItems: Item[], requestUrl: URL, defaultLimit = 20): Response {
  const requestedLimit = Number(requestUrl.searchParams.get("limit") ?? defaultLimit);
  const startIndex = Number(requestUrl.searchParams.get("cursor") ?? 0);
  const pageItems = allItems.slice(startIndex, startIndex + requestedLimit);
  const nextIndex = startIndex + requestedLimit;

  return HttpResponse.json<ListEnvelope<Item>>({
    data: pageItems,
    pagination: {
      next_cursor: nextIndex < allItems.length ? String(nextIndex) : null,
      limit: requestedLimit,
    },
  });
}
