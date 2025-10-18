
## Slash Commands

- [ ] `/implement-endpoint target`
- [ ] `/review-endpoint target`
- [ ] `/next-step`: fetches and prints the next unchecked item from `AGENT_CHECKLIST.md`

## Justfile

- [ ] Security: `scan-secrets`
- [ ] Security: `check-licenses`
- [ ] Security: `vuln-check`
- [ ] Release Prep: `changelog-draft`
- [ ] Release Prep: `version-bump`
- [ ] Provider Utils: `schema-diff`
- [ ] Provider Utils: `import-test`
- [ ] Provider Utils: `sweep`
- [ ] cicd: `ci-local`
- [ ] cicd: `ci-acceptance-setup`
- [ ] debug: `clean`
- [ ] debug: `trace-logs`
- [ ] debug: `debug-provider`
- [ ] code-gen: `generate`
- [ ] code-gen: `generate-docs`
- [ ] code-gen: `verify-generated`
- [ ] quality: `check-docs`
- [ ] quality: `lint`
- [ ] quality: `lint-fix`
- [ ] quality: `validate-examples`
- [ ] testing: `test-unit`
- [ ] testing: `test-acceptance`
- [ ] testing: `test-specific`
- [ ] testing: `test-coverage`
- [ ] sdlc: `init`
- [ ] sdlc: `setup`
- [ ] sdlc: `build`
- [ ] sdlc: `install-local`

## Zed Tasks

- [ ] Convert all justfile commands into Zed tasks in `.zed/tasks.json`

## Project Utilities

- [ ] Pagination helper utility (shared).
- [ ] HTTP client improvements; Retry helper wrapping Curl; timeout, User-Agent, retry on 429 with backoff+jitter.
- [ ] Add Shared model structs with a `internal/model` package with typed Go structs (one per endpoint) mirroring JSON; this lets you:
    - Reuse unmarshalling.
    - Ensure consistency between data sources and resources.
    - Provide helper normalization (e.g. convert null strings to `types.StringNull()`).
- [ ] Unit test harness (mock RoundTripper).
- [ ] Golangci-lint integration.
- [ ] Acceptance test scaffold (TF_ACC gating).
- [ ] Golden file tests for representative JSON payloads.
- [ ] Add environment variable default logic using `Config` plan modifiers (optionally).
- [ ] Add `tools.go` if you standardize codegen or lint deps.
- [ ] Add GitHub Actions (if desired) for `go vet`, `golangci-lint`, `terraform validate`.
- [ ] Add `internal/test` with acceptance scaffolding.


## Project Cleanup

- [ ] Move license text to top-level LICENSE file.
- [ ] Single source version usage (inject from internal/about/version.go).
- [ ] Consistency pass to remove “HashiCups” references.
- [ ] Rename provider address (to what?)
- [ ] Replace `path.Root("key")` → `path.Root("api_key")`.
- Expand `README.md` with installation, local dev instructions:
  - `go build -o terraform-provider-addy`
  - Terraform CLI `.terraformrc` development override
  - Example usage.

## Context Engineering

- [ ] Clarify Data Sources vs Resources classification pattern
  - Usually: single object = resource (if mutable), list endpoint becomes data source returning collection.
  - If an endpoint is read-only (e.g. `app-version`), data source only.
- [ ] Enhance Workflow section with per-step artifacts and gating conditions.
- [ ] Expand Checklist with clearer Testing, Documentation, and Build steps.
- [ ] Add code snippets/examples for common patterns (e.g., schema definition, error handling).
- [ ] Create a FAQ or Troubleshooting section for common pitfalls.
- [ ] Define conventions for attribute naming, error handling, and logging.
- [ ] Add links to relevant documentation (Terraform Plugin SDK, Go best practices).
- [ ] Add Example: Alias resource
- [ ] Add Example: Recipient resource showing encryption toggle
- [ ] Add Example: Rule resource with nested conditions/actions block
- [ ] Add Example: data source usage example (`data "addy_api_token_details" "current" {}`)
- [ ] Add Example: compose example (create domain, then create alias referencing domain).
- [ ] Add Example: Pagination example for list data source (once implemented): show `page_size`.
- [ ] Tell the agent how to turn sample JSON into Terraform schema types:
- Scalar string → `schema.StringAttribute{Computed: true}` if server-created.
    - Boolean flags actionable by user (e.g. `active`, `catch_all`) → `Optional` + `Computed` with PlanModifier to preserve known state.
    - Nested recipient objects → `schema.ListNestedAttribute`.
    - Conditions/actions for rules → `schema.ListNestedAttribute` with `type`, `match`, `values`, `value`.
- [ ] Error Handling Implementation; guide the agent to load `errors.toml` once:
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
