import type { ReactNode } from "react";
import Link from "next/link";

import { NebulaLogo } from "@/components/brand/nebula-logo";

export default function AuthLayout({ children }: { children: ReactNode }) {
  return (
    <div className="flex flex-1 flex-col items-center px-4 py-10 sm:py-16">
      <NebulaLogo />
      <div className="mt-8 w-full max-w-md rounded-2xl border border-border bg-surface/80 p-6 shadow-2xl shadow-black/40 backdrop-blur sm:p-8">
        {children}
      </div>
      <Link href="/" className="mt-6 text-xs text-subtle transition-colors hover:text-muted">
        Back to build preview
      </Link>
    </div>
  );
}
