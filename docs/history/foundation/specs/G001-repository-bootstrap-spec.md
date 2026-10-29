> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G001 — Repository bootstrap work specification

Goal: [G001-repository-bootstrap.md](../goals/G001-repository-bootstrap.md)
Epic: EPIC-001 — Repository and local platform
State: Prepared for Claude implementation planning; implementation not authorized

## Authority and actual starting state

Follow ARCHITECTURE.md (not included), AGENTS.md (not included), CLAUDE.md (not included), the roadmap (not included), and the Goal's budget/execution policy. This Spec defines required outcomes; Claude's reviewed Plan chooses implementation details.

The inspected starting baseline is commit `bf0f20b3aec2967b3afd74c94402e523baed2a77` (`G000: project-archetype Skill implementation and correction loop 1`). The working tree was clean before these G001 planning artifacts. G000 is accepted in PROGRESS.md. The repository contains governance, canonical documentation, management artifacts, and the committed reusable Skill. Platform root `web/`, `core/`, `intelligence/`, `contracts/`, and `infra/` do not yet exist. Root README, environment example, and ignore rules are also not yet present.

The accepted renderer supplies service skeletons, Compose networks, database roles, local storage initialization, health/readiness, redacted logging, runtime identity, check containers, intelligence contracts, and foundation documentation. Its Go template has pgx/S3 dependencies but no MCP SDK initialization. It supplies no migration runner, accessible web shell, or CI workflow. Its isolation fixtures demonstrate contracts; they are not platform runtime implementations. The G000 Result distinguishes template-enforced, fixture-tested, and later prescribed requirements; preserve that distinction.

## Requirements

### R1 — Consume and preserve the accepted archetype

Use `skills/project-archetype/generator/render.sh` to validate, dry-run, and render the platform skeleton into this repository after Plan approval. Record the starting commit, Skill version, resolved non-secret inputs, output ID, manifest, planned actions, actual actions, and any reviewed reconciliation in the Result.

Preserve existing architecture, governance, management, and Skill files. Keep generated foundation guidance under `docs/foundation/`. Retain `.archetype/outputs/<output_id>/manifest.json` as the initial renderer provenance. Identify subsequent G001 project additions separately; do not falsify the manifest to imply they came from an unchanged template. Repeat rendering must preserve files and report expected conflicts for deliberate project customization. Any unexpected collision requires review before reconciliation.

Do not amend the accepted Skill for this consumer goal. Containerized project validation must be maintained outside its package.

### R2 — Configuration and bootstrap defaults

These are assigned local defaults, not production provisioning or permanent product/data decisions:

| Renderer input | G001 value |
|---|---|
| `product_name` | `Beginify` |
| `project_slug` | `beginify` |
| `platform_base_domain` | `foundation.test` |
| `custom_domain_cname_target` | `custom.foundation.test` |
| `database_name` | `beginify` |
| `intelligence_enabled` | `true` |
| `identity_provider` | `oidc` (configuration label only; no selected IdP integration) |
| `production_object_store` | `r2` (documented future adapter; no production account/implementation) |
| `ports` | Accepted standard+100 defaults below |
| `ai_models` | Omitted; no enabled model capability or gateway key required |

Runtime identity comes from `PRODUCT_NAME`, `PLATFORM_BASE_DOMAIN`, `CUSTOM_DOMAIN_CNAME_TARGET`, `DATABASE_NAME`, and `COMPOSE_PROJECT_NAME`. Identity literals belong in environment configuration and renderer provenance/planning evidence, not application source or identity-dependent generated docs. Preserve input validation, including explicit-boolean intelligence flags and null rejection, environment fail-closed checks, configurable ports, and safe local secret creation. No hardcoded model identity, production credential, custom-domain DNS action, or required `AI_GATEWAY_DEFAULT` is added.

### R3 — Service boundaries and local runtime

| Boundary | Required bootstrap outcome |
|---|---|
| `web/` | Adopt the rendered Next.js surface, semantic tokens, and server-side core bridge; no domain policy, datastore clients, or browser access to intelligence. |
| `core/` | Adopt the Go executable/config/health/logging skeleton; add migration integration and dormant MCP initialization within its responsibility. |
| `intelligence/` | Adopt the internal FastAPI skeleton and contracts; no datastore credentials, authoritative mutations, or public browser integration. |
| `contracts/` | Adopt cross-service schemas/documentation; no shared business logic or speculative product schemas. |
| `infra/` | Adopt Compose, images, role initialization, scoped MinIO setup, checks; add bounded migration/validation support. |
| `docs/`, `management/`, `skills/` | Preserve canonical and historical content; add only reviewed project setup documentation and G001 lifecycle artifacts. |

Use responsibility-based names; `web/app`, `core/cmd/server`, `core/internal/mcp`, and `core/migrations` are appropriate. Do not create generic tier roots or empty placeholder domain modules.

