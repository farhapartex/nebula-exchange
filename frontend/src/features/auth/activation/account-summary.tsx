import { Input } from "@/components/ui/input";
import type { ActivationPreview } from "@/features/auth/api/auth-types";

export function AccountSummary({ activationPreview }: { activationPreview: ActivationPreview }) {
  return (
    <div className="space-y-4 rounded-xl border border-border bg-background/50 p-4">
      <Input label="Email" value={activationPreview.email} readOnly disabled />
      <Input label="Username" value={activationPreview.username} readOnly disabled className="font-mono" />
    </div>
  );
}
