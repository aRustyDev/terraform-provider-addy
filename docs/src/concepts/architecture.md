# Architecture Overview

This page provides a user-facing conceptual overview of the Terraform Addy provider’s internal architecture and the guarantees it aims to deliver.
It is intentionally narrative; authoritative operational MUST/SHALL rules live in the project’s policy file (`.rules`).
For design rationale details (the “why”) see `CLAUDE.md` (internal contributor doc).
This page helps you (a provider user) understand how the provider behaves and what to expect across versions.

---

## High-Level Goals

1. Accurate, idempotent representation of Addy objects (domains, aliases, recipients, rules, usernames) in Terraform state.
2. Minimal surprise: feature toggles (e.g., activation, catch-all) expressed as boolean attributes rather than separate imperative steps.
3. Safe, incremental growth: one endpoint (resource or data source) per minor release → predictable adoption curve.
4. Clear error surface: upstream API failures always surface non‑ambiguous diagnostics.
5. Observability-first: structured logs (debug) without leaking secrets; metrics & tracing may come in later phases.

---

## Layered Components

| Layer | Responsibility | Notes |
|-------|----------------|-------|
| Terraform Plugin Framework Layer | Delegates schema definitions, CRUD method invocation | Handles diff planning, state serialization |
| Provider Configuration | Captures API key, optional settings | API key precedence: config attribute overrides environment variable (policy) |
| HTTP Client Wrapper | Adds User-Agent, retry-on-429, timeout, error wrapping | Centralizes all network communication logic |
| Error Mapping Module | Loads static error meaning table | Produces human-friendly diagnostics with truncated body snippet |
| Model Translation Layer | Converts JSON → internal structs → Terraform attribute values | Ensures null vs empty distinctions preserved |
| Toggle Reconciliation | Decides if enable/disable endpoints must be invoked | Driven by diff (desired vs current) in Update path |
| Pagination Helpers (initial) | Pass-through of page_number / page_size parameters | Future: controlled “all” iteration with limit guard |
| Logging Hooks | Structured entries for each CRUD + retry attempt | Sanitization / size truncation enforced |
| Documentation & Examples | mdBook narrative + generated schema docs | Avoids duplication; schema tables live in generated docs |

---

## Request / Response Flow

1. **Plan Phase**
   Terraform config and prior state produce a desired diff. ForceNew attributes (immutable) trigger replacement if changed.

2. **Apply Phase (Create/Update/Delete)**
   a. Input model marshaled to JSON payload(s).
   b. HTTP client sends request with User-Agent `terraform-provider-addy/<version>`.
   c. On HTTP 429: centralized exponential backoff with jitter (bounded attempts & cumulative sleep).
   d. Non-2xx → error map consulted → diagnostic emitted (includes status code meaning + truncated body).
   e. Success path always ends with a **Read** for authoritative normalization.

3. **Read Phase (State Refresh)**
   a. GET endpoint invoked.
   b. JSON decoded → internal models.
   c. Null / missing fields represented as Terraform `null` (not empty strings).
   d. Booleans for toggles reflect server truth (source of authority).
   e. Terraform state updated atomically.

---

## Idempotency & Drift Avoidance

| Mechanism | Purpose |
|-----------|---------|
| Post-mutation Read | Absorb server defaults / normalizations (e.g., canonical casing) |
| ForceNew tagging | Ensures immutable fields trigger replacement instead of in-place mutation |
| Toggle reconciliation | Prevents unnecessary PATCH/POST calls when boolean already matches server |
| Distinct parse vs HTTP error diagnostics | Helps identify serialization bugs vs upstream failures |

---

## Error Handling Model

- **Single Load Mapping**: A static TOML (error codes → meaning) is loaded once; unknown codes labeled “Unexpected status code <n>”.
- **Diagnostic Fields**: Status code, mapped meaning, truncated body (≤300 bytes), operation context.
- **Parse Errors**: Labeled “Parse Error” (distinct from transport or API validation).
- **No Silent Recovery**: Retries only for 429; all terminal failures surface to Terraform.

---

## Retry & Rate Limit Strategy

| Aspect | Policy |
|--------|--------|
| Scope | Only HTTP 429 |
| Attempts | Up to 5 (exponential backoff) |
| Base Delay | 500ms * 2^(attempt-1) |
| Jitter | + random 0–100ms |
| Abort Conditions | Non-429 status or cumulative sleep > 15s |
| POST Retry | Only if endpoint documented idempotent |

Reasoning: targeted fairness with Addy API; avoids unintended duplicate creations.

---

## Pagination Behavior (Early Phase)

- User sets explicit `page_number` and `page_size` on collection data sources.
- Provider returns that page plus optional server metadata (when exposed) such as last_page / total_items.
- Deep auto-pagination & “all” semantics intentionally deferred until guard rails (limit) and performance considerations are validated.
- Rationale: predictable plan size; avoids large accidental state expansions.

---

## Logging & Observability (Phase 1)

Structured log fields (debug level typical):
- operation: create | read | update | delete | plan
- endpoint: API path (without host)
- status_code
- entity_id (when available)
- duration_ms (optional)

