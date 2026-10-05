import Image from "next/image";

const emberPositions = [
  { left: "14%", delay: "0s", size: "size-1.5" },
  { left: "30%", delay: "1.4s", size: "size-1" },
  { left: "44%", delay: "2.8s", size: "size-2" },
  { left: "58%", delay: "0.7s", size: "size-1" },
  { left: "72%", delay: "3.6s", size: "size-1.5" },
  { left: "86%", delay: "2.1s", size: "size-1" },
  { left: "22%", delay: "4.4s", size: "size-1" },
  { left: "66%", delay: "5.2s", size: "size-1.5" },
];

const edgeFadeMask =
  "linear-gradient(to right, transparent, black 14%, black 86%, transparent), linear-gradient(to bottom, transparent, black 10%, black 80%, transparent)";

export function HeroCharacter() {
  return (
    <div className="relative mx-auto aspect-[2/3] w-full max-w-sm lg:max-w-md">
      <div
        className="absolute inset-0"
        style={{
          maskImage: edgeFadeMask,
          WebkitMaskImage: edgeFadeMask,
          maskComposite: "intersect",
          WebkitMaskComposite: "source-in",
        }}
      >
        <Image
          src="/landing/fighter.webp"
          alt="A thin, scarred boy in a burning village, looking back over his shoulder with a vengeful glare"
          fill
          priority
          sizes="(min-width: 1024px) 28rem, 24rem"
          className="object-cover"
        />
      </div>
      <div aria-hidden="true" className="pointer-events-none absolute inset-0 overflow-hidden">
        {emberPositions.map((ember) => (
          <span
            key={`${ember.left}-${ember.delay}`}
            className={`absolute bottom-16 rounded-full bg-orange-400 shadow-[0_0_8px_2px] shadow-orange-500/70 motion-safe:animate-ember-rise ${ember.size}`}
            style={{ left: ember.left, animationDelay: ember.delay }}
          />
        ))}
      </div>
    </div>
  );
}