All services, builds, checks, migration commands, and project-runtime tests run in Docker. Host prerequisites are Git, Docker with Compose 2.24+, and the shell/coreutils required by the accepted scripts; authoring agents are separate tooling. No host Go, Python, Node, PostgreSQL, Redis, or MinIO installation is required. Preserve pinned dependencies/images and lockfiles; any additions must be identified and pinned in the Plan.

| Service | Internal port | Default loopback host port |
|---|---:|---:|
| Web | 3000 | 3100 |
| Go core | 8180 | 8180 |
| Intelligence | 8000 | 8100 |
| PostgreSQL | 5432 | 5532 |
| Redis | 6379 | 6479 |
| MinIO API | 9000 | 9100 |
| MinIO console | 9001 | 9101 |

Retain edge/data/inference separation and service DNS on internal ports. Web cannot reach data stores or intelligence; intelligence cannot reach data stores; core can reach the required services. Published development ports bind to loopback. PostgreSQL and MinIO persist across ordinary stop/start; Redis remains ephemeral. Retain source-built local MinIO; no replacement vendor is authorized.

Adopt existing health/readiness behavior: core liveness stays available during dependency outages, PostgreSQL/Redis/object-store failures gate readiness and identify the dependency, intelligence availability is reported without gating core readiness, and web readiness depends on core. Preserve redaction and correlation scaffolding without expanding G003 observability scope.

### R4 — Minimal database migration foundation

The Plan must select and pin a migration tool/library, describe its container commands and metadata schema, and demonstrate migration status/up/down/reapply on a disposable database. Runtime startup and migration execution must have distinct privileges.

Preserve `<db>_owner` as NOLOGIN, the separate migrator, runtime, and retention roles, and the initialization-only superuser. Migration-created objects and migration metadata must be owned by the intended owner role through the migrator's bounded `SET ROLE` path. Runtime must have no superuser, BYPASSRLS, table ownership, migration authority, or ability to assume privileged roles. Migrator credentials must not be supplied to core/web/intelligence runtime containers.

Provide only migration infrastructure/metadata and a minimal empty-schema baseline; no Identity, Organization, Group, BusinessEntity, RoleBinding, audit, or product table is created. A reversible synthetic migration used to prove DDL ownership and rollback belongs only to disposable validation. Record that migration metadata is system infrastructure, not a tenant-owned domain table. Tenant policy and schema migrations remain G005–G008.

### R5 — Dormant MCP initialization

Use the official MCP Go SDK for a minimal server-construction boundary in core, with an empty tool catalog and in-process SDK lifecycle tests. Pin the SDK dependency and document its adoption. No network MCP endpoint, public transport mount, globally shared authenticated catalog, account/context tool, development identity adapter, or authentication bypass is authorized in G001.

The accepted follow-up guidance gates transport mounting on authentication. Preserve that gate: dormant construction is the bounded G001 initialization; authentication is G004 and context/discovery/reauthorization is G014/G015. In-process synthetic protocol peers used only in tests must not become a runtime authentication path. Tests must establish that construction works and anonymous HTTP requests cannot initialize MCP through the running core.

### R6 — Usable project commands and foundation handoff

Provide a root setup entry point and foundation docs describing container-only configure/start/check/stop, migration status/up/down, local ports, reset implications, and remaining foundation work. Commands must work as written from the actual repository. Preserve existing foundation stages and label only obligations genuinely completed by G001; dormant MCP construction does not complete identity or MCP control-plane stages.

Keep the intelligence-enabled root skeleton and validate a disposable intelligence-disabled consumer variant. The disabled variant must start/check without a Python service or AI credentials. Do not implement a destructive toggle that deletes the root intelligence service. Keep model configuration and the OpenRouter-first gateway scaffolding unchanged; use deterministic contract tests and no live providers.

## Security invariants and explicit non-goals

Preserve the canonical global Identity / Organization / Group / BusinessEntity model, shared RoleBinding authorization, server-resolved context, PostgreSQL RLS defense in depth, exact-object storage authorization, and controlled audit requirements as later foundation obligations. An empty bootstrap database, health checks, or the G000 fixtures do not establish tenant isolation for a running product. Expose no tenant or product operation before its security foundation exists.

Secrets remain empty in committed examples, local in ignored `.env` with restricted permissions, and injected without command-line disclosure. Exclude local environment files from Docker build contexts. No secret may appear in logs, manifests, command reports, or result evidence. Runtime services receive only their own credentials; web receives no datastore secrets and intelligence receives no datastore configuration. No real data/provider accounts are used.

No accessible web shell/shadcn initialization, CI workflow, new metrics subsystem, hot reload, production deployment, tenant schemas, audit ledger, signed URLs, dynamic MCP catalog, hostname authorization, onboarding, retrieval/public AI streaming, product features, or MinIO replacement is included. Inherited tests/schemas remain inherited scaffolding rather than completion claims for later goals.

## Expected artifacts

