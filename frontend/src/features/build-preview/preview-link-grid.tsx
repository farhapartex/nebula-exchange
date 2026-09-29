import type { ReactNode } from "react";

type PreviewLinkGridProps = {
  title: string;
  description: string;
  children: ReactNode;
};

export function PreviewLinkGrid({ title, description, children }: PreviewLinkGridProps) {
  return (
    <section className="mb-10">
      <h2 className="text-lg font-semibold text-foreground">{title}</h2>
      <p className="mt-1 mb-4 text-sm text-muted">{description}</p>
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">{children}</div>
    </section>
  );
}
