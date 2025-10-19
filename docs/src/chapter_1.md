# Terraform Addy Provider Documentation (Index)

Welcome to the user-facing documentation for the Terraform Addy provider. This mdBook focuses on conceptual guidance, upgrade strategy, security posture, and operational behavior. For contributor-facing internal policies and design rationale, see the repository files (`.rules`, `CLAUDE.md`, `AGENT_CHECKLIST.md`).

## Quick Start

```hcl
terraform {
  required_providers {
    addy = {
      source  = "registry.terraform.io/addy-io/addy"
      version = "~> 1.0"
    }
  }
}

provider "addy" {
  # Prefer passing via variable or environment variable ADDY_API_KEY
  api_key = var.addy_api_key
}
```

## Documentation Layers

| Layer | Purpose | Location |
|-------|---------|----------|
| Narrative concepts (you are here) | Explain behavior & usage patterns | mdBook (`docs/`) |
| Schema reference | Argument & attribute tables | Generated provider docs |
| Internal policies | Normative contributor rules | `.rules` |
| Design rationale | WHY decisions were made | `CLAUDE.md` |
| Execution workflow | HOW to implement endpoints | `AGENT_CHECKLIST.md` |
| Troubleshooting (dev) | Contributor pitfalls | `.claude/FAQ.md` |

## Concept Guides

- Architecture Overview (concepts/architecture.md)
- Versioning & Upgrade Guide (concepts/versioning.md)
- Pagination & Collection Retrieval (concepts/pagination.md)
- Retry & Backoff Strategy (concepts/retry.md)
- Security & Sensitive Data Handling (concepts/security.md)
- Observability & Diagnostics (concepts/observability.md)

## Future / Reserved Sections

- Resource Pages (resources/...) – added as each endpoint is implemented
- Data Source Pages (data-sources/...) – added per read-only endpoint
- Migrations (migrations/...) – populated only for MAJOR releases

## Release Cadence (Summary)

One new endpoint (resource or data source) per MINOR version. PATCH releases contain fixes or computed-only additions. See the Versioning guide for details.

## Reporting Issues

When filing an issue, include:
1. Provider version (and `terraform version`)
2. Relevant config snippet (redact secrets)
3. Plan/apply output or diagnostic
4. Any debug log lines (sanitized)

## Contributing Docs

If you add or change user-facing behavior:
- Update or add a concept page (avoid duplicating schema tables).
- Add new resource/data source page with scenario examples.
- Update this index if navigation categories change.

---

_Last updated: 2025-10-19_
