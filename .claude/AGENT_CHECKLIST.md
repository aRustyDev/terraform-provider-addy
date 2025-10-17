# Agent Iteration Checklist (Authoritative)

Complete ALL items for one endpoint/resource before moving to the next. This prevents half-finished implementations accumulating technical debt.

## 1. Selection & Classification
- Identify a single endpoint from `openapi.yaml`.
- Decide: data source (read-only) or resource (mutable).
- Confirm no overlap with already implemented functionality.

## 2. Spec & Example Extraction
- Locate the matching path in `openapi.yaml`.
- Note all response fields, nullable vs non-nullable.
- Gather example JSON under `.claude/examples/responses/**` if available.
- Record required request parameters (query, headers, body).

## 3. Schema Draft
- Map fields → Terraform attributes; tag each as Required / Optional / Computed.
- Define nested blocks for structured arrays (e.g., rule conditions/actions).
- Select collection types (List vs Set) with ordering rationale.
- Add `MarkdownDescription` for every attribute.
- Validate naming consistency (snake_case).

## 4. Implementation Skeleton
- Create or update file under `internal/data` or `internal/resource`.
- Add type struct and `Metadata`, `Schema` methods.
- Remove any placeholder `panic("not implemented")` in modified scope.

## 5. HTTP Integration
- Use `utils.Curl` (refactored version if available) to perform request(s).
- Handle multi-step operations (e.g., activation toggles) cleanly.
- Add timeout and retry (429) automatically if infrastructure exists.

## 6. Error Handling
- On non-2xx: map status → meaning via loaded error map.
- Build diagnostic with status + mapped meaning + truncated body snippet.
- For parsing errors: dedicated diagnostic, no panic.

## 7. State Population
- Parse JSON into typed Go model struct(s).
- Convert to Terraform framework types (`types.StringValue`, `types.BoolValue`, etc.).
- Represent missing/nullable fields with `types.<Type>Null()`.

## 8. Idempotency & Drift
- For resources: Re-run Read after Create/Update to capture server normalization.
- Verify subsequent plan shows no unintended changes.
- Consider `ignore_changes` guidance if server normalizes fields.

## 9. Logging
- Add `tflog.Debug` at start/end of CRUD/Read with endpoint + method.
- Use `tflog.Trace` only for deep internals (optional).
- Ensure no credential leakage.

## 10. Testing (Initial)
- Unit test stub: mock HTTP success response + one error variant.
- If complex nested parsing: add targeted unmarshal test.
- Add TODO for acceptance test if endpoint requires live interaction.

## 11. Documentation & Examples
- Add or update `.claude/examples/terraform` with a usage example referencing new data source/resource.
- Update README section list (manually or leave placeholder).
- Include notes for required preconditions (e.g., verified recipient before alias attachment).

## 12. Build & Diagnostics
- Run `go build ./...` (no compile errors).
- Search for remaining `panic(` calls in touched files.
- Ensure provider address not reverted to tutorial placeholder.

## 13. Code Quality
- Confirm no unused imports (`go vet`, optional linter).
- Ensure version references central constant only.
- Validate attribute naming matches conventions.

## 14. Acceptance Test Prep
- If endpoint mutates state, outline acceptance test steps:
  - Preconditions.
  - Create → Read → Update → Read → Destroy sequence.
  - Mark with `// TODO: implement TF_ACC acceptance test`.

## 15. Commit Preparation
- Stage only related changes (avoid mixing unrelated refactors).
- Summarize endpoint implemented + tests + example additions.
- Ensure no secrets contained in diff.

## 16. Exit Criteria
- Endpoint fully operable (plan/apply lifecycle).
- Example config present and accurate.
- No panics / silent failures.
- Diagnostics meaningful.
- Logging standardized.
- Mapped errors functioning.
- Ready to proceed to next endpoint.


<!--# Workflow for Agent
1. Read OpenAPI spec section for target endpoint.
2. Locate/create corresponding file under `internal/data` or `internal/resource`.
3. Define schema (start minimal).
4. Implement request using `utils.Curl`.
5. Parse JSON → model struct → populate `types.*Value`.
6. Add logging.
7. Add test stub (if test harness present).
8. Run diagnostics (build) before next feature.
9. Update README & examples after each resource.-->
