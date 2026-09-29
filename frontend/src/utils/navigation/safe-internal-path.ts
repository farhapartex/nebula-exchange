export function safeInternalPath(candidatePath: string | null | undefined): string | null {
  if (
    !candidatePath ||
    !candidatePath.startsWith("/") ||
    candidatePath.startsWith("//") ||
    candidatePath.includes("\\")
  ) {
    return null;
  }
  return candidatePath;
}
