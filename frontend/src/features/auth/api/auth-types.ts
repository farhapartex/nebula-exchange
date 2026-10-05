export type AccountStatus = "UNVERIFIED" | "PENDING_PAYMENT" | "ACTIVE" | "FROZEN" | "BANNED";

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
  two_factor_enabled: boolean;
  created_at: string;
  last_login_at: string | null;
};

export type LoginRequest = {
  email: string;
  password: string;
};

export type EstablishedSession = {
  two_factor_required?: false;
  access_token: string;
  access_token_expires_at: string;
  user: UserProfile;
};

export type TwoFactorChallenge = {
  two_factor_required: true;
  challenge_token: string;
  challenge_expires_at: string;
};

export type LoginResponse = EstablishedSession | TwoFactorChallenge;

export function isTwoFactorChallenge(loginResponse: LoginResponse): loginResponse is TwoFactorChallenge {
  return loginResponse.two_factor_required === true;
}

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
