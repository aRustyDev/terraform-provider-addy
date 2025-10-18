Below is a detailed explanation of the classification guidance I referenced earlier: how to decide whether an Addy.io API endpoint should be exposed as a Terraform resource or a data source.

---

## Core Terraform Concepts

- Resource: Represents something you can create, update, delete, or otherwise actively manage. Terraform owns lifecycle.
- Data Source: Reads external data; does not declare ownership. Used for lookup, introspection, inventory, or derived context.

---

## Classification Pattern Explained

The earlier summary:
> “Usually: single object = resource (if mutable), list endpoint becomes data source returning collection. If an endpoint is read-only (e.g. `app-version`), data source only.”

means:

1. If the API allows you to create (POST), update (PATCH/PUT), and/or delete (DELETE) an entity, expose a Terraform resource for an individual instance of that entity.
2. If the API only lets you read (GET) something and never mutate it—either because it’s inherently immutable (e.g. app version info) or not user-controlled—represent it as a data source.
3. List endpoints (e.g., `/api/v1/domains`) generally return collections; these are better modeled as data sources that give you a set/list of IDs or objects you can feed into `for_each` if you want Terraform-managed resources per item.
4. Activation/deactivation or toggle endpoints (e.g., “active-<entity>”) should not become separate resources; instead they become boolean attributes (`active`, `catch_all`, etc.) on the main resource and are managed in Create/Update.

---

## Decision Tree (Agent-Friendly)

Ask these questions for any endpoint:

1. Does the endpoint return a collection (plural)?
   - Yes → Likely a list data source (e.g., `addy_domains`, `addy_aliases`).
   - No → Continue.

2. Can the user create instances of this entity (POST exists)?
   - Yes → Provide a resource for a single instance (e.g., `addy_domain`).
   - No → Continue.

3. Can the user modify or delete it (PATCH/PUT/DELETE exists)?
   - Yes (even if no POST, but modifiable) → Resource (rare case; not typical here).
   - No → Data source (read-only).

4. Is the endpoint specialized (toggles: activate/deactivate, enable/disable)?
   - Yes → Fold into an attribute of an existing resource, not a separate resource.

5. Is the endpoint a derived/system-level constant (e.g., version, config options, token info)?
   - Yes → Data source.

6. Is the “single object” representing something not user-provisioned (e.g., account summary)?
   - Yes → Data source (e.g., `account-details`).

---

## What “Single Object = Resource (If Mutable)” Means

A “single object endpoint” refers to operations on a specific entity instance at a path like `/api/v1/domains/{id}`. If there is a corresponding creation path `/api/v1/domains` (POST) and potentially update/delete paths, you model that as a resource because Terraform can manage it.

If a “single object endpoint” only has GET and no mutation semantics (e.g., `/api/v1/app-version`), Terraform can only read it, so it becomes a data source.

---

## Addy.io Endpoint Classification Examples

| Endpoint (Base) | Methods Present | Mutability | Suggested Terraform Construct | Notes |
|-----------------|-----------------|------------|-------------------------------|-------|
| `/api/v1/domain-options` | GET only | Read-only | Data source (`addy_domain_options`) | Used to choose domain defaults |
| `/api/v1/app-version` | GET only | Read-only | Data source (`addy_app_version`) | Self-hosted version info |
| `/api/v1/api-token-details` | GET only | Read-only | Data source (`addy_api_token_details`) | Current token metadata |
| `/api/v1/account-details` | GET only | Read-only | Data source (`addy_account_details`) | User account info |
| `/api/v1/account-notifications` | GET only | Read-only | Data source (`addy_account_notifications`) | Notifications list |
| `/api/v1/domains` / `{id}` | POST, GET list/single, PATCH, DELETE + activate/catch-all toggles via other endpoints | Mutable | Resource (`addy_domain`) + List data source (`addy_domains`) | Active + catch_all modeled as booleans |
| `/api/v1/aliases` / `{id}` | POST, GET list/single, PATCH, DELETE + activate/limit recipients via other endpoints | Mutable | Resource (`addy_alias`) + List data source (`addy_aliases`) | Recipients list (Set or nested blocks) |
| `/api/v1/recipients` / `{id}` | POST, GET list/single, DELETE + encryption/protected headers toggles | Mutable | Resource (`addy_recipient`) + List data source (`addy_recipients`) | Encryption flags as attributes |
| `/api/v1/rules` / `{id}` | POST, GET list/single, PATCH, DELETE + activation + reorder | Mutable | Resource (`addy_rule`) + List data source (`addy_rules`) | Nested conditions/actions blocks |
| `/api/v1/usernames` / `{id}` | POST, GET list/single, PATCH, DELETE + activation/catch-all/login toggles | Mutable | Resource (`addy_username`) + List data source (`addy_usernames`) | Toggle endpoints → attributes |
| `/api/v1/failed-deliveries` / `{id}` | GET list/single, DELETE, resend/download endpoints | Mostly read + auxiliary actions | Data source (list + single) first; resource optional later if managing deletion | Start as data source unless managing purge/resend states becomes important |
| Bulk endpoints (e.g., `/api/v1/aliases/activate/bulk`) | POST only affecting multiple existing objects | Batch actions | Defer; optionally future helper functions or specialized data source | Not core lifecycle management |

