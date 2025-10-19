# Versioning & Upgrade Guide

This page explains how the Terraform Addy provider is versioned, what each segment of the version means, how frequently you can expect releases, and how to safely upgrade.

---

## 1. Semantic Versioning Overview

The provider uses **Semantic Versioning**: `MAJOR.MINOR.PATCH`

| Segment | Meaning | Examples |
|---------|---------|----------|
| MAJOR   | Backward incompatible changes to user-visible schema or behavior (attribute removal/rename/type change) | 2.0.0 |
| MINOR   | Addition of a new endpoint (resource or data source) or significant new nested block structure | 1.4.0 |
| PATCH   | Bug fixes, performance improvements, docs-only updates, adding computed (non-configurable) attributes | 1.3.2 |

Key principle: **Exactly one new endpoint per MINOR release** (atomic growth). This ensures you can easily audit what changed just by looking at the MINOR bump.

---

## 2. Release Cadence & Philosophy

| Aspect | Policy / Expectation |
|--------|----------------------|
| Atomic Growth | One new endpoint per MINOR release enables small, reviewable diffs |
| Frequency | MINOR releases can be frequent as new endpoints are integrated |
| Stability Bias | Avoid MAJOR bumps unless a clear, necessary breaking change arises |
| Transparency | Each release has a CHANGELOG entry summarizing Added / Changed / Fixed |
| Pre-Releases | Local builds may use `-dev`, `-alpha.N`, `-rc.N` suffixes; official tags are CI-created only |

Why small MINOR steps? Easier regression isolation, lower upgrade friction, simple mental model: “MINOR = 1 new capability.”

---

## 3. Provider Version vs Terraform State

The provider version you select determines:
- Which endpoints (resources/data sources) you can configure.
- Which attributes (and semantics) exist for each entity.

Terraform state contains resource instances keyed by type & provider version constraints. Upgrading the provider:
- Does NOT inherently force resource recreation (unless schema changes require it).
- May surface new computed fields or default values after refresh.

---

## 4. Pinning & Upgrading

### Recommended Pinning Pattern

```hcl
terraform {
  required_providers {
    addy = {
      source  = "registry.terraform.io/addy-io/addy"
      version = "~> 1.3" # Accepts >=1.3.0, <2.0.0
    }
  }
}
```

| Strategy | Use When | Trade-Off |
|----------|----------|-----------|
| Exact `= 1.3.2` | Reproducibility critical (e.g., regulated workload) | No automatic bug fix adoption |
| Pessimistic `~> 1.3` | Stable but accept patch & later minor updates | New endpoints appear (one per minor) |
| Broad `>= 1.3.0` | Rapid adoption / internal testing | Risk of unreviewed capability drift |

### Safe Upgrade Workflow

1. Review CHANGELOG for target version range.
2. Apply in a test workspace (or run plan) to see drift.
3. If adding new endpoint: introduce config incrementally (optional).
4. For MAJOR (rare): read migration guide (under `migrations/` directory) before applying.

---

## 5. What Triggers a Breaking (MAJOR) Change?

| Change Type | Classification | Justification |
|-------------|----------------|---------------|
| Remove attribute | MAJOR | Config referencing it becomes invalid |
| Rename attribute | MAJOR | Plan fails unless config updated |
| Change attribute type (e.g., string -> bool) | MAJOR | Incompatible state & config interpretation |
| Behavior change causing different plan for identical config (non-toggle) | Case-by-case | Assessed if user intent interpretation shifts |
| Mark attribute deprecated but still accepted | NON-BREAKING (deprecation phase) | Provide warning, future migration path |

The goal is to *avoid* MAJOR bumps unless they unblock long-term maintainability or align with upstream API shifts.

---

## 6. Deprecation Policy

| Phase | Description | User Action |
|-------|-------------|-------------|
| Announcement | Attribute or pattern flagged “Deprecated” in docs | None required yet |
| Dual Support | Old + new attribute work; warnings may appear | Migrate config at convenience |
| Removal (MAJOR) | Deprecated element removed | Must have migrated before upgrading |

Deprecations will appear in:
- Release CHANGELOG (Deprecated section)
- Attribute descriptions (Deprecated: use `new_attr`)

---

## 7. Computed Attribute Additions

Adding purely computed (read-only) attributes is a **PATCH** release if:
- No new user-supplied argument is introduced.
- No drift is forced (i.e., defaults or server-supplied values already existed; now just visible).

