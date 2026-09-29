import Link from "next/link";

export function NebulaLogo() {
  return (
    <Link href="/" className="group flex items-center gap-2.5" aria-label="Nebula Exchange home">
      <span className="relative flex size-8 items-center justify-center rounded-lg bg-linear-to-br from-accent to-highlight shadow-[0_0_24px_-4px] shadow-accent/60">
        <span className="size-3 rounded-full bg-background ring-2 ring-foreground/80" />
      </span>
      <span className="text-sm font-semibold tracking-[0.18em] text-foreground uppercase">
        Nebula <span className="text-accent-soft">Exchange</span>
      </span>
    </Link>
  );
}
