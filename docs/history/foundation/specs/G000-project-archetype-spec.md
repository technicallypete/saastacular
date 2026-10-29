> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G000 — Project archetype work specification

Goal: [G000-project-archetype.md](../goals/G000-project-archetype.md)  
Epic: EPIC-000  
State: Prepared for implementation planning; implementation not started

## Scope and authority

Deliver a generic repository-owned multi-tenant SaaS project archetype Skill, supporting resources, and reproducible validation evidence. This specification does not authorize implementing the platform application. The authoritative source is ARCHITECTURE.md (not included), using section titles because its numbering repeats. Follow AGENTS.md (not included) and the roadmap (not included) for plan review and acceptance.

The Skill must establish or guide the architectural foundations below. Templates and executable examples are allowed only inside the Skill package or disposable validation destinations. They must be minimal and generic, not a complete production foundation. For each requirement, identify whether it is enforced by a template, exercised by a fixture, or prescribed for a later consumer implementation. Guidance alone must not be reported as a passing runtime test.

## Architectural requirements

### Skill contract and configuration

- Provide a valid `SKILL.md` with name/description metadata, clear activation conditions, input collection, ordered workflow, references, output contract, and verification steps. Keep essential invariants in the entry point; place detailed commodity guidance in supporting resources.
- Accept product name, platform base domain, database name, custom-domain CNAME target, chosen identity/infrastructure providers, host-port overrides, and whether intelligence is enabled. Document required versus optional values and validate invalid/missing values before producing output.
- Resolve runtime product identity through `PRODUCT_NAME`, domains through `PLATFORM_BASE_DOMAIN`/`CUSTOM_DOMAIN_CNAME_TARGET`, and database naming through `DATABASE_NAME`. Use neutral sample defaults. Product literals may not be embedded in runtime templates, prompts, metadata, or generated text.
- Produce an explicit manifest of generated files and follow-up foundation requirements. A guidance-first implementation is acceptable if the workflow and templates yield verifiable outputs in the disposable trials.
- Inspect the destination before writes. Do not silently overwrite existing files, delete content, replace repository instructions, embed secrets, or provision external resources. Specify collision behavior and demonstrate safe repeat invocation.
- Compose suitable existing skills for commodity setup when available. Record what is reused, why, access/version assumptions, and unavailable-skill fallbacks. Do not require machine-specific absolute paths or uncatalogued skills without documenting the dependency.

### Monorepo service boundaries and non-redundant naming

| Directory | Responsibility and required boundary |
|---|---|
| `web/` | Next.js UI and thin BFF/session bridge. Tailwind + shadcn/ui with Base UI and semantic white-label tokens. No authoritative policy or domain rules. |
| `core/` | Go modular monolith; application/domain services, authorization, identity/tenant control plane, official MCP Go SDK transport, durable state, audit, and public streaming lifecycle. MCP handlers delegate to services. |
| `intelligence/` | Optional internal Python/FastAPI inference execution. Receives authenticated, Go-authorized envelopes; cannot establish tenant authority or independently mutate authoritative domain state. |
| `contracts/` | Cross-service protocol/schema/event contracts only; no shared business-logic dumping ground. |
| `infra/` | Docker Compose, images, database/storage setup, and deployment guidance. |
| `skills/`, `docs/`, `management/` | Reusable guidance, architecture/runbooks, and goal lifecycle artifacts respectively. |

Top-level names describe responsibility. Do not generate `frontend/`, `backend/`, `api/`, `app/`, or ambiguous `services/` tiers. Avoid redundant paths such as `api/api`, `app/app`, `core/internal/core`, `core/cmd/core`, and `ai/ai`. Prefer `web/app`, `core/cmd/server`, `core/internal/<responsibility>`, `intelligence/src`, and `infra/compose/local.yml`. A genuine framework/packaging exception must be explained in the plan.

Do not populate directories with placeholder product modules. With intelligence disabled, no Python service/runtime or AI credential requirement may remain in the normal startup/check path; document whether its directory is omitted or retained as guidance only.

### Docker-first local development and ports

All project services, builds, checks, migrations, and fixture tests must run through Docker images/Compose. Host prerequisites are Git and Docker Engine/Desktop with Compose and applicable WSL integration; Claude Code/Codex are authoring tools. **No host Go, Python, Node, PostgreSQL, Redis, or MinIO installation is required by default.** Any optional debugger/IDE/harness exception must name the concrete workflow and retain a container-only normal path.

