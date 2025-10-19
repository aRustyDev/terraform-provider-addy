# Observability & Diagnostics

This concept page explains how the Terraform Addy provider surfaces runtime insight (logs now; metrics later) so you can understand, troubleshoot, and trust operations.
It is **user-facing narrative (non‑normative)**. Normative MUST/SHALL rules live in the project policy file (`.rules`). Internal contributor rationale resides in `CLAUDE.md`.

---

## 1. Objectives

| Goal | Description |
|------|-------------|
| Transparency | Make provider actions (CRUD, retries, pagination) visible without leaking secrets. |
| Root Cause Speed | Provide enough contextual fields to isolate issues quickly. |
| Performance Insight (Future) | Evolve toward lightweight metrics without premature complexity. |
| Incremental Adoption | Phase in observability components so stability remains high. |

---

## 2. Phased Roadmap

| Phase | Included | Deferred |
|-------|----------|----------|
| Phase 1 (Current) | Structured logging (CRUD, retry, error diagnostics) | Metrics, tracing |
| Phase 2 | Metrics (latency histogram, retry counts, rate-limit counters) | Distributed tracing |
| Phase 3 (Optional) | OpenTelemetry traces/span correlation | Advanced sampling, custom exporters |

Phase transitions will be announced via CHANGELOG and updates to this page.

---

## 3. Current Capabilities (Phase 1)

### Structured Logs
Each significant operation (create, read, update, delete) emits structured key/value logs:
- `operation` (create | read | update | delete | plan)
- `endpoint` (API path sans host, e.g. `/v1/aliases`)
- `status_code` (for network responses)
- `entity_id` (if known post-create/read)
- `duration_ms` (optional timing)
- Retry context (attempt number, delay) for 429 handling

### Diagnostics
Error diagnostics surfaced to Terraform users include:
- HTTP status code
- Mapped meaning (from static error table)
- Safe truncated body snippet (≤300 bytes)
- Distinction between *parse errors* and *HTTP errors*

---

## 4. Log Levels (Conceptual Mapping)

| Level | Intended Use (Provider) | Examples |
|-------|-------------------------|----------|
| ERROR | Terminal operation failure | HTTP 422, 404 (where unexpected), parse error |
| WARN  | Potentially recoverable anomalies (future) | Partial pagination truncation (if ever), approaching limit |
| INFO  | High-level lifecycle milestones (may remain minimal) | Resource created (optional) |
| DEBUG | Standard structured CRUD + retry detail | `operation=update endpoint=/v1/domains status_code=200` |
| TRACE | Future deep HTTP composition or payload shape introspection | Raw request/response metadata (sanitized) |

Current focus: DEBUG suffices for most troubleshooting; TRACE deferred until Phase 2+.

---

## 5. Privacy & Redaction Principles

| Principle | Outcome |
|-----------|---------|
| No secret echo | API keys / auth headers never logged. |
| Truncate large bodies | Prevent accidental leakage & large log volume. |
| Sanitization pass | Mask known sensitive tokens if echoed by upstream. |
| Minimal default verbosity | Avoid overwhelming users with noise. |

---

## 6. Sample (Illustrative) Log Lines

```
ts=2025-10-19T01:05:00Z level=debug operation=create endpoint=/v1/aliases status_code=201 entity_id=al_123 duration_ms=182
ts=2025-10-19T01:05:00Z level=debug operation=read endpoint=/v1/aliases/al_123 status_code=200 entity_id=al_123 duration_ms=34
```

Retry example:
```
ts=... level=debug operation=read endpoint=/v1/aliases status_code=429 attempt=1 next_delay_ms=500
ts=... level=debug operation=read endpoint=/v1/aliases status_code=200 attempt=2 duration_ms=95
```

Error diagnostic (Terraform output—not a log line):
```
Error: HTTP 422 Unprocessable Entity (validation failed); body: {"error":"alias already exists"...}
```

---

## 7. Metrics (Future Phase 2 Design Targets)

| Metric | Type | Purpose | Notes |
|--------|------|---------|-------|
| `addy_request_latency_ms` | Histogram | Identify slow endpoints | Tag by endpoint, operation |
| `addy_rate_limit_retries_total` | Counter | Monitor frequency of 429 recovery | Increments per retry attempt |
| `addy_requests_total` | Counter | Volume tracking | Partitioned by status class |
| `addy_errors_total` | Counter | Surface non-2xx failures | Tag by status code class |
| `addy_retry_exhaustions_total` | Counter | Detect chronic throttling | Trigger investigation |

