export type AccountStatus = "UNVERIFIED" | "ACTIVE" | "FROZEN" | "BANNED";

export type SignupRequest = {
  email: string;
  username: string;
  password: string;
  accepts_terms: boolean;
};

export type SignedUpAccount = {
  id: string;
  email: string;
  username: string;
  status: AccountStatus;
  activation_link_expires_at: string;
};

export type UserProfile = {
  id: string;
  email: string;
  username: string;
  status: AccountStatus;
  is_active: boolean;
  is_admin: boolean;
  created_at: string;
  last_login_at: string | null;
};

export type CurrentPlayer = {
  name: string;
  email: string;
  current_level: number;
  current_level_price: string | null;
  is_current_level_paid: boolean;
  story_level: number;
  total_win: number;
  total_lose: number;
  current_level_win: number;
  current_level_lose: number;
};

export type LoginRequest = {
  email: string;
  password: string;
};

export type EstablishedSession = {
  access_token: string;
  access_token_expires_at: string;
  user: UserProfile;
};

export type ActivationPreview = {
  email: string;
  username: string;
  is_activated: boolean;
};

export type ActivatedAccount = {
  email: string;
  username: string;
  status: AccountStatus;
  activated_at: string;
};
