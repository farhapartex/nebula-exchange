import { describe, expect, it } from "vitest";

import { resetPasswordSchema } from "@/features/auth/reset-password/reset-password-schema";

describe("resetPasswordSchema", () => {
  it("accepts matching passwords of valid length", () => {
    expect(
      resetPasswordSchema.safeParse({ password: "Mining4Crystal!Moon", confirmPassword: "Mining4Crystal!Moon" })
        .success,
    ).toBe(true);
  });

  it("rejects short and mismatched passwords", () => {
    const shortResult = resetPasswordSchema.safeParse({ password: "short", confirmPassword: "short" });
    const mismatchResult = resetPasswordSchema.safeParse({
      password: "Mining4Crystal!Moon",
      confirmPassword: "Mining4Crystal!Mars",
    });
    expect(shortResult.success).toBe(false);
    expect(mismatchResult.success).toBe(false);
    expect(mismatchResult.error?.flatten().fieldErrors.confirmPassword).toBeDefined();
  });
});
