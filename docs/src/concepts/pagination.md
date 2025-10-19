# Pagination & Collection Retrieval

This page explains how the Terraform Addy provider currently handles pagination for collection
data sources (e.g., lists of domains, aliases, recipients) and how the approach will evolve.
It is written for provider *users* (consumer perspective). Internal design rationale
and normative rules live in `.rules` and `CLAUDE.md`.

---

## 1. Goals

| Goal | Description |
|------|-------------|
| Predictability | Terraform plan size and state growth must be explicit and controlled. |
| Safety | Avoid unbounded enumeration that could bloat state or exceed API limits. |
| Extensibility | Provide a clear path to future “iterate all pages” capabilities with guard rails. |
| Observability | Surface enough metadata (e.g., last page, total items) for users to reason about coverage. |

---

## 2. Current Model (Initial Phase)

The initial implementation deliberately opts for **explicit single-page retrieval**:

| Attribute | Purpose |
|-----------|---------|
| `page_number` (optional) | 1-based page selector; defaults to 1 if omitted. |
| `page_size` (optional) | Items per page (subject to API caps). |
| (Optional Computed) `last_page` | Server-reported last page index, if available. |
| (Optional Computed) `total_items` | Server-reported total item count, if available. |

Behavior:
1. Provider sends exactly one GET request with the provided (or default) pagination parameters.
2. The response items for that page populate the data source state.
3. No implicit iteration or secondary requests are performed.
4. If API does not supply pagination metadata, computed fields remain null.

Why single-page first?
- Prevents accidental large state diffs when collections are large or growing.
- Establishes baseline semantics before introducing continuous iteration modes.

---

## 3. Future Evolution (Planned / Deferred)

A controlled “fetch all pages” mode is on the roadmap. Anticipated attributes (naming subject to finalization):

| Proposed Attribute | Type | Planned Semantics |
|--------------------|------|-------------------|
| `all` | Optional bool | When true, iterate pages sequentially until completion (guarded). |
| `limit` | Optional number | Hard cap on total items collected when `all = true`. |
| `start_page` | Optional number | Advanced use: start enumeration at page N (may be gated). |

Guard Rails (Planned):
- Both `all = true` AND a finite `limit` required to enable “fetch all”.
- Hard fail if `limit` exceeds provider internal maximum (e.g., 10k) to avoid runaway state.
- Explicit metrics/log entries (future observability phase) indicating total pages traversed.

Not yet implemented—users should not rely on these attributes until they appear in generated docs & CHANGELOG.

---

## 4. Example: Basic Single Page Data Source

```hcl
data "addy_aliases" "page1" {
  page_number = 1
  page_size   = 25
}

output "alias_ids_page1" {
  value = [for a in data.addy_aliases.page1.aliases : a.id]
}
```

If `last_page` is exposed:

```hcl
output "aliases_pagination_meta" {
  value = {
    page_requested = data.addy_aliases.page1.page_number
    last_page      = data.addy_aliases.page1.last_page
    total_items    = data.addy_aliases.page1.total_items
  }
}
```

---

## 5. Designing Your Own Iteration (Workaround Pattern)

Until native safe “all pages” support lands, you can manually step through pages in *separate* data source blocks:

```hcl
locals {
  pages = [1, 2, 3] # Manually maintained
}

data "addy_aliases" "paged" {
  for_each    = toset(local.pages)
  page_number = each.value
  page_size   = 50
}

# Flatten results (schema permitting)
locals {
  all_alias_ids = flatten([
    for p in data.addy_aliases.paged : [
      for a in p.aliases : a.id
    ]
  ])
}

output "all_alias_ids" {
  value = local.all_alias_ids
}
```

Caveats:
- You must keep the `pages` list aligned with actual `last_page`.
- Over-fetching (setting a page number > last_page) may yield empty sets or diagnostics depending on API behavior.

---

## 6. Anti-Patterns to Avoid (Current Phase)

| Anti-Pattern | Why Problematic | Safer Alternative |
|--------------|-----------------|------------------|
| Using `count` over a guessed range of 1..N without metadata | Might overshoot; wasted requests or failures | Manually inspect `last_page` first |
| Embedding external scripts to pre-enumerate pages | Increases complexity; brittle when API rate limits shift | Wait for provider’s guarded “all” mode |
| Attempting to simulate “all” by massive `page_size` (e.g., 10k) | Risk of timeouts, large memory payloads | Use modest page sizes and explicit multi-block pattern |

---

## 7. Drift & State Implications

