import { cn } from "@/utils/class-names";

type PilotAvatarProps = {
  username: string;
  className?: string;
};

export function PilotAvatar({ username, className }: PilotAvatarProps) {
  const initials = username.slice(0, 2).toUpperCase();

  return (
    <span
      aria-hidden="true"
      className={cn(
        "flex size-8 shrink-0 items-center justify-center rounded-full bg-linear-to-br from-accent-strong to-highlight/70 text-xs font-semibold text-white",
        className,
      )}
    >
      {initials}
    </span>
  );
}