Compose guidance/templates must include PostgreSQL, Redis, MinIO, Go core, web, and optionally intelligence, with health/readiness, persistent development volumes, and documented start/check/stop behavior. Redis is ephemeral/cache infrastructure; PostgreSQL is the durable system of record. ObjectStore uses S3-compatible APIs with MinIO locally and configurable production adapters. Host-bound services must not unintentionally expose databases or internal intelligence beyond local development.

Use conventional service port +100 for host exposure where a standard exists; internal ports stay conventional where practical:

| Service | Container port | Default host port | Configuration |
|---|---:|---:|---|
| Web | 3000 | 3100 | `WEB_PORT` |
| Go core | 8180 (chosen default) | 8180 | `CORE_PORT` |
| Intelligence | 8000 | 8100 | `INTELLIGENCE_PORT` |
| PostgreSQL | 5432 | 5532 | `POSTGRES_PORT` |
| Redis | 6379 | 6479 | `REDIS_PORT` |
| MinIO API | 9000 | 9100 | `MINIO_API_PORT` |
| MinIO Console | 9001 | 9101 | `MINIO_CONSOLE_PORT` |

Go has no single standard service port, hence the explicit 8180 default. Validate port ranges/collisions and use Docker service DNS/internal ports for container-to-container traffic. Changing host mappings must not break internal connections.

### Global identity and Organization / Group / BusinessEntity model

- One global Identity may belong to multiple Organizations with different roles. Integrate identity providers behind an interface, with session/revocation and MFA/SSO capability guidance; do not build a custom authentication provider.
- Organization is the persisted tenant/security boundary. Use `organization_id` on tenant-owned records; `TenantContext` is a trusted security envelope, not a second persisted tenant entity.
- Group is an Organization-owned collaboration scope with immutable UUID and editable kebab-case slug. Require `UNIQUE (organization_id, slug)` and reserved-route validation. Slugs and hostnames never grant access.
- BusinessEntity is reusable Organization-owned business identity/context (Business Profile in generic UI guidance), associated with multiple Groups through GroupBusinessEntity. Enforce same-Organization associations with database integrity constraints as well as service validation.
- Specify OrganizationMembership, GroupMembership, and direct BusinessEntityMembership/grants where needed. BusinessEntity is not a one-Group child profile.
- Describe generic Group onboarding: create Group, create or select an existing BusinessEntity under sufficient owner/admin authority, establish the association, then complete onboarding. Unauthorized IDs cannot complete the flow.
- Preserve extensibility for BusinessEntity-owned multiple billing profiles without implementing billing schemas/workflows in the archetype. Generic label guidance, if supplied, validates owner/target scope and defaults Organization-private metadata to private.

### RoleBinding authorization and PostgreSQL RLS

- Define shared `RoleBinding(subject, scope_type, scope_id, role, source)` and one policy/capability system. Derive effective permissions server-side from scoped membership, direct grants, capability-specific Group inheritance, and Organization policy; do not mirror Group grants onto BusinessEntity records.
- Deny by default. A Group administrator's ordinary Business Profile editing authority must not imply ownership or other privileged administrative authority. The permission matrix must distinguish these capabilities using generic administration examples.
- Resolve request context from authenticated session and validated Organization/Group grants. Browser/model-supplied identifiers are selectors, not authority. Hostname routing accepts configured platform subdomains or verified custom domains and still requires authenticated authorization.
- Prescribe PostgreSQL RLS for tenant-owned tables and Group/resource scope defense in depth. Cover reads and insert/update/delete checks, same-tenant foreign keys, and global identity tables explicitly rather than applying tenant policies blindly.
- Separate migration/owner and runtime roles. Runtime access must not be superuser or `BYPASSRLS`; table-owner bypass must be prevented by role separation or appropriate forced policies. Scope security-definer functions narrowly and document any privileged administrative path.
- Set tenant context safely for each transaction; missing/invalid context fails closed. Demonstrate reset/isolation with pooled connections and prevent connection context from carrying across requests.
- ObjectStore operations and signed URLs require exact resource authorization; prefixes alone do not isolate data. Credentials are scoped and server-resolved.

### Audit foundation and operational hooks

Prescribe controlled durable AuditEvent records with actor/service identity, Organization, optional Group, resource identity/type, action, outcome, timestamp, and trace/request correlation. Include security denials, context changes, membership/role changes, and administrative mutations. Restrict modification/deletion through controlled roles and retention policy; an append-only application API alone is insufficient protection.

