export const minimumPasswordLength = 10;
export const maximumPasswordLength = 128;

export type PasswordStrengthLevel = 0 | 1 | 2 | 3 | 4;

export type PasswordStrength = {
  level: PasswordStrengthLevel;
  label: "Too short" | "Weak" | "Fair" | "Good" | "Strong";
  hints: string[];
};

const commonPasswordFragments = ["password", "123456", "qwerty", "letmein", "nebula", "welcome", "admin"];

function countCharacterClasses(password: string): number {
  const characterClassPatterns = [/[a-z]/, /[A-Z]/, /\d/, /[^A-Za-z0-9]/];
  return characterClassPatterns.filter((characterClassPattern) => characterClassPattern.test(password)).length;
}

function containsCommonFragment(password: string): boolean {
  const lowercasePassword = password.toLowerCase();
  return commonPasswordFragments.some((commonFragment) => lowercasePassword.includes(commonFragment));
}

export function evaluatePasswordStrength(password: string): PasswordStrength {
  const hints: string[] = [];
  if (password.length < minimumPasswordLength) {
    hints.push(`Use at least ${minimumPasswordLength} characters`);
    return { level: 0, label: "Too short", hints };
  }

  const characterClassCount = countCharacterClasses(password);
  let score = 1;
  if (password.length >= 14) {
    score += 1;
  }
  if (characterClassCount >= 3) {
    score += 1;
  }
  if (characterClassCount === 4 || password.length >= 20) {
    score += 1;
  }
  if (containsCommonFragment(password)) {
    score = Math.min(score, 1);
    hints.push("Avoid common words like password or qwerty");
  }

  if (characterClassCount < 3) {
    hints.push("Mix upper and lower case letters, numbers and symbols");
  }
  if (password.length < 14) {
    hints.push("Longer passwords are much harder to guess");
  }

  const level = Math.min(score, 4) as PasswordStrengthLevel;
  const labels: Record<PasswordStrengthLevel, PasswordStrength["label"]> = {
    0: "Too short",
    1: "Weak",
    2: "Fair",
    3: "Good",
    4: "Strong",
  };
  return { level, label: labels[level], hints };
}
