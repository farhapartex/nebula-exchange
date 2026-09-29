import { CircleCheck, Info, LoaderCircle, type LucideIcon } from "lucide-react";

import { cn } from "@/utils/class-names";

export type ActivationStage = "activating" | "activated" | "already_activated";

const stageAppearance: Record<
  ActivationStage,
  { icon: LucideIcon; iconClassName: string; title: string; description: string }
> = {
  activating: {
    icon: LoaderCircle,
    iconClassName: "animate-spin bg-accent/15 text-accent-soft",
    title: "Activating your account",
    description: "Hold tight while we confirm your email.",
  },
  activated: {
    icon: CircleCheck,
    iconClassName: "bg-up/10 text-up",
    title: "Account activated",
    description: "Your email is confirmed. You can log in now.",
  },
  already_activated: {
    icon: Info,
    iconClassName: "bg-info/10 text-info",
    title: "Already activated",
    description: "This account was activated earlier. Just log in.",
  },
};

export function ActivationStatusHeader({ stage }: { stage: ActivationStage }) {
  const appearance = stageAppearance[stage];
  const StageIcon = appearance.icon;

  return (
    <div className="text-center" aria-live="polite">
      <span
        className={cn("mx-auto mb-4 flex size-12 items-center justify-center rounded-full", appearance.iconClassName)}
      >
        <StageIcon className="size-6" />
      </span>
      <h1 className="text-xl font-semibold text-foreground">{appearance.title}</h1>
      <p className="mt-1 text-sm text-muted">{appearance.description}</p>
    </div>
  );
}
