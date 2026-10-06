import Image from "next/image";

import { isRemoteImage } from "@/utils/images/is-remote-image";

const sceneFadeMask =
  "linear-gradient(to left, black 45%, transparent 100%), linear-gradient(to top, transparent 0%, black 18%, black 85%, transparent 100%)";

export function SlideSceneImage({ source }: { source: string }) {
  return (
    <div
      aria-hidden="true"
      className="pointer-events-none absolute inset-y-0 right-0 w-full opacity-35 motion-safe:animate-fade-in lg:w-[58%] lg:opacity-100"
      style={{
        maskImage: sceneFadeMask,
        WebkitMaskImage: sceneFadeMask,
        maskComposite: "intersect",
        WebkitMaskComposite: "source-in",
      }}
    >
      <Image
        src={source}
        alt=""
        fill
        priority
        unoptimized={isRemoteImage(source)}
        sizes="(min-width: 1024px) 58vw, 100vw"
        className="object-cover object-[50%_20%]"
      />
    </div>
  );
}