Provide secrets-provider boundaries, structured redacted logs, health/readiness, metrics/trace hooks, access provisioning/deprovisioning, vulnerability/dependency/secret scanning, CI/change evidence, encryption guidance, backup/restore procedures/tests, retention/deletion hooks, incident evidence, and vendor inventory conventions. Identify consumer implementation and production configuration responsibilities. These are SOC 2 readiness foundations, not a compliance claim.

### MCP Organization/Group switching and dynamic tool exposure

Use the small authenticated control-plane surface: `account.get_current_user`, `account.list_organizations`, `account.get_current_context`, `account.switch_organization`, `context.list_groups`, `context.switch_group`, and logout.

The flow must authenticate, validate Organization membership, optionally validate Group membership/effective grant, resolve capabilities, and invalidate/recalculate the catalog. Switching Organizations clears incompatible Group/resource context. Normal capability arguments cannot switch tenants.

`tools/list` exposes only capabilities available in the active context. Every execution independently validates identity, active Organization/Group, current membership/grants, resource ownership, required capability, and applicable policy/approval. Hidden tools reject direct calls. Membership/role/policy changes and session expiry invalidate stale capabilities. Fixtures must show a user privileged in one Organization and restricted in another without authority bleed.

### Optional AI configuration and OpenRouter-first ModelGateway

Use exactly the capability-oriented convention:

```text
AI_MODEL_<CAPABILITY>=<provider>:<model>
AI_MODEL_<CAPABILITY>_FALLBACKS=<provider>:<model>,<provider>:<model>
```

Initial capabilities are CHAT, FAST, REASONING, EXTRACTION, and EMBEDDING. Validate syntax and missing configuration for enabled capabilities; do not hardcode feature-specific model names. Illustrative model identities must be clearly samples, not claims of current provider availability. Disabled intelligence requires no model credentials or live model calls.

Application capability → ModelResolver → ModelGateway → provider adapter. OpenRouter is the initial default gateway behind ModelGateway. Translate canonical `provider:model` identities to gateway identifiers in the adapter; do not make the gateway a domain concept or require `AI_GATEWAY_DEFAULT`. Credentials such as `OPENROUTER_API_KEY` remain secret provider configuration. Alternative direct-provider adapters are extension points, not required implementations.

Go owns tenant/policy checks, public streams, cancellation/deadlines, approvals, audit, and durable usage attribution. Intelligence receives trusted envelopes and emits provider-independent typed events (text delta, tool call/result, citation, structured output, usage, error, done). Browser/BFF never calls intelligence directly in the normal application path. Authorized callbacks cannot bypass core policy.

Prescribe pre-model retrieval filtering and scoped memory/context. Usage hooks distinguish capability, configured provider/model, actual gateway/provider/model, tokens/units, latency, estimated cost, trace, user, and applicable Organization/Group/BusinessEntity dimensions. External observability excludes secrets and tenant content by default; it is not the authoritative durable usage store. No product AI workflows are included.

## Expected artifacts

| Artifact | Required contents |
|---|---|
| `skills/project-archetype/SKILL.md` | Valid Skill entry point, activation/input/output contract, ordered workflow, essential invariants, support-resource links. |
| Supporting resources under `skills/project-archetype/` | Architecture/security guidance; neutral configuration and Docker/service-layout templates; commodity-skill composition notes; validation and acceptance checklists. Exact filenames are fixed in the reviewed plan. No application implementation outside this package. |
| Disposable validation fixtures/harness under the Skill package or temporary destinations | Generic inputs, output manifests, minimal executable isolation examples, and containerized repeatable validation commands. Generated fixtures must not become this platform's application. |
| `management/v0/plans/G000-project-archetype-plan.md` | Proposed files, guidance-versus-template decisions, composition dependencies, test strategy, destination safety, exclusions, and evidence mapping for Codex review before execution. |
| `management/v0/results/G000-project-archetype-result.md` | Final artifact manifest, acceptance matrix, exact commands/outcomes, executed versus prescribed checks, fixture configuration, limitations, and review outcome. |

The plan/result and Skill are future implementation deliverables. Preparing this specification creates none of them. Do not install into a global Skill directory or provision a provider account as part of G000.

## Security requirements

Apply least privilege and fail-closed behavior at every boundary described above. Samples must use synthetic identities/data and local-only credentials, never real tenant content or production secrets. Document safe secrets injection and redaction; no credentials in committed templates, logs, generated manifests, or traces.