- Single-page collection state will *not* reflect newly created items in later pages unless you adjust `page_number`.
- Plans remain stable because you control exactly which subset is included.
- Transition to future “all” mode (once released) will likely produce a plan showing additions for the newly included pages—this is expected.

---

## 8. Performance Considerations

| Factor | Impact |
|--------|--------|
| Large page sizes | Higher latency per request and larger JSON decode overhead. |
| Many discrete page data sources | Terraform evaluation overhead; but keeps each state fragment smaller. |
| Future “all” iteration | Will include safety checks to avoid pathological fetches. |

Recommendation: Start with moderate page sizes (25–100) unless you have confirmed API performance at higher thresholds.

---

## 9. Error Handling Specific to Pagination

| Scenario | Outcome |
|----------|---------|
| Requesting page beyond `last_page` | API may return empty list or 404/422; provider surfaces diagnostic accordingly. |
| Invalid `page_size` (too large) | API typically returns validation error (422); diagnostic includes meaning & truncated body. |
| Rate limited mid-request (future multi-page) | Retry logic (429 only) applies per request; persistent limit triggers failure after bounded backoff. |

---

## 10. Planned Compatibility Guarantees

| Aspect | Guarantee (Subject to Policy) |
|--------|-------------------------------|
| Existing single-page attributes (`page_number`, `page_size`) | Will remain stable when “all” mode added. |
| Introduction of `all` | MINOR release with clear CHANGELOG entry. |
| Requirement of `limit` with `all` | Enforced to prevent accidental unbounded fetch. |
| Removal or renaming of pagination attributes | Would require MAJOR (not anticipated). |

---

## 11. Migration Strategy (When “All” Arrives)

If/when `all = true` is released:
1. Start by setting `limit` conservatively (e.g., expected number of items + buffer).
2. Run a plan—expect additions representing newly included pages.
3. Monitor apply time & any rate limit messages.
4. Increment `limit` cautiously if you under-estimated collection size.
5. Avoid setting a limit equal to remote total if near provider maximum; leave headroom for growth.

---

## 12. Troubleshooting

| Symptom | Possible Cause | Resolution |
|---------|----------------|-----------|
| Empty list unexpectedly | Selected page has no items yet or is beyond last page | Check `last_page` (if exposed) or try lower `page_number` |
| Validation error for pagination params | Out-of-range `page_size` or `page_number` | Adjust to allowed bounds per API docs |
| Frequent 429 responses | Page size too large or high concurrency in other resources | Reduce page_size; stagger queries; rely on retry backoff |
| Drift in downstream resources referencing data source | Underlying alias/domain created on a different page than selected | Adjust `page_number` or manually replicate multi-page pattern |

---

## 13. FAQ

**Q: Why not auto-fetch all pages by default?**
To prevent unexpected massive state expansions & performance issues. Explicit control avoids surprise diffs.

**Q: Can I force a huge page size to capture everything?**
Not recommended; once size exceeds practical limits the API may reject it or responses may slow dramatically.

**Q: Will switching from single-page to future “all” mode recreate resources?**
Data sources don’t manage objects, so they won’t “recreate,” but downstream references may change if they depend on enumerated IDs. This is expected and should be reviewed in plan output.

**Q: How do I know the total I should set for `limit` (future)?**
Start with a safe upper bound informed by `total_items` (if exposed) plus a buffer; adjust iteratively.

---

## 14. Example: Incremental Manual Aggregation Pattern

```hcl
# Discover first page
data "addy_recipients" "p1" {
  page_number = 1
  page_size   = 50
}

# Use computed last_page (when available) to decide further blocks manually later.
output "recipients_page_meta" {
  value = {
    last_page   = try(data.addy_recipients.p1.last_page, null)
    total_items = try(data.addy_recipients.p1.total_items, null)
  }
}
```

You can iteratively add `page2`, `page3` data sources only if needed—keeping state lean.

---

## 15. Design Principles (User Perspective)

| Principle | Implication |
|-----------|-------------|
| Explicitness | You decide which slice of data enters state. |
| Predictability | Plan changes remain tied to known, user-defined page scopes. |
| Minimal Surprise | Provider will not silently expand state breadth across upgrades. |
| Progressive Disclosure | Advanced bulk enumeration arrives only with safe guard rails. |

---

## 16. Change Log (Local to Pagination Page)

| Version | Change |
|---------|--------|
| 1 | Initial pagination concept documentation (single-page model, future roadmap, troubleshooting). |

*(This table tracks conceptual doc changes; refer to global CHANGELOG for provider release notes.)*

---

_Last updated: (update when multi-page “all” mode or guard rails land)._
_Status: Single-page mode active; multi-page deferred._
