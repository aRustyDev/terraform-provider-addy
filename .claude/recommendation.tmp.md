Recommendation
Yes—put the actionable iteration checklist in a dedicated file (`.claude/AGENT_CHECKLIST.md`) and reference it from `CLAUDE.md`. This separation:
- Prevents duplication and “drift” when tweaking process.
- Keeps `CLAUDE.md` readable (strategic context) while `AGENT_CHECKLIST.md` remains purely operational.
- Makes it easier for a slash-command or Zed Agent to extract only the step list.

### For Slash Commands / Zed Agent Threads
- You can implement a command like `/next-step` that fetches and prints the next unchecked item from `AGENT_CHECKLIST.md`.
- Provide anchors (you could add `## [Step 1] Selection & Classification` style headings if machine parsing helps).
- Avoid copying the checklist into multiple places—point everything at the canonical path.

### For Claude-Code vs Zed Agent Approaches
- Claude-code agent: reads the checklist for procedural next actions.
- Zed agent thread: can hyperlink or auto-open `.claude/AGENT_CHECKLIST.md` when starting a new iteration.
- Only reference from `CLAUDE.md`—no need to embed the entire list again.
- Optionally create a minimal pointer in README for humans: “Development process defined in `.claude/AGENT_CHECKLIST.md`.”

### Optional Enhancement
Add a small machine-readable header at top of checklist:
```
# meta: format=iterative-checklist; version=1
```
So tools can confirm they’re parsing the right doc.

## Does It Need To Be In More Than One Place?
No. Single source of truth is superior. Redundant copies increase maintenance burden and risk of contradictory instructions. Use references/links only:
- `CLAUDE.md` → “See `.claude/AGENT_CHECKLIST.md`”
- README (optional) → “Development process: `.claude/AGENT_CHECKLIST.md`”

## Next Suggested Actions
1. Replace existing `CLAUDE.md` content with the revised version above.
2. Add the new `AGENT_CHECKLIST.md`.
3. Update README to mention both documents.
4. Begin implementing HTTP client improvements before adding more resources (foundation first).

If you want, I can also draft an updated README or the HTTP client retry refactor next—just let me know.


---

> Critique of Existing TODO List

Critique
- Lacks grouping by phase (foundation vs features vs polish).
- “Enable provider variables to be sourced from ENV vars” partly done; refine to include precedence & doc updates.
- Missing tasks for:
  - README sections: installation, environment variable usage, known limitations.
  - Provider documentation generation (for eventual registry).
  - License relocation (top-level LICENSE).
  - Consistency pass to remove “HashiCups” references.
  - Implement test scaffolding (unit + acceptance structure).
  - Add linting tools (e.g. `golangci-lint`).
  - Release pipeline (tagging, changelog).
  - Data source modeling for app-version, api-token-details enhancements (null handling test).
  - Add importers: specify which resources (domain, alias, recipient, rule, username).
  - Schema stability review before 0.1.0 release.
  - Observability improvements (standardized logging fields).
  - Pagination doc + example.


---

. CLAUDE.md Critique & Enhancements

### Observations
- “Repository Overview” section is empty placeholder—should be concretely filled.
- Data Sources vs Resources classification currently lists both list and single endpoints redundantly; decide pattern:
  - Usually: single object = resource (if mutable), list endpoint becomes data source returning collection.
  - If an endpoint is read-only (e.g. `app-version`), data source only.
- The section “Examples Directory (to create)” now outdated because examples exist in `.claude/examples/terraform/`.
- Workflow section good but can be made more precise with per-step artifacts and gating conditions.
- Checklist truncated (“Ran (or prepared)”) and missing test/doc/build clarity.
- Non-Goals useful but can be refined to prevent scope creep explicitly (e.g., “No speculative endpoints” separate from “No rewriting OpenAPI”).

---

---

Fixes to Explicitly Mention

You want CLAUDE.md to instruct the agent to implement these early:
- Replace `path.Root("key")` → `path.Root("api_key")`.
- Rename provider address.
- Remove legacy “HashiCups” strings.
- Add environment variable default logic using `Config` plan modifiers (optionally).
- Implement timeouts & user-agent header in `NewClient`.

---

Mapping Examples to Schema (Agent Guidance)

Tell the agent how to turn sample JSON into Terraform schema types:
- Scalar string → `schema.StringAttribute{Computed: true}` if server-created.
- Boolean flags actionable by user (e.g. `active`, `catch_all`) → `Optional` + `Computed` with PlanModifier to preserve known state.
- Nested recipient objects → `schema.ListNestedAttribute`.
- Conditions/actions for rules → `schema.ListNestedAttribute` with `type`, `match`, `values`, `value`.

## Suggest Centralized Model Definitions

Add a `internal/model` package with typed Go structs (one per endpoint) mirroring JSON; this lets you:
- Reuse unmarshalling.
- Ensure consistency between data sources and resources.
- Provide helper normalization (e.g. convert null strings to `types.StringNull()`).

## Suggested Enhancements to Repo (Document in CLAUDE.md)

