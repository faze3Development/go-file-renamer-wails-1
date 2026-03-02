# ADR: Log Retention And Export Limits

- **Date:** 2025-10-28
- **Status:** Accepted
- **Context:** We are introducing a centralized log store that aggregates frontend diagnostics and backend events for user-facing troubleshooting. Retention controls must balance observability needs against memory pressure, storage consumption, and export ergonomics for both desktop and future cloud integrations.

## Decision

1. **Default Retention:** Keep the in-memory ring buffer at **1,000 entries**. This holds common support cases (~1-2 MB) without overwhelming the UI.
2. **Soft User Limit:** Allow the user to raise retention up to **5,000 entries** via settings. This covers short-lived bursts while staying performant on mid-tier hardware.
3. **Hard Ceiling:** Clamp retention to **10,000 entries**, which caps memory usage near **25 MB** assuming ~2.5 KB per serialized row. The UI enforces this and trims any overflow.
4. **Export Guardrails:** Exports mirror the active ring buffer so their size naturally respects the 10k~25 MB ceiling. Before writing to disk we will add a free-space check to block exports when the target volume cannot accommodate the snapshot.

## Rationale

- **Performance:** Profiling shows list virtualization keeps 5k entries fluid; beyond 10k rendering slumps and GC churn spikes on low-power laptops.
- **Storage:** 25 MB fits comfortably within typical support artifacts (e-mail, ticket systems) while avoiding multi-hundred-MB dumps.
- **Predictability:** Fixed ceilings eliminate runaway growth when emitters misbehave, aligning with guardrails documented in `docs/logging-guardrails.md`.
- **User Experience:** Exposing a conservative default with a clear soft cap keeps the settings surface approachable while still enabling power users to stretch capacity when needed.

## Consequences

- Emitters must prune or summarize chatter before forwarding to avoid hitting the ceiling prematurely.
- Future disk persistence should reuse the same hard ceiling or introduce chunking/rotation.
- QA will track scenarios at default, soft max, and hard max to ensure trimming and serialization remain stable.
