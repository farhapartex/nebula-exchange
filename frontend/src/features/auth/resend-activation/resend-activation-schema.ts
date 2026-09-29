import { z } from "zod";

export const resendActivationSchema = z.object({
  email: z.email("Enter a valid email address"),
});

export type ResendActivationFormValues = z.infer<typeof resendActivationSchema>;

export const resendActivationFieldNames = ["email"] as const;
