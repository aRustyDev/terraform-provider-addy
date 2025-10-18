---
meta: format=iterative-checklist; version=2
---

# Agent Iteration Checklist (Authoritative)

Complete ALL items for exactly one endpoint or resource per iteration before moving on. This enforces incremental, fully integrated progress and prevents half-finished implementations accumulating technical debt.

## 1. Selection & Classification
- Pick ONE logical entity (e.g. domain, alias, recipient, rule, username) from `openapi.yaml`.
- Collect all related OpenAPI paths (base path + any activation/toggle variants).
- Decide: data source vs resource.
  - Resource if there is at least a POST (create) OR PATCH/PUT (update) OR DELETE (destroy).
  - Data source if only GET operations exist.
- Confirm no existing implementation already covers it.

## 2. CRUD Capability Derivation & Planning
Produce a “Method Plan Table” describing how Terraform methods map to API operations.

1. Enumerate operations:
   - Create: POST /entities
   - Read: GET /entities/{id}
   - Update: PATCH (or PUT) /entities/{id}
   - Delete: DELETE /entities/{id}
   - Ancillary toggles (e.g., /active-*, /catch-all-*): map to attributes (e.g., `active`, `catch_all`).
2. Decide for each Terraform method:
   - Implement? If missing PATCH, mark mutable-looking fields as ForceNew or confirm immutability.
   - If no DELETE endpoint: resource may be “orphanable” (avoid unless necessary) or reclassify as data source.
3. Attribute classification rules:
   - Required: Field appears in POST body and is mandatory.
   - Optional: Field appears in POST body but is not required.
   - Optional+Computed: Server may default or normalize; resource sets but refreshes exact server value.
   - Computed: Field appears only in responses (never in POST/PATCH).
   - ForceNew: Field appears in Create (POST) but not accepted in Update (PATCH/PUT) OR API docs indicate immutable.
   - Sensitive: API key only (for provider); rarely needed at resource attr level unless secrets appear later.
4. Activation / toggles:
   - Map paired POST/DELETE (activate/deactivate) or PATCH toggles to a single boolean attribute.
   - Implement inside Create/Update logic: if plan wants true and object is false → call activate endpoint; inverse for false.
5. Derive error expectations:
   - List typical status codes (201 create success, 200 read/update, 204 delete).
   - Note special error codes (404 after delete, 422 validation).
6. Data normalization planning:
   - Identify any fields likely to be reformatted by server (timestamps, counts).
   - Plan to always perform a Read after Create/Update to synchronize state.
7. Output artifact (add as code comment at top of file or doc snippet):

Example Method Plan Table (inline comment format):
```
Method Plan (alias):
# Create: POST /api/v1/aliases -> requires domain (required), description (optional)
# Read:   GET /api/v1/aliases/{id}
# Update: PATCH /api/v1/aliases/{id} (description, from_name)
# Delete: DELETE /api/v1/aliases/{id}
# Toggle active: POST /api/v1/active-aliases (activate), DELETE /api/v1/active-aliases/{id} (deactivate) -> attribute: active (Optional+Computed)
# Immutable: local_part, domain (ForceNew)
# Computed: emails_forwarded, emails_blocked, created_at,
```
Do not proceed until this table is written.

## 3. Spec & Example Extraction
- Extract exact field list from `openapi.yaml` path(s).
- Record which response fields are nullable; treat `null` as Terraform Null.
- Gather example JSON payload from `.claude/examples/responses/**` (if present).
- Record required request parameters (query, headers, body).
- If implementing list data source, Note additional query parameters (pagination, filters).

## 4. Schema Draft
- Translate Method Plan Table into a Terraform schema; Map fields → Terraform attributes.
- Assign `Required`, `Optional`, `Computed`, `Optional + Computed`, `ForceNew` flags.
- Define nested blocks for structured arrays (e.g., rule conditions/actions).
- Use sets for unordered ID collections; lists for ordered or repeated structured blocks (document ordering rationale).
- Add `MarkdownDescription` for every attribute.
- Confirm naming: snake_case, consistent with existing provider style.

## 5. Implementation Skeleton
- Create or update `<entity>.go` under `internal/data` or `internal/resource`.
- Add struct types.
- Implement `Metadata` and `Schema` methods.
- Remove any placeholder `panic("not implemented")` in modified scope.

