import { describe, expect, it } from "vitest";

import { evaluatePasswordStrength } from "@/utils/security/password-strength";

describe("evaluatePasswordStrength", () => {
  it("rejects passwords under the minimum length", () => {
    expect(evaluatePasswordStrength("short1!").level).toBe(0);
  });

  it("rates longer and more varied passwords higher", () => {
    expect(evaluatePasswordStrength("aaaaaaaaaa").level).toBe(1);
    expect(evaluatePasswordStrength("aaaaaaaaaaaaaa").level).toBe(2);
    expect(evaluatePasswordStrength("Mining4Ore!x").level).toBe(3);
    expect(evaluatePasswordStrength("Mining4Crystal!Moon").level).toBe(4);
  });

  it("caps passwords with common words at weak", () => {
    const strength = evaluatePasswordStrength("MyPassword2026!!");
    expect(strength.level).toBe(1);
    expect(strength.hints).toContain("Avoid common words like password or qwerty");
  });
});
