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
