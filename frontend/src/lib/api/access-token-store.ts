type AccessTokenRefresher = () => Promise<string | null>;

let currentAccessToken: string | null = null;
let accessTokenRefresher: AccessTokenRefresher | null = null;
let inFlightRefresh: Promise<string | null> | null = null;

export function getAccessToken(): string | null {
  return currentAccessToken;
}

export function setAccessToken(accessToken: string | null): void {
  currentAccessToken = accessToken;
}

export function registerAccessTokenRefresher(refresher: AccessTokenRefresher | null): void {
  accessTokenRefresher = refresher;
}

export function refreshAccessTokenOnce(): Promise<string | null> {
  if (!accessTokenRefresher) {
    return Promise.resolve(null);
  }
  inFlightRefresh ??= accessTokenRefresher().finally(() => {
    inFlightRefresh = null;
  });
  return inFlightRefresh;
}