## 6. HTTP Integration
- Add necessary imports & model structs.
- Use `utils.Curl` (refactored version if available) to perform request(s).
- Implement CRUD/Read methods:
  - **Create**: Build body from planned attributes; call POST; then Read.
  - **Read**: Fetch canonical state; gracefully handle 404 if resource intended to be destroyed.
  - **Update**: Compare planned changes; if ForceNew attribute changed, Terraform framework should plan replace automatically; otherwise call PATCH; finalize with Read.
  - **Delete**: Call DELETE (or perform toggle endpoints if a “soft” disable).
- Activation toggles: Branch on desired vs current state and invoke appropriate endpoints.
- Add retry for 429 via shared helper (if implemented).
- Add User-Agent header if available.

##  7. Error Handling
- Wrap all non-2xx responses: include HTTP code + mapped meaning from `errors.toml`.
- Shorten body snippet (truncate >300 chars).
- For unexpected JSON, diagnostic with parse error context (status + mapped meaning + truncated body snippet).
- For parsing errors: dedicated diagnostic, no panic.

## 8. State Population
- Map response JSON → typed internal Go model struct(s) → Terraform state objects.
- Use/Convert to Terraform framework types (`types.StringValue`, `types.BoolValue`, etc.)
- Use `types.<Type>Null()` for absent/missing/nullable fields.
- Sort sets deterministically before assigning (if needed).
- Guarantee Read method always sets all declared attributes (avoid partial state).

## 9. Idempotency & Drift Control
- After Create/Update always Re-run Read, to capture server normalization.
- Verify subsequent plan is a no-op (no drift).
- If server mutates canonical field (e.g., case normalization), accept server value and document in attribute description.
- Consider `ignore_changes` guidance if server normalizes fields.

## 10. Logging
- `tflog.Debug` at method entry & exit with keys: `operation`, `endpoint`, `status_code`.
- `tflog.Debug` at start/end of CRUD/Read with endpoint + method.
- `tflog.Trace` for request payload size / raw endpoint building (avoid sensitive).
- Ensure no credential or API key leakage/exposure.

## 11. Testing (Initial)
- Unit: Mock success + error (e.g., 422 validation).
- If activation logic present: test path with active=true → POST activation endpoint called.
- If complex nested parsing: add targeted unmarshal test.
- Add TODO block for acceptance test (`TF_ACC`) including required preconditions, if endpoint requires live interaction.

## 12. Documentation & Examples
- Add or update Terraform usage example referencing new resource/data source (See `.claude/examples/terraform/{datasource,resource}`)
    - **Resource** goes in `examples/resource/<entity>`.
    - **Data Source** goes in `examples/data-source/<entity>`.
- Update README “Supported Resources & Data Sources” list.
- Inline resource comments: ForceNew fields, toggles, server-normalized values.
- If list data source: show example with pagination query args.
- Include notes for required preconditions (e.g., verified recipient before alias attachment).

##  13. Build & Diagnostics
- Run `just build` (no compile errors).
- No `panic(` left in updated code.
- Ensure Provider address still correct.
- No unused imports (`just code-qa` clean).

## 14. Code Quality & Consistency
- Attribute names match schema draft.
- Types correct (string vs bool vs int).
- ForceNew semantics enforced by marking attributes `Required` + absence from Update logic (or explicit plan modifier if needed).
- Version reference uses central version constant only.
- Validate attribute naming matches conventions; snake_case, consistent with existing provider style.

## 15. Acceptance Test Plan (If Mutable Endpoint)
- Outline steps in code comment:
  - Preconditions.
  - Precheck (env var set).
  - Create → Read → Update → Read → Destroy sequence.
    - Create config.
    - Update attribute.
    - Verify toggles.
    - Destroy.
- Mark with `// TODO: implement acceptance test (TF_ACC)`

## 16. Commit Preparation
- Single logical commit (or minimal stack) per entity.
    - Stage only related changes (avoid mixing unrelated refactors).
- Commit message
    - includes: `[feature] add <entity> (<resource|data source>) + tests + example`.
    - Summarizes endpoint implemented + tests + example additions.
    - use message template if available (`.github/templates/commit.*.md`)
- Ensure Diff free of secrets and debug prints.

## 17. Exit Criteria (All Must Be True)
- CRUD/Read behavior matches Method Plan Table.
- Example usable without modification (besides secrets).
- Example config present and accurate.
- No drift on second plan.
- No panics / silent failures.
- Error mappings correct.
- Diagnostics meaningful.
- Logging present, structured, & standarized.
- Endpoint fully operable (plan/apply lifecycle; unless endpoint is down).
- Checklist items satisfied.
- Ready to proceed to next endpoint.

---
End of checklist.
