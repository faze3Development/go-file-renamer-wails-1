export type ErrorContext = string | undefined;

/**
 * Supported severities for loggable errors, ordered by increasing criticality.
 *
 * - `debug`: Development-time diagnostics and verbose traces.
 * - `info`: Normal operational events worth surfacing in activity feeds.
 * - `warn`: Recoverable problems or unexpected states that need attention.
 * - `error`: Failures that impact the requested operation.
 * - `fatal`: Irrecoverable failures that leave the app in a bad state.
 */
export type ErrorSeverity = "debug" | "info" | "warn" | "error" | "fatal";

/**
 * Canonical schema for error payloads emitted throughout the frontend.
 *
 * This schema intentionally mirrors the shared log-entry contract so payloads can be
 * serialized directly into the ring buffer without losing context. Non-serializable
 * values in `details` or `metadata` should be sanitized before emitters forward them
 * to the backend or persistence layers.
 */
export interface ErrorPayload {
  /** Optional subsystem or operation label, e.g. "watcher". */
  context?: ErrorContext;
  /** Human-readable summary describing what went wrong. */
  message: string;
  /** Free-form structured data with additional diagnostic context. */
  details?: unknown;
  /** Severity classification; defaults to `error` when omitted. */
  severity?: ErrorSeverity;
  /** High-level origin of the event (frontend, backend, watcher, etc.). */
  source?: string;
  /** Machine-friendly metadata such as IDs, counts, or timing data. */
  metadata?: Record<string, unknown>;
  /** Optional ISO timestamp captured closer to the producer. */
  timestamp?: string;
  /** Optional producer-supplied identifier to promote deduplication. */
  id?: string;
}

export function toErrorPayload(error: unknown, context?: ErrorContext, severity: ErrorSeverity = "error"): ErrorPayload {
  if (typeof error === "string") {
    return { context, message: error, severity };
  }

  if (error instanceof Error) {
    return {
      context,
      message: error.message || "Unexpected error",
      details: error,
      severity,
    };
  }

  return {
    context,
    message: "Unexpected error",
    details: error,
    severity,
  };
}
