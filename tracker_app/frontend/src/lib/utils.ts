export type Aggregation = {
  groupers: Record<string, any>;
  duration: number;
};

export type DataPoint = {
  label: string;
  value: number;
  category?: string;
};

export function formatDuration(seconds: number, hoursOnly?: boolean): string {
  const hours = Math.floor(seconds / 3600);
  if (hoursOnly) return hours < 1 ? "<1h" : `${hours}h`;
  const minutes = Math.floor((seconds % 3600) / 60);
  if (hours > 0) return `${hours}h ${minutes}m`;
  return `${minutes}m`;
}
