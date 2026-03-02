import type { ErrorPayload, ErrorSeverity } from "./errorBus";

export type LogSeverity = ErrorSeverity;
export type LogSource = "frontend" | "backend" | string;
export type LogMetadata = Record<string, unknown> | undefined;

export interface LogEntry {
  id: string;
  timestamp: string;
  severity: LogSeverity;
  message: string;
  context?: string;
  details?: unknown;
  source?: LogSource;
  metadata?: LogMetadata;
}

export interface LogEntryInput {
  message: string;
  severity?: LogSeverity;
  context?: string;
  details?: unknown;
  source?: LogSource;
  metadata?: LogMetadata;
  timestamp?: string;
  id?: string;
}

export interface ExportOptions {
  pretty?: boolean;
  severities?: LogSeverity[];
}

const DEFAULT_MAX_ENTRIES = 1_000;
const HARD_MAX_ENTRIES = 10_000;

const randomId = (): string => {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    return crypto.randomUUID();
  }
  return `log-${Date.now()}-${Math.random().toString(16).slice(2)}`;
};

export const clampLimit = (requested: number): number => {
  if (!Number.isFinite(requested) || requested <= 0) {
    return DEFAULT_MAX_ENTRIES;
  }
  return Math.min(Math.max(Math.floor(requested), 1), HARD_MAX_ENTRIES);
};

export const normalizeLogEntry = (entry: LogEntryInput): LogEntry => {
  const severity = entry.severity ?? "info";
  const timestamp = entry.timestamp ?? new Date().toISOString();
  return {
    id: entry.id ?? randomId(),
    timestamp,
    severity,
    message: entry.message,
    context: entry.context,
    details: entry.details,
    source: entry.source,
    metadata: entry.metadata,
  };
};

const filterEntriesBySeverity = (entries: LogEntry[], severities?: LogSeverity[]): LogEntry[] => {
  if (!severities || severities.length === 0) return entries;
  const allowed = new Set(severities);
  return entries.filter((entry) => allowed.has(entry.severity));
};

export const serializeEntries = (entries: LogEntry[], options?: ExportOptions): string => {
  const filtered = filterEntriesBySeverity(entries, options?.severities);
  const spacing = options?.pretty ? 2 : 0;
  return JSON.stringify(filtered, null, spacing);
};

export const fromErrorPayload = (
  payload: ErrorPayload,
  overrides?: Partial<Omit<LogEntryInput, "message" | "severity" | "context" | "details">>,
): LogEntryInput => ({
  message: payload.message,
  severity: payload.severity ?? "error",
  context: payload.context,
  details: payload.details,
  source: payload.source,
  metadata: payload.metadata,
  timestamp: payload.timestamp,
  id: payload.id,
  ...overrides,
});

// Append one or more log entries to an existing array, respecting the max limit.
export const appendLog = (current: LogEntry[], entry: LogEntryInput | LogEntryInput[], maxEntries: number): LogEntry[] => {
  const newEntries = Array.isArray(entry) ? entry : [entry];
  const normalized = newEntries.map(normalizeLogEntry);
  const combined = current.concat(normalized);
  const limit = clampLimit(maxEntries);
  const excess = combined.length - limit;
  return excess > 0 ? combined.slice(excess) : combined;
};

export const LOG_STORE_DEFAULT_LIMIT = DEFAULT_MAX_ENTRIES;
export const LOG_STORE_HARD_LIMIT = HARD_MAX_ENTRIES;
