import type { EstablishedSession, LoginRequest, UserProfile } from "@/features/auth/api/auth-types";
import { requestData } from "@/lib/api/api-client";

export function logIn(loginRequest: LoginRequest): Promise<EstablishedSession> {
  return requestData<EstablishedSession>("/auth/login", {
    method: "POST",
    body: loginRequest,
    skipSessionRefresh: true,
  });
}

export function refreshSession(): Promise<EstablishedSession> {
  return requestData<EstablishedSession>("/auth/refresh", { method: "POST", skipSessionRefresh: true });
}

export function fetchCurrentUser(signal?: AbortSignal): Promise<UserProfile> {
  return requestData<UserProfile>("/me", { signal });
}

export function logOut(): Promise<{ logged_out: boolean }> {
  return requestData<{ logged_out: boolean }>("/auth/logout", { method: "POST", skipSessionRefresh: true });
}
