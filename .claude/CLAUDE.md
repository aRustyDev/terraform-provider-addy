---
doc_type: design-guidance
scope: terraform-provider-addy
purpose: architectural-and-modeling-rationale
authoritative_policies: ./.rules
authoritative_checklist: .claude/AGENT_CHECKLIST.md
method_tables_dir: .claude/examples/method-tables/
open_work_items: .claude/TODO.md
version: 3
last_updated: 2025-10-18
precedence: ".rules > CLAUDE.md (this file) > AGENT_CHECKLIST.md"
---

# CLAUDE.md
<!-- anchor:top -->

This document provides DESIGN RATIONALE for the Terraform provider for Addy.io.
It deliberately avoids duplicating normative operational rules (logging, security, versioning, testing) which now live in the project policy file (`.rules`).
Procedural iteration steps (HOW to implement each endpoint) live in `.claude/AGENT_CHECKLIST.md`.
Entity‑specific Method Plan Tables live in `.claude/examples/method-tables/`.

If any conflict arises:
1. `.rules` (policy / enforcement)
2. `CLAUDE.md` (design reasoning)
3. `AGENT_CHECKLIST.md` (execution workflow)

---

## Document Map
<!-- anchor:document-map -->
| Purpose | File |
|---------|------|
| Policies & enforcement | `.rules` |
| Design & modeling rationale | `CLAUDE.md` |
| Iteration workflow | `.claude/AGENT_CHECKLIST.md` |
| Per-entity method plans | `.claude/examples/method-tables/` |
| Roadmap / backlog | `.claude/TODO.md` |

---

## Design Principles
<!-- anchor:design-principles -->
1. Declarative Terraform state must reflect stable, server‑authoritative object snapshots.
2. JSON field fidelity: preserve semantics; map to Terraform attributes using snake_case.
3. Minimize surprise: toggles implemented as boolean attributes rather than separate resources.
4. Idempotency: every Create/Update ends with a Read to absorb server normalization.
5. Additive evolution: each new endpoint integration stands alone (supports frequent minor version increments).
6. Safety first: bulk destructive endpoints deferred until single-object lifecycle is proven stable.
7. Clear separation of concerns:
   - Policy (must / must not) → `.rules`
   - Rationale (why) → this file
   - Steps (how) → checklist.

---

## Endpoint Classification Rationale
<!-- anchor:endpoint-classification -->
| Category | Criteria | Examples |
|----------|----------|----------|
| Resource | Has POST or PATCH/DELETE enabling lifecycle mutation | domains, aliases, recipients, rules, usernames |
| Data Source (read-only) | Only GET endpoints, no mutation semantics | api-token-details, app-version, domain-options |
| Collection Data Source | GET list endpoints; may offer filtering/pagination | domains list, aliases list, recipients list |
| Deferred / Actions | Imperative batch or transient effect endpoints | bulk alias actions, restore/forget operations |

Decision Tree (simplified):
1. Does endpoint support creation or modification? → Resource.
2. Only retrieval? → Data source.
3. Returns a list? → Collection data source.
4. Performs non-stateful batch action? → Defer / consider future ephemeral resource.

---

## Toggle Endpoint Modeling
<!-- anchor:toggle-modeling -->
Activation / deactivation and capability toggle endpoints (e.g., active-aliases, catch-all-domains, encrypted-recipients) are **folded into boolean attributes**:
- Advantages: fewer Terraform objects; natural diff-based reconciliation.
- Implementation Pattern:
  - Desired vs current boolean evaluated during Update.
  - Invoke enable or disable endpoint accordingly.
  - Final Read to consolidate server state.
- Rationale: Terraform plan semantics remain clear; toggle operations become idempotent attribute transitions.

---

## Schema Modeling Rationale
<!-- anchor:schema-modeling -->
Attribute classification logic (normative rules in `.rules`, rationale here):

| Classification | Purpose | Example |
|----------------|---------|---------|
| Required + ForceNew | Immutable identifier; change requires replacement | `domain`, `username`, `alias.local_part (custom)` |
| Optional | User-managed mutable property | `description`, `from_name` |
| Optional + Computed | User can set; server may default/normalize | `active`, `catch_all`, `can_login` |
| Computed | Server-only / metrics / timestamps | `aliases_count`, `emails_forwarded`, `created_at` |
| Sensitive | Secrets or private material | `api_key`, possibly `public_key` (depending on exposure policy) |
| Nested Blocks | Structured arrays with validation | `conditions`, `actions`, `recipients` |

**Nullable Fields:** Represent with Terraform null (`types.StringNull()`, etc.) not empty strings.

**Recipient Sets vs Ordered Lists:**
- Use a Set for unordered unique IDs (`recipient_ids`).
- Use ListNestedAttribute when order matters (rule conditions sequence).

**Validation (Design):** Keep early validation minimal (enum sets & required presence). Defer complex regex or semantic validation to later phases to avoid premature brittleness.

