/**
 * Type definitions for backend events and payloads
 * Based on official TypeScript/Svelte type safety best practices
 */

/**
 * Structured log entry from backend logger
 */
export interface LogEntryPayload {
  message: string;
  severity?: 'debug' | 'info' | 'warn' | 'error' | 'fatal';
  source?: string;
  context?: string;
  details?: string;
  metadata?: Record<string, any>;
  timestamp?: string;
  id?: string;
}

/**
 * Statistics update payload
 */
export interface StatsPayload {
  scanned?: number;
  renamed?: number;
  skipped?: number;
  errors?: number;
  filesProcessed?: number;
  lastUpdate?: string;
  [key: string]: any;
}

/**
 * Type guard to check if value is LogEntryPayload
 */
export function isLogEntryPayload(value: unknown): value is LogEntryPayload {
  return (
    typeof value === 'object' &&
    value !== null &&
    'message' in value &&
    typeof (value as any).message === 'string'
  );
}

/**
 * Type guard to check if value is StatsPayload
 */
export function isStatsPayload(value: unknown): value is StatsPayload {
  return (
    typeof value === 'object' &&
    value !== null &&
    ('filesProcessed' in value || 'errors' in value)
  );
}
