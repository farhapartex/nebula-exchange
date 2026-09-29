import { requestData } from "@/lib/api/api-client";

export type TwoFactorSetupDetails = {
  secret: string;
  otpauth_url: string;
};

export type TwoFactorToggleResult = {
  two_factor_enabled: boolean;
};

export function beginTwoFactorSetup(): Promise<TwoFactorSetupDetails> {
  return requestData<TwoFactorSetupDetails>("/auth/2fa/setup", { method: "POST" });
}

export function enableTwoFactor(code: string): Promise<TwoFactorToggleResult> {
  return requestData<TwoFactorToggleResult>("/auth/2fa/enable", { method: "POST", body: { code } });
}

export function disableTwoFactor(code: string): Promise<TwoFactorToggleResult> {
  return requestData<TwoFactorToggleResult>("/auth/2fa/disable", { method: "POST", body: { code } });
}