---

## Why Separate List Data Sources From Resources?

- Terraform state is clearest when each managed object is one resource instance.
- List endpoints (`GET /entities`) return arbitrary numbers of objects—Terraform shouldn’t “own” all of them implicitly; that would cause churn if external items appear/disappear.
- A list data source gives you a filtered inventory; you can then selectively manage specific items as resources using `for_each` with stable keys.
- This separation avoids Terraform destroying objects you didn’t intend to control just because they weren’t in configuration.

---

## Handling Toggle Endpoints

Examples:
- Activate alias: `POST /api/v1/active-aliases`, deactivate alias: `DELETE /api/v1/active-aliases/{id}`
- Enable catch-all domain: `POST /api/v1/catch-all-domains`, disable: `DELETE /api/v1/catch-all-domains/{id}`
- Enable protected headers for recipient: `POST /api/v1/protected-headers-recipients`

Pattern:
- Represent as a boolean attribute on the resource (`active`, `catch_all`, `protected_headers`).
- During Create/Update:
  - If desired state is true and current is false → call activation endpoint.
  - If desired state is false and current is true → call deactivation endpoint.
- This avoids cluttering provider with many micro-resources whose lifecycle is conceptually part of the parent entity.

---

## Edge Cases & Guidance

1. Single-object endpoints with only GET (e.g., `api-token-details`) → Always data source.
2. Entities that can be created but not updated (if hypothetical) → Resource with ForceNew attributes; any change triggers replacement.
3. Entities that can be updated partially—only expose attributes actually patchable; mark others ForceNew.
4. Endpoints performing actions (resend failed delivery) without changing resource identity:
   - Action endpoints are often better implemented via an attribute diff (e.g., setting a `resend_at` timestamp) or deferred until you introduce “action” resources (rare in Terraform). Usually postpone or leave unmodeled.

---

## Practical Implementation Flow Per Entity

1. Start with resource schema (only core fields + toggles).
2. Implement Create (POST) + Read (GET).
3. Add Update (PATCH) where documented.
4. Add toggle handling inside Update (or Create if initial state differs).
5. Add Delete (DELETE).
6. Add auxiliary list data source returning summaries (IDs + key attributes).
7. Add single-object data source only if you anticipate referencing an unmanaged existing instance.

---

## Common Mistakes to Avoid

- Modeling list endpoints as a single resource that contains all items (hard to reconcile and causes drift).
- Creating separate resources for every toggle endpoint (bloats provider).
- Marking every field as Required when server can default values (makes updates painful).
- Treating read-only endpoints as resources (Terraform will expect lifecycle operations that don’t exist).

---

## Suggested Schema Attribute Categories (Example: Alias Resource)

| Attribute | Category | Reason |
|-----------|----------|--------|
| `local_part` | ForceNew (Required) | Cannot patch; changing implies new alias |
| `domain` | ForceNew (Required) | Bound at creation |
| `description` | Optional (Patchable) | Can update via PATCH |
| `from_name` | Optional (Patchable) | PATCH-supported |
| `active` | Optional+Computed (bool) | Managed via activation endpoints |
| `attached_recipients_only` | Optional+Computed | Toggle endpoints exist |
| `recipients` | Computed (Nested) | Returned from API; modify via separate attach endpoints or PATCH if supported |
| Counts (`emails_forwarded`, etc.) | Computed | Server metrics |
| Timestamps (`created_at`, `updated_at`, etc.) | Computed | Server-generated |
| `deleted_at` | Computed or Optional+Computed? (Nullable) | May appear after deletion restore operations |

---

## Summary

The classification pattern is a structured approach to keep Terraform semantics aligned with the API’s capabilities:
- Resource = Something you can materially manage (create/update/delete/toggle).
- Data Source = Something you look up (read-only, inventory, system metadata).
- List endpoints = Data sources; individual items = resources (if mutable).
- Toggle endpoints = Attributes within resources, not standalone entities.

If you want, I can next generate Method Plan Tables for each major entity to jumpstart implementation—just let me know which one you’d like first (e.g., domain, alias, recipient, rule, username).