Redactions & Constraints:
- API key, bearer tokens, large/raw full bodies are never logged.
- Bodies (when needed) truncated to a safe byte limit and sanitized.

Future Phases (not yet guaranteed):
- Metrics (latency histogram, rate limit counters)
- Trace spans (OpenTelemetry) if/when stability achieved.

---

## Versioning & Release Cadence

| Change Type | Version Bump | Notes |
|-------------|--------------|-------|
| New resource or data source | MINOR | One per MINOR release (atomic) |
| Add computed-only attribute | PATCH | No new user-configurable surface |
| Breaking schema removal/rename | MAJOR | Avoid unless justified |
| Bug fix / performance improvement | PATCH | Must not add config knobs |

Release Tagging:
- Only CI creates tags after verifying tests, lint, coverage, and changelog.
- Local builds can use `-dev` or metadata suffixes; do not push tags manually.

---

## Documentation Layering (User-Facing Summary)

| Content | Location | Key Purpose |
|---------|----------|-------------|
| Narrative concepts (this page, pagination, retry, security, versioning) | mdBook (`docs/`) | Understand provider behavior & patterns |
| Schema attribute details | Generated provider docs | Exact argument/computed status, types |
| Internal design rationale | `CLAUDE.md` | Contributor reasoning; may evolve rapidly |
| Operational policies | `.rules` | Normative constraints for contributors |
| Troubleshooting (developer-facing) | `.claude/FAQ.md` | Internal pitfall resolution |

Avoid duplication: This mdBook page does not copy full attribute tables—link to generated docs instead.

---

## Security Considerations

- API key precedence ensures explicit Terraform config can override environment (useful for CI clarity).
- Sensitive fields never logged; detection relies on schema marking.
- No disk persistence of raw responses to reduce accidental leakage.
- Future enhancements may introduce validation of least-privilege scopes (if Addy API releases granular tokens).

---

## Extensibility & Future Enhancements

| Area | Potential Direction |
|------|---------------------|
| Pagination | Add safe “all” mode with explicit item cap |
| Retry | Adaptive jitter strategy, 5xx selective retries (opt-in) |
| Importers | Support `terraform import addy_domain.example <id>` etc. |
| Bulk Actions | Possible ephemeral action resources (create-only) |
| Metrics | Latency histograms, retry counters |
| Code Generation | Partial OpenAPI-driven scaffolding while retaining hand-tuned schema docs |

---

## Typical Lifecycle Example (Conceptual)

1. You define an `addy_domain` resource with desired toggles (e.g., `catch_all = true`).
2. Terraform plan marks resource for creation (no prior state).
3. Apply: provider issues Create → receives server response.
4. Immediate Read normalizes: server might supply timestamps, counters.
5. Subsequent apply with unchanged config results in “no changes”.
6. If you flip a boolean (e.g., `catch_all = false`), plan shows an in-place update triggering toggle endpoint(s); final Read re-synchronizes.

---

## Failure Scenario Examples

| Scenario | Surface Behavior | User Action |
|----------|------------------|-------------|
| 422 validation error | Diagnostic with status + mapped meaning + snippet | Adjust invalid attribute value |
| 404 on Read (after previously existing) | Provider interprets as drift (deleted externally) → plan may recreate (resource) | Re-apply if recreation desired, or remove from config |
| 429 rate limit spike | Automatic bounded retries with backoff | Usually no action; if persistent, reduce apply concurrency |
| Parse error (unexpected JSON shape) | “Parse Error” diagnostic | File issue—may indicate upstream API change |

---

## Migration Guidance (Preview Pattern)

For future MAJOR releases:
- A dedicated `migrations/<major>.md` page will describe schema changes, deprecated attributes, replacement patterns, and state migration steps.
- Minor releases will list newly added endpoints in a “Capabilities Growth” summary section in the mdBook landing page.

---

## How to Contribute Documentation (User-Level)

If you notice conceptual gaps:
1. Open an issue or PR adding / refining a page under `docs/src/concepts/`.
2. Keep examples scenario-focused; do *not* replicate provider schema tables.
3. Link to the generated docs for deep attribute details.
4. Provide anchor-friendly headings to enable stable cross-references.

---

## Glossary

| Term | Meaning |
|------|--------|
| ForceNew | Terraform must destroy & recreate resource if attribute changes |
| Toggle Reconciliation | Logic that enables/disables server flags to match desired state |
| Normalization Read | Post-mutation Read to absorb server canonical representation |
| Drift | Difference between Terraform state and current server resource state |
| Ephemeral Action Resource (future) | Create-only resource modeling transient batch operations |

---

## Feedback

Have suggestions or found discrepancies between behavior and documentation?
Open an issue in the repository with:
- Reproduction steps (Terraform config + provider version)
- Expected vs actual
- Relevant log snippet (sanitized)

Consistent reporting helps maintain accuracy and evolve this architecture coherently.

---

*Last updated:* (synchronize with future edits when material behavior changes.)
*Scope:* Conceptual, non-normative user guidance.
