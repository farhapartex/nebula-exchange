import { KeyRound, LogIn, MailCheck, Send, UserPlus, type LucideIcon } from "lucide-react";

export type AuthPage = {
  href: string;
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
    href: "/activate?token=demo-pending-activation-token",
    label: "Activate account (mock)",
    description: "A fresh activation link: loader, success message, then login after 3 seconds.",
    icon: MailCheck,
  },
  {
    href: "/activate?token=demo-already-activated-token",
    label: "Activate: already done (mock)",
    description: "A link for an account that is already active.",
    icon: MailCheck,
  },
  {
    href: "/activate?token=demo-expired-activation-token",
    label: "Activate: expired link (mock)",
    description: "An invalid or expired link redirects to the home page.",
    icon: MailCheck,
  },
  {
    href: "/resend-activation",
    label: "Resend activation",
    description: "Request a new activation link when the old one expired.",
    icon: Send,
  },
  {
    href: "/login",
    label: "Log in",
    description: "Sign in with email and password. Only activated accounts can log in.",
    icon: LogIn,
  },
  {
    href: "/forgot",
    label: "Forgot password",
    description: "Get a single-use reset link by email.",
    icon: KeyRound,
    plannedTask: "T-017",
  },
];
