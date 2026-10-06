const visibleLeadingCharacters = 6;
const visibleTrailingCharacters = 4;

export function shortenAddress(address: string): string {
  if (address.length <= visibleLeadingCharacters + visibleTrailingCharacters) {
    return address;
  }
  return `${address.slice(0, visibleLeadingCharacters)}…${address.slice(-visibleTrailingCharacters)}`;
}
