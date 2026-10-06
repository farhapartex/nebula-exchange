import type { NextConfig } from "next";

import { buildSecurityHeaders } from "./src/lib/security/security-headers";

const nextConfig: NextConfig = {
  poweredByHeader: false,
  async headers() {
    return [
      {
        source: "/:path*",
        headers: buildSecurityHeaders({
          isDevelopment: process.env.NODE_ENV !== "production",
          apiBaseUrl: process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080/api/v1",
          assetOrigin: process.env.NEXT_PUBLIC_ASSET_ORIGIN,
          chainRpcUrl: process.env.NEXT_PUBLIC_CHAIN_RPC_URL ?? "http://localhost:8545",
        }),
      },
    ];
  },
};

export default nextConfig;
