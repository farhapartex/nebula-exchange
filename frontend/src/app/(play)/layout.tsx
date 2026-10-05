import type { ReactNode } from "react";

import { RequireAuthentication } from "@/features/auth/session/require-authentication";

export default function PlayLayout({ children }: { children: ReactNode }) {
  return <RequireAuthentication>{children}</RequireAuthentication>;
}
