import type { Metadata } from "next";

import { SignupForm } from "@/features/auth/signup/signup-form";

export const metadata: Metadata = {
  title: "Sign up",
};

export default function SignupPage() {
  return <SignupForm />;
}
