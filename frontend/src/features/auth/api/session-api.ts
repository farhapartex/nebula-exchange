import type { CurrentPlayer, EstablishedSession, LoginRequest } from "@/features/auth/api/auth-types";
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

export function fetchCurrentPlayer(signal?: AbortSignal): Promise<CurrentPlayer> {
  return requestData<CurrentPlayer>("/me", { signal });
}

export function logOut(): Promise<{ logged_out: boolean }> {
  return requestData<{ logged_out: boolean }>("/auth/logout", { method: "POST", skipSessionRefresh: true });
}
