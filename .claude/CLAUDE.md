# CLAUDE.md

## Purpose
Implement and evolve a Terraform provider for Addy.io using the provided OpenAPI specification and example payloads contained under `.claude/examples/**`. This document defines operating rules, scope boundaries, design decisions, and references for autonomous or semi-autonomous agent iterations.

## High-Level Goals
- Deliver a correct, maintainable Terraform provider exposing Addy.io functionality (domains, aliases, recipients, rules, usernames, token details, etc.).
- Base all modeling decisions strictly on `openapi.yaml` and example responses.
- Avoid speculative fields and unimplemented endpoint assumptions.
- Provide incremental development guidance with a clear iteration workflow.

## Non-Goals (Extended)
- No guessing undocumented fields.
- No speculative endpoints absent from `openapi.yaml`.
- No rewriting OpenAPI spec.
- No modification or paraphrasing of the upstream license beyond relocation.
- No excessive production API calls (avoid hammering endpoints).
- No dynamic code generation until explicitly prioritized.
- No UI components or CLI wrappers outside Terraform provider scope.
- No premature performance micro-optimizations (connection pooling tweaks, concurrency) before correctness.
- No bulk endpoints implementation until single-object resources are stable.
- No transformation of error semantics into “friendlier” wording—surface raw meaning plus mapped explanation.
- No storing secrets or full raw API responses permanently in the repo.
- No tests that perform destructive actions outside acceptance runs gated by environment variable (e.g. `TF_ACC=1`).
- No introducing external dependencies unless needed.

# Repository Overview

## Repository Overview
- `build/main.go` — Entrypoint; serves provider; sets address and debug; address must be updated from tutorial placeholder.
- `internal/provider/` — Provider definition and configuration: schema, Metadata, DataSources, Resources registration, API key resolution.
- `internal/data/` — Data source implementations; most currently stubbed except `api_token_details`.
- `internal/resource/` — Resource implementations (currently only domain stub).
- `internal/utils/client.go` — HTTP client creation and raw request helper (`Curl`); needs timeout, retry, error mapping upgrades.
- `internal/about/version.go` — Version constant (single source of truth).
- `internal/about/license.go` — Embedded license text (should move to top-level `LICENSE` later).
- `.claude/examples/openapi.yaml` — Canonical API contract for modeling.
- `.claude/examples/errors.toml` — HTTP error code mapping to human meaning.
- `.claude/examples/responses/**` — Sample JSON payloads for schema modeling/inference.
- `.claude/examples/terraform/**` — Usage examples for Terraform configs (provider, domain).
- `.claude/TODO.md` — Roadmap tasks (group into phases).
- `README.md` — Currently minimal; must be expanded (installation, usage, env vars).

## Design Decisions
1. **Classification**
   - Mutable entities (domain, alias, recipient, rule, username) → Terraform resources.
   - Read-only informational endpoints (e.g. `api-token-details`, `app-version`, `domain-options`) → data sources.
   - List endpoints produce collection data sources (e.g. `domains`, `aliases`, `recipients`, `rules`, `usernames`, `failed-deliveries`).

2. **Activation / Toggle Endpoints**
   - Activation/deactivation (e.g., active domains, active aliases) modeled as boolean attributes (`active`, `catch_all`) rather than separate resources. Backend POST/DELETE or PATCH endpoints are abstracted under Create/Update methods.

3. **Encryption / Recipient Flags**
   - Recipient encryption-related toggles (inline encryption, protected headers, allowed reply/send) are represented as attributes in the recipient resource for cohesion.

4. **Complex Nested Structures**
   - Rules: `conditions` and `actions` implemented as nested blocks (`ListNestedAttribute`) with explicit `type`, `match`, `values` or `value` validation.

5. **Pagination**
   - Data sources listing large sets accept `page_size`, `page_number` initially; optional “full pagination” mode (iterate until last page) added later with safety guard (max pages or `limit`).

6. **Error Handling**
   - Centralized loader for `errors.toml` builds a map. All non-2xx responses map: status code + meaning + truncated body snippet.
   - 429 responses retried with exponential backoff + jitter (policy below).

7. **Environment vs Configuration**
   - API key resolution order: config attribute `api_key` overrides environment variable `ADDY_API_KEY`. Absence of either yields diagnostic error.

8. **Resource Import**
   - Planned (post-stability) for domain, alias, recipient, rule, username using natural IDs or external identifiers.

## Retry Policy (429 Only)
- Attempts: max 5
- Base delay: 500ms * 2^(attempt-1)
- Jitter: random 0–100ms added to each delay
- Abort if cumulative sleep exceeds 15s or non-429 encountered
- Do not retry POST that is not idempotent unless endpoint documented as safe; prefer implementing safe idempotency tokens later if required.

## Pagination Pattern (Pseudo-Code)

```
for page = 1; page <= max_pages; page++ {
  resp = GET /endpoint?page[number]=page&page[size]=size
  accumulate items
  if page >= last_page break
}
```
Guard with `max_pages` or user `limit` attribute to avoid unbounded loops.

# Pagination Strategy (OLD)
- Accept optional `page_size`, `page_number` arguments in collection data sources.
- For “list all” operations requiring multiple pages, iterate until last page (guard with max pages / user-specified limit).


## Versioning (OLD)
- Single source of truth: `internal/about/version.go`.
- Inject version into provider via `build/main.go`.

