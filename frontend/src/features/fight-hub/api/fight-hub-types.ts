export type FighterProfile = {
  fighter_level: number;
  experience: number;
  experience_to_next_level: number;
  wins: number;
  losses: number;
  coins: number;
};

export type StoryLevelStatus = "completed" | "current" | "locked";

export type StoryLevel = {
  id: string;
  number: number | null;
  title: string;
  teaser: string;
  status: StoryLevelStatus;
  best_stars: number | null;
};

export type StoryChapter = {
  id: string;
  number: number;
  title: string;
  is_free: boolean;
  is_unlocked: boolean;
  levels: StoryLevel[];
};

export type ToolKey = "bare_fists" | "iron_pipe" | "street_blade" | "scrap_shield";

export type EquippedTool = {
  id: string;
  key: ToolKey;
  name: string;
  mastery_percent: number;
};

export type LoadoutSlot = {
  slot_number: number;
  tool: EquippedTool | null;
};
