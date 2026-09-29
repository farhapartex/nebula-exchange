import type { ActivatedAccount, ActivationPreview } from "@/features/auth/api/auth-types";
import { requestData } from "@/lib/api/api-client";

export function fetchActivationPreview(activationToken: string, signal?: AbortSignal): Promise<ActivationPreview> {
  return requestData<ActivationPreview>(`/auth/activations/${encodeURIComponent(activationToken)}`, {
    signal,
    skipSessionRefresh: true,
  });
}

export function activateAccount(activationToken: string): Promise<ActivatedAccount> {
  return requestData<ActivatedAccount>("/auth/activations", {
    method: "POST",
    body: { token: activationToken },
    skipSessionRefresh: true,
  });
}

export type ActivationEmailRequestResult = {
  accepted: boolean;
};

export function requestActivationEmail(emailAddress: string): Promise<ActivationEmailRequestResult> {
  return requestData<ActivationEmailRequestResult>("/auth/activation-emails", {
    method: "POST",
    body: { email: emailAddress },
    skipSessionRefresh: true,
  });
}