---

## Retry & Backoff Design
<!-- anchor:retry-backoff -->
Rationale (policy specifics in `.rules`):
- Only 429 (rate limit) triggers exponential backoff + jitter.
- POST operations that are non-idempotent are **not** retried unless documented safe; better to implement idempotency tokens later if API evolves.
- Design Goals:
  - Prevent provider hammering the service.
  - Provide transparent diagnostics on terminal failure.
  - Keep implementation simple—no circuit breakers yet.

---

## Pagination Design
<!-- anchor:pagination -->
Rationale (implementation safety rules in `.rules`):
- Initial support: explicit `page_number`, `page_size`.
- Extended support: `all = true` with defensive `limit` to cap total items.
- Use server response metadata (e.g., `meta.last_page`) to stop iteration rather than speculative loops.
- Expose optional computed pagination metadata (total_items, last_page) for observability.
- Avoid deep pagination for massive collections until memory impact tested.
- Deterministic ordering: preserve server order; do not sort unless required for Set semantics.

---

## Error Handling Design
<!-- anchor:error-handling -->
Implementation principles (normative specifics in `.rules`):
- Central error map loaded from `errors.toml` gives human meaning per status code.
- Represent upstream errors verbatim + truncated body snippet to retain transparency.
- Parsing failures are logically distinct from HTTP failures (diagnostic: Parse Error).
- Unknown status codes surfaced plainly (no silent fallback).

---

## Security & Logging (Rationale)
<!-- anchor:security-logging -->
Policies enforced in `.rules`; rationale here:
- Logging scope minimized to operational metadata (operation, endpoint, status code).
- Security posture: never echo token or key material; sensitive attributes flagged to prevent accidental output.
- Progressive observability: start with structured logs; add metrics/tracing only after stability proven.

---

## Bulk Endpoints Design (Deferred)
<!-- anchor:bulk-endpoints -->
Bulk alias/domain operations (activate, delete, restore, recipient modifications) are **not** modeled initially:
- They represent transient batch actions rather than durable managed objects.
- Risk of large unintended state churn if integrated prematurely.
- Future Option: ephemeral “action” resource with ForceNew semantics per invocation (create-only).
- Current Action: Document deferral; keep single-object lifecycle robust first.

---

## Versioning Rationale
<!-- anchor:versioning -->
Normative mechanics in `.rules`; rationale here:
- Frequent MINOR increments (each endpoint integration) provides granular traceability of capability growth.
- Encourages small, reviewable PRs.
- PATCH reserved for non-surface changes (bug fixes, perf).
- Pre-release suffix (`-dev`, `-alpha.N`) clarifies readiness stage without fragmenting policy.
- CI-only tagging ensures reproducible releases and prevents inconsistent local tagging.

---

## Method Plan Tables
<!-- anchor:method-plan-tables -->
Per-entity detailed lifecycle mapping lives under:
`{{ method_tables_dir }}` (e.g., `domain.md`, `alias.md`, `recipient.md`, `rule.md`, `username.md`).

Each table defines:
- Endpoint → Terraform method mapping.
- Attribute ForceNew rationale.
- Toggle reconciliation flow.

Agents MUST consult the relevant method table before drafting schema code (see Step 2 in `AGENT_CHECKLIST.md`).

---

## Agent Workflow Pointer
<!-- anchor:workflow -->
Implementation procedure is NOT duplicated here.
See `.claude/AGENT_CHECKLIST.md` for:
1. CRUD capability derivation.
2. Schema drafting.
3. HTTP integration.
4. State population & drift control.
5. Exit criteria.

Design sections in this file are treated as read-only references during those steps.

---

## Future Enhancements (Design Outlook)
<!-- anchor:future-enhancements -->
1. **Importers:** Add import capability per resource once CRUD stable (expected order: domain, alias, recipient, rule, username).
2. **Selective Pagination Modes:** Allow user to opt into hash-based state diff suppression for large collections.
3. **State Delta Optimization:** Track ETag/Last-Modified for conditional GET (deferred).
4. **Retry Policy Extensions:** Introduce jitter distribution tuning & optional max cumulative delay override.
5. **Observability Phase 2:** Metrics: request latency histogram, rate limit counter.
6. **Bulk Reconsideration:** Evaluate demand; may add ephemeral action resource pattern.
7. **Automated Model Generation:** Evaluate partial codegen from OpenAPI without sacrificing maintainability clarity.

---

## Appendix: Anchor Index
<!-- anchor:appendix-anchors -->
Anchors (for programmatic retrieval):
- top
- document-map
- design-principles
- endpoint-classification
- toggle-modeling
- schema-modeling
- retry-backoff
- pagination
- error-handling
- security-logging
- bulk-endpoints
- versioning
- method-plan-tables
- workflow
- future-enhancements
- appendix-anchors

---
END OF CLAUDE.md
