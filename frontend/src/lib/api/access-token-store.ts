type AccessTokenRefresher = () => Promise<string | null>;

const crossTabRefreshLockName = "nebula-session-refresh";

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
  const refresher = accessTokenRefresher;
  inFlightRefresh ??= runAcrossTabsExclusively(refresher).finally(() => {
    inFlightRefresh = null;
  });
  return inFlightRefresh;
}

async function runAcrossTabsExclusively(work: () => Promise<string | null>): Promise<string | null> {
  if (typeof navigator === "undefined" || !navigator.locks) {
    return work();
  }
  return await navigator.locks.request(crossTabRefreshLockName, work);
}
