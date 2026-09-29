import { z } from "zod";

import { maximumPasswordLength, minimumPasswordLength } from "@/utils/security/password-strength";

export const usernamePattern = /^[A-Za-z0-9_]{3,20}$/;

export const signupSchema = z.object({
  email: z.email("Enter a valid email address").max(254, "Email is too long"),
  username: z.string().regex(usernamePattern, "Use 3 to 20 letters, numbers or underscores"),
  password: z
    .string()
    .min(minimumPasswordLength, `Use at least ${minimumPasswordLength} characters`)
    .max(maximumPasswordLength, `Use at most ${maximumPasswordLength} characters`),
  acceptsTerms: z.boolean().refine((hasAccepted) => hasAccepted, "You must be 18 or older and accept the Terms"),
});

export type SignupFormValues = z.infer<typeof signupSchema>;

export const signupFieldNames = ["email", "username", "password", "acceptsTerms"] as const;
