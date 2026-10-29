> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G000 — Project archetype implementation plan

Goal: [G000-project-archetype.md](../goals/G000-project-archetype.md)  
Spec: [G000-project-archetype-spec.md](../specs/G000-project-archetype-spec.md)  
Author: Claude Code (implementer)  
Plan status: **Resubmitted for Codex plan review after correction pass 2. Implementation has not started.** MinIO is retained as specified (R1). Replacing it remains a product-owner escalation, and no replacement or partial implementation is assumed.  
Revision: r5, 2026-10-04. Execution budget amendment (Goal amendment A1, authorized by the product owner): a ceiling of 15 cycles with checkpoints at cycles 8 and 12. No scope change. Previously r4, correction pass 2. It answers the remaining part of Codex correction 6 (cycle accounting and reporting); Codex confirmed corrections 1–5 in its r3 re-review. The r1–r3 history is preserved in §12.

Architecture references below use **section titles**, because the architecture's section numbers repeat.

## 0. Summary

Deliver `skills/project-archetype/`, a **self-contained, project-agnostic** Skill. It contains:

- an entry point (`SKILL.md`) that carries the essential invariants;
- reference guidance;
- a generic requirements catalog;
- consumer checklists;
- render templates;
- a small deterministic **renderer**: Go standard library only, run in the official `golang` image through one wrapper script;
- a **validation package**: a harness, neutral inputs and minimal isolation fixtures.

The renderer makes the Skill testable. It turns a validated input file into a destination tree plus a machine-readable manifest. It never overwrites or deletes existing content, and invalid input fails before any write.

`SKILL.md` wraps the renderer in an agent workflow: preflight, input collection, destination inspection, dry run, render, conflict reconciliation, local secret creation, executable start/check commands, and follow-up hand-off.

Evidence comes from one containerized harness, `skills/project-archetype/validation/run.sh`. Its suites are self-contained:

| Suite | What it covers |
|---|---|
| static | Syntax and lint, Skill structure, traceability integrity, leakage, per-service secret scope |
| generator | Input validation and destination safety |
| intelligence | Model configuration, the OpenRouter adapter contract, internal authentication, typed events |
| isolation | Minimal tenant/RLS/policy/MCP/storage/audit fixtures, run against real PostgreSQL, real MinIO (built from the final upstream source release, R1) and the real MCP SDK. Runtime assertions use non-bypass roles; privileged connections are separated and guarded |
| trials | Two disposable generated projects with different product names, domains and databases. One has intelligence enabled and default ports; the other has it disabled and overridden ports. Covers executed start/check docs, a rename through configuration only, network boundaries, health/readiness, outages and persistence |

Repository changes are limited to `skills/project-archetype/**` and the G000 lifecycle artifacts listed in §3. Platform root `web/`, `core/`, `intelligence/`, `contracts/` and `infra/` remain absent. Generated trees exist only in temporary directories outside the repository.

## 1. Repository state at planning time

- **Git.** There is one commit (`99e2c79 Initial planning`), containing `AGENTS.md`, `CLAUDE.md`, `docs/` and `management/`. The only untracked path is this plan. There is no application code and no `skills/` directory.
- **No ADRs.** `docs/SECURITY.md`, `docs/DECISIONS.md` and `management/v0/BACKLOG.md` are empty, so there are no accepted ADRs to apply.
- **Host (WSL2).**
  - Docker Engine 27.5.1 and Compose v2.32.4.
  - **Host `python3` (/usr/bin) and `node` (nvm) are installed. Go is absent.** WSL interop may also expose Windows executables through `/mnt/*` PATH entries.
  - The Docker-only trial must therefore *prove* that no host runtime is invoked (T-DOCKER-ONLY).
- **Images and ports.** `golang:1.27.1-alpine` pulls successfully. None of the default archetype host ports (3100/8180/8100/5532/6479/9100/9101) is currently bound.
- **MinIO.** No maintained image can be pulled (R1). Building the final upstream releases from source in a container was verified to work during correction pass 1 (§10, R1, source 7).
- **Host reachability.** Probes from the host were verified during correction pass 1:
  - bash `/dev/tcp` and host `curl` both reached a container port published on `127.0.0.1`;
  - the host's non-loopback address was refused (T-PORTS).

## 2. Implementation approach

### 2.1 Enforcement classes and where traceability lives

| Method | Meaning | Reported as |
|---|---|---|
| **T — Template-enforced** | The renderer emits it, and a named harness check verifies it in generated output or a running stack | Executed check |
| **F — Fixture-exercised** | A minimal fixture in `validation/` runs it against real containers or the real SDK, or as a labelled unit test | Executed check, labelled Real or Unit |
| **P — Prescribed** | Guidance plus a concrete checklist item, a named *foundation stage* that owns it, and its future verification method | **Not** a passing test. Reported as a later consumer obligation |

Traceability is split so the reusable Skill carries no knowledge of this platform.

- **Inside the Skill.** `references/requirements-catalog.md` gives stable generic IDs: `BND-*`, `CFG-*`, `DKR-*`, `TEN-*`, `POL-*`, `RLS-*`, `STO-*`, `AUD-*`, `OPS-*`, `MCP-*`, `AI-*` and `SAFE-*`. Each ID has:
  - a T/F/P class;
  - the Skill resource that states it;
  - for T and F, the verifying check or test ID;
  - for P, the foundation stage and the verification method.

  Foundation stages are generic: `STAGE-BOOTSTRAP`, `STAGE-WEB-SHELL`, `STAGE-OBSERVABILITY`, `STAGE-IDENTITY`, `STAGE-TENANCY`, `STAGE-POLICY`, `STAGE-RLS`, `STAGE-STORAGE`, `STAGE-AUDIT`, `STAGE-OPS`, `STAGE-MCP`, `STAGE-DOMAINS`, `STAGE-ONBOARDING`, `STAGE-LABELS` and `STAGE-AI`. There are **no `G0xx` IDs inside the Skill**.
- **In management artifacts.** These map each spec bullet to a catalog ID and its evidence, and each foundation stage to this platform's G001–G046 goals. The initial mapping is Appendix C; the final mapping goes in the Result.
- **Integrity check (S-TRACE).** A static test parses the catalog and fails in any of these cases:
  - an F row names a test that does not exist;
  - a T row names a check that `run.sh` does not define;
  - a resource link does not resolve;
  - a P row lacks a stage or a verification method.

  This stops a requirement from being claimed as tested when it is not.

What the templates emit is deliberately minimal, matching the spec's "not a complete production foundation":

- **T, templated:**
  - the env contract;
  - Compose with three networks;
  - Dockerfiles;
  - PostgreSQL role and schema hardening;
  - a Go core skeleton: config, a secret-provider boundary, readiness, redacting structured logs, and request and trace correlation;
  - a Next.js skeleton: runtime config, health and ready routes, and semantic CSS tokens;
  - an optional FastAPI skeleton: config, model configuration, the OpenRouter adapter contract, internal service authentication, an execute endpoint that validates the envelope and emits typed events, and a privacy-safe observability hook;
  - cross-service JSON Schemas;
  - foundation docs: getting started (with executable blocks), the follow-up checklist, and the vendor inventory.
- **P, prescribed and not templated:**
  - tenancy schema;
  - RoleBinding/policy engine;
  - **MCP server mount.** This is deliberately excluded so that no MCP endpoint exists before an identity provider does;
  - identity provider integration;
  - migration tooling;
  - ObjectStore implementation;
  - audit tables;
  - metrics export;
  - dependency/vulnerability scanning;
  - shadcn/ui initialization;
  - hot-reload configuration;
  - production role provisioning.

  Generated `core/internal/` contains only `config/`, `health/` and `observability/`. No placeholder domain modules and no empty directories are emitted.
- **F, fixtures.** The tenancy, RLS, policy, hostname, MCP, storage and audit fixtures live under `validation/isolation/` as a reference Go module. They are never rendered into a destination.

### 2.2 Skill package rules (reusability and leakage)

1. **Self-contained.** Every relative link resolves *inside* `skills/project-archetype/` (S-LINKS). The Skill never links to, or cites section numbers from, this repository's `docs/`, `management/`, `AGENTS.md` or `CLAUDE.md`.
2. **No platform knowledge.** No brand, domain, product-domain term, roadmap or goal ID appears in the Skill.
   - The denylist is **not** stored in the Skill. This platform's denylist is the lifecycle artifact `management/v0/results/G000-validation-denylist.txt`, passed with `run.sh --denylist <file>`. It is scanned with no self-exclusion.
   - Without `--denylist`, the input-derived literal checks (T-PORTABILITY) still run.
