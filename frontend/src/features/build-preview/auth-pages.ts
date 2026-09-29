import { KeyRound, LogIn, MailCheck, Send, UserPlus, type LucideIcon } from "lucide-react";

export type AuthPage = {
  href: `/${string}`;
  label: string;
  description: string;
  icon: LucideIcon;
  plannedTask?: string;
};

export const authPages: AuthPage[] = [
  {
    href: "/signup",
    label: "Sign up",
    description: "Create a pilot account with email, username and password.",
    icon: UserPlus,
  },
  {
    href: "/activate",
    label: "Activate account",
    description: "Opened from the activation link in the signup email.",
    icon: MailCheck,
    plannedTask: "T-013",
  },
  {
    href: "/resend-activation",
    label: "Resend activation",
    description: "Request a new activation link when the old one expired.",
    icon: Send,
    plannedTask: "T-013",
  },
  {
    href: "/login",
    label: "Log in",
    description: "Sign in with email and password, plus a TOTP code if 2FA is on.",
    icon: LogIn,
    plannedTask: "T-014",
  },
  {
    href: "/forgot",
    label: "Forgot password",
    description: "Get a single-use reset link by email.",
    icon: KeyRound,
    plannedTask: "T-017",
  },
];
