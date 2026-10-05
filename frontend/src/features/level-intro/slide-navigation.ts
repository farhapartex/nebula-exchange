export function nextSlideIndex(currentIndex: number, slideCount: number): number {
  return Math.min(currentIndex + 1, slideCount - 1);
}

export function previousSlideIndex(currentIndex: number): number {
  return Math.max(currentIndex - 1, 0);
}

export function isFinalSlide(currentIndex: number, slideCount: number): boolean {
  return currentIndex >= slideCount - 1;
}