- Create top-level `LICENSE` file (move content out of `internal/about/license.go`).
- Expand `README.md` with installation, local dev instructions:
  - `go build -o terraform-provider-addy`
  - Terraform CLI `.terraformrc` development override
  - Example usage.
- Add `Makefile`:
  - `make build`
  - `make test`
  - `make lint`
- Add `tools.go` if you standardize codegen or lint deps.
- Add GitHub Actions (if desired) for `go vet`, `golangci-lint`, `terraform validate`.
- Add `internal/test` with acceptance scaffolding.

---

Error Handling Implementation Suggestion

You can guide the agent to load `errors.toml` once:

```
package errors

import (
  "os"
  "github.com/pelletier/go-toml"
)

type ApiErrorMeaning struct {
  Message string `toml:"message"`
  Meaning string `toml:"meaning"`
}

type Errors struct {
  Error map[int][]ApiErrorMeaning `toml:"error"`
}

var codes Errors

func Init() error {
  b, err := os.ReadFile(".claude/examples/errors.toml")
  if err != nil { return err }
  return toml.Unmarshal(b, &codes)
}

func Lookup(status int) string {
  if entries, ok := codes.Error[status]; ok && len(entries) > 0 {
    return entries[0].Meaning
  }
  return "Unknown error"
}
```
Then wrap `Curl` errors to append human meaning.

## Fixes to Explicitly Mention

You want CLAUDE.md to instruct the agent to implement these early:
- Replace `path.Root("key")` → `path.Root("api_key")`.
- Rename provider address.
- Remove legacy “HashiCups” strings.
- Add environment variable default logic using `Config` plan modifiers (optionally).
- Implement timeouts & user-agent header in `NewClient`.

## Testing Section Guidance

Explain:
- Unit: mock transport with `RoundTripFunc`.
- Acceptance: requires live API; skip tests when `ADDY_API_KEY` missing.
- Use `terraform-plugin-testing` (optional future) for plan/apply checks.

---

Resource State Pattern Example (Domain)

Show a sample schema pattern:

```go
resp.Schema = schema.Schema{
  Attributes: map[string]schema.Attribute{
    "id": schema.StringAttribute{Computed: true},
    "domain": schema.StringAttribute{Required: true},
    "description": schema.StringAttribute{Optional: true, Computed: true},
    "from_name": schema.StringAttribute{Optional: true, Computed: true},
    "active": schema.BoolAttribute{Optional: true, Computed: true},
    "catch_all": schema.BoolAttribute{Optional: true, Computed: true},
    "aliases_count": schema.Int64Attribute{Computed: true},
    "default_recipient": schema.StringAttribute{Optional: true, Computed: true},
    "auto_create_regex": schema.StringAttribute{Optional: true, Computed: true},
    "created_at": schema.StringAttribute{Computed: true},
    "updated_at": schema.StringAttribute{Computed: true},
    "domain_verified_at": schema.StringAttribute{Computed: true},
    "domain_mx_validated_at": schema.StringAttribute{Computed: true},
    "domain_sending_verified_at": schema.StringAttribute{Computed: true},
  },
}
```

Roadmap List (Refine TODO.md)

Add actionable tasks you can reference:

- [ ] Fix provider metadata/address.
- [ ] Correct api_key error path.
- [ ] Implement error loader and wrap Curl.
- [ ] Refactor HTTP client with timeout, retry on 429.
- [ ] Implement `domain_options` data source.
- [ ] Implement `domain` resource.
- [ ] Implement `alias` resource.
- [ ] Implement `recipient` resource (with encryption flags).
- [ ] Implement `rule` resource (complex nested modeling).
- [ ] Implement `username` resource.
- [ ] Add pagination support for list data sources.
- [ ] Add Makefile + tests + examples.
- [ ] Expand README.
- [ ] Add importers (optional).
- [ ] Acceptance tests.

## Agent Operating Rules (Meta Section)

Add a section telling the agent:
- Always inspect existing file before overwriting.
- Use consistent code style (no unused imports).
- Never leave `panic("not implemented")` in final commits.
- Add logging at start/end of Read/Create/Update/Delete.
- Run build after each major change.

## Helpful Non-Goals

Clarifying non-goals helps keep scope tight:
- No UI components.
- No rewriting OpenAPI spec.
- No performance optimization beyond simple retries.
- No dynamic code generation unless later explicitly added.

---

Final Checklist to Put at Bottom of CLAUDE.md

A short checklist the agent can tick off each cycle:
- Parsed OpenAPI section for endpoint.
- Mapped fields to schema.
- Implemented CRUD / Read method.
- Added logging.
- Added error mapping.
- Added (or updated) example.
- Built successfully.
- Ran (or prepared) test.

---

Suggested Additional Examples To Add
- Alias resource (once implemented).
- Recipient resource showing encryption toggle.
- Rule resource with nested conditions/actions block.
- A data source usage example (`data "addy_api_token_details" "current" {}`).
- A compose example (create domain, then create alias referencing domain).
- Pagination example for list data source (once implemented): show `page_size`.

---

"Acceptance tests for each."