## Schema Modeling Guidelines
- Use `Computed` for server-populated timestamps and derived counts.
- Use `Optional + Computed` for toggles defaulted server-side (e.g., `active`, `catch_all`) with plan modifiers to retain explicit user intent.
- Use `Set` for unordered unique IDs (e.g., `recipient_ids`).
- Preserve JSON field naming where possible; convert dash/underscore to snake_case Terraform attribute naming.
- Represent nullable fields as Terraform null types (e.g., `types.StringNull()`).
- Provide `MarkdownDescription` for all attributes (supports future registry docs generation).
- Nested blocks for complex structures (conditions, actions, recipients).
- Sensitive fields: API key only.

## Logging Standards
- At start/end of each CRUD/Read: `tflog.Debug`.
- Include structured fields: `endpoint`, `method`, `status_code`, `duration_ms` (if measured).
- Never log API key or Authorization header.
- Use `tflog.Trace` for verbose request building diagnostics only when necessary.

# Logging (OLD)
- Use `tflog.Debug` for request lifecycle; avoid logging sensitive headers.
- Provide correlation fields (endpoint, method, status).

## Error Handling Standards
- All non-2xx: create diagnostic containing:
  - HTTP status
  - Mapped meaning (from `errors.toml`)
  - Short excerpt of body (first 300 bytes, sanitized)
- For JSON parsing errors: diagnostic “Parse Error” + original error message.
- For unknown status codes: fallback “Unexpected status code”.

# Error Handling (OLD)
- Load `errors.toml` once into a map.
- Wrap API errors: show HTTP code + mapped meaning.
- Distinguish retriable vs terminal errors (429 retriable with exponential backoff).

## Security Practices
- API key is sensitive attribute; never output in logs.
- Avoid persisting raw response bodies beyond immediate parsing.
- Consider adding User-Agent header (`terraform-provider-addy/<version>`).

# Security (OLD)
- API key resolution order: config attribute → environment variable.
- Add validation (length > 0).
- Do not print key.

## File & Naming Conventions
- Data source filenames: `<entity>_data_source.go`
- Resource filenames: `<entity>_resource.go`
- Type names: `<Entity>DataSource`, `<Entity>Resource`
- Provider type name: `addy`
- Attributes: snake_case
- Avoid abbreviations unless conventional (e.g., `id`, `api_key`).

# Provider Address (OLD)
- Use canonical address or development override documented in README.

# Rate Limits & Retries (OLD)
- Backoff on 429: e.g. base 500ms doubling up to max 8s.
- Abort on 4xx (non-429) unless remediation clear.

# Bulk Endpoints (OLD)
- Defer until core single-object resources complete.
- Possibly expose as separate data sources returning aggregated status objects.

# Code Quality & Conventions (OLD)
- File naming: `*_data_source.go`, `*_resource.go`.
- Type names: `<prefix><Entity>DataSource`.
- Consistent snake_case in Terraform attribute names.
- Keep OpenAPI field naming unless conflicts (then document translation).

## Workflow
Reference external checklist for iteration steps (see `AGENT_CHECKLIST.md`). Each pass must complete all required criteria before proceeding.

## Examples
Current examples live at `.claude/examples/terraform`. Expand gradually:
- Provider config with env var fallback.
- Minimal domain resource, then extended version.
- Alias creation referencing domain (once implemented).
- Recipient with encryption flags.
- Rule with nested conditions/actions.

## Testing Strategy
- Unit tests: isolate Curl wrapper, mock transport via custom `http.RoundTripper`.
- Acceptance tests: gated by `TF_ACC=1` environment variable; require valid `ADDY_API_KEY`.
- Golden file state snapshots (optional) for complex nested structures; tests for schema JSON state outputs using example responses.
- Linting: integrate `golangci-lint` for static checks.

## Future Enhancements (Deferred)
- Bulk endpoints modeling once base CRUD stable.
- Code generation from OpenAPI (evaluate benefit vs maintenance).
- Conditional GET (ETag/Last-Modified) for bandwidth efficiency.
- Incremental refresh logic (only fetch if timestamp changed).
- Structured metrics (rate limit counters, cache hits).

## Roadmap Reference
See `.claude/TODO.md` for grouped, phased actionable tasks.

## Checklist Location
The operational iteration checklist is maintained in `.claude/AGENT_CHECKLIST.md`. Agents should treat that as the authoritative step sequence. This document references it but does not duplicate it to prevent divergence.

---
<!--OLD STUFF BELOW-->
---
# Purpose
Explain the goal: Implement a Terraform provider for Addy.io using only assets in `.claude/examples/`.



# Data Sources vs Resources
List endpoints and classify:
- Data Sources: api-token-details, account-details, account-notifications, domain-options, app-version, aliases (read-only list), alias (single), domains (list), domain (single), recipients (list), rules (list), usernames (list), failed-deliveries (list/single).
- Resources: domain, alias, recipient, rule, username (CRUD + activation toggles), encryption configurations (could be nested attributes on recipient instead of standalone resources).

# Implementation Roadmap
1. Fix provider address & attribute path bug.
2. Implement shared HTTP client wrapper features (timeouts, retry on 429, mapping errors.toml).
3. Complete Api Token Details (already done; refine).
4. Implement Domain Options data source (schema from sample).
5. Implement Domain resource (Create/Read/Update/Delete + catch_all, activation).
6. Add Alias resource.

10.
11. Documentation & examples.




# Future Enhancements
- Code generation from OpenAPI.
- Bulk action resources or data sources.
- Import support.
- State refresh optimization (conditional retrieval based on ETag or timestamp).
