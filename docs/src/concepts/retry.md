# Retry & Backoff Strategy

This concept page explains how the Terraform Addy provider handles **rate limiting (HTTP 429)** and why
retry logic is deliberately narrow in scope. It is **user-facing narrative** (non‑normative).
Normative MUST/SHALL rules live in the project policy file; internal design rationale lives in `CLAUDE.md`.

---

## 1. Goals

| Goal | Description |
|------|-------------|
| Safety | Avoid hammering the Addy API when rate limits are encountered. |
| Predictability | Bounded, exponential delays with a strict upper cap on cumulative sleep time. |
| Transparency | Clear diagnostics when retries exhaust without success. |
| Simplicity | Centralized helper; no ad hoc sleep loops scattered across resources. |
| Idempotency Awareness | Avoid duplicating non-idempotent POST actions that could produce duplicates. |

---

## 2. Scope (What Is Retried)

| Status / Error Type | Retried? | Reason |
|---------------------|----------|-------|
| 429 (Rate Limit) | Yes | Explicit signal to slow down; safe to retry after backoff. |
| 5xx (Server Errors) | No (current phase) | Out of scope initially to reduce complexity & unintended load. |
| Network timeouts | Not (initial) | Handled by timeouts; may be revisited in future phases. |
| Non-2xx validation (422) | No | Immediate user/config issue; retrying wastes quota. |
| Parse / JSON errors | No | Indicates upstream format shift or bug; manual intervention preferred. |

---

## 3. Policy Summary

| Parameter | Value |
|-----------|-------|
| Applicable Status | 429 only |
| Max Attempts | 5 (initial call + up to 4 retries, or define as 5 total including first? -> Here: 5 total attempts) |
| Backoff Formula | `delay = 500ms * 2^(attempt-1)` |
| Jitter | Uniform random 0–100ms added per retry |
| Cumulative Sleep Cap | 15s (abort if exceeded before next attempt) |
| Abort Conditions | Non-429 response OR cumulative sleep exceeded |
| POST Retries | Only if endpoint is documented idempotent (otherwise aborted) |
| Logging | Each attempt logs operation, endpoint, status_code, attempt index, and delay applied |
| Body Inspection | Truncated safe snippet (≤300 bytes) for diagnostics; not for control flow |

> Note: “Attempt” numbering begins at 1 for the first (non-delayed) request; backoff applies starting before attempt 2.

---

## 4. Backoff Timeline Example

Example of a full 5-attempt sequence (excluding jitter):

| Attempt | Base Delay | Jitter (0–100ms) | Cumulative (approx, no jitter) |
|---------|-----------:|-----------------|--------------------------------|
| 1 (initial) | 0ms | 0 | 0ms |
| 2 | 500ms | +ε | ~0.5s |
| 3 | 1000ms | +ε | ~1.5s |
| 4 | 2000ms | +ε | ~3.5s |
| 5 | 4000ms | +ε | ~7.5s |

Even in the worst case (with jitter), total deliberate sleep stays well under the 15s cap, leaving margin for
possible future adjustments or network latency. If future phases expand attempt count or add 5xx handling,
the cumulative limit ensures bounded impact.

---

## 5. Why Only 429?

| Rationale | Explanation |
|-----------|-------------|
| Signal Strength | 429 unambiguously instructs the client to slow down; semantics are clear. |
| Avoid Masking Outages | Blanket 5xx retries can hide systemic platform issues and delay user awareness. |
| Complexity Control | Narrow scope simplifies reasoning and testing (deterministic coverage). |
| Fair Usage | Fewer aggressive retry loops prevents accidental self‑amplification under congestion. |

Future phases could selectively include idempotent-safe 503/504, but only with careful safeguards (see §10).

---

## 6. Idempotency & POST Requests

Retries for POST are risky if the server *might* have created the resource before the failure or responded 429 after partial processing.

| Scenario | Action |
|----------|--------|
| POST documented safe / idempotent (e.g., server uses request token) | May retry under 429 |
| POST not documented idempotent | Do NOT retry; emit diagnostic after first failure |
| GET / PATCH / DELETE (idempotent or safe enough) | Eligible for 429 retry per policy |

