export type DataEnvelope<Payload> = {
  data: Payload;
};

export type PageInfo = {
  next_cursor: string | null;
  limit: number;
};

export type ListEnvelope<Item> = {
  data: Item[];
  pagination: PageInfo;
};

export type ApiErrorCode =
  | "VALIDATION_FAILED"
  | "UNAUTHORIZED"
  | "FORBIDDEN"
  | "NOT_FOUND"
  | "METHOD_NOT_ALLOWED"
  | "CONFLICT"
  | "ACCOUNT_NOT_ACTIVE"
  | "INSUFFICIENT_FUNDS"
  | "INSUFFICIENT_ITEMS"
  | "MARKET_HALTED"
  | "BID_TOO_LOW"
  | "LIMIT_EXCEEDED"
  | "WALLET_REQUIRED"
  | "TWO_FA_REQUIRED"
  | "RATE_LIMITED"
  | "ENGINE_BUSY"
  | "INTERNAL_ERROR"
  | "SERVICE_UNAVAILABLE"
  | "NETWORK_ERROR";

export type ErrorEnvelope = {
  error: {
    code: ApiErrorCode;
    message: string;
    details?: Record<string, string> | Record<string, unknown>;
  };
};

export type PaginationParameters = {
  cursor?: string | null;
  limit?: number;
};