After upgrade, a `terraform plan` may show “no changes,” but `terraform state show` will include new fields after refresh.

---

## 8. Handling New Endpoints (MINOR Increase)

When you upgrade to a new MINOR:
- Existing configurations remain valid; no forced changes.
- You can begin using the newly added resource or data source.
- If you do nothing, the plan should still report “no changes.”

---

## 9. Migration Guides

For each MAJOR release (`2.0.0`, `3.0.0`, ...):
- A migration document appears in `docs/src/migrations/<major>.md`.
- It covers removed/renamed attributes, replacement patterns, and state handling notes.
- MAJOR releases are deliberately infrequent.

If a prospective change *might* be MAJOR:
- A “Migration Preview” section may be added to the mdBook ahead of time.

---

## 10. Example: Assessing an Upgrade

Suppose you run `terraform init -upgrade` and the provider moves from `1.2.0` → `1.4.0`:
1. Check CHANGELOG for 1.3.0 and 1.4.0 entries.
2. For each new endpoint (one per MINOR), decide if you want to adopt it now or later.
3. Run `terraform plan`.
   - EXPECTED: No forced diffs unless you add new config referencing new resource/data source or a bug fix modifies a previously incorrect computed value.
4. Apply confidently if plan is empty or intentional.

---

## 11. Versioning & Retry / Pagination Semantics

Capability evolution (e.g., adding future `all` pagination or refined retry coverage) will:
- Start as a MINOR if it introduces new user-exposed arguments (e.g., `all = true`).
- Remain PATCH if internal logic improves without surface change (e.g., refined jitter distribution).
- Leverage documentation updates (pagination or retry concept pages) so behavior changes are explicit.

---

## 12. State Compatibility Guarantees

| Guarantee | Scope |
|-----------|-------|
| Pessimistic constraint (`~> 1.x`) protects against accidental MAJOR upgrade | Terraform config |
| No silent data loss on PATCH/MINOR | Managed attributes |
| ForceNew semantics only expand (never silently removed) in PATCH/MINOR | Resource lifecycles |
| MAJOR change includes explicit migration instructions | Migrations page |

---

## 13. Interaction With Terraform Core

| Requirement | Value / Policy |
|-------------|----------------|
| Minimum Terraform Core version | 1.6.0 |
| Reason | Ensures availability of current Plugin Framework features |
| Future increases | Will be MINOR if non-breaking for existing state; MAJOR if it breaks older workflows |

---

## 14. Local Development / Pre-release Tags

You may encounter doc references to `-dev` or `+dev.<sha>` versions:
- These are not published release tags.
- Intended for internal or contributor testing.
- Do not pin production workloads to pre-release versions.

---

## 15. Frequently Asked Questions

**Q: I upgraded and saw a replacement plan for a resource. Is that a bug?**
A: Possibly. Replacements should only occur when a ForceNew attribute actually changed in your config. If not, open an issue with plan output.

**Q: A MINOR release added an attribute I want to set—but plan shows a replacement.**
A: If the attribute is ForceNew (immutable server-side), initial introduction might require recreate. Check the release notes to confirm.

**Q: How can I test a new resource without affecting production?**
A: Use a separate workspace or a narrower version constraint (`= 1.x.y`) in production until validation completes.

---

## 16. Quick Upgrade Checklist

| Step | Action |
|------|--------|
| 1 | Read CHANGELOG between your current and target version |
| 2 | Scan for Deprecated or Breaking sections |
| 3 | Run `terraform init -upgrade && terraform plan` in a staging workspace |
| 4 | Investigate any unexpected replacements |
| 5 | Apply if plan is acceptable |
| 6 | (MAJOR only) Follow migration guide instructions carefully |
| 7 | Update internal docs / runbooks if new endpoints adopted |

---

## 17. Reporting Version Issues

Include in any issue:
- Current version & target version.
- Output of `terraform version`.
- Relevant plan diff (redacted secrets).
- Whether ForceNew attributes changed in configuration.

Fast, precise reports accelerate fixes and help maintain versioning integrity.

---

## 18. Summary

- MINOR = one new endpoint or major additive structure.
- PATCH = safe fixes and computed attribute visibility.
- MAJOR = rare, only for deliberate breaking schema changes.
- Atomic, documented evolution reduces risk and improves auditability.
- Pin versions according to your risk tolerance, and always review CHANGELOG entries.

---

_Last updated: (update this page when introducing new versioning behaviors or after a MAJOR release)._
