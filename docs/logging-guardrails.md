# Logging Guardrails

## Severity Taxonomy
- `debug`: High-volume diagnostic details useful during development or when explicit verbose logging is enabled.
- `info`: Routine operational events (e.g., successful actions, state transitions) that are useful for tracing user workflows.
- `warn`: Recoverable issues or unexpected states that merit attention but do not halt execution.
- `error`: Failures that prevent an operation from completing; surfaced prominently to end users.
- `fatal`: Critical failures that require immediate attention and typically stop the process or feature.

These severities are surfaced through `frontend/src/lib/errorBus.ts` and will drive filtering, highlighting, and routing decisions across the logging pipeline.

## Retention and Export Limits
- **Default retention**: Keep the most recent 1,000 log entries in memory. This strikes a balance between useful history and frontend performance.
- **User-configurable retention**: Allow the user to configure retention up to a soft ceiling of 5,000 entries. Beyond that, the UI warns that higher values may degrade responsiveness.
- **Hard ceiling**: Enforce an absolute cap of 10,000 entries regardless of user input to avoid runaway memory usage when the application runs for extended periods.
- **Export size limit**: Block exports larger than 25 MB (after serialization) to prevent disk exhaustion and keep downloads manageable.
- **Free-space check**: Before writing an export, verify at least 2× the intended file size is available on disk; cancel the export with guidance if the safety threshold is not met.

### Rationale
- The 1,000-entry default comfortably covers common debugging sessions without incurring noticeable UI cost during filtering or search.
- The 5,000-entry soft ceiling gives power users headroom while allowing the application to warn about performance implications.
- The 10,000-entry hard ceiling aligns with memory guardrails that keep the in-memory store under ~8–10 MB for typical log payloads, even on machines with limited RAM.
- The 25 MB export limit and 2× free-space guard provide a deterministic upper bound while accommodating extended sessions where logs are needed for support or forensic analysis.
