# Migration Guide Template

This template is used to author a migration document for each **MAJOR** version of the Terraform Addy provider.
Copy this file to: `docs/src/migrations/<major-version>.md` (e.g. `2.0.0.md`) and replace all placeholder sections.

> IMPORTANT: Keep this document user-facing, concise, and free of internal-only rationale.
> All normative policies still live in `.rules`; this guide explains how to move from one major version to the next safely.

---

## 1. At-a-Glance Summary

| From | To | Release Date | Migration Complexity | Requires Manual State Manipulation |
|------|----|--------------|----------------------|------------------------------------|
| `<previous_major>.x` | `<new_major>.0` | `<YYYY-MM-DD>` | Low / Medium / High | Yes / No |

**Key Headline Changes:**
- (Bullet 1)
- (Bullet 2)
- (Bullet 3)

---

## 2. Why This Major Version Exists (User Perspective)

Briefly explain the user-facing motivation (e.g., unify naming, remove deprecated attributes, align with upstream API changes).
Avoid internal implementation details—focus on “what” and “why it matters” to users.

---

## 3. Breaking Changes

Describe all breaking changes succinctly.

| Category | Description | Old Behavior | New Behavior | User Action |
|----------|-------------|--------------|--------------|-------------|
| Attribute Removal | `<resource>.<attribute>` removed | Attribute accepted | Configuration fails | Remove attribute; optionally adopt replacement |
| Attribute Rename | `<old>` → `<new>` | Old name accepted | Old name invalid | Rename in config |
| Type Change | `string` → `bool` | String interpreted loosely | Strongly typed boolean | Update variable / value |
| ForceNew Semantics | `<attribute>` now ForceNew | In-place update | Replacement | Plan for recreation |
| Resource Split/Merge | `<resource>` split/merged | Single resource | Multiple / consolidated | Refactor config |

Add rows for every breaking item.

---

## 4. Deprecated Items Removed

| Removed Item | Introduced Deprecation In | Removed In | Replacement / Alternative | Notes |
|--------------|---------------------------|-----------|---------------------------|-------|
| `<attr>` | `1.5.0` | `2.0.0` | `<new_attr>` |  |
| ... | ... | ... | ... | ... |

If nothing was removed, state: “None (no deprecations reached removal threshold).”

---

## 5. New Capabilities in This Major

| Capability | Type (Resource/Data Source/Concept) | Brief Description | Config Impact |
|------------|-------------------------------------|-------------------|---------------|
| `<name>` | Resource |  | Optional adoption |
| `<name>` | Data Source |  | Optional |
| Pagination Enhancement | Concept | e.g., added `all` + `limit` | Opt-in |

---

## 6. Attribute Rename Mapping

If any renames occurred, provide a machine-friendly table:

| Resource | Old Attribute | New Attribute | Transformation Notes |
|----------|---------------|---------------|----------------------|
| `<resource>` | `old_name` | `new_name` | Direct rename |
| ... | ... | ... | ... |

---

## 7. ForceNew Changes

List attributes whose semantics changed to (or from) ForceNew.

| Resource | Attribute | Old Semantics | New Semantics | Reason | User Impact |
|----------|----------|---------------|---------------|--------|-------------|
| `<resource>` | `<attr>` | In-place update | Replacement | Identity clarified | Recreate on change |

---

## 8. Behavioral Adjustments (Non-Schema)

| Topic | Prior Behavior | New Behavior | Trigger | Action |
|-------|----------------|--------------|---------|--------|
| Retry Policy | 429 only | (If expanded) 429 + 503 | Upgrade to `<new_major>` | None / review applies |
| Pagination | Single-page only | Added `all` with `limit` guard | Config sets `all=true` | Add safe limit |
| Logging | Basic | Added metrics counters | Enable metrics opt-in variable | Optional |

If none, state “No behavioral adjustments affecting existing configurations.”

---

## 9. State Migration Guidance

Only include if users must manipulate state or perform multi-step changes.

### 9.1 Simple (Recommended) Path
1. Upgrade provider version in `required_providers`.
2. Run `terraform init -upgrade`.
3. Run `terraform plan`.
4. Review replacements / deletions.
5. Apply when acceptable.

### 9.2 Advanced: Minimizing Downtime
If applicable (e.g., resource rename requiring two-step deployment), outline steps:
1. Add new resource alongside old (with `lifecycle { ignore_changes = [...] }` if necessary).
2. Import existing remote object into new resource (`terraform import ...`).
3. Remove old resource block.
4. Plan & apply.

### 9.3 Manual State Edits (Only If Unavoidable)
Provide `terraform state mv` or `terraform state rm` commands with explicit caution.

---

## 10. Upgrade Checklist

| Step | Required | Done |
|------|----------|------|
| Review CHANGELOG for `<new_major>.0` | Yes | [ ] |
| Update provider version constraint | Yes | [ ] |
| Remove deprecated attributes | If used | [ ] |
| Rename attributes per mapping table | If used | [ ] |
| Handle ForceNew replacements (plan review) | Yes | [ ] |
| Run plan in staging workspace | Yes | [ ] |
| Backup state (remote backend snapshot or export) | Recommended | [ ] |
| Apply and validate outputs | Yes | [ ] |
| Remove any temporary migration code | If created | [ ] |

---

## 11. Verification Steps Post-Upgrade

1. Run `terraform state list` to confirm expected resources remain.
2. Validate no unintended replacements occurred beyond those documented.
3. Check logs for unexpected warnings.
4. (If metrics enabled) Verify latency/retry counters appear as expected.
5. Run a no-op `terraform plan`—should report “No changes.”

---

## 12. Rollback Strategy

| Scenario | Rollback Action | Caveat |
|----------|-----------------|--------|
| Pure schema rename failure | Pin previous MAJOR version; revert attribute names | Reintroduce deprecated attrs temporarily |
| Forced replacement undesirable | Restore pre-upgrade state backup & version constraint | Data changed during window may need reconciliation |
| Unhandled error codes appear | Downgrade provider; open issue | Provide logs & repro config |

---

## 13. Frequently Asked Questions

**Q: The plan wants to recreate X—expected?**
See ForceNew table; if listed, yes. Otherwise file an issue with plan diff.

**Q: A removed attribute still in my config—how do I find all occurrences?**
Use a recursive grep (`grep -R "<attr_name>" .`) before upgrading.

**Q: Can I skip intermediate minor versions before jumping to `<new_major>.0`?**
Yes; read all intervening CHANGELOG notes anyway for context and deprecation timing.

---

## 14. Known Issues / Caveats

| Issue | Affected Resources | Workaround | Target Fix Version |
|-------|--------------------|------------|--------------------|
| (Populate if any) |  |  |  |

If none: “No known migration-specific issues at release.”

---

## 15. Change Log (For This Migration Document)

| Version | Change |
|---------|--------|
| 1 | Initial migration template created. |

---

## 16. How to Use This Template (Contributor Notes – Remove in Final Guide)

- Replace all angle-bracket placeholders (`<...>`).
- Remove this section before publishing actual migration guide.
- Ensure any removal/rename items are also reflected in main CHANGELOG.
- Cross-link new/changed concept pages (pagination, retry) if relevant.

---

_Last updated: (set date when finalizing)._
_Status: Template — not an actual migration guide._