Any fixture identity adapter is explicitly development-only and must not accidentally enable production authentication bypass. Restrict local service exposure, authenticate internal execution contracts, and document encryption requirements for production. Audit, storage authorization, RLS, and context-switch checks are mandatory even when another provider supplies permissions.

Major infrastructure deviations require an ADR. Preserve the no-product-domain constraint across references, templates, examples, generated resources, tool names, and prompts. Exclusions may be stated in planning documents but must not become archetype implementation concepts.

## Validation/testing requirements

Validate the actual Skill workflow and resource links, not only whether requirement words appear. Use Docker for project-runtime checks. No live AI/provider account is needed; use contract fixtures and deterministic test adapters. The reviewed plan must identify minimal fixtures that demonstrate the following without building the later platform:

| Check | Required evidence |
|---|---|
| Skill structure and usability | Valid metadata, resolved relative resources, documented composition/fallbacks, and a complete disposable invocation from supplied inputs to output manifest. |
| Portability trial | Generate/guide two neutral projects with differing product names, domains, and database names. Verify outputs use their configuration and contain no hidden brand/hostname dependencies. |
| Docker-only path | Startup/build/check commands succeed using Git/Docker prerequisites with host Go/Python/Node unavailable to those commands. Report required images and tooling explicitly. |
| Intelligence toggle | Exercise enabled and disabled configurations. Disabled normal startup/check paths do not start Python or demand AI credentials. |
| Compose and port behavior | Validate Compose, configured defaults/internal DNS, port overrides, health/readiness and dependency outage behavior for provided runnable skeletons. |
| Destination safety | Reinvoke against an existing destination; existing sentinel content survives and collisions are reported or safely reconciled as documented. Invalid input fails before unsafe output. |
| Model configuration | Fixture parses canonical identities/fallbacks, rejects malformed values, translates through an OpenRouter adapter contract, and preserves configured/actual identity attribution without a live request. |
| Tenant isolation | At least two Organizations, multiple Groups, a shared BusinessEntity within one Organization, and an identity with differing roles. Positive controls establish permitted access; negative cases deny foreign IDs and unauthorized same-Organization Group access. |
| RLS and database roles | Minimal PostgreSQL fixture exercises real non-bypass runtime roles for reads/writes, absent/invalid context, cross-Organization links, and pooled/connection-reuse context leakage. Do not run these assertions only as owner/superuser. |
| Shared policy | Demonstrate ordinary inherited BusinessEntity editing versus restricted ownership administration; role removal changes effective access without mirrored permission rows. |
| MCP catalog/execution | Minimal contract fixture verifies switching clears scope, catalogs differ by active context, hidden direct calls fail, and stale/revoked permissions cannot execute. Clearly label mocked transport versus real SDK checks. |
| Storage and audit | Minimal adapter/policy fixtures demonstrate foreign-object signed access denial and controlled audit attribution/modification. Report what is mocked versus exercised against MinIO/PostgreSQL. |
| Security/maintainability | Review secret/brand/domain leakage, non-redundant directory structure, service dependency direction, and CI/security/evidence guidance. Check supported resource syntax with containerized tools as appropriate. |

Do not claim full production security or completion of G004–G020 from these bounded examples. Requirements intentionally prescribed for later goals still need a concrete implementation checklist and future verification method. If execution is unavailable, record the exact limitation and leave the affected acceptance item unmet for review.

## Acceptance criteria

1. All expected Skill artifacts exist and are consumable without this project's product knowledge; links, metadata, configuration contract, and workflow are valid.
2. Every architectural/security requirement maps to a specific Skill resource and enforcement/validation method. Later consumer obligations are explicit.
3. Disposable portability, intelligence-toggle, Docker-only, and destination-safety trials have recorded passing evidence.
4. Minimal tenant/RLS/policy/MCP/storage/audit fixtures provide the positive and negative evidence above, with mock/runtime distinctions and no bypass-role false positives.
5. Folder boundaries, standard+100 host mappings, canonical tenant/RoleBinding model, audit, AI configuration, and OpenRouter boundary match the architecture without introducing product feature domains.
6. The diff changes only the Skill/support fixtures and goal-lifecycle artifacts authorized by the reviewed plan. The platform application remains unimplemented by G000.
7. No major unapproved infrastructure dependency, production provisioning, global installation, real secret, or live model dependency is introduced.
8. Codex reviews architecture, security, diff, and evidence and returns **ACCEPT**. Any unmet criterion remains a correction or architecture decision; tests alone do not confer acceptance.
