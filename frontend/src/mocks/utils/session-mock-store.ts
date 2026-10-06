export function readSessionMockState<Value>(storageKey: string, fallbackValue: Value): Value {
  try {
    const storedValue = window.sessionStorage.getItem(storageKey);
    return storedValue ? (JSON.parse(storedValue) as Value) : fallbackValue;
  } catch {
    return fallbackValue;
  }
}

export function writeSessionMockState(storageKey: string, value: unknown) {
  try {
    window.sessionStorage.setItem(storageKey, JSON.stringify(value));
  } catch {}
}
