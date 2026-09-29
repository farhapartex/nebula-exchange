import type { SignedUpAccount, SignupRequest } from "@/features/auth/api/auth-types";
import { requestData } from "@/lib/api/api-client";

export function signUp(signupRequest: SignupRequest): Promise<SignedUpAccount> {
  return requestData<SignedUpAccount>("/auth/signup", { method: "POST", body: signupRequest });
}
