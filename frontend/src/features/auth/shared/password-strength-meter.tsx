import { cn } from "@/utils/class-names";
import { evaluatePasswordStrength, type PasswordStrengthLevel } from "@/utils/security/password-strength";

const levelBarClassNames: Record<PasswordStrengthLevel, string> = {
  0: "bg-border-strong",
  1: "bg-down",
  2: "bg-warning",
  3: "bg-info",
  4: "bg-up",
};

const levelLabelClassNames: Record<PasswordStrengthLevel, string> = {
  0: "text-subtle",
  1: "text-down",
  2: "text-warning",
  3: "text-info",
  4: "text-up",
};

const meterSegmentCount = 4;

export function PasswordStrengthMeter({ password }: { password: string }) {
  if (!password) {
    return null;
  }
  const strength = evaluatePasswordStrength(password);

  return (
    <div className="mt-1 space-y-1.5" aria-live="polite">
      <div className="flex gap-1" aria-hidden="true">
        {Array.from({ length: meterSegmentCount }, (_, segmentIndex) => (
          <span
            key={segmentIndex}
            className={cn(
              "h-1 flex-1 rounded-full transition-colors",
              segmentIndex < strength.level ? levelBarClassNames[strength.level] : "bg-border",
            )}
          />
        ))}
      </div>
      <p className="text-xs text-subtle">
        <span className={cn("font-medium", levelLabelClassNames[strength.level])}>{strength.label}</span>
        {strength.hints.length > 0 && <span> · {strength.hints[0]}</span>}
      </p>
    </div>
  );
}
