# Security & Sensitive Data Handling

This concept page explains how the Terraform Addy provider manages authentication material,
sensitive attributes, logging redaction, and error reporting. It is **user-facing narrative** and
non‑normative. Authoritative (MUST / MUST NOT) rules reside in the project policy file (`.rules`).
Internal design rationale for contributors lives in `CLAUDE.md`.

---

## 1. Goals

| Goal | Description |
|------|-------------|
| Least Exposure | Minimize persistence and propagation of secrets (API key) beyond what Terraform needs. |
| Deterministic Precedence | Clear, documented resolution order for API key value. |
| Redaction by Default | Structured logs never emit raw secrets or full HTTP bodies. |
| Transparent Failures | Errors include actionable context without leaking sensitive content. |
| Evolvability | Foundation supports later observability (metrics/tracing) without weakening secrecy posture. |

---

## 2. Authentication Model (Current Phase)

| Aspect | Behavior |
|--------|----------|
| Auth Mechanism | Single API key (string) provided to Addy API via header (e.g., `Authorization: <key>`) |
| Required at | Provider configuration (init) |
| Storage | Held in memory by the provider runtime only; not written to disk by the provider |
| Terraform State | API key is NOT stored in state (only in provider config) |
| Diagnostics | Never echo the full key; generic wording used for invalid or missing key |

---

## 3. API Key Resolution Order

Current normative policy (see `.rules`):
**Terraform configuration attribute `api_key` overrides the `ADDY_API_KEY` environment variable.**

| Source | Priority | Example |
|--------|----------|---------|
| Provider block attribute `api_key` | High | `provider "addy" { api_key = var.addy_api_key }` |
| Environment variable `ADDY_API_KEY` | Fallback | `export ADDY_API_KEY=...` |
| Absent sources | Error | Provider initialization diagnostic |

Why this order?
- Explicit configuration in code is treated as an intentional override (supports per-workspace divergence).
- Environment variable acts as a convenient default for local workflows and CI.

Potential future change (tracked as backlog): invert to prefer env for ephemeral override—would be clearly documented in CHANGELOG if adopted.

---

## 4. Sensitive Attributes & Schema Marking

| Attribute | Sensitivity | State Persistence |
|-----------|-------------|-------------------|
| `api_key` (provider config) | Secret | Not stored in resource state |
| Future cryptographic materials (e.g., public/private key pairs if introduced) | Evaluated individually | Only non-sensitive portions may appear |

Terraform schema marks sensitive fields (`Sensitive: true`) so CLI output and logs avoid plain display.
Users should still treat plan / apply logs as potentially sensitive if custom debug prints are introduced (avoid adding those).

---

## 5. Logging & Redaction

| Logging Element | Included? | Notes |
|-----------------|-----------|-------|
| Operation (create/read/update/delete/plan) | Yes | Structured field |
| Endpoint path (sans host) | Yes | e.g., `/v1/aliases` |
| HTTP status code | Yes | For success & errors |
| Duration (ms) | Optional | Helpful for performance tuning |
| Request/response bodies | Truncated (≤300 bytes) and sanitized | Only at debug or trace level |
| Authorization header / API key | Never | Explicitly excluded |
| Full untruncated body | Never | Risk of leaking sensitive data |

Truncation ensures uniform maximum size; sanitization strips or masks known secret markers (e.g., `"api_key":"***"` if echoed by server).

---

## 6. Error Diagnostics

| Principle | Implementation |
|-----------|----------------|
| Map status codes to human meaning | Static error map loaded once (TOML) |
| Truncate bodies | ≤300 bytes; binary or huge payloads cut off |
| Distinguish parse vs HTTP errors | “Parse Error” diagnostic separate from non-2xx codes |
| Avoid secret leakage | Body snippet filtered to remove token-like substrings |

Diagnostic Example (illustrative):

```
Error: addy_domain.domain_x: HTTP 422 Unprocessable Entity (validation failed); body: {"error":"invalid domain"...}
```

No secret fields appear; truncated with ellipsis when necessary.

---

## 7. Transport & Network Safeguards (Baseline)

| Safeguard | Purpose | Status |
|-----------|---------|--------|
| Timeout (reasonable default) | Prevent indefinite hangs | Implemented in HTTP client wrapper |
| Central retry (429 only) | Avoid ad hoc loops & duplication | Active (see Retry page) |
| TLS verification | Security baseline | Follows Go defaults (no custom disable flags) |
| Connection reuse | Performance optimization | Handled by Go `http.Client` pooling |

No unsafe toggles (e.g., “skip TLS verify”) are exposed to users at this phase to maintain secure defaults.

---

