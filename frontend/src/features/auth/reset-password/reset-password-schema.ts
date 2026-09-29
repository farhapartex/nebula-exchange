import { z } from "zod";

import { maximumPasswordLength, minimumPasswordLength } from "@/utils/security/password-strength";

export const resetPasswordSchema = z
  .object({
    password: z
      .string()
      .min(minimumPasswordLength, `Use at least ${minimumPasswordLength} characters`)
      .max(maximumPasswordLength, `Use at most ${maximumPasswordLength} characters`),
    confirmPassword: z.string(),
  })
  .refine((formValues) => formValues.password === formValues.confirmPassword, {
    path: ["confirmPassword"],
    message: "Passwords don't match",
  });

export type ResetPasswordFormValues = z.infer<typeof resetPasswordSchema>;

export const resetPasswordFieldNames = ["password", "confirmPassword"] as const;
