import type { ReactNode } from "react";

type PreviewSectionProps = {
  title: string;
  description: string;
  children: ReactNode;
};

export function PreviewSection({ title, description, children }: PreviewSectionProps) {
  return (
    <section className="rounded-2xl border border-border bg-surface/70 p-5 sm:p-6">
      <header className="mb-5">
        <h2 className="text-base font-semibold text-foreground">{title}</h2>
        <p className="mt-1 text-sm text-muted">{description}</p>
      </header>
      {children}
    </section>
  );
}