If you add a new POST endpoint to the provider and are unsure of idempotency guarantees, assume **NOT safe** and disable retry until confirmed.

---

## 7. Logging During Retries

Each retry attempt logs:
- operation (e.g., read, update)
- endpoint (path without host)
- attempt number
- planned backoff delay
- status_code (from prior attempt’s response)

Sensitive headers or payload data are never logged. The presence of rate limit headers (if any) may be logged in sanitized form in a future observability phase.

---

## 8. Diagnostics on Exhaustion

If all attempts fail with repeated 429:
- Final diagnostic includes: status code, meaning (“Rate limit exceeded” if mapped), last truncated body snippet, attempt count.
- Suggestion: reduce concurrent Terraform operations or space out applies.

If a non-429 terminates the sequence early:
- Immediate diagnostic surfaces that response (no further delay).

---

## 9. Testing Considerations (User Awareness)

Although internal tests mock retry behavior, as a user you might still observe:
- Slight apply delay when stressing rate limits.
- Multiple structured log lines for a single planned action.

You typically need not intervene; intervention is only required if repeated plans systematically hit the attempt cap, which may indicate over-concurrency or an upstream service quota change.

---

## 10. Future Enhancements (Not Yet Implemented)

| Potential Feature | Benefit | Considerations |
|-------------------|---------|----------------|
| Selective 5xx retry (e.g., 503, 504) | Handles transient upstream outages | Must ensure no hidden duplication side-effects |
| Adaptive Jitter Distribution | Better fairness under bursty contention | Complexity; marginal benefit until scale grows |
| Retry Budget Metrics | Help users tune concurrency | Requires metrics subsystem (observability Phase 2) |
| Idempotency Tokens | Safe POST retries by injecting client token | Dependent on Addy API feature support |
| Per-Endpoint Retry Overrides | Fine-grained control for advanced users | Risk of misconfiguration increasing load |

Any expansion will be clearly documented in CHANGELOG & this page.

---

## 11. Manual Mitigation Strategies for Users

| Symptom | Mitigation |
|---------|------------|
| Frequent 429 on applies involving many resources | Reduce `-parallelism` in `terraform apply` |
| Repetitive exhaustion on single large resource | Stagger changes; verify no external process is simultaneously modifying resources |
| Mixed workload interference (CI + local) | Schedule heavy plans to avoid overlap |

---

## 12. Troubleshooting Matrix

| Issue | Possible Cause | Action |
|-------|----------------|-------|
| Diagnostic: “Rate limit exceeded” after max attempts | Sustained quota pressure | Lower plan frequency; open ticket if persistent |
| No retry happened for 429 on POST | Endpoint not marked idempotent | Confirm upstream docs; open an issue if truly safe |
| Unexpected duplicate resource upstream | Non-idempotent POST was manually retried outside provider | Verify resource creation semantics; avoid manual re-apply loops |
| Long apply time spikes | Multiple retries due to 429 bursts | Inspect logs; tune concurrency or sequence large changes |

---

## 13. FAQ

**Q: Why not retry 500 or 502 errors?**
To avoid masking systemic outages; users should see failures promptly so they can decide whether to re-apply later.

**Q: Can I force more attempts via configuration?**
Not currently. A future enhancement might expose a capped override. For now, deterministic limits keep behavior predictable.

**Q: Does retry logic change plan results?**
No. It only affects the timing of network attempts. A final success yields the same state as if the first try succeeded.

**Q: How does this interact with Terraform’s own retry semantics?**
Terraform core does not automatically retry provider HTTP calls; this provider-level logic is the only retry layer for 429.

**Q: Are DELETE requests retried?**
Yes, if they encounter 429 and are considered safe (idempotent). A second attempt at deletion of the same resource should be harmless if the first was accepted—or will yield a not-found, which is final.

---

## 14. Change Log (Local to This Concept Page)

| Version | Change |
|---------|--------|
| 1 | Initial documentation of retry/backoff (429-only, exponential with jitter, POST idempotency constraints). |

---

_Last updated: (update this page if retry scope expands or parameters change)._
_Status: Phase 1 – 429-only retry active; broader transient error handling deferred._