## 8. State & Drift Security Considerations

| Concern | Mitigation |
|---------|------------|
| External deletion causing hidden drift | Read returning 404 triggers resource recreation or state removal during subsequent plan |
| Toggle mismatches (e.g., re-enabled externally) | Update & Read phases reconcile server truth back into state |
| Replay / duplicate create on retry | POST not retried unless proven idempotent |

---

## 9. Example Minimal Secure Configuration

```hcl
variable "addy_api_key" {
  type      = string
  sensitive = true
}

provider "addy" {
  api_key = var.addy_api_key
}

# Optionally: store var.addy_api_key in a Terraform Cloud workspace variable or CI secret store.
```

Environment-based alternative (fallback):
```bash
export ADDY_API_KEY="sensitive_value_here"
terraform apply
```
(Do *not* hardcode secrets in VCS-tracked `.tf` files.)

---

## 10. Common Pitfalls & Resolutions

| Symptom | Likely Cause | Resolution |
|---------|--------------|-----------|
| “Missing or empty API key” diagnostic | Neither provider `api_key` nor env set | Set attribute or export `ADDY_API_KEY` |
| “Unauthorized” / 401 | Invalid / revoked key | Rotate key; update provider config |
| Repeated 429 | High concurrency or external workload collisions | Reduce apply parallelism; wait for quota recovery |
| Sensitive value appears in logs | Custom output (user code) or upstream echo | File an issue if provider leaked; audit custom tooling |

---

## 11. Threat Model Snapshot (Simplified)

| Threat | Provider Mitigation | Out of Scope |
|--------|---------------------|--------------|
| Accidental secret logging | Redaction + omission of headers | User-added debug prints |
| Overexposed error content | Truncation + filtering | Upstream server including secret in nonstandard field |
| State tampering (local) | Terraform handles state locking | External compromise of Terraform backend |
| Network MITM | TLS verification | User overriding system trust store maliciously |
| Replay of non-idempotent POST via retry | POST retries disabled unless safe | External script replays requests |

---

## 12. Future Enhancements (Planned / Considered)

| Feature | Benefit | Considerations |
|---------|---------|----------------|
| Secret scanning in CI (policies) | Prevent accidental commit of keys | Requires additional tooling integration |
| Metrics: auth failures count | Operations visibility | Observability phase prerequisite |
| OTel tracing (with span redaction) | Deep latency insight | Must guarantee zero secret leakage |
| Per-resource encryption flags (structured) | Finer user control / clarity | Driven by upstream API expansion |
| Idempotency tokens for safe POST retries | Reliability under transient rate limits | Requires upstream support |

Each enhancement would be surfaced in CHANGELOG + dedicated concept page updates.

---

## 13. User Responsibilities

| Area | Expectation |
|------|-------------|
| Secret Storage | Use secret managers or Terraform Cloud sensitive variables; avoid plaintext in VCS |
| Rotation | Update provider config promptly when rotating keys |
| Principle of Least Privilege | If Addy introduces scoped keys, choose minimal scopes needed for planned resources |
| Logging Hygiene | Do not pipe debug logs to unsecured channels |
| Issue Reporting | Redact keys when opening GitHub issues; share minimal failing config |

---

## 14. FAQ

**Q: Why does the provider prefer config attribute over environment variable?**
To ensure explicit Terraform code wins over ambient environment—reduces surprises when different shells have different env values.

**Q: Can I disable body truncation for deep debugging?**
Not currently; truncation avoids accidental leakage. You can request a feature for opt‑in extended diagnostics (still redacted).

**Q: Is the API key ever written to disk?**
Not by the provider. Terraform may cache plan files; avoid embedding keys directly in `.tf` to reduce exposure risk.

**Q: How do I rotate keys with minimal downtime?**
1. Add new key (variable) alongside old inside secret store.
2. Update `api_key` variable value.
3. Run `terraform apply`.
4. Revoke old key once applies succeed.

---

## 15. Local vs CI Considerations

| Context | Recommendation |
|---------|---------------|
| Local developer machine | Use environment variable or a local (ignored) `.tfvars` file with proper OS-level secret store integration |
| CI pipeline | Inject `ADDY_API_KEY` from secure secret manager; avoid printing its value |
| Shared workspaces | Prefer remote variable sets over embedding secrets in code |

---

## 16. Change Log (Local to Security Concept Page)

| Version | Change |
|---------|--------|
| 1 | Initial security concept documentation (key precedence, logging redaction, error truncation, threat snapshot). |

---

_Last updated: (update this line when security posture or precedence changes)._

_Status: Phase 1 – foundational secret handling & redaction active; metrics & tracing deferred._
