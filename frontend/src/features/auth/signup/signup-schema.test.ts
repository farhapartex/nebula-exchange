import { describe, expect, it } from "vitest";

import { signupSchema } from "@/features/auth/signup/signup-schema";

const validSignup = {
  email: "pilot@nebula.test",
  username: "pilot_nova",
  password: "Mining4Crystal!Moon",
  acceptsTerms: true,
};

function fieldErrorsFor(signupValues: unknown) {
  const parseResult = signupSchema.safeParse(signupValues);
  return parseResult.success ? {} : parseResult.error.flatten().fieldErrors;
}

describe("signupSchema", () => {
  it("accepts a valid signup", () => {
    expect(signupSchema.safeParse(validSignup).success).toBe(true);
  });

  it("enforces the PRD username rules", () => {
    expect(fieldErrorsFor({ ...validSignup, username: "ab" }).username).toBeDefined();
    expect(fieldErrorsFor({ ...validSignup, username: "pilot-nova" }).username).toBeDefined();
    expect(fieldErrorsFor({ ...validSignup, username: "a".repeat(21) }).username).toBeDefined();
    expect(fieldErrorsFor({ ...validSignup, username: "Pilot_123" }).username).toBeUndefined();
  });

  it("requires a 10 character password, a valid email and accepted terms", () => {
    expect(fieldErrorsFor({ ...validSignup, password: "short" }).password).toBeDefined();
    expect(fieldErrorsFor({ ...validSignup, email: "not-an-email" }).email).toBeDefined();
    expect(fieldErrorsFor({ ...validSignup, acceptsTerms: false }).acceptsTerms).toBeDefined();
  });
});
