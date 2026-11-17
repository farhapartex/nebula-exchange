export function safeInternalPath(candidatePath: string | null | undefined): string | null {
  if (!candidatePath || !candidatePath.startsWith("/") || candidatePath.startsWith("//")) {
    return null;
  }
  return candidatePath;
}
