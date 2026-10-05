export function toWorldX(simulationX: number, arenaWidth: number, worldUnitsPerPixel: number): number {
  return (simulationX - arenaWidth / 2) * worldUnitsPerPixel;
}

export function toWorldLength(simulationLength: number, worldUnitsPerPixel: number): number {
  return simulationLength * worldUnitsPerPixel;
}
