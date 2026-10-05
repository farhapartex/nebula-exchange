export const apiBaseUrl = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080/api/v1";

export const areApiMocksEnabled = process.env.NEXT_PUBLIC_API_MOCKS === "true";

export const clientIdentificationHeader = { name: "X-Nebula-Client", value: "web" } as const;

export const idempotencyKeyHeaderName = "Idempotency-Key";

export function buildApiUrl(path: string): string {
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;
  return `${apiBaseUrl}${normalizedPath}`;
}
