type StatFormatter = (value: number) => string;

function signed(value: number): string {
  return value > 0 ? `+${value}` : `${value}`;
}

function percent(value: number): string {
  return `${Math.round(Math.abs(value) * 100)}%`;
}

const statFormatters: Record<string, StatFormatter> = {
  damage_bonus: (value) => `${signed(value)} damage`,
  reach_bonus: (value) => `${signed(value)} reach`,
  swing_speed_modifier: (value) => `${percent(value)} ${value < 0 ? "slower" : "faster"} swings`,
  block_damage_reduction: (value) => `${percent(value)} less damage when blocking`,
  block_stamina_cost_reduction: (value) => `${percent(value)} less stamina to block`,
};

function readableStatName(statKey: string): string {
  const spacedName = statKey.replaceAll("_", " ");
  return spacedName.charAt(0).toUpperCase() + spacedName.slice(1);
}

export function toolStatLines(baseStats: Record<string, number>): string[] {
  return Object.entries(baseStats).map(([statKey, statValue]) => {
    const formatter = statFormatters[statKey];
    return formatter ? formatter(statValue) : `${readableStatName(statKey)}: ${statValue}`;
  });
}
