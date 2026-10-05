import Image from "next/image";

const backdropFadeMask =
  "linear-gradient(to right, transparent, black 18%, black 82%, transparent), linear-gradient(to bottom, transparent, black 10%, black 90%, transparent)";

export function CallToActionBackdrop({ source }: { source: string }) {
  return (
    <div
      aria-hidden="true"
      className="pointer-events-none absolute inset-0 flex justify-center motion-safe:animate-fade-in"
    >
      <div
        className="relative h-full w-full sm:aspect-[864/1203] sm:w-auto"
        style={{
          maskImage: backdropFadeMask,
          WebkitMaskImage: backdropFadeMask,
          maskComposite: "intersect",
          WebkitMaskComposite: "source-in",
        }}
      >
        <Image
          src={source}
          alt=""
          fill
          priority
          sizes="(min-width: 640px) 70vh, 100vw"
          className="object-cover object-bottom"
        />
      </div>
      <div className="absolute inset-x-0 top-0 h-[55%] bg-linear-to-b from-background/80 via-background/40 to-transparent" />
    </div>
  );
}
