import type { ApiErrorCode, ErrorEnvelope } from "@/lib/api/api-types";

export class ApiError extends Error {
  readonly statusCode: number;
  readonly code: ApiErrorCode;
  readonly details?: ErrorEnvelope["error"]["details"];

  constructor(statusCode: number, code: ApiErrorCode, message: string, details?: ErrorEnvelope["error"]["details"]) {
    super(message);
    this.name = "ApiError";
    this.statusCode = statusCode;
    this.code = code;
    this.details = details;
  }

  get retryAfterSeconds(): number | null {
    const retryAfterValue =
      this.details && "retry_after_seconds" in this.details ? this.details.retry_after_seconds : null;
    return typeof retryAfterValue === "number" && retryAfterValue > 0 ? retryAfterValue : null;
  }

  get isThrottled(): boolean {
    return this.code === "RATE_LIMITED" || this.code === "LOGIN_LOCKED";
  }

  get fieldErrors(): Record<string, string> {
    if (this.code !== "VALIDATION_FAILED" || !this.details) {
      return {};
    }
    return Object.fromEntries(
      Object.entries(this.details).filter((detailEntry): detailEntry is [string, string] => {
        return typeof detailEntry[1] === "string";
      }),
    );
  }
}

export function isApiError(error: unknown): error is ApiError {
  return error instanceof ApiError;
}

export function isErrorEnvelope(body: unknown): body is ErrorEnvelope {
  if (typeof body !== "object" || body === null || !("error" in body)) {
    return false;
  }
  const errorBody = (body as { error: unknown }).error;
  return typeof errorBody === "object" && errorBody !== null && "code" in errorBody && "message" in errorBody;
}
