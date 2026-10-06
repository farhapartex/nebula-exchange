type SecurityHeaderOptions = {
  isDevelopment: boolean;
  apiBaseUrl: string;
  assetOrigin?: string;
};

type ResponseHeader = { key: string; value: string };

function webSocketOriginFor(httpOrigin: string): string {
  return httpOrigin.replace(/^http/, "ws");
}

export function buildContentSecurityPolicy({ isDevelopment, apiBaseUrl, assetOrigin }: SecurityHeaderOptions): string {
  const apiOrigin = new URL(apiBaseUrl).origin;
  const imageSources = ["'self'", "data:", "blob:", ...(assetOrigin ? [new URL(assetOrigin).origin] : [])];
  const scriptSources = ["'self'", "'unsafe-inline'", ...(isDevelopment ? ["'unsafe-eval'"] : [])];
  const connectSources = [
    "'self'",
    apiOrigin,
    webSocketOriginFor(apiOrigin),
    ...(isDevelopment ? ["ws://localhost:*", "http://localhost:*"] : []),
  ];

  const policyDirectives: Record<string, string[]> = {
    "default-src": ["'self'"],
    "script-src": scriptSources,
    "style-src": ["'self'", "'unsafe-inline'"],
    "img-src": imageSources,
    "font-src": ["'self'"],
    "connect-src": connectSources,
    "worker-src": ["'self'"],
    "object-src": ["'none'"],
    "base-uri": ["'self'"],
    "form-action": ["'self'"],
    "frame-ancestors": ["'none'"],
  };

  return Object.entries(policyDirectives)
    .map(([directiveName, directiveSources]) => `${directiveName} ${directiveSources.join(" ")}`)
    .join("; ");
}

export function buildSecurityHeaders(options: SecurityHeaderOptions): ResponseHeader[] {
  const securityHeaders: ResponseHeader[] = [
    { key: "Content-Security-Policy", value: buildContentSecurityPolicy(options) },
    { key: "X-Content-Type-Options", value: "nosniff" },
    { key: "X-Frame-Options", value: "DENY" },
    { key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
    { key: "Permissions-Policy", value: "camera=(), microphone=(), geolocation=(), payment=()" },
  ];
  if (!options.isDevelopment) {
    securityHeaders.push({ key: "Strict-Transport-Security", value: "max-age=63072000; includeSubDomains; preload" });
  }
  return securityHeaders;
}