- Rendered `web/`, `core/`, `intelligence/`, `contracts/`, `infra/`, `.env.example`, `.gitignore`, `.archetype/outputs/`, and `docs/foundation/` resources from the accepted Skill.
- Project-owned migration commands/baseline, dormant MCP construction/tests, and bounded bootstrap validation within `core/` and `infra/`; exact paths belong in the reviewed Plan.
- Root setup documentation and any necessary foundation-document amendments, identifying generated content versus consumer additions.
- Claude-owned `management/v0/plans/G001-repository-bootstrap-plan.md` before execution and `management/v0/results/G001-repository-bootstrap-result.md` after validation. This specification creates neither.
- Local `.env` and disposable runtime resources are execution artifacts, never committed deliverables.

## Acceptance criteria and required tests

| ID | Acceptance criterion | Required evidence |
|---|---|---|
| AC1 | Actual repository is bootstrapped from the accepted Skill with preserved existing content and attributable additions. | Starting commit/path inventory, resolved inputs, dry-run/render output and manifest; protected-file hashes; repeat-render test with zero unapproved writes and all customization conflicts explained. |
| AC2 | Normal configure/start/check/stop and migration paths require no host language/datastore runtimes. | Execute documented commands verbatim against the root implementation using disposable local credentials; shadow/remove host runtimes from the project command path, record zero invocations, and pass containerized builds/lint/types/unit tests/secret scanning. |
| AC3 | Identity/domain/database/port settings remain configurable. | Actual root starts at assigned defaults. A disposable copy of its implementation demonstrates rename/re-domain without rebuilding, fresh-volume database rename with correct role names, and overridden host ports with working internal DNS. Compare image IDs/source hashes; do not use only pristine Skill templates as evidence. |
| AC4 | Existing service health, isolation, least privilege, and persistence survive adoption. | Actual implementation passes dependency outage/recovery, liveness/readiness, PostgreSQL/MinIO persistence and Redis ephemerality checks. Host protocol probes confirm all published ports on loopback and refusal on a non-loopback address. Independently probe each datastore's own port by name and every IP from forbidden namespaces; negative controls detect PostgreSQL-only, Redis-only, and MinIO-only reachability, with core positive controls. |
| AC5 | Migration commands work under intended roles without introducing domain schema or granting runtime migration authority. | Disposable status/up/down/reapply tests, owner/role catalog assertions, and negative runtime DDL/privileged-role tests. Use a synthetic disposable migration for DDL proof, exclude it from the normal migration set, and verify rollback leaves expected metadata/baseline state. |
| AC6 | MCP initialization is functional and dormant. | Official SDK in-process initialization/empty-catalog tests pass; running core exposes no unauthenticated MCP transport. Source/routes and HTTP probes corroborate the boundary; no fixture authenticator or G000 isolation package appears in runtime imports. |
| AC7 | Optional intelligence and secret boundaries remain intact. | Enabled root passes existing intelligence checks without a gateway key/live request; disabled disposable variant starts/checks without Python or AI settings. Verify scoped environments/build contexts, empty committed secrets, and generated-secret canaries absent from logs/reports/manifests/committed files. |
| AC8 | Diff, documentation, Result, and budget evidence support bounded acceptance. | Review actual tracked/untracked paths against Plan; only bootstrap directories/docs/config and G001 lifecycle artifacts change. Result maps AC1–AC8 to exact commands/outcomes, distinguishes inherited/new/deferred work, and reports shared B5 usage/checkpoint/blockers. Codex returns ACCEPT. |

## Required validation and evidence handling

Test the adopted repository and its bootstrap additions, not only the already accepted generator. The Plan must map every acceptance criterion to concrete commands/tests, identify real network/database checks versus test doubles, and name all expected artifacts/dependencies. Run appropriate containerized static checks, project tests, migration checks, and integrated bootstrap trials; do not reduce validation to fit the budget. Unavailable or skipped required checks remain unmet.

Do not run the G000 repository-clean assertion unchanged against G001: platform root directories were forbidden for G000 and are intentional G001 deliverables. Use a G001-specific path/scope check against the reviewed Plan and inspect the full git diff, including untracked files. Preserve G000's historical guard rather than weakening the reusable harness to accommodate this consumer. If the unchanged Skill itself needs regression validation, run its harness in an isolated copy without G001 platform roots. Any external management-file exception must name the exact path, ownership, and inspected diff; broad allow-lists cannot substitute for scope review.

Use uniquely named disposable Compose projects/volumes for destructive rename, migration rollback, outage, and persistence tests. Preserve any existing local developer `.env`, volumes, and data; starting from an empty application tree does not authorize deleting unrelated Docker resources. Record cleanup ownership and commands. Evidence must contain no secrets.

Claude returns the Result for Codex review. Only Codex updates progress or accepts completion; preparing this Spec and Goal does not authorize rendering, implementation, G002/G003 work, or product features.