Export path: standard OpenTelemetry metric pipeline (once implemented).
Initial configuration may allow opt-in via environment variable (e.g. `ADDY_METRICS=1`).

---

## 8. Tracing (Phase 3 Optional)

Potential span structure (illustrative):

| Span | Parent | Attributes |
|------|--------|------------|
| `terraform.apply` | (Terraform process) | workspace, run_id |
| `addy.provider.call` | `terraform.apply` | operation, endpoint, attempt, status_code |
| `addy.provider.retry` | `addy.provider.call` | attempt, backoff_ms |

OpenTelemetry exporters would be optional to prevent forcing additional dependencies.

---

## 9. Performance Considerations

| Concern | Mitigation |
|---------|------------|
| Log flood on large plans | Keep logs at DEBUG (user opt-in), avoid per-item verbose loops |
| High contention 429 storms | Central retry caps attempts & cumulative sleep |
| Metrics overhead (future) | Use lightweight sampling / static buckets |
| Tracing overhead (future) | Opt-in; disabled by default |

---

## 10. Drift vs Observability Signals

| Symptom | Observability Aid | Next Action |
|---------|-------------------|------------|
| Unexpected replacement plan | DEBUG logs show prior Read state; verify ForceNew logic | Audit config vs server normalization |
| Repeated 429 warnings | Retry log sequence length | Reduce apply concurrency |
| Slow apply | `duration_ms` anomaly for specific endpoint | Open issue if persistent |
| Unknown status code | Diagnostic entry with raw code | File issue; add to error map table |

---

## 11. Local Debugging Workflow

1. Set `TF_LOG=DEBUG` (or provider-specific env if added later).
2. Re-run `terraform plan` / `apply`.
3. Filter logs by `endpoint` to isolate problematic entity.
4. Examine last error diagnostic snippet (avoid copying secrets).
5. Reproduce with minimized configuration block.
6. If unresolved: file issue including anonymized log excerpt + provider version.

---

## 12. Anti-Patterns

| Anti-Pattern | Why Harmful | Preferred Approach |
|--------------|-------------|--------------------|
| Adding ad hoc `fmt.Println` in provider fork | Unstructured, may leak secrets | Use structured logging layer |
| Disabling truncation locally | Risk of secret leakage in shared logs | Preserve truncation; request feature if insufficient |
| Blind retries via shell scripts around `terraform apply` | Masks underlying cause; compounds load | Rely on built-in 429 retry & analyze logs |
| Parsing provider logs via brittle regex for metrics | Fragile against format shifts | Wait for native metrics phase |
| Forcing huge concurrency to “speed up” | Heightens 429 frequency | Right-size Terraform `-parallelism` |

---

## 13. Troubleshooting Matrix

| Issue | Likely Cause | Observability Clue | Resolution |
|-------|--------------|--------------------|-----------|
| “Parse Error” | Upstream response shape change | Error diagnostic label | Open issue; include truncated body |
| Frequent retry exhaustion | Sustained rate limit | Series of 429 debug logs ending with error | Throttle operations; contact support if abnormal |
| Slow create operations | API processing or network latency | High `duration_ms` for create | Compare vs baseline; escalate if persistent |
| Missing `entity_id` in logs | Create failed pre-ID extraction or read step | Absence in log | Inspect preceding error diagnostic |
| Intermittent 404 after create | Eventual consistency / replication delay | 404 read after 201 create | Re-running apply may resolve; open issue if frequent |

---

## 14. FAQ

**Q: Can I disable all logging?**
Minimal informational output is already low. Full suppression may impair support—feature not prioritized.

**Q: Will future metrics require additional dependencies?**
Likely optional modules; base provider will remain functional without them.

**Q: How large can truncated bodies get?**
Policy currently caps at ≤300 bytes (subject to change; clearly documented if updated).

**Q: Are request payloads ever logged?**
Not by default. Only high-level metadata is logged; future TRACE mode may log sanitized structural outlines (never full raw JSON).

---

## 15. Change Log (Local to Observability Page)

| Version | Change |
|---------|--------|
| 1 | Initial observability concept (phased plan, logging fields, future metrics & tracing). |

---

_Last updated: (update when metrics or tracing shipped).
Status: Phase 1 (structured logging only); metrics & tracing pending._