3. **No machine-specific paths.** No absolute paths (`/home/`, `~/.claude`, `C:\`) and no required uncatalogued skills (S-PATHS).
4. **Invocation by path, no installation.**
   - Consumers read `skills/project-archetype/SKILL.md` and run `skills/project-archetype/generator/render.sh`.
   - The wrapper locates its own directory. Commands do not use the Claude-specific `${CLAUDE_SKILL_DIR}`, so other agents can run them.
   - Nothing is installed into a global or user skill directory. Copying the package into an agent's skill directory is described as optional.
5. **Size discipline.** `SKILL.md` stays under about 400 lines and holds the essential invariants and the workflow. Detail goes into references kept one level deep.

### 2.3 `SKILL.md` outline

- **Frontmatter:**
  - `name: project-archetype`. It must equal the directory name and be 1–64 characters of `[a-z0-9-]`.
  - `description`: 1024 characters or fewer, saying what the Skill does and when to activate it.
  - `metadata.version`.
  - Only fields valid in both the agentskills.io specification and Claude Code. Parsed with a real YAML parser in a container (S-FRONTMATTER).
- **When to use and not use:**
  - Use it to bootstrap a multi-tenant SaaS repository, or to re-apply the foundation layout to an existing repository.
  - Do not use it for product features or cloud provisioning.
- **Prerequisites:**
  - Git, and Docker Engine/Desktop with Compose ≥ 2.24 (WSL integration on Windows), plus base OS shell utilities.
  - No host Go, Python, Node, PostgreSQL, Redis or object-store server.
  - Authoring agents and optional IDE plugins may carry their own runtimes. They are a named exception and never part of the project's normal path.
- **Essential invariants.** Numbered, each linked to its reference and catalog IDs. The last bullet is **mandatory even when another provider supplies permissions**.
  - Service boundaries and dependency direction (web→core→{data, intelligence}).
  - Non-redundant naming.
  - Configuration-only product, domain and database identity, with no literals in source or generated text.
  - Docker-first operation.
  - Standard+100 host ports bound to loopback, with internal service DNS.
  - Global identity, with Organization as the tenant boundary; tenant rows carry `organization_id`; TenantContext is a non-persisted envelope.
  - Group UUID and slug rules.
  - A BusinessEntity can be shared across Groups, with same-Organization integrity.
  - One RoleBinding policy system:
    - deny by default;
    - capability-specific inheritance;
    - no mirrored rows;
    - sensitive operations require direct authority.
  - Context is resolved on the server. IDs from clients and models are selectors only.
  - RLS with role separation, and transaction-local context that fails closed.
  - Exact-object authorization before any signed URL is issued.
  - Audit that is controlled and attributed.
  - The MCP control plane: a per-context catalog and per-execution reauthorization.
  - Optional internal intelligence:
    - driven by envelopes;
    - no data-store credentials;
    - typed events;
    - `AI_MODEL_<CAPABILITY>` configuration;
    - OpenRouter behind ModelGateway;
    - no `AI_GATEWAY_DEFAULT`.
  - Secrets never appear in templates, logs or manifests.
  - SOC 2 readiness is not compliance.
  - Destination-safety rules.
  - Audit, storage authorization, RLS and context-switch checks.
- **Inputs.** A summary table linking to `references/configuration.md`, plus a neutral sample input (`examples/neutral.input.json`).
- **Workflow:**
  1. **Preflight.** Check Docker and Compose versions, and that the destination exists and is not the Skill. Recommend a clean git working tree or a new branch.
  2. **Collect inputs.** Ask for every required value and never invent product identity. Write the input file.
  3. **Validate.** Run `render.sh validate`.
  4. **Inspect and plan.** Run `render.sh plan`, a dry run, and show the planned creates and conflicts.
  5. **Confirm** with the user.
  6. **Render.** Run `render.sh render`.
  7. **Reconcile conflicts.** Show a diff for each conflict and let the user decide. Never overwrite, and never edit `AGENTS.md` or `CLAUDE.md`.
  8. **Create `.env`** from `.env.example` with secrets generated locally. Never commit or print them.
  9. **Start and check.** Run the executable blocks in `docs/foundation/GETTING-STARTED.md`.
  10. **Hand off** the foundation stages in `docs/foundation/FOLLOW-UP.md`.
- **Output contract:**
  - the generated file set;
  - the manifest location and schema;
  - per-file actions;
  - exit codes:
    - `0`: success;
    - `2`: invalid input or unsafe destination, nothing written;
    - `3`: rendered with conflicts;
    - `1`: internal error.
- **Verification.** What the consumer runs (the checks profile and the optional validation harness), and what a pass does and does not prove.
- **Index.** References index and composition notes.

### 2.4 Renderer

- **Entry point:** `generator/render.sh {validate|plan|render} --input <file> --dest <dir>` (bash). It runs:

  ```
  docker run --rm --user "$(id -u):$(id -g)" \
    -e HOME=/tmp -e GOCACHE=/cache/build -e GOMODCACHE=/cache/mod -e GOTOOLCHAIN=local \
    -v <named cache volume>:/cache \
    -v <skill>:/skill:ro -v <dest>:/dest \
    golang:<pinned> go run /skill/generator …
  ```

  The **harness calls the same wrapper**. The renderer uses only the standard library, so after the image is present it works offline.
- **Templating rule:**
  - Only `*.tmpl` files are processed. **Source code is copied verbatim** and contains no product, domain, database, CNAME or slug literals; those are resolved from the environment at runtime.
  - Templated files:
    - `.env.example`: the only file that receives identity values;
    - `infra/compose/local.yml`: conditional on intelligence, and identity-free;
    - `docs/foundation/*`: conditional on intelligence and provider choice, and identity-free. Documentation refers to `$PRODUCT_NAME` and similar variables, never their values.
  - Module and package names are neutral: Go module `core`, npm package `web` (private), Python project `intelligence`.
- **Validation before write.** All input errors and destination-safety errors are collected and printed together. The renderer then exits 2 having written nothing; this is verified by hashing the destination before and after (G-INVALID).
- **Collision policy:**

  | Action | Condition | Result |
  |---|---|---|
  | `create` | Target is absent | Written |
  | `unchanged` | Target exists and is byte-identical | No write |
  | `conflict` | Target exists and differs | **Not written.** The proposal goes to `.archetype/outputs/<output_id>/proposed/<path>`; exit 3 |
  | `ensure-lines` | `.gitignore` only | Missing required lines (`.env`, `.env.local`, `.archetype/outputs/*/proposed/`) are appended under a marked block. Existing bytes are kept as a prefix |
- **Refusals (exit 2).** The renderer refuses when:
  - the destination is missing, is not a directory, is the filesystem root, or is the Skill or lies inside it;
  - any path escapes the destination;
  - an existing target or parent is a symlink;
  - an emitted path is reserved.

  The reserved paths are `AGENTS.md`, `CLAUDE.md`, `README.md`, `.env`, `docs/ARCHITECTURE.md`, `management/`, `skills/` and `.git/`. The renderer never deletes anything and never writes secrets.
- **Output identity and invocation actions** (corrected in r3). r2 hashed per-file *actions* into the run ID, so the ID changed when `create` became `unchanged` on a repeat run. r3 separates two things:
  - **Output identity: `output_id`.** It is `sha256(canonical non-secret inputs ‖ Skill version ‖ sorted (target path, intended-content sha256, write mode) list)[:16]`. It depends only on what the renderer *would* produce, never on the state of the destination. The same inputs and Skill version therefore give the same `output_id` whether the targets end up created, unchanged or in conflict.
    - For the `ensure-lines` `.gitignore`, the intended content is the required-lines block, not the merged file.
  - **Invocation actions.** A run's per-path actions (`create`, `unchanged`, `conflict`, `ensure-lines`) describe *that* run against *that* destination state.
    - They are printed to stdout as `<action>\t<path>`, so the harness can assert them without host `jq`.
    - With `--report <file>`, they are also written as JSON to a caller-chosen path. A path inside the destination is refused with exit 2.
    - They are **never written into the destination**, so recording them can never cause a write.
- **Destination artifacts**, content-addressed by `output_id` and subject to the normal collision policy:
  - `.archetype/outputs/<output_id>/manifest.json` contains the non-secret inputs, the Skill version, and the intended file list with content hashes and write modes. It also records the follow-up checklist path. It contains no action states, timestamps, secrets or absolute paths.
  - `.archetype/outputs/<output_id>/proposed/<path>` holds the intended content for each conflicting target. It is deterministic for a given `output_id`.
- **Repeat semantics:**

  | Repeat run | Outcome |
  |---|---|
  | Same inputs, same Skill version, destination unchanged since the last run | Same `output_id`. Every target, including the manifest and any proposals, is `unchanged`. **Zero writes.** Exits 0 if no conflicts remain, or 3 if unresolved conflicts remain; they are reported again but nothing is rewritten |
  | Changed inputs or Skill version | A new `output_id` and a new output directory. Earlier output directories are never modified or deleted |

### 2.5 Generated project shape (trial outputs only)

```text
<destination>/
├── .env.example                      # the ONLY file carrying identity values; secrets empty; AI model lines commented as samples
├── .gitignore                        # ensure-lines merge
├── core/                             # module "core": cmd/server, internal/{config,health,observability}, *_test.go, go.mod/go.sum
├── web/                              # Next.js App Router: app/{layout,page,globals.css,api/health,api/ready}, lib/{config,core-client}.ts (server-only), lockfile
├── intelligence/                     # ONLY when enabled: pyproject.toml, uv.lock, src/{main.py,models,providers,security,observability,streaming}, tests/
├── contracts/                        # README; intelligence/{execution-envelope,stream-event}.schema.json ONLY when enabled
├── infra/
│   ├── compose/local.yml             # postgres, redis, minio + minio-init, core, web, [intelligence]; profile "checks"
│   ├── docker/{core,web,minio[,intelligence]}.Dockerfile   # runtime + "check" targets; minio built from pinned upstream source (R1)
│   ├── postgres/init/10-roles.sh     # role separation + schema hardening; role names derived from $DATABASE_NAME at init
│   └── minio/init.sh                 # mc: private bucket + scoped non-root service user/policy
└── docs/foundation/{GETTING-STARTED.md,FOLLOW-UP.md,VENDORS.md}
```

**Packaging and naming exceptions** (explained as the spec requires):

- **Python.** `intelligence/src/` holds top-level modules (`models`, `providers`, `security`, `observability`, `streaming`), following the architecture's "Monorepo Naming and Non-Redundant Directory Convention". They run with `PYTHONPATH=src`. There is no `src/intelligence/` package and no installed distribution, so generic names cannot shadow site-packages.
- **Entry point.** The FastAPI entry point is `src/main.py`.
- **Dockerfiles** live in `infra/docker/`. Each build uses its service directory as the context.
- **Stream-event schema.** It lives under `contracts/intelligence/` so that it does not collide with later domain or public event contracts.

**Architecture inconsistencies, resolved in favour of the spec and noted for review** (N1, §10):

- "Frontend UI and Design-System Architecture" shows `web/src/app`; the naming convention and the spec use `web/app`.
- "Go as the Core MCP Backend" shows `/cmd/platform`; the naming convention and the spec use `core/cmd/server`.
- The control-plane `logout` name: see Q3.

**Intelligence-disabled policy.** These are **omitted**, not kept as guidance:

- `intelligence/`, its Dockerfile, its Compose service and the `inference` network;
- the intelligence contracts;
- `INTELLIGENCE_*`, `AI_MODEL_*` and `OPENROUTER_API_KEY`.

Guidance for enabling it later stays in the Skill. Re-rendering with intelligence enabled creates the intelligence files and reports `.env.example` and `local.yml` as conflicts with proposals; the reconciliation is manual and documented (G-TOGGLE-RERENDER).

**Runtime images.** Tags are exact and never `latest`. They are re-verified at implementation, and the Result records digests (§11):

- `golang:1.27.1-alpine` as the builder, with a distroless static runtime for core;
- `node:24-slim` for web (Next.js `output: "standalone"`);
- `python:3.13-slim` plus a pinned `uv` for intelligence;
- `pgvector/pgvector:0.8.7-pg17`;
- `redis:8.10.2-alpine`;
- **MinIO, retained as specified (R1).** No maintained image is published, so `infra/docker/minio.Dockerfile` follows upstream's documented route ("clone the source and build the latest container") and builds the final upstream releases from source:
  - the server, `github.com/minio/minio@RELEASE.2025-10-15T17-29-55Z`;
  - the client, `github.com/minio/mc@RELEASE.2025-08-13T08-35-41Z`.

  The build stage uses `golang:1.27.1-alpine` with `CGO_ENABLED=0`, and modules are verified against the Go checksum database. There are two targets:
  - a static non-root runtime that includes `mc` for the healthcheck;
  - a `minio-init` target that adds a minimal shell for `infra/minio/init.sh`.

  The image is used only for local development and validation; production object storage remains R2 behind `ObjectStore`, per the architecture. The residual risk is recorded in R1 and in the generated `VENDORS.md`.

**Compose (`infra/compose/local.yml`):**

- **Single source of configuration.** Every variable, including ports, comes from `.env` via `${VAR:?message}`. There are no literal defaults, so missing values fail closed (S-FAILCLOSED-ENV).
- **Ports.** Host ports bind `${BIND_ADDRESS}` (`127.0.0.1` in `.env.example`). Inside containers, services use their conventional ports; core uses 8180.
- **Networks:**
  - `edge`: web and core;
  - `data`: core, postgres, redis, minio and minio-init;
  - `inference`: core and intelligence.

  web is on neither `data` nor `inference`. intelligence is not on `data` and gets **no** database, Redis or object-store credentials. This is the default; any later AI stage that needs data access must justify it.
- **Per-service secrets.** Each service receives only what it needs (S-ENV-SCOPE):
  - core gets the runtime database password, the Redis password, the scoped object-store key and the intelligence token;
  - core never gets the superuser password, the migrator password or the MinIO root credential;
  - web gets no secrets;
  - intelligence gets only its service token and the optional `OPENROUTER_API_KEY`.
- **Health and startup.** Every long-running service has a healthcheck. Core depends on healthy postgres, redis and minio, and on a completed `minio-init`.
- **Volumes.** Postgres and MinIO have named volumes. Redis is ephemeral (`--save "" --appendonly no`) and requires `REDIS_PASSWORD`.
- **Profile `checks`.** It builds `check` targets and makes no bind-mounted writes:
  - `core-check`: `gofmt -l`, `go vet`, `go test`;
  - `web-check`: lint and typecheck;
  - `intelligence-check`: ruff and pytest;
  - `secret-scan`: gitleaks, read-only.

**Core skeleton (Go):**

- **Config.** Env config is validated at startup and fails fast. It covers `PRODUCT_NAME`, `PLATFORM_BASE_DOMAIN`, `CUSTOM_DOMAIN_CNAME_TARGET`, `DATABASE_NAME` and `APP_ENV`, plus separate DB host, name, user and password fields so URLs never need escaping. It also covers Redis host, port and password, and the object-store endpoint, region, bucket, path-style flag and credentials.
- **Secrets** are read through a `SecretProvider` interface with an env implementation.
- **Endpoints.** `/healthz` is liveness. `/readyz` checks PostgreSQL (pgx, as the **runtime** role), Redis (RESP `AUTH`/`PING`) and the object store (S3 `HeadBucket` through `aws-sdk-go-v2`, which is provider-neutral). Any of those failing returns 503 naming the dependency. Intelligence status is **reported but not blocking**, and the payload carries `PRODUCT_NAME`. It never includes secrets or DSNs.
- **Logging.** `log/slog` JSON output goes through a redacting handler, with request-ID and W3C `traceparent` correlation.
- **Healthcheck.** `server healthcheck` is the container healthcheck.
- **Tests** cover config validation, redaction, readiness with fakes, and traceparent parsing.

**Web skeleton (Next.js App Router):**

- **Request-time config.** The server reads `PRODUCT_NAME` and domain values at request time. Pages and `generateMetadata` render dynamically, so nothing is captured at build time.
- **No public variables.** There are no `NEXT_PUBLIC_*` identity values or backend URLs. Server-only `lib/core-client.ts` uses `CORE_INTERNAL_URL` on `edge`.
- **Routes.** `app/api/health` and `app/api/ready`; the latter returns 503 when core is unready or unreachable.
- **Tokens.** `app/globals.css` defines the semantic tokens `--background`, `--foreground`, `--primary`, `--primary-foreground`, `--accent`, `--muted-foreground` and `--border`, wired to Tailwind 4.
- **shadcn/ui with Base UI** is prescribed for `STAGE-WEB-SHELL` (`shadcn init --base base --template next … --yes`).
- **Hand-written skeleton.** It is not generated with `create-next-app`, which now writes `AGENTS.md`/`CLAUDE.md`. Any guidance that does scaffold with it requires `--no-agents-md`.

**Intelligence skeleton (FastAPI, when enabled):**

- **Health.** `/healthz` and `/readyz` are unauthenticated and content-free.
- **Fail-closed internal routes.** If `INTELLIGENCE_SERVICE_TOKEN` is unset or empty, `/internal/*` refuses every request.
- **`POST /internal/v1/execute`:**
  - requires the bearer token;
  - validates the request body against `execution-envelope.schema.json` (422 otherwise);
  - streams NDJSON typed events conforming to `stream-event.schema.json`.

  If the envelope's capability is not configured, it streams a typed `error` (`capability_not_configured`) followed by `done`, and makes no model call. When configured, it goes through ModelGateway. Tests inject the deterministic stub transport through FastAPI dependency overrides. There is **no runtime switch that selects a stub**.
- **ModelResolver.** A capability is enabled exactly when its `AI_MODEL_<CAP>` is set. The resolver rejects:
  - unknown capability names;
  - fallbacks without a primary;
  - malformed identities;
  - a gateway named as the provider (Q8).
- **OpenRouter adapter (ModelGateway):**
  - splits the identity on the first `:`;
  - maps the provider through an alias table to the OpenRouter author slug;
  - uses an injectable transport;
  - normalizes usage as described in §4.3, with `actual_provider` set to `unknown` unless router metadata supplies it.

  The static service token is a **fixture-level default**. Envelope integrity (core-signed, short-lived, audience-bound) is prescribed for `STAGE-AI`.
- **Observability.** The hook's default export and the service logs exclude prompts, completions, retrieved content and secrets. Content capture needs an explicit flag that defaults to off.

### 2.6 Isolation fixtures (reference Go module under `validation/isolation/`)

The seed data is synthetic and uses `example.test` identities:

| Item | Detail |
|---|---|
| Organizations | `org-alpha` (slug `alpha`) and `org-bravo` (slug `bravo`) |
| Groups | GA1 and GA2 in alpha; GB1 in bravo |
| BusinessEntities | BE-A, shared by GA1 and GA2; BE-A2, linked to GA2 only; BE-B, in bravo |
| `ada@example.test` | Alpha `org_owner`; bravo `org_member` |
| `lin@example.test` | GA1 `group_admin` only |
| `sam@example.test` | GA2 `group_member` only |
| `kai@example.test` | Direct `be_owner` of BE-A only |

Rules that every test follows:

- **Paired controls.** Every negative test is paired with a **positive control in the same test and on the same connection**: an exact seeded row count proving that the target exists and is visible under the correct context.
- **Mechanism assertions.** Tests assert *how* access was blocked, not just that something failed:
  - a hidden row gives 0 rows and no error;
  - a WITH CHECK violation gives SQLSTATE 42501 with a row-level security message;
  - a missing privilege gives 42501 with a permission message.

  These are distinguished from one another, so a missing GRANT can never pass as RLS.

**Authority model: canonical membership semantics with one mutable authority store** (correction 3).

The architecture defines `OrganizationMembership(user_id, organization_id, role)`, `GroupMembership(user_id, group_id, role)`, `BusinessEntityMembership(user_id, business_entity_id, role)` and the shared `RoleBinding(subject, scope_type, scope_id, role, source)`. The fixture keeps **all four canonical shapes, including the membership `role`**, and derives membership roles from RoleBinding, so there is no duplicate mutable grant:

- **`role_bindings` is the only mutable store of roles.** Each row has the canonical columns plus typed scope columns (`organization_id`, nullable `group_id`/`business_entity_id`) with composite same-Organization FKs and a consistency CHECK.
  - `source` records how the binding arose. `membership` means a membership role at that scope. `direct_grant` means a direct BusinessEntity grant. Later stages may add values such as `invitation` or `system`.
  - A partial unique index allows at most one `membership` role per subject and scope.
- **The membership entities are derived, read-only projections of RoleBinding.** They are views with `security_invoker = true`, so RLS applies, and they keep their canonical columns:

  | Membership view | Derived from bindings where |
  |---|---|
  | `organization_memberships` | `scope_type = 'organization'` and `source = 'membership'` |
  | `group_memberships` | `scope_type = 'group'` and `source = 'membership'` |
  | `business_entity_memberships` | `scope_type = 'business_entity'`, and `source` is `membership` or `direct_grant` |

  The runtime role has SELECT only on these views. Any INSERT, UPDATE or DELETE through them fails with 42501.
- **One write per change.** Granting, changing or removing a membership role is a single write to a single `role_bindings` row. In the same transaction, that write changes:
  - the membership projection;
  - group-scope RLS visibility;
  - every capability the evaluator inherits from it.

  No other table stores a role, and no BusinessEntity row mirrors a Group grant.
- **Room for lifecycle attributes.** If a later stage needs membership attributes (status, invited_by, joined_at), they belong in a membership-attributes table keyed by the binding that carries **no role**, so the single-authority rule holds.
- **Evidence:** Appendix A case 24 (`Authority_SingleStore`, read-only views, single-row propagation) and the `policy` PostgreSQL-backed store test. This reconciles the two canonical concepts without redefining memberships. Q7 asks Codex to confirm the interpretation.

**Test connections and guards** (correction 5). Every database test obtains its connection from exactly one `internal/fixture` helper. Each helper verifies its identity on every acquire, and the test aborts on any mismatch:

| Helper | Login → effective role | Guard (checked on every acquire) | Used only for |
|---|---|---|---|
| `fixture.RuntimeDB` | `<db>_runtime` → same | `session_user = current_user = <db>_runtime`. `rolsuper`, `rolbypassrls`, `rolcreaterole` and `rolcreatedb` are all false. Not a member of owner, migrator, retention or policy-reader | All runtime assertions; the default |
| `fixture.OwnerTx` | `<db>_migrator` → `<db>_owner` through `SET LOCAL ROLE` inside each transaction | `session_user = <db>_migrator` and `current_user = <db>_owner`. Neither is superuser or BYPASSRLS. `relforcerowsecurity = true` on the tables under test | Owner and FORCE-RLS assertions; audit-trigger assertions on the owner path |
| `fixture.RetentionDB` | `<db>_retention` → same | `current_user = <db>_retention`; not superuser or BYPASSRLS; not a member of owner or runtime | Retention assertions |
| `fixture.SuperuserSetup` | image init superuser → same | Asserts `rolsuper = true`, so it can never be mistaken for a test connection | Seed data and `Control_*` tests only |

- **Naming.** Test names carry the connection: `TestRuntime_*`, `TestOwner_*`, `TestRetention_*` and `TestControl_*`. The evidence label (Real/PG, Real/S3, Real/SDK-inmem, Real/SDK-http or Unit) is recorded per test through `fixture.Label(t, …)`. The Result tabulates the tests by connection and by label.
- **Static guard (`TestGuard_ConnectionScope`).** An AST scan of the isolation module fails if a helper is used outside its scope:
  - `SuperuserSetup` may be referenced only from `internal/fixture/seed.go` and `TestControl_*` functions;
  - `OwnerTx` only from `TestOwner_*`;
  - `RetentionDB` only from `TestRetention_*`.
- **Owner tests** contain both an **authorized positive control** (own-scope rows visible and writable under the matching context) and a **foreign-scope denial** (another Organization's rows invisible, and a write fails 42501). This is Appendix A case 2.
- **Retention tests** contain both a positive deletion of an expired row and denials. This is Appendix A case 14.
- **Schema application.** The harness applies the schema as the migrator through `psql`. Schema application is never an assertion connection.

| Package | What it demonstrates | Label |
|---|---|---|
| `sql/` | **Tables.** The global `identities` table, plus these tenant tables:<br>• `organizations`: globally unique slug with a reserved-name check, used for subdomains;<br>• `groups`: immutable UUID, kebab-case and reserved-route CHECKs (a fixture list; later configuration-driven), `UNIQUE(organization_id, slug)`;<br>• `business_entities`;<br>• `group_business_entities`: composite same-org FKs;<br>• `role_bindings`: the only mutable role store; see "Authority model" above;<br>• `audit_events`;<br>• `stored_objects`.<br>**Derived views.** `organization_memberships`, `group_memberships` and `business_entity_memberships` are read-only views with `security_invoker` set, derived from `role_bindings`.<br>**Policies.** RLS is ENABLEd and FORCEd on every tenant table, with separate per-command policies. Context helpers use `NULLIF(current_setting(…, true), '')::uuid`.<br>**Definer helpers.** Group-scope visibility needs membership lookups that would make the policies recurse, so it uses **narrowly scoped SECURITY DEFINER STABLE helpers** (e.g. `app.my_group_ids()`). They are owned by a dedicated NOLOGIN, NOBYPASSRLS `<db>_policy_reader` role, which reads `role_bindings` through a role-specific SELECT-only policy. Each helper pins `search_path`, has EXECUTE revoked from PUBLIC and granted to the runtime role only, and is checked against an allowlist.<br>**Triggers.** `organization_id` and `id` are immutable. `audit_events` is append-only, with a retention exception for the retention role only. | Real/PG |
| `rls/` (pgx) | **Cases.** Appendix A, each marked with its connection.<br>**Connections.** Through the guarded helpers above. `TestMain` checks that all four DSNs are present and resolve to four distinct roles, and that the runtime DSN passes the `RuntimeDB` guard before any test runs.<br>**Helper.** `WithTenantTx` is the pattern consumers copy.<br>**Pooling.** Pooled reuse runs with `MaxConns=1` and asserts the same `pg_backend_pid`.<br>**Owner.** `TestOwner_ForcedRLS` (an assertion) runs through `OwnerTx`. It covers an authorized positive control and a foreign-scope denial, showing FORCE RLS applies to the owner path.<br>**Backup.** X-BKP dumps with `pg_dump`, restores into a second instance initialized from the same rendered role script, and re-runs a subset of cases. | Real/PG |
| `policy/` | **Evaluator.** Implements Appendix B: deny by default; Organization, Group and direct BusinessEntity bindings; capability-specific Group→BusinessEntity inheritance computed at evaluation time; an Organization policy toggle.<br>**Tests:**<br>• no cross-Organization bleed for ada;<br>• a Group admin can edit the BE-A profile but cannot transfer ownership or manage privileged settings;<br>• lin cannot edit BE-A2, and **the shared BE-A gives lin no visibility of GA2 or its members**;<br>• kai, as a direct owner, can transfer ownership;<br>• turning the policy toggle off removes inherited editing;<br>• a member attempting to grant a role is denied;<br>• forged or foreign selectors give the same result as a nonexistent ID;<br>• **service-level** same-Organization association validation;<br>• onboarding *selection* of an existing BusinessEntity requires owner or admin authority;<br>• onboarding *completion* requires the GroupBusinessEntity row to exist;<br>• `TenantContext` is built only from the authenticated identity and validated grants.<br>**PostgreSQL-backed store test** (`RuntimeDB`). Revoking a Group binding is one `role_bindings` row change. In the same transaction it changes the `group_memberships` projection and the effective BusinessEntity access, with no mirrored rows, and the audit event commits atomically with it.<br>**Doc-sync test.** The matrix in `references/tenancy-authorization.md` must equal the evaluator's matrix. | Unit + Real/PG |
| `hostname/` | **Selector-only resolution** from the configured `PLATFORM_BASE_DOMAIN` plus a verified-domain list:<br>• normalization (case, trailing dot, port);<br>• `<slug>.<base>` resolves;<br>• unknown, unverified and look-alike hosts (`alpha.<base>.evil.test`) resolve to nothing;<br>• a resolved Organization grants nothing without membership;<br>• a foreign Group's slug grants nothing. | Unit |
| `mcp/` | **Server shape.** Uses the official MCP Go SDK v1.8.x with **one `Server` per authenticated session context**, because SDK `tools/list` returns every tool registered on that server. The tool set is reconciled from effective capabilities on a switch, on a per-request **policy/version marker** change, and on expiry or logout. The client waits on `list_changed` through a channel-synchronized handler, with no sleeps.<br>**Tools.** The spec's control-plane tools plus generic `groups.list`, `groups.update`, `business_entities.get`, `business_entities.update`, `business_entities.transfer_ownership` and `members.manage`. A name-pattern assertion is included, because the SDK only logs invalid names.<br>**Tests:**<br>• catalogs differ for ada in alpha and in bravo;<br>• an Organization switch clears Group scope;<br>• a Group switch requires a grant;<br>• `account.list_organizations` returns only the caller's own Organizations and exposes no other identity's memberships;<br>• a hidden tool's direct call fails exactly like an unknown tool;<br>• `organization_id` or `group_id` arguments cannot change scope;<br>• an approval-required capability is rejected without approval;<br>• logout invalidates the session;<br>• denials, switches and role changes emit audit events.<br>**Reauthorization isolated (H6).** A test hook **suspends catalog reconciliation** so the tool stays registered. The test then revokes a grant, downgrades a role, toggles Organization policy and expires the session. Each case asserts a **handler policy denial distinct from unknown-tool**, plus a denial audit event.<br>**HTTP smoke.** An in-process streamable-HTTP test (`httptest`, stateless, `getServer` per authenticated request) gives two bearer identities different `tools/list` results and returns 401 to anonymous callers. | Real/SDK-inmem; Real/SDK-http (in-process); the **dev-only fixture authenticator** lives in test-only files that are never rendered |
| `storage/` | **Components.** An `ObjectStore` interface, an S3-compatible adapter (`aws-sdk-go-v2`, path-style addressing, offline presigning) and an authorizer.<br>**Unit tests:**<br>• issuing a signed URL requires a `stored_objects` row visible under RLS plus the scope's read capability;<br>• a foreign object ID behaves like a nonexistent one;<br>• a guessed key under the caller's own prefix with no row is denied;<br>• client-supplied keys, path traversal and encoded slashes are rejected;<br>• the TTL is capped;<br>• issuing a URL is audited.<br>**Against a real server:**<br>• an exact-key presigned GET works;<br>• a sibling key fails the signature check;<br>• an expired URL fails;<br>• anonymous GET and LIST are denied;<br>• the scoped credential cannot change the bucket policy or delete the bucket. | Unit + Real/S3 (source-built MinIO, R1) |
| audit cases (in `rls/`, `policy/`, `mcp/`) | **Attribution fields.** Actor type and ID, organization, optional group, resource type and ID, action, outcome, `occurred_at`, `trace_id` and `request_id`. Organization may be null only for security events that occur before any context exists.<br>**Event categories.** Security denial, context change, membership or role change, and administrative mutation (a Group slug update).<br>**Controls:**<br>• `RuntimeDB` can only INSERT; its UPDATE, DELETE, TRUNCATE and SELECT are denied;<br>• `OwnerTx`: reading own-scope events is the positive control; foreign events are invisible; UPDATE is blocked by the trigger;<br>• `RetentionDB` can delete only rows older than the retention window. Its UPDATE, INSERT and content-column SELECT are denied. This is executed in the tests, not just documented. | Real/PG |

**Privileged administrative paths** are documented in `references/database-rls.md` and are not callable by the runtime role:

- Organization provisioning: bootstrap or administration through migrator/owner tooling, audited;
- audit retention;
- the policy-reader helpers.

**Fixture design choices.** Each is labelled as a recommended default for its owning stage, not a decision:

- the typed `role_bindings` scope columns;
- a nullable audit Organization;
- the definer-helper pattern and the `<db>_policy_reader` role;
- no Organization creation at runtime;
- role names `<db>_owner`, `<db>_migrator`, `<db>_runtime`, `<db>_retention` and `<db>_policy_reader`;
- schema `app`;
- the reserved-slug list;
- membership entities as derived read-only projections of RoleBinding, with canonical shapes and roles preserved (Q7);
- the connection-helper scheme and its static guard;
- `business_entities.*` tool names;
- the static internal service token.

## 3. Files and modules to create

Everything below is new. Nothing outside these paths changes.

```text
skills/project-archetype/
├── SKILL.md
├── examples/neutral.input.json              # neutral sample input (reserved .example/.test names)
├── references/
│   ├── service-boundaries.md                # boundaries, naming, dependency direction, packaging exceptions
│   ├── configuration.md                     # input/env contracts, validation rules, collision policy, exit codes
│   ├── docker-runtime.md                    # host prerequisites + named IDE/authoring exceptions, Compose topology/networks, ports, health, start/check/stop, hot reload (P)
│   ├── tenancy-authorization.md             # identity/IdP boundary, sessions/revocation, MFA/SSO; Org/Group/BusinessEntity; memberships; RoleBinding; permission matrix; onboarding; labels; billing-profile extensibility; hostname routing
│   ├── database-rls.md                      # roles, RLS patterns, WithTenantTx, pooling, same-org FKs, definer/view/FK/unique pitfalls, privileged paths, production role provisioning (P)
│   ├── storage-audit.md                     # ObjectStore boundary, signed URLs, AuditEvent model, controls, retention
│   ├── mcp-control-plane.md                 # switching flow, per-context catalog, version markers, per-execution reauthorization, HTTP mounting pattern
│   ├── intelligence-ai.md                   # AI_MODEL_* convention, Resolver→Gateway→OpenRouter, envelope, callbacks, typed events, retrieval filtering, scoped memory, usage attribution, observability privacy
│   ├── security-operations.md               # SOC 2 readiness hooks: secrets, logs, metrics/traces, scanning, CI evidence, encryption, backup/restore, retention/deletion, incident, vendor inventory, access provisioning
│   ├── commodity-composition.md             # per-skill evaluation, reuse decision (deviation, R8), assumptions, fallbacks, optional authoring aids
│   └── requirements-catalog.md              # generic IDs → T/F/P → resource → check/stage
├── checklists/
│   ├── invocation.md                        # operator/agent run checklist
│   ├── foundation-follow-up.md              # P-items by generic foundation stage (source of FOLLOW-UP.md)
│   └── verification.md                      # consumer verification of a generated project
├── templates/
│   ├── base/…                               # §2.5; *.tmpl only for .env.example, local.yml, docs/foundation/*
│   └── intelligence/…                       # intelligence/, contracts/intelligence/, intelligence.Dockerfile
├── generator/
│   ├── render.sh                            # the single documented entry point
│   ├── go.mod                               # stdlib only
│   ├── main.go  input.go  render.go  write.go  manifest.go
│   └── *_test.go                            # validation, collisions, refusals, determinism, manifest; Skill structure tests (links, paths, catalog integrity, frontmatter rules on yq output)
└── validation/
    ├── run.sh                               # suites: static, generator, intelligence, isolation, trials, all; --denylist, --keep
    ├── lib.sh                               # docker wrappers, hashing, /dev/tcp HTTP, PATH hygiene + runtime shadows, busybox probes, secret generation, docs-block extraction
    ├── compose.yml                          # validation-only: tool runners, fixture postgres ×2, MinIO built from the rendered minio.Dockerfile; no published host ports
    ├── inputs/{alpha.json,bravo.json,invalid/*.json}
    └── isolation/                           # Go module: sql/, rls/, policy/, hostname/, mcp/, storage/, internal/fixture/ (+ go.mod/go.sum)

management/v0/plans/G000-project-archetype-plan.md          # this file
management/v0/results/G000-project-archetype-result.md      # after implementation
management/v0/results/G000-validation-denylist.txt          # platform brand/product-domain denylist (kept out of the Skill)
```

Go module paths are local-only and neutral (e.g. `example.invalid/archetype/generator`).

## 4. Contracts and interfaces

### 4.1 Skill input contract (`*.input.json`)

| Field | Req. | Default | Validation |
|---|---|---|---|
| `product_name` | **required** | — | 1–64 printable characters. Rejects control characters, `"`, `\`, `$` and backtick (env and Compose interpolation safety) |
| `project_slug` | optional | kebab-case of `product_name` | `^[a-z][a-z0-9-]{1,38}[a-z0-9]$`. Becomes `COMPOSE_PROJECT_NAME` and the bucket name, both through env |
| `platform_base_domain` | **required** | — | Lowercase RFC 1123 hostname with ≥2 labels and no scheme, port or wildcard |
| `custom_domain_cname_target` | optional | `custom.<platform_base_domain>` | Hostname rules, and must not equal the base domain |
| `database_name` | **required** | — | `^[a-z_][a-z0-9_]{0,47}$`, which leaves room for role suffixes within 63 characters. Not `postgres`, `template0` or `template1` |
| `intelligence_enabled` | **required**, explicit boolean | — | Must be a JSON boolean |
| `identity_provider` | optional | `oidc` | Slug `^[a-z][a-z0-9-]{1,31}$`. Rendered as `IDENTITY_PROVIDER` plus commented generic OIDC placeholders. **No vendor templates and no integration** |
| `production_object_store` | optional | `s3-compatible` | Enum `r2`, `s3` or `s3-compatible`. Selects production guidance, vendor-inventory rows and placeholders. Local development always uses MinIO, as specified |
| `ports.{web,core,intelligence,postgres,redis,minio_api,minio_console}` | optional | 3100/8180/8100/5532/6479/9100/9101 | Integers 1024–65535, unique among enabled services |
| `ai_models.{chat,fast,reasoning,extraction,embedding}`, `ai_models.<cap>_fallbacks` | optional, allowed only when intelligence is enabled | Absent: `.env.example` shows commented sample lines marked as samples, with no claim of availability | `provider:model`, split on the first `:`. Fallbacks use the same syntax with no duplicates. A gateway name as provider is rejected (Q8) |

Missing required fields, non-boolean flags, malformed JSON, unknown fields and secret-looking fields are all rejected, so mistakes fail closed.

### 4.2 Generated environment contract (`.env.example`)

- **Identity and domains:** `PRODUCT_NAME`, `PLATFORM_BASE_DOMAIN`, `CUSTOM_DOMAIN_CNAME_TARGET`, `DATABASE_NAME`, `COMPOSE_PROJECT_NAME`, `APP_ENV=development`, `BIND_ADDRESS=127.0.0.1`.
- **Host ports:** `WEB_PORT`, `CORE_PORT`, `INTELLIGENCE_PORT`*, `POSTGRES_PORT`, `REDIS_PORT`, `MINIO_API_PORT` and `MINIO_CONSOLE_PORT`, as specified.
- **Internal wiring:**
  - `CORE_INTERNAL_URL` (web);
  - `INTELLIGENCE_URL`* (core);
  - `REDIS_HOST`, `REDIS_PORT_INTERNAL`;
  - `OBJECT_STORE_ENDPOINT` (`http://minio:9000` locally), `OBJECT_STORE_REGION`, `OBJECT_STORE_BUCKET`, `OBJECT_STORE_FORCE_PATH_STYLE`, `OBJECT_STORE_ACCESS_KEY`. These configure core's provider-neutral `ObjectStore` adapter; the MinIO-specific names belong to the MinIO service.
- **Secrets.** Empty in the template and required via `:?`:
  - `POSTGRES_SUPERUSER_PASSWORD` (init only);
  - `DB_MIGRATOR_PASSWORD`, `DB_RUNTIME_PASSWORD`, `DB_RETENTION_PASSWORD`;
  - `REDIS_PASSWORD`;
  - `MINIO_ROOT_USER` and `MINIO_ROOT_PASSWORD` (root credential; used only by minio and minio-init);
  - `OBJECT_STORE_SECRET_KEY` (scoped);
  - `INTELLIGENCE_SERVICE_TOKEN`*.
- **AI**\*: commented `AI_MODEL_{CHAT,FAST,REASONING,EXTRACTION,EMBEDDING}` and `AI_MODEL_<CAP>_FALLBACKS` sample lines. `OPENROUTER_API_KEY` is empty and not required at startup.
- **Identity:** `IDENTITY_PROVIDER` plus commented placeholders. There is **no** `AI_GATEWAY_DEFAULT`.

\* present only when intelligence is enabled.

### 4.3 Cross-service contracts (rendered only when intelligence is enabled)

- **`contracts/intelligence/execution-envelope.schema.json`.** The Go-authorized envelope contains:
  - `request_id` and `trace_id`;
  - the capability;
  - an opaque `authorization_ref`;
  - the Organization, optional Group, optional BusinessEntity and user (attribution for prompt and context construction only);
  - **allowed tools and capabilities** with a read/propose/commit level;
  - a deadline;
  - pre-filtered context references;
  - an optional opaque, short-lived, audience-bound `callback_grant_ref`.

  It contains no credentials (`additionalProperties: false`), and intelligence cannot modify or widen it.
- **`contracts/intelligence/stream-event.schema.json`.** Defines the typed events `text_delta`, `tool_call`, `tool_result`, `citation`, `structured_output`, `usage`, `error` and `done`. Every event carries `request_id`. The `usage` event carries:
  - the capability;
  - configured provider and model;
  - the gateway;
  - actual provider and model;
  - input and output tokens or units;
  - latency;
  - estimated cost;
  - `trace_id`.

  It **cannot carry tenant or user dimensions**. The schema rejects them, so intelligence cannot assert attribution.
- **Prescribed for `STAGE-AI`: durable usage record (core).** Core builds the record by joining the `usage` event to *its own* issued envelope by `request_id`. The user, Organization, Group and BusinessEntity dimensions come from the envelope, not from intelligence. Vendor observability is never the ledger. **Verification:** a forged attribution field in an event is rejected, and the record's dimensions equal the envelope's.
- **Prescribed for `STAGE-AI`: authorized callbacks.** A callback presents its `callback_grant_ref`. Core reauthorizes it against *current* grants, and any tenant IDs that intelligence supplies are ignored. **Verification:** after a revoke, the callback is denied; a forged scope is ignored.

Schemas are validated against the meta-schema and against good and bad samples (I-SCHEMA). The runtime execute endpoint is exercised in I-EXECUTE and T-BOUNDARIES.

### 4.4 Public API / MCP implications

None for this platform. G000 creates no public API and no MCP server. Tool names in the fixture follow the "MCP context switching terminology" section of the canonical-semantics override and are illustrative; `STAGE-MCP` owns the real contract.

## 5. Database, schema and migrations

- **Platform database.** None. G000 creates no platform schema or migrations.
- **Template: `infra/postgres/init/10-roles.sh`.** Runs once, when the volume is first initialized, as the image's init superuser. It reads `DATABASE_NAME` and the passwords from the environment and derives every role name at init time.
  - **Roles:**
    - `<db>_owner`: NOLOGIN NOBYPASSRLS;
    - `<db>_migrator`: LOGIN, granted owner `WITH INHERIT FALSE, SET TRUE`;
    - `<db>_runtime`: LOGIN NOSUPERUSER NOBYPASSRLS NOCREATEROLE NOCREATEDB NOINHERIT, with no owner membership;
    - `<db>_retention`: LOGIN, used only for the audit-retention path.
  - **Schema `app`** is owned by the owner role.
  - **Privileges:**
    - `REVOKE ALL ON DATABASE … FROM PUBLIC`; CONNECT is granted to migrator, runtime and retention;
    - CREATE on `public` and TEMP are revoked from PUBLIC;
    - USAGE on `app` is granted to runtime and retention;
    - `ALTER DEFAULT PRIVILEGES` revokes function EXECUTE from PUBLIC and grants nothing on tables. New tables are therefore unreachable until explicitly granted (fail closed).
  - **Connections.** The core runtime connects as `<db>_runtime`. The superuser password is used only during initialization.
- **Fixture-only role.** `<db>_policy_reader` (NOLOGIN NOBYPASSRLS) is created by the fixture schema, not the template. It is a fixture choice.
- **Production role provisioning** is prescribed for `STAGE-RLS`/`STAGE-OPS`, because managed PostgreSQL does not run init scripts.
- **Migration tooling** is not chosen. It is a lasting dependency decision, prescribed for `STAGE-BOOTSTRAP`/`STAGE-RLS`.
- **Fixture schema and data:**
  - The schema is applied with `psql` inside the fixture container, as `<db>_migrator` (`SET ROLE <db>_owner`), into a disposable database.
  - Seed data is loaded by the init superuser. It is labelled as setup and never used as an assertion.
  - Everything is destroyed with `down -v`, and every run uses a unique project name.

## 6. Implementation sequence, budget and bounded execution

### 6.1 Commitments (correction 6)

- **Budget** (amended in r5 per Goal amendment A1). A **hard ceiling of 15 meaningful implement → test → diagnose → fix → retest cycles for the entire G000 goal**, with **mandatory checkpoints at cycles 8 and 12** (§6.2, "Checkpoints").
  - The ceiling is a **single goal-wide counter**, shared by the main agent and every subagent, and across every work package, milestone and suite.
  - The limit applies to the goal as a whole. It is **not** per agent, per subagent, per work package or per milestone.
  - It is never reset, and I will not extend it without Codex.
- **What counts** (revised in r4). **Any** meaningful implement → test → diagnose → fix → retest cycle, by **any** agent, at **any** stage. That includes:
  - cycles a subagent performs inside its work package **before integration**;
  - cycles during integration;
  - fixes that follow review findings;
  - close-out.

  Concretely:
  - **When a cycle is counted.** A test or validation run of implemented work fails materially, and the failure is diagnosed, fixed and retested. The retest closes the cycle. If the retest fails materially again, the next diagnose → fix → retest is another cycle.
  - **Concurrent cycles.** Cycles that run at the same time in different agents count separately.
  - **Doubtful cases are counted.**
- **What does not count** (routine, non-diagnostic commands):
  - simple file reads;
  - formatting;
  - lint-only checks with no fix loop;
  - one-off validation commands that do not trigger an implementation/debug cycle, for example a test or validation run that passes;
  - environment inspection, such as tool or image versions, port availability and image pulls.
- **Recording.** The Result's execution log records each counted cycle with:
  - its goal-wide number (1–5);
  - the agent (main agent, or a subagent's ID);
  - the work package;
  - what the cycle attempted;
  - the failing check or test IDs;
  - the diagnosis and fix;
  - the retest outcome.
- **Exhaustion.** No agent may start a counted cycle that would push the goal-wide total past 15. If required checks still fail materially once the total reaches 15, I stop. The Result then records the blockers and the approaches attempted, and control returns to Codex. Tests and validation are never reduced to fit the budget.
- **Scope.** All required evidence stays in G000, including every isolation fixture. **No split is proposed.** A split would need a Codex-approved Goal/Spec revision, which this plan does not request.

### 6.2 Bounded execution strategy (resolves the context-size concern, R2)

- **Fixed work packages.** Ten work packages, each with fixed file ownership, explicit inputs and outputs, and a size estimate (§6.3). No package needs another package's full source in context; packages interact only through the contracts frozen in WP1/WP2.
- **Contract-first freeze.** WP1 and WP2 freeze these before any delegation:
  - input schema;
  - env contract;
  - template data model;
  - `output_id` and manifest semantics;
  - fixture API: seed IDs, connection helpers and guards;
  - the public `policy` types.

  A change to a frozen contract after delegation is itself a change batch, and is recorded in the execution log.
- **Delegated briefs.** Each delegated package gets a written brief in the session scratchpad, not the repository. It contains:
  - the owned paths;
  - the frozen contracts;
  - the test IDs that must pass;
  - the done criteria;
  - **the §6.1 counting rules, verbatim;**
  - **the subagent's cycle allowance** under the shared-budget control below.

  I review each package's diff: the file list first, then targeted reading of the security-relevant files (guards, policies, stale-catalog tests, authentication). Delegated source is not loaded wholesale into the integrator's context.
- **Shared-budget control** (added in r4). The goal-wide limit is 15 cycles (r5), and subagents run in parallel, so they cannot draw from the budget ad hoc.
  - **Allocation.** Before spawning subagents, I allocate allowances out of the 15 and record the allocation in the execution log. The initial allocation (r5) is:
    - subagents 1, 2a and 2b: at most **two** counted cycles each;
    - the main agent: the remaining **nine**, for integration and close-out.
  - **Reaching the allowance.** A subagent that has used its allowance and still has material failures does **not** start another cycle. It stops and hands back.
  - **Reallocation.** I may reallocate unused allowance between agents, and I record each reallocation. The goal-wide total never exceeds 15.
  - **Read-only subagents.** Subagents 3, 4 and 5 do no implementation and consume no cycles. A meaningful fix cycle prompted by their findings is a main-agent cycle.
- **Mandatory handoff report** (added in r4). Every delegated subagent (1, 2a, 2b) returns this report on handoff, whatever its outcome:
  1. **Cycles consumed.** The number of meaningful implement → test → diagnose → fix → retest cycles, counted by the §6.1 rules (zero is a valid answer).
  2. **Each cycle.** What it attempted, the failing tests or checks, the diagnosis, the fix, and the retest result.
  3. **Final status.** `complete`, `incomplete` or `blocked`, with the passing test IDs and the files changed.
  4. **Unresolved issues.** Remaining failures, known gaps, and dependencies or contract changes it requested.

  I copy the report into the Result's execution log as soon as the handoff arrives, before any integration work. If the report is missing or incomplete, the package is not integrated until the report is supplied.
- **Compaction-safe state.** The Result file's append-only **execution log** records, after every package and every cycle:
  - status;
  - cycle number;
  - failing and passing check IDs;
  - decisions and open issues.

  Work can therefore resume after a context compaction without re-deriving state. This log is the only Result content written before close-out.
- **Integration order**, chosen so that later packages consume earlier contracts rather than rework them: WP1 → WP2 → (WP3, WP4, WP5, WP6, WP7 in parallel) → WP8 → WP9 → WP10.
- **Checkpoints.** Checkpoints are self-checks; none is a partial-authorization gate.
  - **C1, after WP8's `static` and `generator` suites.** The renderer and destination-safety contract is stable before the trial runs.
  - **C2, after the `intelligence` and `isolation` suites.**
  - **C3, after `trials`.**
  - **Early warning.** If any required suite is still failing materially at the end of cycle 5, the execution log records the forecast and the remaining risk, and work continues within the budget.
  - **Mandatory budget checkpoints, cycles 8 and 12 (r5).** When cumulative cycle 8, and then cycle 12, is reached, the execution log records:
    - cycles consumed per agent;
    - the status of every suite;
    - the remaining work;
    - a forecast against the 15-cycle ceiling.

    If the forecast shows the ceiling cannot cover the remaining required evidence, I stop at that checkpoint and report under §6.1 "Exhaustion" rather than spending further cycles.

### 6.3 Work packages

| WP | Content | Owner | Size (approx. lines) |
|---|---|---|---|
| WP1 | Contracts; `generator/` + `render.sh`; manifest and output identity; Skill structure tests | Me | 1,000 Go/sh |
| WP2 | `10-roles.sh`; isolation `go.mod` pre-seeded with every planned dependency (pgx, MCP SDK, aws-sdk-go-v2); `internal/fixture` (seed data, connection helpers and guards, labels); public `policy` types | Me | 600 |
| WP3 | Core skeleton and its tests | Me | 600 Go |
| WP4 | Web skeleton and lockfile | Me | 300 TS (+ lockfile) |
| WP5 | Infra: Compose, Dockerfiles (including `minio.Dockerfile`), MinIO init, `.env.example`, docs with executable blocks | Me | 500 |
| WP6 | Intelligence templates | Subagent 1 | 900 Py/JSON |
| WP7a | `sql/`, `rls/`, `storage/` fixtures | Subagent 2a | 1,600 SQL/Go |
| WP7b | `policy/`, `hostname/`, `mcp/` fixtures | Subagent 2b | 1,600 Go |
| WP8 | Harness: `run.sh`, `lib.sh`, validation Compose, inputs | Me | 900 sh |
| WP9 | `SKILL.md`, references, requirements catalog, checklists | Me | 2,500 md |
| WP10 | Reviews (subagents 3, 4, 5), the final `run.sh all`, the Result | Me | — |

### 6.4 Sequence

1. **WP1 and WP2.** Freeze the contracts, then spawn subagents 1, 2a and 2b (§8) in parallel.
2. **WP3, WP4 and WP5,** while WP6 and WP7 proceed in parallel.
3. **WP8.** Build the harness, then run `static` and `generator` (C1). Integrate WP6 and run `intelligence`. Integrate WP7a/b and run `isolation` (C2). Run `trials` (C3).
4. **WP9.** Complete the guidance, iterating until `S-TRACE` is clean.
5. **WP10.** Run subagents 3, 4 and 5 and fix findings, but only within scope. Then run `run.sh all --denylist management/v0/results/G000-validation-denylist.txt` from a clean state and write the Result. The Result must contain:
   - exact commands and outcomes;
   - image tags and digests;
   - the cycle accounting required by §8 ("Cycle reporting in the Result");
   - executed versus prescribed evidence;
   - the finalized Appendix C;
   - `git status --porcelain`.

## 7. Test and validation strategy

All checks run through `skills/project-archetype/validation/run.sh <suite>`.

- **Self-contained suites.** Each suite renders its own temporary destinations, so suites have no order dependence. `all` runs static → generator → intelligence → isolation → trials.
- **Host requirements.** bash, base coreutils, git, and Docker with Compose ≥ 2.24 (checked in preflight). No host `curl`, `jq` or `openssl`. Secrets come from `/dev/urandom` and are hex-encoded.
- **PATH hygiene.** At startup the harness removes `/mnt/*` entries from `PATH`, so WSL interop cannot expose Windows runtimes. It then puts the runtime-shadow directory first (T-DOCKER-ONLY).
- **Isolation of runs.**
  - Destinations are created with `mktemp -d` outside the repository, which is asserted.
  - `COMPOSE_PROJECT_NAME` values are unique, and the validation Compose publishes no host ports.
  - A trap runs `down -v` and `rm -rf`, falling back to a container-based removal, unless `--keep` is set.
  - Containers never write into bind mounts as root.
- **No secret echo.** Secrets are never echoed, and `set -x` is never used around them.
- **Network.** Image pulls and Go, npm and PyPI downloads need the network (R3).

### 7.1 `static`

| ID | Check |
|---|---|
| S-GO | `gofmt -l`, `go vet` and `go test` for `generator/` |
| S-FRONTMATTER | `SKILL.md` frontmatter is parsed with a pinned `yq` container and then checked against the stricter agentskills.io rules. Required section headings exist |
| S-LINKS | Every relative link resolves **inside** the package, and anchors exist |
| S-PATHS | No absolute or machine-specific paths, no `${CLAUDE_SKILL_DIR}` in commands, and no citations of the host repository's documents or section numbers |
| S-TRACE | Requirements-catalog integrity (§2.1) |
| S-SHELL | `shellcheck` on `render.sh`, `validation/*.sh` and rendered shell scripts |
| S-DOCKERFILE | `hadolint` on rendered Dockerfiles |
| S-COMPOSE | `docker compose config -q` on all four render variants (§7.5) |
| S-FAILCLOSED-ENV | Rendered Compose with any single required variable removed from `.env` fails `docker compose config` with that variable named |
| S-ENV-SCOPE | Run `docker compose config --format json` on the rendered variants and check each service's environment:<br>• core has no superuser, migrator, retention or MinIO root secret;<br>• web has no secret;<br>• intelligence has no database, Redis or object-store setting or credential |
| S-NAMING | Path linter on the rendered trees:<br>• no forbidden root names;<br>• no child directory repeating its parent's name;<br>• none of the listed redundant paths;<br>• no empty directories |
| S-DEPS | Service dependency direction:<br>• `contracts/` contains only `*.json` and `*.md`;<br>• web `package.json` has no PostgreSQL, Redis or S3 client;<br>• intelligence `pyproject.toml` has no database, Redis or S3 client;<br>• core does not import `internal/health` from `internal/config` |
| S-SECRETS | `gitleaks` over the Skill and rendered trees, plus a check that every secret variable in `.env.example` is empty |
| S-LEAK | Case-insensitive scan with `--denylist` over the Skill and rendered trees, with no self-exclusion. Requires zero hits |

### 7.2 `generator` (destination safety)

| ID | Check |
|---|---|
| G-INVALID | Each file in `inputs/invalid/*.json` makes the renderer exit 2 and leaves the destination's tree hash unchanged. Cases:<br>• malformed JSON;<br>• missing `product_name`, `platform_base_domain`, `database_name` or `intelligence_enabled`;<br>• non-boolean `intelligence_enabled`;<br>• an unknown field;<br>• a bad domain;<br>• a CNAME equal to the base domain;<br>• a reserved or overlong database name;<br>• a port collision;<br>• a port out of range;<br>• a malformed `provider:model`;<br>• a gateway named as the provider;<br>• a duplicate fallback;<br>• AI models while intelligence is disabled;<br>• injection characters in the product name;<br>• the destination is the Skill;<br>• the destination is missing |
| G-SENTINEL | A *platform-shaped* destination is seeded with `README.md`, `AGENTS.md`, `CLAUDE.md`, `docs/ARCHITECTURE.md`, `management/x.md`, `skills/other/SKILL.md`, `.git/HEAD`, a sentinel `.env`, a custom `.gitignore` and a differing `infra/compose/local.yml`. After rendering:<br>• every sentinel hash is unchanged;<br>• the original `.gitignore` bytes form a prefix of the new file;<br>• `local.yml` is reported as a conflict and its proposal is written to `.archetype/outputs/<output_id>/proposed/infra/compose/local.yml`;<br>• the exit code is 3 |
| G-REPEAT-CLEAN | Render into an empty destination: everything is `create`, exit 0. Render again: the same `output_id`, and every target, including `.archetype/outputs/<output_id>/manifest.json`, is `unchanged`. The tree hash **and every file's mtime** are unchanged (zero writes). Exit 0 |
| G-REPEAT-CONFLICT | Render into the G-SENTINEL destination: conflicts, exit 3. Render again: the same `output_id`, and the same conflicts reported again. The proposals and manifest are `unchanged`, the tree hash and mtimes are unchanged (zero writes), and the exit code is 3 |
| G-OUTPUT-ID | **State independence.** `output_id` is identical for the same inputs rendered into an empty destination, the sentinel destination and an already-rendered destination. **Sensitivity.** `output_id` changes when any single input field or the Skill version changes (one mutation per field). Changing inputs creates a new output directory and leaves earlier ones byte-identical |
| G-REPORT | The `--report` JSON equals the stdout action lines and the expected per-path action for `create`, `unchanged`, `conflict` and `ensure-lines`. Invocation actions never appear inside the destination, and a `--report` path inside the destination is refused with exit 2 |
| G-REFUSE | A symlinked target, a symlinked parent, or an escaping path is refused with exit 2 and no writes |
| G-DETERMINISM | Identical input rendered into two empty destinations gives byte-identical trees, identical manifests and the same `output_id` |
| G-TOGGLE-RERENDER | Rendering with intelligence disabled, then enabled, gives a new `output_id`. It creates the intelligence files and reports conflicts for `.env.example` and `local.yml`, with proposals in the new output directory. Nothing is overwritten, and the earlier output directory is untouched. The manual reconciliation is documented |

### 7.3 `intelligence` (pytest in the rendered intelligence `check` target; no network)

| ID | Check |
|---|---|
| I-CONFIG | All five capabilities and fallbacks parse. Rejected:<br>• malformed values;<br>• fallbacks without a primary;<br>• unknown capability names;<br>• a gateway named as the provider;<br>• duplicate fallbacks.<br>An unconfigured capability raises `CapabilityNotConfigured`. Zero configured capabilities is valid |
| I-GATEWAY | The OpenRouter adapter, run through the stub transport:<br>• translates the model identity;<br>• applies the provider alias table;<br>• preserves variant suffixes;<br>• carries the fallback list;<br>• records a canned response whose `model` differs from the configured one as the actual model;<br>• sets `actual_provider` to `unknown` without router metadata, and to the provider when metadata is present.<br>No `AI_GATEWAY_DEFAULT` is read, and `OPENROUTER_API_KEY` is unset |
| I-AUTH | `/internal/*` returns 401 when the token is missing or wrong, and refuses all requests when the token is unset or empty. Health endpoints carry no content |
| I-EXECUTE | With a valid token and envelope, the stub gateway streams schema-valid `text_delta → usage → done`. An invalid envelope (missing `authorization_ref`, a credential field, or an unknown field) returns 422. An unconfigured capability streams `error(capability_not_configured)` then `done` |
| I-OBSERVABILITY | After a stub invocation, the captured logs and the default observability export contain no prompt text, completion text, retrieved content or key. Content appears only when the capture flag is set |
| I-SCHEMA | Both schemas are valid against the meta-schema. Good samples pass and bad ones fail, including a `usage` event that carries tenant or user dimensions |
| I-LINT | `ruff` |

### 7.4 `isolation` (validation compose: two fixture PostgreSQL instances initialized from a freshly rendered role script, MinIO built from the rendered `minio.Dockerfile`, and a golang runner)

- `go test ./... -count=1`. Every test name carries a label: `Real/PG`, `Real/S3`, `Real/SDK-inmem`, `Real/SDK-http` or `Unit`. The Result tabulates the results per label.
- Every database test uses exactly one guarded connection helper (§2.6 "Test connections and guards"). `TestGuard_ConnectionScope` enforces this statically.
  - Runtime assertions use `RuntimeDB`.
  - Owner and FORCE-RLS assertions use `OwnerTx`. Each includes an authorized positive control and a foreign-scope denial.
  - Retention assertions use `RetentionDB`.
  - `TestControl_*` tests use `SuperuserSetup` or session-level `set_config`. They are reported separately and are never counted as assertions.
- Cases: §2.6, Appendices A and B, and X-BKP.

### 7.5 `trials`

Inputs use neutral, reserved names:

| Trial | Product | Domain | Database | Intelligence | Ports |
|---|---|---|---|---|---|
| alpha | "Sample Atlas" | `atlas.example.test` | `atlas_app` | **enabled** | defaults |
| bravo | "Harbor Desk" | `harbor.example.org` | `harbor_desk` | **disabled** | overridden: 3200/8280/5632/6579/9200/9201 |

Two render-only variants complete a 2×2 matrix: alpha with intelligence disabled, and bravo with it enabled.

| ID | Check |
|---|---|
| T-DOCKER-ONLY | The whole harness runs with `/mnt/*` stripped from `PATH` and a shadow directory first in `PATH`. The shadow directory wraps `go`, `gofmt`, `python`, `python3`, `pip`, `pip3`, `pytest`, `uv`, `uvicorn`, `node`, `npm`, `npx`, `corepack`, `pnpm`, `yarn`, `bun`, `deno`, `psql`, `postgres`, `pg_ctl`, `redis-cli`, `redis-server`, `mc` and `aws`, plus their `.exe` names. Each wrapper logs and exits 127; the log must be empty. A static grep confirms that no script calls an absolute runtime path. The required images are printed for the Result |
| T-PORTABILITY | Across the 2×2 matrix, source trees are **byte-identical** except `.env.example`, `.archetype/**` and the intelligence- and provider-conditional files. **Literal confinement:** each trial's product name, domain, CNAME, database name and slug, in all their case and slug variants, appear **only** in `.env.example` and `.archetype/outputs/**`, and never in the other trial's output |
| T-DOCS-EXEC | The harness extracts the marked executable blocks (`<!-- archetype:start|check|stop -->`) from the rendered `GETTING-STARTED.md` and runs them **verbatim** under the shadowed `PATH`. alpha runs on default ports, after a preflight that **fails clearly** if a port is busy and never silently re-maps. bravo runs on its override ports |
| T-RENAME | (1) Change `PRODUCT_NAME`, `PLATFORM_BASE_DOMAIN` and `CUSTOM_DOMAIN_CNAME_TARGET` in alpha's `.env` and run `up -d --no-build`. Image IDs are unchanged, and the web `<title>`/body and core `/readyz` show the new values. (2) Change `DATABASE_NAME` and run `down -v` then `up --wait --no-build`. PostgreSQL then has the new database and `<newdb>_*` roles, and core is ready. Together these demonstrate a rename and re-domain **without re-rendering, rebuilding, or search/replace** |
| T-PORTS | **Both parts are required. Configuration evidence is never reported as reachability.**<br>**(1) Configuration.** `docker compose port` returns `127.0.0.1:<configured>` for every published service.<br>**(2) Reachability, executed from the host.** For every published port, a host-side probe connects to `127.0.0.1:<port>` and performs a protocol-level exchange:<br>• web `GET /api/health` → 200; core `GET /healthz` → 200; intelligence `GET /healthz` → 200 (alpha only);<br>• MinIO API `GET /minio/health/live` → 200; MinIO console `GET /` → any HTTP response;<br>• Redis: `PING` → `-NOAUTH`, which proves the port is Redis and that authentication is enforced;<br>• PostgreSQL: an 8-byte SSLRequest → a single `N` or `S` byte.<br>**(3) Loopback-only.** The same probes against the host's first non-loopback IPv4 address are refused for every published port.<br>**(4) Internal wiring.** Internal DNS and container ports keep working under bravo's overridden host ports.<br>**Probe implementation.** The primary probe is bash `/dev/tcp`. An equivalent fallback is used only if `/dev/tcp` is unavailable: host `curl`, using HTTP for the HTTP ports and `telnet://` with the same byte exchanges for Redis and PostgreSQL. If neither is available, T-PORTS is recorded as **unmet** with the reason. Both probes were verified on this host during planning (§1) |
| T-BOUNDARIES | Probes use a pinned `busybox` container sharing each service's network namespace (`--network container:<id>`), so no tools are needed in the distroless or app images:<br>• **web** cannot resolve or TCP-connect (`nc -z`) to `intelligence:8000`, `postgres:5432`, `redis:6379` or the object store, whether by name **or by IP** (IPs taken from `docker inspect`);<br>• **intelligence** cannot connect to the data services;<br>• **core** can connect to every data service and to `intelligence:8000`;<br>• from core's namespace, `POST /internal/v1/execute` returns 401 without the token, 422 with a bad envelope, and a schema-valid `error → done` stream with a valid envelope and no configured capability. There is no live model call |
| T-OUTAGE | Stopping redis makes core `/readyz` return 503 naming `redis`, while `/healthz` stays 200; restarting it returns 200. The same holds for postgres and minio. Stopping intelligence leaves core `/readyz` at 200, reporting intelligence as unavailable (the defined non-blocking behaviour). Stopping core makes web `/api/ready` return 503 |
| T-PERSIST | A marker table in postgres and a marker object in MinIO survive `down` + `up`. A Redis key does not (ephemeral) |
| T-TOGGLE | For bravo:<br>• `config --services` and `config --images` contain no intelligence or Python image;<br>• the rendered `.env.example` and Compose files contain no `AI_MODEL_*`, `OPENROUTER_API_KEY` or `INTELLIGENCE_*`;<br>• no Python container runs;<br>• startup and checks succeed.<br>For alpha, intelligence is healthy |
| T-CHECKS | Every service check passes in both trials through the executed `check` docs block |
| T-SECRET-CANARY | None of the generated secret values appears in compose logs, manifests, rendered files or harness output (exact-match scan) |
| T-REPO-CLEAN | `git status --porcelain` shows only plan-authorized paths, and the platform root has no `web/`, `core/`, `intelligence/`, `contracts/` or `infra/` |

### 7.6 Evidence from agents (required for AC1, but not automated gates)

- **Cold-read usability (subagent 4).** The Skill is copied to a temporary directory outside the repository.
  - **Preferred method:** run a headless agent session (`claude -p`) from that directory, so no project instructions load.
  - **Fallback:** a subagent with an explicit instruction to read only the copy. In that case the limitation that project context may still be loaded is stated.
  - **Request:** "a project called Lumen Notes at lumen.example.net, database lumen, no AI".
  - **Checks on the result:** its destination is checked with the deterministic tools: the manifest is present, S-NAMING, S-COMPOSE, S-ENV-SCOPE and literal confinement.
  - **Reporting:** the Result reports deviations, friction and any need for platform knowledge.
- **Security review (subagent 3)** over the full diff. Read-only.
- **Spec-trace verification (subagent 5).** An independent re-trace of every spec requirement to its evidence, before the Result is finalized.

## 8. Subagent / delegation strategy

**During planning**, four read-only subagents ran:
- commodity-skill evaluation;
- library and version research;
- RLS/security test design;
- an adversarial review of r1 against the spec.

I also checked the MCP SDK source myself, in a container. The subagents made no decisions. Every choice labelled as a fixture default is mine and is listed for review.

**Implementation** follows plan ACCEPT. Subagents start only after WP1 and WP2 have frozen the contracts and stubs (§6.2–§6.4). Their work packages are WP6, WP7a and WP7b, and every subagent consumes the shared 15-cycle goal budget (r5).

| Agent | Scope and file ownership | Integration |
|---|---|---|
| **Me (integrator)** | `SKILL.md`, `examples/`, `references/`, `checklists/`, `templates/base/`, `generator/`, `validation/{run.sh,lib.sh,compose.yml,inputs}`, isolation `go.mod`/`go.sum` (pre-seeded), `internal/fixture`, `policy` public types, management artifacts | All wiring, dependency changes, final runs and the Result |
| 1. Intelligence templates | `templates/intelligence/**` only | I render through the generator and run `intelligence` and T-BOUNDARIES |
| 2a. Database fixtures | `validation/isolation/{sql,rls,storage}/**`: Appendix A, X-BKP, audit cases, storage | Uses only pre-seeded dependencies; any new dependency is reported back to me rather than edited in. I verify the connection-helper guards and `TestGuard_ConnectionScope`, the mechanism assertions, and that each test uses the correct helper |
| 2b. Policy/MCP fixtures | `validation/isolation/{policy,hostname,mcp}/**` except the public types: Appendix B, MCP behaviours including the reconciliation-suspend hook, HTTP smoke, doc-sync | Same dependency rule as 2a. Its PostgreSQL-backed store test uses 2a's schema through `internal/fixture`. I review the stale-catalog tests specifically for unknown-tool false positives |
| 3. Security review | Read-only, whole diff | I triage findings; fixes stay in scope |
| 4. Cold-read usability | Writes only to temporary locations | AC1 evidence |
| 5. Spec-trace verifier | Read-only | Checks the final Appendix C before submission |

Subagents 1, 2a and 2b run in parallel on non-overlapping paths. Each one returns the mandatory handoff report (§6.2) and works within the cycle allowance it was given out of the shared 15-cycle goal budget (§6.1, r5). No subagent edits management artifacts, makes architecture or scope decisions, or reports acceptance. I review every delegated diff before integrating it.

**Cycle reporting in the Result** (added in r4). The final `management/v0/results/G000-project-archetype-result.md` must contain:

1. **The total goal cycle count**, out of 15, together with the checkpoint records made at cycles 8 and 12.
2. **The main-agent cycle count.**
3. **Per-subagent cycle counts** for subagents 1, 2a and 2b, taken from their handoff reports and reconciled with the execution log. Subagents 3, 4 and 5 are listed with 0, or with any count if they unexpectedly performed implementation.
4. **A reconciliation check:** total = main agent + the sum of the subagents. Any discrepancy is explained.
5. **The allowance allocation** and every reallocation (§6.2).
6. **A concise execution log** showing what counted as each meaningful cycle. Each entry gives the goal-wide cycle number, the agent, the work package, what was attempted, the failing check or test IDs, the diagnosis and fix, and the retest outcome.
7. **Each subagent's handoff report**, summarized, including its final status and unresolved issues.

## 9. Acceptance-criteria mapping

| # | Criterion | Plan evidence |
|---|---|---|
| AC1 | Artifacts exist and are consumable without product knowledge; links, metadata, configuration contract and workflow are valid | §3; S-FRONTMATTER, S-LINKS, S-PATHS, S-LEAK; the documented `render.sh` used by G-* and T-*; executed docs (T-DOCS-EXEC); cold-read trial with deterministic checks of its output (§7.6) |
| AC2 | Every requirement maps to a resource and an enforcement method; later obligations are explicit | `requirements-catalog.md` plus S-TRACE; `foundation-follow-up.md` → `FOLLOW-UP.md`; Appendix C finalized in the Result and verified by subagent 5 |
| AC3 | Portability, toggle, Docker-only and destination-safety trials pass | T-PORTABILITY, T-RENAME, T-TOGGLE, G-TOGGLE-RERENDER, T-DOCKER-ONLY, T-DOCS-EXEC, G-INVALID, G-SENTINEL, G-REPEAT-CLEAN, G-REPEAT-CONFLICT, G-OUTPUT-ID, G-REPORT, G-REFUSE, G-DETERMINISM |
| AC4 | Tenant/RLS/policy/MCP/storage/audit fixtures show positive and negative evidence, distinguish mock from runtime, and give no bypass-role false positives | §2.6 and Appendices A–B: guarded, separated connections (`RuntimeDB`, `OwnerTx`, `RetentionDB`, `SuperuserSetup` limited to controls) with the static scope guard; owner positive control plus foreign denial; paired controls; mechanism assertions (SQLSTATE); owner FORCE assertion; same-backend pool test; the reconciliation-suspend MCP test; `Control_*` reported separately; Real/Unit labels |
| AC5 | Boundaries, +100 ports, tenant/RoleBinding model, audit, AI configuration and OpenRouter boundary match the architecture, with no product domains | S-NAMING, S-DEPS, S-ENV-SCOPE, T-PORTS, T-BOUNDARIES, Appendix B, audit cases, I-CONFIG, I-GATEWAY, I-SCHEMA, S-LEAK |
| AC6 | The diff touches only the Skill, fixtures and lifecycle artifacts; the platform application stays unimplemented | §3; T-REPO-CLEAN; destinations outside the repository |
| AC7 | No unapproved infrastructure, provisioning, global install, real secret or live model dependency | Only images the architecture names, plus validation-only tool images (§11). No installation. S-SECRETS, T-SECRET-CANARY. The stub transport needs no key. MinIO is retained as specified, built from pinned upstream source. Its replacement is escalated, not adopted (R1) |
| AC8 | Codex ACCEPT | The Result is submitted as "ready for review". I do not self-accept |

## 10. Risks and questions for Codex review

**Risks**

- **R1: MinIO distribution status. MinIO is retained as specified; replacement is escalated (correction 1).**
  - **Rule applied.** The specified MinIO behaviour stays unless the product owner approves a replacement through an accepted ADR **and** the corresponding canonical updates are made. Those updates cover:
    - the `docs/ARCHITECTURE.md` stack, object-storage, port and Docker-first sections;
    - the G000 spec's port table and MinIO wording;
    - a reviewed plan revision.

    Until then, nothing in this plan uses or partially implements a replacement.
  - **Supporting sources** (accessed 2026-10-04):
    1. **GitHub API, `repos/minio/minio`.** `"archived": true`, licence `AGPL-3.0`. <https://api.github.com/repos/minio/minio>
    2. **Upstream README.** It states: "THIS REPOSITORY IS NO LONGER MAINTAINED." and "The MinIO community edition is now distributed as source code only. We will no longer provide pre-compiled binary releases for the community version." <https://github.com/minio/minio/blob/master/README.md>
    3. **Final release `RELEASE.2025-10-15T17-29-55Z`** (published 2025-10-16). This is a security release fixing GHSA-jjjj-jwhf-8rgr. Its notes give `go install -v github.com/minio/minio@RELEASE.2025-10-15T17-29-55Z` and say "For container environments, please clone the source and build the latest container." <https://github.com/minio/minio/releases/tag/RELEASE.2025-10-15T17-29-55Z>
    4. **GitHub API, `repos/minio/mc`.** `"archived": true`. Its final release is `RELEASE.2025-08-13T08-35-41Z`. <https://api.github.com/repos/minio/mc>, <https://github.com/minio/mc/releases/tag/RELEASE.2025-08-13T08-35-41Z>
    5. **Docker Hub API, `v2/repositories/minio/minio/`.** HTTP 404. <https://hub.docker.com/v2/repositories/minio/minio/>
    6. **Local registry check.** `docker manifest inspect` fails for `minio/minio:latest`, `minio/mc:latest`, `quay.io/minio/minio:latest` and `quay.io/minio/minio:RELEASE.2025-09-07T16-13-09Z`. The only copy on this machine is a September 2025 cached image, which other consumers cannot reproduce, so it is not used.
    7. **Local feasibility check.** In `golang:1.27.1-alpine` with `CGO_ENABLED=0`, `go install github.com/minio/minio@RELEASE.2025-10-15T17-29-55Z` and `go install github.com/minio/mc@RELEASE.2025-08-13T08-35-41Z` both succeeded through the Go module proxy and checksum database, in 107 s. Both binaries ran (`--version`).
  - **How the plan retains the specified behaviour:**
    - **Everything the spec names stays as written:** the MinIO API and console, `MINIO_API_PORT=9100`/`MINIO_CONSOLE_PORT=9101`, `MINIO_ROOT_*`, `infra/minio/init.sh` using `mc`, and S3 compatibility behind `ObjectStore`.
    - **Supply.** Upstream's documented source route, in `infra/docker/minio.Dockerfile`, as in §2.5:
      - the final releases above, pinned;
      - module checksums verified;
      - static, non-root binaries;
      - `mc` included for the healthcheck and init.
    - **Scope of use.** Local development and validation only. Production object storage remains R2, per the architecture.
    - **Not a vendor change.** It is the specified software, supplied the way upstream now distributes it.
    - **Why Codex should review it.** It adds a first-build cost (about 2 minutes, then cached) and the residual risk below.
  - **Residual risk, recorded in the generated `VENDORS.md` and `references/security-operations.md`:**
    - the software is unmaintained, and no further security fixes will follow the final release;
    - the licence is AGPL-3.0;
    - the console's functionality was reportedly reduced in 2025 community releases. The port is kept as specified and is only probed for an HTTP response (T-PORTS);
    - consumers must re-evaluate before any non-local use.
  - **Escalation to Codex and the product owner** (an ADR proposal, not an approval). The recommendation is to replace local MinIO with a maintained, Apache-2.0, version-tagged S3-compatible server (for example SeaweedFS) for local development only. Other options considered:
    - Garage (AGPL);
    - RustFS (young);
    - `chainguard/minio` (`latest`-only, not pinnable).

    The adapter and readiness check are already S3-neutral, which limits the effect of a future replacement. That would need the ADR and canonical updates listed above, plus a plan revision, before any implementation.
  - **If Codex does not accept the source-build supply method,** there is no compliant way to provide MinIO. G000 implementation then waits for a product-owner decision. **The plan assumes no partial implementation in that case.**
- **R2: Goal size (correction 6).** Roughly 10–12k lines across Markdown, Go, Python, SQL and shell, excluding lockfiles. This is handled by the committed 5-cycle budget for the whole goal and by the bounded execution strategy in §6.1–§6.3:
  - fixed work packages built on frozen contracts;
  - delegated briefs;
  - review per work package;
  - a compaction-safe execution log;
  - self-check checkpoints.

  All required evidence, including the isolation fixtures, stays in G000, and no split is proposed. If the budget is exhausted with material failures, I stop and report under AGENTS.md, without reducing validation.

  Since r4, subagent cycles before integration count toward the same goal total. Goal amendment A1 (r5) raised that total to a hard ceiling of 15, with checkpoints at cycles 8 and 12. The §6.2 allowances and the mandatory handoff reports keep parallel subagents from using up the budget unnoticed. Overriding the budget at goal level is Codex's decision, and this plan does not request it.
- **R3 — Network-dependent validation.** Validation is reproducible from pins and lockfiles but cannot run offline. Unavailable items are recorded as unmet, never skipped silently.
- **R4 — Host runtimes are present.** PATH stripping and runtime shadows catch use of the host or WSL-interop `node`/`python`. Tooling inside containers is permitted.
- **R5 — Redis licensing.** Redis moved to RSALv2/SSPLv1 in 7.4 and added AGPLv3 in 8; Valkey is the BSD-3 fork. The plan follows the architecture and uses `redis`. A substitution would need an ADR. This is recorded only in the generated `VENDORS.md`, as a product-owner candidate alongside R1.
- **R6 — Fixture choices could be mistaken for decisions.** The list is in §2.6. Each is labelled as a recommended default for its owning stage.
- **R7 — Next.js runtime configuration.** Static prerendering would freeze identity at build time. The skeleton renders dynamically, and T-RENAME proves it. Build time and flakiness are mitigated by `--wait` and one build per trial.
- **R8 — Commodity composition: a deviation for review.**
  - **What the architecture and roadmap say.** "Compose existing high-quality skills … rather than duplicating."
  - **Evaluation (2026-10-04, `commodity-composition.md`).** Every available commodity skill is unversioned, user-local and incompatible as a whole:
    - `docker-setup`: root compose file, default ports, superuser migrations, no Redis or RLS;
    - `nextjs-setup`: Bun, an `app/` root, `NEXT_PUBLIC` backend calls, host scaffolding;
    - `django-project` and `bootstrap-fullstack`: a Django core, `api/` and `app/` roots;
    - `auth-resend`: builds custom magic-link auth, which contradicts "do not build a custom authentication provider".
  - **Approach.** The archetype keeps the compatible *patterns* in its own references:
    - healthchecks;
    - init jobs;
    - idempotent role SQL;
    - Docker hot reload;
    - a token-security checklist.

    It names those skills only as optional sources, with fallbacks.
  - **What it composes.** Upstream tools: the official images, the `shadcn` CLI with Base UI for the web-shell stage, `create-next-app --no-agents-md` guidance, and the official MCP Go SDK.
  - **Optional authoring aids.** Plugins such as `context7`, the LSP plugins, `security-guidance` and `playwright` may need host runtimes. They are documented as the spec's named IDE/authoring exception and are never on the normal path.
  - Codex should confirm that this satisfies the composition requirement.
- **R9 — Docker Desktop on WSL (correction 4).** Host reachability depends on Docker Desktop's localhost forwarding. It was verified here: bash `/dev/tcp` and `curl` reached a loopback-published port, and the non-loopback address was refused. In an environment where neither the primary probe nor the equivalent fallback works, T-PORTS is recorded as **unmet**; `docker compose port` output is never substituted as reachability evidence. Root-owned files are avoided as described in §7.
- **R10 — The SDK only logs invalid tool names.** The fixture asserts the name pattern itself. I checked the v1.8.0 source: `[A-Za-z0-9_.-]`, up to 128 characters, so dotted names are valid.
- **R11 — Embeddings through OpenRouter are not verified.** G000 covers only the adapter contract. Whether a direct-provider adapter is needed is an extension point for `STAGE-AI`.
- **R12 — Validation-only tool images.** `shellcheck`, `hadolint`, `gitleaks`, `yq` and `busybox`, all pinned. They are used only by the validation harness; the generated project uses only `gitleaks`, in its checks profile. They are tooling, not infrastructure, and are recorded in `VENDORS.md` and the Result.

**Questions**

- **Q1 — Budget: resolved in r3 (correction 6), then amended in r5.** Goal amendment A1 sets the budget to a hard ceiling of 15 cycles for the entire goal, with checkpoints at cycles 8 and 12 (§6.1).
- **Q2 — Neutral defaults.**
  - **The conflict.** The roadmap calls `Beginify`/`foundation.test` "template defaults". The G000 spec asks for "neutral sample defaults" and forbids product literals in the reusable Skill.
  - **The plan's choice.** The three identity inputs are **required with no default**, so a project is never silently named after a sample. Neutral samples ship as `examples/neutral.input.json` and as commented sample lines; G001 supplies the platform's own values.
  - Please confirm that this satisfies "neutral sample defaults", or tell me to make the samples actual defaults.
- **Q3 — Logout tool name.** `logout` (the canonical-semantics override and the spec) or `account.logout` (the earlier "MCP Context Switching and Dynamic Tool Exposure" section)? The fixture uses `logout`.
- **Q4 — Web runtime.** Node LTS rather than Bun, since the architecture is silent. Please confirm, or flag it for the product owner.
- **Q5 — Intelligence host port.** The plan binds it to `127.0.0.1` only, and web has no network path to it. An alternative is to publish it only under a `debug` profile; that departs from the spec's port table, so it is Codex's call.
- **Q6 — Actual-provider attribution.** The plan records `actual_provider = unknown` when OpenRouter metadata is absent, rather than guessing. Does that satisfy "preserves configured/actual identity attribution"?
- **Q7 — Membership and RoleBinding reconciliation, for confirmation (correction 3).** The architecture defines the membership entities *with* `role` *and* the shared RoleBinding. The plan keeps every canonical shape and derives the membership roles from RoleBinding (§2.6, "Authority model"). Membership entities are read-only projections of the single mutable `role_bindings` store, so there are no duplicate mutable grants, and memberships are **not** redefined as belonging-only. Please confirm that this interpretation is consistent with the architecture, or route it to the product owner if you consider it a design change. It is a recommended default for `STAGE-TENANCY`/`STAGE-POLICY`.
- **Q8 — Gateway as provider.** The fixture rejects gateway names (e.g. `openrouter:`) as the provider in `AI_MODEL_*`. Please confirm this reading of "do not make the gateway a domain concept".
- **N1 — Informational.** These naming inconsistencies were resolved in favour of the spec and the later sections, as noted in §2.5: `web/app` vs `web/src/app`, and `core/cmd/server` vs `/cmd/platform`.

## 11. Library and image research (2026-10-04)

Sources: a planning research subagent (registry, proxy and release APIs, upstream docs), plus my own checks of the MinIO status and the SDK tool-name rules. Every version is re-checked and pinned exactly at implementation, and the Result records digests.

| Component | Planned pin | Notes affecting the plan |
|---|---|---|
| MCP Go SDK | v1.8.0, requires Go ≥1.25 | No per-caller `tools/list`; one Server per context. `AddTool`/`RemoveTools` debounce `list_changed`. `NewInMemoryTransports` is available. `NewStreamableHTTPHandler(getServer, {Stateless:true})` is the HTTP pattern. Names are `[A-Za-z0-9_.-]`, ≤128 characters; invalid names are only logged (checked in the source) |
| Go | `golang:1.27.1-alpine` | Pulled successfully. Go 1.25 is out of support |
| pgx | v5.11.0 | `set_config(…, true)` inside the transaction. The pool destroys connections released mid-transaction. Use `PrepareConn` instead of the deprecated `BeforeAcquire`. Avoid `DISCARD ALL` |
| S3 client | `aws-sdk-go-v2/service/s3` v1.114.x | Path-style addressing and offline presigning; provider-neutral |
| PostgreSQL | `pgvector/pgvector:0.8.7-pg17` | PG16+ role-membership options |
| Redis | `redis:8.10.2-alpine` | R5 |
| MinIO (local) | Source build in `infra/docker/minio.Dockerfile`: `github.com/minio/minio@RELEASE.2025-10-15T17-29-55Z`, `github.com/minio/mc@RELEASE.2025-08-13T08-35-41Z` | Upstream publishes no maintained images and is archived (R1, sources 1–6). The source build was verified in 107 s (R1, source 7). Release tags are recorded as image labels, because `go install` builds report `DEVELOPMENT.GOGET` as their version |
| Node / Next.js / React / Tailwind | `node:24-slim`; Next 16.3.x; React 19.3; Tailwind 4.3.x | `create-next-app` writes AGENTS.md/CLAUDE.md unless given `--no-agents-md` |
| shadcn CLI | 4.21.x, `--base base` (prescribed) | Base UI officially supported |
| Python / FastAPI / Uvicorn | `python:3.13-slim`; FastAPI 0.142.x; Uvicorn 0.54.x; pinned `uv` | `TestClient` for in-process tests |
| OpenRouter | Adapter contract only | IDs are `author/slug` with optional variant suffixes. The response `model` field is the actual model. The provider comes only from opt-in router metadata or the generation endpoint |
| Validation tools | `koalaman/shellcheck:v0.11.0`, `hadolint/hadolint:v2.15.1`, `zricethezav/gitleaks:v8.30.1`, `mikefarah/yq` 4.x, `busybox` 1.37 | Trivy, govulncheck, npm audit and pip-audit are **prescribed** for `STAGE-OPS` and not run here, because their vulnerability databases make results non-deterministic |
| Agent Skills frontmatter | `name` 1–64 characters of `[a-z0-9-]`, equal to the directory name; `description` 1–1024 characters | The stricter agentskills.io rules apply |

## 12. Review history

- **r1 (2026-10-04).** Initial plan. Not reviewed by Codex.
- **r2 (2026-10-04).** Revised in place after the implementer's deep self-review and an independent adversarial review subagent, before any Codex review. Main changes:
  - **Leakage and reuse:**
    - The Skill is now project-agnostic: generic requirement IDs and foundation stages, and no goal IDs.
    - The G000 acceptance checklist moved out of the Skill.
    - The denylist moved to `management/v0/results/` (`--denylist`, no self-exclusion).
    - Links must resolve inside the package.
    - Generated source and docs contain no literals; only `.env.example` carries identity values.
    - New checks: T-PORTABILITY (byte-identical sources plus literal confinement) and T-RENAME (rename and re-domain, including the database, without re-render, rebuild or search/replace).
  - **Boundaries:**
    - A `data` network.
    - Intelligence holds no data credentials.
    - New checks: S-ENV-SCOPE (per-service secrets), S-DEPS (dependency direction), and T-BOUNDARIES (DNS, TCP-by-IP and internal-auth probes through busybox namespace sharing).
    - The MCP mount is deliberately not templated.
  - **Skill structure:**
    - Added the `SKILL.md` outline and the single `render.sh` entry point.
    - Added a requirements catalog with S-TRACE integrity checks, S-NAMING, YAML-parsed frontmatter, link and path rules, exit codes, refusals and per-run proposal directories.
    - Executable GETTING-STARTED blocks run verbatim (T-DOCS-EXEC).
  - **Isolation:**
    - Added the inline case list (Appendix A) and the permission matrix (Appendix B).
    - Added the `TestMain` runtime guard, paired positive controls and SQLSTATE mechanism assertions.
    - The owner's FORCE RLS becomes an assertion; the superuser is the only bypass control.
    - Narrowly scoped definer helpers through a NOLOGIN, non-bypass policy-reader role, with an allowlist check.
    - New cases:
      - identity and `list_organizations` visibility;
      - a shared BusinessEntity is not a bridge;
      - slug and UUID rules with SQLSTATEs;
      - service-level association validation;
      - onboarding selection and completion;
      - role-grant escalation;
      - policy-change and approval cases.
    - **An MCP reconciliation-suspend hook, so stale-permission tests prove handler reauthorization rather than unknown-tool.**
    - An MCP HTTP smoke test and a hostname selector fixture.
    - Executed audit retention and atomicity tests.
    - A doc/evaluator matrix sync test and a backup/restore smoke test.
  - **Operations:**
    - Redis requires a password and is ephemeral.
    - New checks: T-PERSIST, S-FAILCLOSED-ENV, the outage cases (object store, intelligence non-blocking, core→web), T-SECRET-CANARY and T-REPO-CLEAN.
    - A vendor inventory and core skeleton tests.
    - Single-source env.
    - Database connection settings as split components.
  - **AI:**
    - The envelope gains allowed tools, a read/propose/commit level and a callback-grant reference.
    - `usage` events can no longer carry tenant dimensions; core joins them from its envelope.
    - Durable-record and callback rules are prescribed with verification methods.
    - Fail-closed internal authentication and an execute endpoint with typed events (I-EXECUTE).
    - A provider alias map, rejection of gateway-as-provider (Q8) and of unknown capabilities, and an observability-privacy test over captured logs.
    - Sample model identities appear only as comments.
    - The stream-event schema moved under `contracts/intelligence/`.
  - **Docker-first:**
    - `/mnt/*` is stripped from `PATH` and a wider set of runtimes is shadowed, including `.exe` names.
    - `GOTOOLCHAIN`/`GOMODCACHE` are set and cleanup is container-safe.
    - Authoring and IDE plugins are documented as the named exception.
  - **Sequencing and delegation:**
    - Gate A splits into A1 (R1-independent) and A2 (R1); Milestone B no longer waits for A2.
    - Contract-first stubs and a pre-seeded `go.mod`.
    - Subagent 2 splits into 2a and 2b, and a spec-trace verifier is added.
    - The cold-read trial runs outside the repository, and its output is machine-checked.
  - **Fixes:**
    - The section cross-reference error.
    - Stale MinIO assumptions.
    - The untracked-files claim.
    - Architecture references by title.
    - The billing capability renamed to privileged settings.
    - `business_entities.*` tool names.
  - **New items:** Q7, Q8, N1, and R9–R12. R8 is reframed as a deviation for review.
- **Codex plan review of r2 (2026-10-04): CORRECTION LOOP.** Six required corrections:
  1. R1 approval gate;
  2. manifest determinism;
  3. canonical membership semantics;
  4. mandatory connectivity validation;
  5. separated runtime and privileged test connections;
  6. explicit scope and budget commitments.

  Codex modified no files.
- **r3 (2026-10-04), correction pass 1.** This pass addresses only the six items. Content not named below is unchanged from r2.
  1. **R1 approval gate (§10 R1, §0, §1, §2.5, §4.1, §4.2, §6, §7, §9 AC7, §11, Appendix C).**
     - MinIO is **retained as specified**: the API and console, `MINIO_*` names and ports, and `infra/minio/init.sh` with `mc`.
     - It is supplied through upstream's documented source-build route, using the final pinned releases in `infra/docker/minio.Dockerfile`. The build was verified in a container during this pass (107 s).
     - Seven supporting sources were added: GitHub API archive status for `minio` and `mc`, the upstream README notices, the final release notes, the Docker Hub 404, the local registry checks and the local build check.
     - A replacement now requires an accepted ADR, canonical architecture and spec updates, and a plan revision. It remains an escalation (an ADR proposal, not an approval).
     - **Removed:** every `[R1]` step marker, the A1/A2 gate split and the "proceed if Codex accepts the rest first" assumption. If Codex rejects the source build, G000 waits for a product-owner decision, with no partial implementation.
  2. **Manifest determinism (§2.4, §7.2).**
     - The stable `output_id` hashes only inputs, the Skill version and intended content. It is separated from per-invocation actions, which go only to stdout and `--report` and never into the destination.
     - Destination artifacts are content-addressed under `.archetype/outputs/<output_id>/`.
     - Repeat semantics are defined for clean, conflicting and changed-input runs.
     - G-REPEAT is replaced by G-REPEAT-CLEAN and G-REPEAT-CONFLICT, both checking zero writes through tree hashes and mtimes. G-OUTPUT-ID (state independence and sensitivity) and G-REPORT are added. G-DETERMINISM, G-SENTINEL and G-TOGGLE-RERENDER are updated.
  3. **Canonical membership semantics (§2.6 "Authority model", the `sql`/`policy` rows, Appendix A 7/10/11/20/24, Appendix C, Q7).**
     - All canonical shapes are kept, including the membership `role`.
     - Membership entities are read-only `security_invoker` projections derived from the single mutable `role_bindings` store, so there are no duplicate mutable grants. They are no longer belonging-only.
     - Lifecycle attributes, if a later stage needs them, carry no role.
     - New case 24 (`AuthoritySingleStore`). Q7 now asks Codex to confirm the interpretation.
  4. **Mandatory connectivity validation (§1, §7.5 T-PORTS, R9).**
     - Reachability probes run from the host, with protocol-level exchanges, against every published port.
     - A loopback-only negative probe is added.
     - Bash `/dev/tcp` is the primary probe and host `curl` the equivalent fallback; both were verified on this host. Otherwise T-PORTS is **unmet**.
     - `docker compose port` output is never used as reachability evidence.
  5. **Separate runtime and privileged connections (§2.6 "Test connections and guards", the `rls`/audit rows, §7.4, §8 2a, §9 AC4, Appendix A).**
     - There are guarded helpers `RuntimeDB`, `OwnerTx`, `RetentionDB` and `SuperuserSetup`. `SuperuserSetup` is limited to seed data and controls, and the static `TestGuard_ConnectionScope` enforces each helper's scope.
     - Owner tests include an authorized positive control and a foreign-scope denial. Retention tests include a positive deletion and denials.
     - Every Appendix A case is tagged with its connection.
  6. **Scope and budget commitments (§6 rewritten as §6.1–§6.4, R2, Q1).**
     - **Five cycles for the entire goal**, with cycle accounting defined.
     - A concrete bounded execution strategy: ten work packages on frozen contracts, delegated briefs, per-package review, a compaction-safe execution log, and self-check checkpoints that are not authorization gates.
     - The split alternative is withdrawn. All required isolation evidence stays in G000.
  - **Validation of this revision:**
    - an internal cross-reference check: every check ID has a defining row, and every `§` reference resolves;
    - a check that no `[R1]` gating or partial-authorization language remains;
    - the planning probes recorded in §1 and in R1 sources 6–7.
- **Codex re-review of r3 (2026-10-04).** Corrections 1–5 are satisfied. Correction 6 is partially unmet, with four required changes:
  1. remove the blanket exclusion of "authoring-time unit-test runs … before integration" from §6.1;
  2. clarify the shared five-cycle budget: subagent cycles before integration count, routine or non-diagnostic commands do not, and the limit is goal-wide rather than per agent;
  3. require every delegated subagent handoff to report the cycles it consumed, what each cycle attempted, its final status and its unresolved issues (§6.2);
  4. require the Result to include the total, per-subagent and main-agent cycle counts, plus a concise execution log (§8).

  Codex modified no files.
- **r4 (2026-10-04), correction pass 2.** This pass addresses only the four remaining correction-6 changes. All other content is unchanged from r3.
  1. **§6.1.** The blanket exclusion of authoring-time unit-test runs before integration is **removed**, and "Cycle accounting" is replaced:
     - **"What counts":** any meaningful implement → test → diagnose → fix → retest cycle by any agent at any stage, including subagent cycles before integration. Concurrent cycles count separately, and doubtful cases are counted.
     - **"What does not count":** Codex's routine or non-diagnostic list (file reads, formatting, lint-only checks with no fix loop, one-off validation that triggers no debug cycle, environment inspection).
     - **"Budget":** states the single goal-wide counter, which is not per agent or per subagent.
     - **"Recording":** the per-cycle fields for the execution log.
     - **"Exhaustion":** no agent may start a cycle that would exceed the goal total.
  2. **§6.2.** Briefs now carry the counting rules and a cycle allowance. Two bullets are added:
     - **shared-budget control:** an initial allocation of one cycle each to subagents 1, 2a and 2b and two to the main agent, with recorded reallocation, and read-only subagents consuming none;
     - **mandatory handoff report:** cycles consumed, each cycle's attempt, failures, fix and retest, the final status, and unresolved issues. It is copied into the execution log before integration.
  3. **§8.** Adds "Cycle reporting in the Result": the total goal count, the main-agent count, per-subagent counts, a reconciliation check, the allocation history, a concise execution log of every counted cycle, and summarized handoff reports. §6.4 step 5 now points to it.
  4. **R2.** One consistency sentence notes that counting subagent cycles before integration tightens the budget. No override is requested.
  - **Validation of this revision:**
    - re-ran the cross-reference check (check IDs and `§` references);
    - confirmed that the removed exclusion text no longer appears outside this history.
- **Codex plan review of r4 (2026-10-04): APPROVED.**
- **r5 (2026-10-04): execution budget amendment.** The product owner authorized Goal amendment A1 at the start of implementation. The budget changes from 5 to a hard ceiling of 15 meaningful cycles for the entire goal, with mandatory checkpoints at cycles 8 and 12. The updated sections are:
  - §6.1 "Budget" and "Exhaustion";
  - §6.2: allowances of two cycles each for the three implementing subagents and nine for the main agent, plus the checkpoints;
  - §8 reporting;
  - R2;
  - Q1.

  Scope, requirements, acceptance criteria and architecture are unchanged. Counting rules are unchanged from r4.

---

## Appendix A — SQL/RLS cases

Each case is tagged with the **only** connection helper it may use (§2.6, "Test connections and guards"):

| Tag | Helper |
|---|---|
| **[Runtime]** | `RuntimeDB` |
| **[Owner]** | `OwnerTx` |
| **[Retention]** | `RetentionDB` |
| **[Control]** | `SuperuserSetup`, or a deliberately unsafe session setting; reported separately, never an assertion |

Each negative case is paired with a positive control on the same connection, and the mechanism is asserted.

1. **[Runtime] Precondition, enforced by the `RuntimeDB` guard on every acquire.**
   - `session_user = current_user = <db>_runtime`.
   - `rolsuper`, `rolbypassrls`, `rolcreaterole` and `rolcreatedb` are all false.
   - `pg_has_role` shows no membership in owner, migrator, retention or policy-reader.
   - `TestMain` also checks that the four DSNs resolve to four distinct roles.
2. **RLS coverage and FORCE on the owner path.**
   - **[Runtime]** Every `app` table with an `organization_id` column has `relrowsecurity` and `relforcerowsecurity` set, which also catches tables added later. There is no `tenants` table.
   - **[Owner] `TestOwner_ForcedRLS`** (through `OwnerTx`, guard `session_user = <db>_migrator`, `current_user = <db>_owner`):
     - **Authorized positive control.** In context A, the owner sees exactly A's seeded `groups` rows and can insert a group with `organization_id = A`.
     - **Foreign-scope denial.** In the same context, GB1 is invisible (0 rows, no error), and inserting a group with `organization_id = B` fails with 42501 (row-level security).
     - **No context.** With no context, the owner sees 0 tenant rows.
3. **[Runtime] Read isolation.** In context A, looking up GB1 returns 0 rows with no error. Control: GA1 is visible, and the A-only count is exact and greater than 0.
4. **[Runtime] Forged insert.** In context A, inserting into `groups` with `organization_id = B` fails with 42501 (row-level security WITH CHECK).
5. **[Runtime] Moving a row.** In context A, changing GA1's `organization_id` fails, through WITH CHECK or the immutability trigger. Updating `id` also fails.
6. **[Runtime] Foreign update and delete.** In context A, UPDATE and DELETE on GB1 affect 0 rows with no error. **[Control]** `TestControl_ForeignRowUnchanged` confirms through `SuperuserSetup` that GB1 is unchanged.
7. **[Runtime] No context.** With no context or an empty one, every tenant table and membership view returns 0 rows with no error, and INSERT fails with 42501.
8. **[Runtime] Malformed context.** A malformed context raises an invalid-input error, and no rows are ever returned.
9. **[Runtime] Cross-org association.**
   - In context A, `group_business_entities(A, GA1, BE-B)` fails with an FK violation (23503).
   - `(B, GB1, BE-B)` fails with 42501.
   - Control: `(A, GA1, BE-A)` exists.
10. **[Runtime] Cross-org membership binding.** In context A, a `role_bindings` row (`scope_type = 'group'`, `source = 'membership'`) that points at GB1 fails the composite FK (23503). Control: the matching GA1 binding exists and appears in `group_memberships`.
11. **[Runtime] Group scope**, through the narrowly scoped definer helper:

    | Identity | Can see | Cannot see |
    |---|---|---|
    | lin | GA1; BE-A; GA1's `group_memberships` | GA2; BE-A2; GA2's memberships |
    | sam | GA2; BE-A; BE-A2 | GA1 |
    | ada (`org_owner`) | Everything in alpha | — |

12. **[Runtime] Pooled connection reuse** (`MaxConns=1`, same `pg_backend_pid`). After a context-A transaction commits, rolls back or errors, a query without context returns 0 rows, and `current_setting` returns `''`. **[Control]** `TestControl_SessionSetConfigLeaks`: a session-level `set_config` *would* leak.
13. **[Runtime] `audit_events` privileges.**
    - UPDATE, DELETE, TRUNCATE and SELECT fail with 42501 (permission).
    - A valid INSERT succeeds; this is the positive control.
    - An INSERT for a foreign Organization fails with 42501 (row-level security).
14. **Audit protection beyond privileges.**
    - **[Owner]**
      - Positive control: in context A, the owner reads A's seeded events.
      - Foreign-scope denial: B's events are invisible.
      - Any UPDATE is blocked by the append-only trigger.
    - **[Retention]**
      - Positive control: deleting an event older than the retention window succeeds, and exactly that row is removed.
      - Denials:
        - deleting an event inside the window affects 0 rows, with the trigger as a second line of defence;
        - UPDATE and INSERT fail with 42501;
        - SELECT of content columns fails with 42501, because only `id` and `occurred_at` are granted.
15. **[Runtime] Role switching.** `SET ROLE <db>_owner` and `SET SESSION AUTHORIZATION` are denied.
16. **[Runtime] Disabling protections.** `ALTER TABLE … DISABLE/NO FORCE ROW LEVEL SECURITY`, `DROP POLICY` and `DISABLE TRIGGER` fail with "must be owner".
17. **[Runtime] `row_security`.** `SET row_security = off` followed by a tenant SELECT raises an error instead of bypassing RLS.
18. **[Runtime] Object creation.** `CREATE TABLE`, `CREATE FUNCTION` and `CREATE VIEW` in `public` or `app`, and `CREATE TEMP TABLE`, are all denied.
19. **[Runtime] SECURITY DEFINER functions.** The set in `app` exactly equals the allowlist. Each function:
    - is owned by `<db>_policy_reader` (NOLOGIN, NOBYPASSRLS);
    - pins `search_path` in `proconfig`;
    - is STABLE;
    - has EXECUTE revoked from PUBLIC and granted only to the runtime role.

    Every view uses `security_invoker`.
20. **[Runtime] Identities and own memberships.**
    - In context A, the runtime role sees itself and co-members of A, but not an identity that belongs only to B.
    - The caller's own `organization_memberships` across Organizations (a derived view) shows only the caller's rows.
    - There are no credential-like columns.
21. **[Runtime] Slugs.**
    - A slug that exists only in B can be inserted in A; there is no oracle.
    - A duplicate slug within A fails with 23505.
    - A reserved slug fails with 23514, and so does a non-kebab slug.
    - Updating a slug succeeds and leaves the UUID unchanged.
22. **[Control] `TestControl_BypassSeesForeign`.** The `SuperuserSetup` connection sees B's rows for the query in case 3. This shows that the fixture data would reveal a bypass. It is reported as a control.
23. **X-BKP.** The database is dumped with `pg_dump` and restored into a second instance initialized from the same rendered role script. On the restored instance:
    - cases 1, 3, 7, 13 and 19 pass under [Runtime];
    - case 2 passes under [Runtime] and [Owner];
    - the seed counts match.
24. **[Runtime] `TestRuntime_AuthoritySingleStore`** (correction 3).
    - **Catalog.** In `app`, `role_bindings` is the only relation that has a `role` column *and* is a table. `organization_memberships`, `group_memberships` and `business_entity_memberships` are views (`relkind = 'v'`) with `security_invoker` set.
    - **Read-only views.** INSERT, UPDATE and DELETE through each membership view fail with 42501.
    - **Propagation.** In one transaction, changing lin's GA1 binding from `group_admin` to `group_member`, and then deleting it, each takes exactly one row change. Each change is reflected immediately in:
      - `group_memberships`;
      - Group visibility under RLS (case 11);
      - the evaluator's BusinessEntity capabilities, shared with the `policy` store test.

      **[Control]** `TestControl_SingleRowChange` takes a before/after hash of every other `app` table through `SuperuserSetup` and confirms that no other table was modified. It is reported as a control, because the runtime role cannot read every table.

## Appendix B — Fixture permission matrix (generic; derived from RoleBinding; deny by default)

Roles:
- Organization: `org_owner`, `org_admin`, `org_member`.
- Group: `group_admin`, `group_member`.
- BusinessEntity (direct grants): `be_owner`, `be_admin`.

"Via Group" means the capability is inherited through a GroupBusinessEntity association. It is computed at evaluation time and never stored.

| Capability | org_owner | org_admin | org_member | group_admin | group_member | be_owner | be_admin |
|---|---|---|---|---|---|---|---|
| `organization.read` | ✓ | ✓ | ✓ | — | — | — | — |
| `organization.manage_members` / grant roles | ✓ | ✓ | — | — | — | — | — |
| `organization.manage_policy` | ✓ | — | — | — | — | — | — |
| `group.read` | ✓ all | ✓ all | — | ✓ own | ✓ own | — | — |
| `group.update` (name/slug) | ✓ | ✓ | — | ✓ own | — | — | — |
| `group.manage_members` | ✓ | ✓ | — | ✓ own | — | — | — |
| `business_entity.profile.read` | ✓ | ✓ | — | ✓ via Group | ✓ via Group | ✓ | ✓ |
| `business_entity.profile.update` | ✓ | ✓ | — | ✓ via Group **if** Org policy `group_admins_edit_business_profiles` (default on) | — | ✓ | ✓ |
| `business_entity.select_for_onboarding` | ✓ | ✓ | — | — | — | ✓ | ✓ |
| `business_entity.ownership.transfer` (approval required in the MCP fixture) | ✓ | — | — | — (never inherited) | — | ✓ | — |
| `business_entity.privileged_settings.manage` (e.g. future billing-profile, tax and legal settings; capability only, no schema) | ✓ | — | — | — (never inherited) | — | ✓ | ✓ |
| `audit.read` | ✓ | — | — | — | — | — | — |
| `storage.object.read` | the read capability of the object's owning scope, for that exact object | | | | | | |

**Sensitive capabilities are never inherited through a Group:** ownership, privileged settings, policy, audit and role grants.

## Appendix C — Spec requirement coverage index (initial; finalized in the Result)

Class: T = template, F = fixture, P = prescribed (foundation stage → this platform's goal).

| Spec area | Requirement (abridged) | Class | Evidence / owner |
|---|---|---|---|
| Skill contract | Valid `SKILL.md` with metadata, activation, inputs, workflow, references, output and verification; invariants in the entry point | T | S-FRONTMATTER, S-LINKS, T-DOCS-EXEC, cold-read trial |
| | Inputs (CNAME, providers, port overrides, intelligence); required vs optional; validation before output | T | §4.1, G-INVALID |
| | Identity through `PRODUCT_NAME`/`PLATFORM_BASE_DOMAIN`/`CUSTOM_DOMAIN_CNAME_TARGET`/`DATABASE_NAME`; neutral samples; no literals in templates, prompts, metadata or generated text | T | T-PORTABILITY (confinement), T-RENAME, S-LEAK; Q2 |
| | Manifest plus follow-up requirements | T | G-OUTPUT-ID, G-REPEAT-CLEAN, G-DETERMINISM, G-REPORT, `FOLLOW-UP.md` |
| | Inspect the destination; no overwrite, deletion, instruction replacement, secrets or provisioning; collision behaviour; safe repeat | T | G-SENTINEL, G-REPEAT-CLEAN, G-REPEAT-CONFLICT, G-REFUSE, G-TOGGLE-RERENDER, S-SECRETS |
| | Composition: reuse, reasons, assumptions, fallbacks; no machine paths | T/P | `commodity-composition.md`, S-PATHS; R8 deviation |
| Boundaries | Responsibilities of web, core, intelligence, contracts and infra; dependency direction | T | T-BOUNDARIES, S-DEPS, S-ENV-SCOPE |
| | Naming; no redundant paths; exceptions explained | T | S-NAMING, §2.5 |
| | No placeholder modules; intelligence-disabled policy | T | S-NAMING (no empty directories), T-TOGGLE |
| Docker | Everything in containers; Git and Docker on the host; exceptions named | T | T-DOCKER-ONLY, T-DOCS-EXEC, `docker-runtime.md` |
| | Compose services, health and readiness, volumes, start/check/stop; Redis ephemeral; S3 abstraction | T | T-DOCS-EXEC, T-CHECKS, T-PERSIST, T-OUTAGE |
| | +100 ports, configurable, validated; internal DNS unaffected; no unintended exposure | T | G-INVALID, T-PORTS, S-FAILCLOSED-ENV |
| Identity and tenancy | Global identity across Organizations with different roles | F | `policy` and `mcp` (ada) |
| | IdP interface, session and revocation, MFA/SSO guidance; no custom auth | P (F in part) | `tenancy-authorization.md`; MCP logout/expiry (F); STAGE-IDENTITY → G004 |
| | Organization as boundary; `organization_id`; TenantContext not persisted | F | Appendix A 2; `policy` TenantContext test |
| | Group UUID and slug; `UNIQUE(organization_id, slug)`; reserved routes; slugs and hostnames grant nothing | F/P | Appendix A 5, 21; `hostname`; STAGE-TENANCY → G005, STAGE-DOMAINS → G016 |
| | BusinessEntity in many Groups; same-Organization check in the DB and in the service | F | Appendix A 9; `policy` association test |
| | Memberships and direct grants (canonical shapes, with roles derived from RoleBinding) | F | `sql` derived views; Appendix A 10, 11, 20, 24; Appendix B; Q7 |
| | Group onboarding; unauthorized IDs cannot complete it | F/P | `policy` selection and completion tests; STAGE-ONBOARDING → G018 |
| | Billing-profile extensibility; label owner/target and privacy default | P | `tenancy-authorization.md`; STAGE-TENANCY → G006/G038; STAGE-LABELS → G020 |
| Policy and RLS | Shared RoleBinding; derivation; no mirrored rows | F | `policy` plus the PostgreSQL store test |
| | Deny by default; Group-admin editing ≠ ownership; matrix with generic examples | F | Appendix B; doc-sync test |
| | Server-resolved context; selectors are not authority; hostname still needs authorization | F/P | `policy` forged-ID tests; `hostname`; `mcp` argument test; STAGE-DOMAINS → G016 |
| | RLS reads and writes; same-tenant FKs; global identity table | F | Appendix A 3–11, 20 |
| | Role separation; no superuser or BYPASSRLS; owner bypass prevented; definer narrowly scoped; privileged paths documented | T/F | `10-roles.sh`; Appendix A 1, 2, 15–19; `database-rls.md` |
| | Context per transaction; fail closed; pooled reset | F | Appendix A 7, 8, 12 |
| | Exact ObjectStore authorization; prefixes insufficient; credentials scoped and resolved on the server | F (real-server part against source-built MinIO) | `storage`; S-ENV-SCOPE |
| Audit and operations | AuditEvent fields; event categories; controlled modification; retention | F | Appendix A 13–14; `policy`/`mcp` audit cases |
| | Secrets boundary, redacted logs, health and readiness, trace correlation | T | Core skeleton tests; T-SECRET-CANARY; T-OUTAGE |
| | Metrics | P | STAGE-OBSERVABILITY → G003 |
| | Access provisioning and deprovisioning; scanning; CI evidence; encryption | P (secret scan T) | `security-operations.md`; `secret-scan` check; STAGE-OPS → G012 |
| | Backup and restore | F/P | X-BKP; STAGE-OPS → G013 |
| | Retention and deletion; incident evidence | F/P | Appendix A 14; STAGE-OPS → G013 |
| | Vendor inventory | T/P | `VENDORS.md`; STAGE-OPS → G012 |
| MCP | Control-plane tools; switching flow; scope clearing; arguments cannot switch tenants | F | `mcp` |
| | Context-filtered `tools/list`; per-execution revalidation; hidden-tool rejection; staleness (membership, role, policy, expiry); cross-Organization difference | F | `mcp`, including the reconciliation-suspend tests and the HTTP smoke; STAGE-MCP → G014/G015 |
| AI | `AI_MODEL_<CAP>` and `_FALLBACKS`; five capabilities; validation; samples; disabled means no credentials | T | I-CONFIG, T-TOGGLE, S-ENV-SCOPE |
| | Resolver → Gateway → OpenRouter adapter; translation; no `AI_GATEWAY_DEFAULT`; key is a secret | T | I-GATEWAY |
| | Go owns policy and streams; envelope; typed events; no browser-to-intelligence path | T | I-SCHEMA, I-EXECUTE, I-AUTH, T-BOUNDARIES |
| | Authorized callbacks cannot bypass core policy | P | §4.3 rule and verification; STAGE-AI → G033 |
| | Retrieval filtering before the model; scoped memory | P | `intelligence-ai.md`; STAGE-AI → G035 |
| | Usage dimensions; observability privacy; vendor telemetry is not the ledger | T/P | I-SCHEMA (no tenant dimensions in events), I-OBSERVABILITY; durable record rule → G035 |
| Security | Synthetic data, local credentials, redaction; no secrets in templates, logs, manifests or traces | T | S-SECRETS, T-SECRET-CANARY, I-OBSERVABILITY |
| | Development-only fixture identity adapter cannot enable a production bypass | F/P | Test-only and never rendered (S-LEAK and structure checks); guard rule for STAGE-IDENTITY → G004 |
| | Restricted local exposure; authenticated internal contracts; production encryption documented | T/P | T-PORTS, I-AUTH, T-BOUNDARIES; `security-operations.md` |
| | Audit, storage authorization, RLS and context checks mandatory even with a provider's permissions | T | `SKILL.md` invariant; S-FRONTMATTER/section check |
| | No product domain anywhere | T | S-LEAK with the platform denylist |
