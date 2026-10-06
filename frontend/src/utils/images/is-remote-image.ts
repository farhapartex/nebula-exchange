export function isRemoteImage(source: string): boolean {
  return /^https?:\/\//.test(source);
}
