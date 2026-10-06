import { requestData } from "@/lib/api/api-client";

export type NextLevelStatus = "AVAILABLE" | "LOCKED" | "COMING_SOON";

export type NextLevel = {
  status: NextLevelStatus;
  chapter: {
    number: number;
    title: string | null;
    price: string | null;
    is_paid: boolean;
  };
  level: {
    id: string;
    number: number;
    title: string;
    teaser: string;
    time_limit_seconds: number;
    best_stars: number | null;
    attempts: number;
  } | null;
};

export const nextLevelQueryKey = ["me", "next-level"] as const;

export function fetchNextLevel(signal?: AbortSignal): Promise<NextLevel> {
  return requestData<NextLevel>("/me/next-level", { signal });
}
