import { getAccessToken, refreshAccessTokenOnce } from "@/lib/api/access-token-store";
import { buildApiUrl, clientIdentificationHeader, idempotencyKeyHeaderName } from "@/lib/api/api-config";
import { ApiError, isErrorEnvelope } from "@/lib/api/api-error";
import type { DataEnvelope, ListEnvelope, PaginationParameters } from "@/lib/api/api-types";

type HttpMethod = "GET" | "POST" | "PUT" | "PATCH" | "DELETE";

type RequestOptions = {
  method?: HttpMethod;
  body?: unknown;
  query?: Record<string, string | number | boolean | null | undefined>;
  idempotencyKey?: string;
  signal?: AbortSignal;
  skipSessionRefresh?: boolean;
};

const stateChangingMethods: ReadonlySet<HttpMethod> = new Set(["POST", "PUT", "PATCH", "DELETE"]);

export function createIdempotencyKey(): string {
  return crypto.randomUUID();
}

function buildRequestUrl(path: string, query: RequestOptions["query"]): string {
  const requestUrl = buildApiUrl(path);
  if (!query) {
    return requestUrl;
  }
  const searchParameters = new URLSearchParams();
  for (const [parameterName, parameterValue] of Object.entries(query)) {
    if (parameterValue !== undefined && parameterValue !== null && parameterValue !== "") {
      searchParameters.set(parameterName, String(parameterValue));
    }
  }
  const queryString = searchParameters.toString();
  return queryString ? `${requestUrl}?${queryString}` : requestUrl;
}

function buildHeaders(method: HttpMethod, hasBody: boolean, idempotencyKey: string | undefined): Headers {
  const headers = new Headers({ Accept: "application/json" });
  if (hasBody) {
    headers.set("Content-Type", "application/json");
  }
  const accessToken = getAccessToken();
  if (accessToken) {
    headers.set("Authorization", `Bearer ${accessToken}`);
  }
  if (stateChangingMethods.has(method)) {
    headers.set(clientIdentificationHeader.name, clientIdentificationHeader.value);
    if (idempotencyKey) {
      headers.set(idempotencyKeyHeaderName, idempotencyKey);
    }
  }
  return headers;
}

async function readJsonBody(response: Response): Promise<unknown> {
  const responseText = await response.text();
  if (!responseText) {
    return null;
  }
  try {
    return JSON.parse(responseText);
  } catch {
    return null;
  }
}

async function sendRequest(path: string, options: RequestOptions): Promise<unknown> {
  const method = options.method ?? "GET";
  const idempotencyKey = stateChangingMethods.has(method)
    ? (options.idempotencyKey ?? createIdempotencyKey())
    : undefined;

  try {
    return await sendSingleRequest(path, method, idempotencyKey, options);
  } catch (requestError) {
    const isExpiredSession = requestError instanceof ApiError && requestError.statusCode === 401;
    if (!isExpiredSession || options.skipSessionRefresh) {
      throw requestError;
    }
    const refreshedAccessToken = await refreshAccessTokenOnce();
    if (!refreshedAccessToken) {
      throw requestError;
    }
    return sendSingleRequest(path, method, idempotencyKey, options);
  }
}

async function sendSingleRequest(
  path: string,
  method: HttpMethod,
  idempotencyKey: string | undefined,
  options: RequestOptions,
): Promise<unknown> {
  const hasBody = options.body !== undefined;

  let response: Response;
  try {
    response = await fetch(buildRequestUrl(path, options.query), {
      method,
      headers: buildHeaders(method, hasBody, idempotencyKey),
      body: hasBody ? JSON.stringify(options.body) : undefined,
      credentials: "include",
      signal: options.signal,
    });
  } catch (networkFailure) {
    if (networkFailure instanceof DOMException && networkFailure.name === "AbortError") {
      throw networkFailure;
    }
    throw new ApiError(0, "NETWORK_ERROR", "Could not reach the server. Check your connection and try again.");
  }

  const responseBody = await readJsonBody(response);
  if (response.ok) {
    return responseBody;
  }
  if (isErrorEnvelope(responseBody)) {
    const { code, message, details } = responseBody.error;
    throw new ApiError(response.status, code, message, details);
  }
  throw new ApiError(response.status, "INTERNAL_ERROR", "The server returned an unexpected response.");
}

export async function requestData<Payload>(path: string, options: RequestOptions = {}): Promise<Payload> {
  const responseBody = (await sendRequest(path, options)) as DataEnvelope<Payload> | null;
  return responseBody?.data as Payload;
}

export async function requestList<Item>(
  path: string,
  pagination: PaginationParameters = {},
  options: RequestOptions = {},
): Promise<ListEnvelope<Item>> {
  const responseBody = (await sendRequest(path, {
    ...options,
    query: { ...options.query, cursor: pagination.cursor, limit: pagination.limit },
  })) as ListEnvelope<Item>;
  return responseBody;
}

export async function requestAllPages<Item>(path: string, pageSize = 100): Promise<Item[]> {
  const collectedItems: Item[] = [];
  let cursor: string | null = null;
  do {
    const listPage: ListEnvelope<Item> = await requestList<Item>(path, { cursor, limit: pageSize });
    collectedItems.push(...listPage.data);
    cursor = listPage.pagination.next_cursor;
  } while (cursor);
  return collectedItems;
}
