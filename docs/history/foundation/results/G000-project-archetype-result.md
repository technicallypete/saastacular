> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G000 — Project archetype result

Goal: [G000-project-archetype.md](../goals/G000-project-archetype.md) (amendment A1: 15-cycle ceiling)  
Spec: [G000-project-archetype-spec.md](../specs/G000-project-archetype-spec.md)  
Plan: [G000-project-archetype-plan.md](../plans/G000-project-archetype-plan.md) (r5, approved by Codex)

## 1. Status

**Resubmitted for Codex re-review after correction loop 1 (§15).** Codex's first review returned CORRECTION LOOP with two bounded gaps; both are fixed and re-validated in §15. The evidence in §1–§14 is preserved as originally submitted, and the corrections are appended in §15.

**Current shared cycle budget: 8 of 15** (correction loop 1 used 1 cycle; the cycle-8 checkpoint is in §15.5).

**Original submission:** implementation and validation complete. This is not a self-acceptance: only Codex can return ACCEPT, CORRECTION LOOP or ARCHITECTURE DECISION REQUIRED, and only Codex updates PROGRESS.md.

- Final validation run: `validation/run.sh all` gave **60/60 checks PASS** and 0 FAIL/UNMET (§6).
- Shared cycle budget: **7 of 15 used**. The checkpoints at cycles 8 and 12 were not reached (§8).
- No platform application code was created. Root `web/`, `core/`, `intelligence/`, `contracts/` and `infra/` do not exist (T-REPO-CLEAN).

## 2. Implementation summary

`skills/project-archetype/` is a self-contained, project-agnostic Skill. It contains no platform brand, domain, product-domain term or goal ID (S-LEAK, S-PATHS). It is made up of:

- **`SKILL.md`**: frontmatter, activation, prerequisites, 16 essential invariants, inputs, a 10-step workflow, the output contract and verification. It is 106 lines long.
- **11 references** under `references/`:
  - boundaries;
  - configuration;
  - Docker runtime;
  - tenancy and authorization, with the frozen permission matrix;
  - database RLS;
  - storage and audit;
  - MCP control plane;
  - intelligence and AI;
  - security operations;
  - commodity composition;
  - the requirements catalog (76 requirement IDs, each classified T/F/P).
- **3 checklists:**
  - foundation follow-up, rendered as `docs/foundation/FOLLOW-UP.md`, with 15 generic foundation stages;
  - invocation;
  - verification.
- **Renderer** (`generator/`): Go, standard library only, run through `generator/render.sh` in `golang:1.27.1-alpine`. It provides:
  - all-errors-at-once input validation;
  - a stable `output_id` that is independent of the destination's state;
  - content-addressed manifests and proposals under `.archetype/outputs/<id>/`;
  - a never-overwrite collision policy (`create`, `unchanged`, `conflict`, `ensure-lines`);
  - symlink and traversal refusals and reserved paths;
  - reports written only outside the destination;
  - exit codes 0, 1, 2 and 3.
- **Templates** (`templates/base`, 39 files; `templates/intelligence`, 28 files), rendered into a destination:
  - **Single-source config.** `.env.example` is the only file carrying identity values.
  - **Compose (`infra/compose/local.yml`):**
    - networks `edge`, `data` and `inference`;
    - `:?` fail-closed variables;
    - per-service secret scoping;
    - loopback, standard+100 ports;
    - a checks profile.
  - **Dockerfiles:** core (distroless), web (Next.js standalone) and MinIO, which is built from the final upstream source releases (R1) with non-root runtimes.
  - **Init scripts:** PostgreSQL role separation in `10-roles.sh`; MinIO init with a scoped credential.
  - **Go core skeleton:** config, secret-provider boundary, readiness, redacting logs and trace correlation, with tests.
  - **Next.js skeleton:** runtime identity, health and ready routes, semantic tokens.
  - **Optional FastAPI intelligence:**
    - model configuration;
    - OpenRouter adapter contract;
    - fail-closed internal auth;
    - an execute endpoint that streams typed events;
    - privacy-safe observability;
    - JSON Schemas.
  - **Foundation docs:** GETTING-STARTED with executable blocks, FOLLOW-UP, and VENDORS.
- **Validation package** (`validation/`):
  - `run.sh` and `lib.sh`, a harness with five suites;
  - `compose.yml`, an isolation environment with two PostgreSQL instances initialized from a freshly rendered role script, plus MinIO built from the rendered Dockerfile;
  - the inputs (alpha, bravo, 20 invalid);
  - the reference Go isolation module: `sql`, `rls`, `policy`, `hostname`, `mcp`, `storage`, and `internal/fixture` with guarded connection helpers.

## 3. Files changed

| Path | Change |
|---|---|
| `skills/project-archetype/**` | New: 161 files, about 16k lines excluding lockfiles (160 untracked, plus `generator/go.mod`, which was carried by commit `06c8ca7`) |
| `management/v0/results/G000-project-archetype-result.md` | This Result |
| `management/v0/results/G000-validation-denylist.txt` | Platform brand and product-domain denylist, kept outside the Skill and passed with `--denylist` |
| `management/v0/goals/G000-project-archetype.md` | Amendment A1 (15-cycle ceiling) appended at the product owner's direction; the original goal text is unchanged |
| `management/v0/plans/G000-project-archetype-plan.md` | r5 budget amendment, with history preserved |

The Goal and Plan amendments were already committed upstream (`06c8ca7 "Docs update"`). That commit was made outside this session and also touched `AGENTS.md` and `management/BUDGET_POLICY.md`; those are not G000 implementation changes. **In the working tree**, nothing outside the paths above changed (T-REPO-CLEAN inspects the working tree, run with `--repo-allow '^(skills/project-archetype/|management/v0/(goals|plans|results)/G000-)'`).

## 4. Generated and scaffolded artifacts

Generated projects exist only in temporary destinations outside the repository and are deleted after every run. A rendered destination contains:

- `.env.example`, `.gitignore` (merged);
- `core/` (module `core`: `cmd/server`, `internal/{config,health,observability}`, tests, `go.sum`);
- `web/` (`app/`, `lib/`, lockfile);
- `contracts/README.md`;
- `infra/{compose/local.yml, docker/{core,web,minio}.Dockerfile + dockerignores, postgres/init/10-roles.sh, minio/init.sh, security/gitleaks.toml}`;
- `docs/foundation/{GETTING-STARTED,FOLLOW-UP,VENDORS}.md`;
- `.archetype/outputs/<output_id>/manifest.json`;
- when intelligence is enabled: `intelligence/` (src, tests, `uv.lock`), `infra/docker/intelligence.Dockerfile` and `contracts/intelligence/{execution-envelope,stream-event}.schema.json`.

Bravo, with intelligence disabled, renders 41 files; the cold-read trial independently rendered 41 for its project.

## 5. Schema, migration and contract changes

- **Platform:** none. No platform schema, migrations, public API or MCP server.
- **Templates:** a role/schema hardening init script (owner NOLOGIN, migrator via `SET ROLE`, runtime NOSUPERUSER/NOBYPASSRLS/NOINHERIT, retention). Migration tooling is prescribed for STAGE-BOOTSTRAP/STAGE-RLS.
- **Fixture schema** (`validation/isolation/sql`, disposable only):
  - `role_bindings` is the single mutable authority store, with read-only `security_invoker` membership views;
  - ENABLE + FORCE RLS on all nine tables;
  - per-role policies;
  - four narrowly scoped definer helpers owned by a NOLOGIN, non-bypass `policy_reader` role;
  - immutability and append-only audit triggers;
  - column-level audit INSERT;
  - object keys bound to their row by a CHECK.
- **Contracts** (intelligence templates): `execution-envelope.schema.json` and `stream-event.schema.json` (JSON Schema 2020-12). Usage events cannot carry tenant or user dimensions.
- **MCP:** fixture tool names follow the canonical-semantics override (`logout`, per Q3). STAGE-MCP owns the real contract.

## 6. Validation commands and outcomes

### 6.1 Final run

```sh
skills/project-archetype/validation/run.sh \
  --denylist management/v0/results/G000-validation-denylist.txt \
  --repo-allow '^(skills/project-archetype/|management/v0/(goals|plans|results)/G000-)' \
  --report-dir <scratch>/final all
```

- **Environment:** Docker 27.5.1, Compose 2.32.4. Host probe mode: bash `/dev/tcp`.
- **Docker-only:** host `python3` and `node` exist but were shadowed; T-DOCKER-ONLY confirms neither was invoked.
- **Result: exit 0, `RESULT: PASSED`.**

| Check | Status | Evidence |
|---|---|---|
| G-DETERMINISM | PASS | identical input -> byte-identical trees and manifests, same output_id |
| G-INVALID | PASS | 20 invalid inputs + Skill/missing destinations: exit 2, zero writes |
| G-OUTPUT-ID | PASS | output_id equal across empty/sentinel/rendered destinations; changes for each of 9 input mutations |
| G-REFUSE | PASS | symlinked target, parent and destination refused with exit 2 and no writes |
| G-REPEAT-CLEAN | PASS | repeat on clean destination: all unchanged (manifest included), hashes and mtimes identical, exit 0 |
| G-REPEAT-CONFLICT | PASS | repeat with unresolved conflicts: same output_id, conflicts re-reported, zero writes, exit 3 |
| G-REPORT | PASS | --report JSON matches stdout actions; never written inside the destination |
| G-SENTINEL | PASS | platform-shaped destination: sentinels intact, .gitignore prefix preserved, conflict + proposal, exit 3 |
| G-TOGGLE-RERENDER | PASS | enable intelligence later: new output_id, intelligence files created, .env.example/local.yml conflicts proposed, nothing overwritten |
| I-AUTH | PASS | 17 passed (pytest -k i_auth, --network none) |
| I-CONFIG | PASS | 36 passed (pytest -k i_config, --network none) |
| I-EXECUTE | PASS | 17 passed (pytest -k i_execute, --network none) |
| I-GATEWAY | PASS | 16 passed (pytest -k i_gateway, --network none) |
| I-LINT | PASS | ruff check + format check |
| I-OBSERVABILITY | PASS | 3 passed (pytest -k i_observability, --network none) |
| I-SCHEMA | PASS | 52 passed (pytest -k i_schema, --network none) |
| S-COMPOSE | PASS | docker compose config -q for all four variants |
| S-DEPS | PASS | service dependency direction (contracts, web, intelligence, core packages) |
| S-DOCKERFILE | PASS | hadolint clean for rendered Dockerfiles |
| S-ENV-SCOPE | PASS | per-service secrets and networks scoped (core, web, intelligence, data stores) |
| S-FAILCLOSED-ENV | PASS | 32 required variables each fail closed when missing |
| S-FRONTMATTER | PASS | SKILL.md frontmatter (yq-parsed) meets Agent Skills rules and section headings exist |
| S-GO | PASS | generator gofmt, vet and unit tests |
| S-LEAK | PASS | zero denylist hits in the Skill and all rendered variants |
| S-LINKS | PASS | relative links resolve inside the Skill package |
| S-NAMING | PASS | no forbidden roots, redundant segments or empty directories |
| S-PATHS | PASS | no machine-specific paths or host-repository citations |
| S-SECRETS | PASS | gitleaks clean (Skill, alpha, bravo); every secret empty in .env.example |
| S-SHELL | PASS | shellcheck on renderer, harness and rendered scripts |
| S-TRACE | PASS | requirements catalog integrity |
| T-BOUNDARIES-alpha | PASS | web isolated from data and intelligence; intelligence isolated from data stores and holds no data credentials; core reaches intelligence; execute endpoint 401/422/typed stream |
| T-BOUNDARIES-bravo | PASS | web isolated from data (intelligence disabled, so the intelligence clauses do not apply; the empty trailing clause in the harness message is cosmetic) |
| T-CHECKS-ALPHA | PASS | documented check block (core, web, intelligence, secret-scan) |
| T-CHECKS-BRAVO | PASS | documented check block (core, web, secret-scan) |
| T-DOCKER-ONLY | PASS | no host Go/Python/Node/psql/redis-cli/mc/aws invocation (shadowed, /mnt/* stripped); no absolute runtime paths |
| T-DOCS-EXEC | PASS | start/check/stop blocks executed verbatim for alpha and bravo |
| T-DOCS-EXEC-START-ALPHA | PASS | documented start block (alpha) |
| T-DOCS-EXEC-START-BRAVO | PASS | documented start block (bravo, overridden ports) |
| T-DOCS-EXEC-STOP-ALPHA | PASS | documented stop block (alpha) |
| T-DOCS-EXEC-STOP-BRAVO | PASS | documented stop block (bravo) |
| T-OUTAGE | PASS | redis/postgres/minio outages -> core 503 naming the dependency, liveness 200, recovery 200; intelligence outage non-blocking; core outage -> web /api/ready 503 |
| T-PERSIST | PASS | postgres table and MinIO object survive down/up; Redis key does not (ephemeral) |
| T-PORTABILITY | PASS | 2x2 matrix: sources byte-identical except allowlist; identity literals confined to .env.example/.archetype |
| T-PORTS-CONFIG-alpha | PASS | docker compose port: every service on 127.0.0.1 at its configured port (configuration evidence only) |
| T-PORTS-CONFIG-bravo | PASS | docker compose port: every service on 127.0.0.1 at its configured port (configuration evidence only) |
| T-PORTS-alpha | PASS | host reachability (devtcp): web, core, intelligence, MinIO API/console HTTP; Redis -NOAUTH; PostgreSQL SSLRequest; refused on <host-private-ip>; internal DNS/ports intact |
| T-PORTS-bravo | PASS | host reachability (devtcp): web, core, MinIO API/console HTTP; Redis -NOAUTH; PostgreSQL SSLRequest; refused on <host-private-ip>; internal DNS/ports intact |
| T-RENAME | PASS | rename + re-domain via .env only (no rebuild: image IDs unchanged); DATABASE_NAME change yields new database and <db>_* roles; no re-render or search/replace |
| T-REPO-CLEAN | PASS | repository changes limited to --repo-allow paths; no root web/core/intelligence/contracts/infra |
| T-SECRET-CANARY | PASS | 51 generated secrets absent from harness output, compose logs, manifests and rendered files |
| T-TOGGLE | PASS | bravo: no intelligence service, Python image, AI configuration or intelligence directory; stack healthy without them |
| X-BKP | PASS | pg_dump -> restore into second instance (same rendered role script); 7 restore-subset tests pass |
| X-FIXTURE | PASS | go test ./internal/fixture: ok   |
| X-HOSTNAME | PASS | go test ./hostname: ok   |
| X-MCP | PASS | go test ./mcp: ok   |
| X-POLICY | PASS | go test ./policy: ok   |
| X-RLS | PASS | go test ./rls: ok   |
| X-SCHEMA | PASS | fixture SQL applied (owner objects via migrator SET ROLE; superuser for roles, definer ownership and seed only) |
| X-SKIPS | PASS | no isolation test skipped |
| X-STORAGE | PASS | go test ./storage: ok   |

### 6.2 Test inventory

| Suite | Tests | Notes |
|---|---|---|
| Generator (Go) | 22 tests (`input_test.go`, `render_test.go`, `skill_test.go`) | Inside S-GO; S-FRONTMATTER, S-LINKS, S-PATHS and S-TRACE run individually |
| Core template (Go) | 16 tests | config, health, observability; T-CHECKS-ALPHA, T-CHECKS-BRAVO |
| Web template | lint + typecheck + `next build` | T-CHECKS-ALPHA, T-CHECKS-BRAVO |
| Intelligence template (pytest) | 143 tests: I-CONFIG 36, I-GATEWAY 16, I-AUTH 17, I-EXECUTE 17, I-OBSERVABILITY 3, I-SCHEMA 52, I-LINT 2 | Run with `--network none` |
| Isolation (Go) | 76 passed, 0 skipped | — |

**Isolation tests by connection:**

| Connection | Tests |
|---|---|
| Runtime | 29 |
| Owner | 2 |
| Retention | 1 |
| Control | 4 (reported separately) |
| S3 | 6 |
| Unit | 33 |
| Guard | 1 |

**Isolation evidence-label records:**

| Label | Records |
|---|---|
| Real/PG | 39 |
| Real/S3 | 6 |
| Real/SDK-inmem | 13 |
| Real/SDK-http | 1 |
| Unit | 20 |
| Control | 4 |

### 6.3 Required images (pinned; digests as resolved on this host)

| Image | Digest |
|---|---|
| alpine:3.22 | (build stage only; no local repo digest) |
| busybox:1.37.0 | sha256:bdf57e528e45e4433820e045b29b4597825a1c9e38353532d90a01445013f82e |
| gcr.io/distroless/static-debian12:nonroot | (build stage only; no local repo digest) |
| ghcr.io/astral-sh/uv:0.12.23 | sha256:61d393e44e249f2e4b526b6c7ddcecce245946826e608e11c93ad4f5bba55b21 |
| golang:1.27.1-alpine | sha256:8a5910f31396cd4d89662f56c68b3ae31d374308270a1c3bd96672ee5ed43414 |
| hadolint/hadolint:v2.15.1 | sha256:32dac94127fd60b7b7e3fbfc65e1383b9b5e25c9bfd7b8536de7a539fe68a12d |
| koalaman/shellcheck:v0.11.0 | sha256:61862eba1fcf09a484ebcc6feea46f1782532571a34ed51fedf90dd25f925a8d |
| mikefarah/yq:4.48.1 | sha256:75b3a60dce1fa4641b0f02bdb4465eddc5da95e62028118a292239ca8e3921e5 |
| node:24-slim | sha256:0e0ff40c39bc087845bfb27465a0df4ea419520094bc35842ff83dd8cbe6f9b6 |
| pgvector/pgvector:0.8.7-pg17 | sha256:ac08538c6f8b9904c33c8224c5e5706dbe760aca29db1d096972b4052c22a75d |
| python:3.13.16-slim | sha256:3dd7cc108ec1493442514f5c2a871af6af0ec31d768ff6e378a93340c3b3db5f |
| redis:8.10.2-alpine | sha256:3811787313eba226a2ef38658c6ccb91cd5e110edc89c37767de373120a0e5a0 |
| zricethezav/gitleaks:v8.30.1 | sha256:c00b6bd0aeb3071cbcb79009cb16a60dd9e0a7c60e2be9ab65d25e6bc8abbb7f |

`alpine:3.22` and `gcr.io/distroless/static-debian12:nonroot` were pulled only as build stages. They have no local repo digest, so pin them by digest when stage evidence needs it.

### 6.4 Supplementary agent evidence (AC1)

- **Cold-read usability trial (subagent 4).** The Skill was copied to `/tmp`, outside any git repository. A headless `claude -p` session started from there, so no project instructions loaded. It was asked for "Lumen Notes, lumen.example.net, database lumen, no AI".
  - **Result:** completed end to end with no blocking issue:
    - validate exited 0;
    - plan showed 41 `create`;
    - render produced 41 files (output_id `75f6227ab506747d`);
    - `.env` was created with hex secrets;
    - `up --build --wait` brought all services healthy;
    - `/readyz`, `/api/ready` and the health endpoints returned 200;
    - the `lumen_*` roles were checked as non-superuser and non-BYPASSRLS;
    - core-check, web-check and secret-scan passed;
    - a repeat plan gave 41 `unchanged`;
    - "lumen" appeared only in `.env`, `.env.example` and the manifest;
    - `down -v` was clean.
  - **Friction it reported, and what I did** (fixed after the trial; static and generator suites green afterwards):

    | Friction | Action |
    |---|---|
    | Hex/URL-safe secrets not stated | Now stated in SKILL.md, `.env.example` and GETTING-STARTED |
    | "POSIX shell" was inaccurate | Now "bash" |
    | Git handling unclear | Preflight now covers the git-repository decision |
    | No port preflight | Added to the workflow |
    | Resolved defaults not shown before writing | `validate` now prints the resolved configuration |
    | Health verification undefined | A readiness command is documented |
    | First-initialization trap for role passwords and `DATABASE_NAME` | Documented with the `down -v` fix |
    | CI needs a full `.env` | Documented |
    | First-run cost not mentioned | Documented |
    | Cleanup and leftovers not mentioned | Documented |
    | Commit and `--report` guidance gaps | Filled |
- **Security review (subagent 3).** No High findings. M1, M2 and L1–L7 were fixed in cycle 5 (§8).

## 7. Acceptance-criteria checklist

| # | Criterion | Status | Evidence |
|---|---|---|---|
| AC1 | Artifacts exist and are consumable without this project's knowledge; links, metadata, configuration contract and workflow are valid | Met | S-FRONTMATTER, S-LINKS, S-PATHS, S-LEAK, S-TRACE; `render.sh` driven end to end by G-* and T-*; T-DOCS-EXEC; cold-read trial §6.4 |
| AC2 | Every requirement maps to a resource and an enforcement method; later obligations are explicit | Met | `references/requirements-catalog.md` (76 IDs, integrity checked by S-TRACE); `checklists/foundation-follow-up.md` rendered as FOLLOW-UP (15 stages); Appendix C below |
| AC3 | Portability, intelligence toggle, Docker-only and destination-safety trials pass | Met | T-PORTABILITY, T-RENAME, T-TOGGLE, G-TOGGLE-RERENDER, T-DOCKER-ONLY, T-DOCS-EXEC, G-INVALID, G-SENTINEL, G-REPEAT-CLEAN, G-REPEAT-CONFLICT, G-OUTPUT-ID, G-REPORT, G-REFUSE, G-DETERMINISM |
| AC4 | Tenant/RLS/policy/MCP/storage/audit fixtures with positive and negative evidence, mock/runtime distinctions, and no bypass-role false positives | Met | X-RLS, X-POLICY, X-HOSTNAME, X-MCP, X-STORAGE, X-FIXTURE (incl. `TestGuard_ConnectionScope`), X-BKP, X-SKIPS. Runtime assertions run only through `RuntimeDB` (guard on every acquire). Owner tests have an authorized positive control and a foreign denial. Controls are labelled. SQLSTATE mechanism assertions. The MCP suspended-catalog tests prove a handler policy denial, not unknown-tool |
| AC5 | Folder boundaries, +100 ports, tenant/RoleBinding model, audit, AI configuration and OpenRouter boundary match the architecture; no product domains | Met | S-NAMING, S-DEPS, S-ENV-SCOPE, T-PORTS-alpha/bravo, T-BOUNDARIES-alpha/bravo, I-CONFIG, I-GATEWAY, I-SCHEMA, S-LEAK, the permission-matrix doc-sync test |
| AC6 | The diff touches only the Skill, fixtures and lifecycle artifacts; the platform application stays unimplemented | Met | T-REPO-CLEAN; §3 |
| AC7 | No unapproved infrastructure, provisioning, global installation, real secret or live model dependency | Met | Images are those the architecture names (MinIO retained, source-built per R1), plus validation-only tool images. No Skill installation. S-SECRETS and T-SECRET-CANARY (51 generated secrets absent from logs and artifacts). The stub transport makes no live calls |
| AC8 | Codex ACCEPT | **Pending Codex review** | — |

## 8. Cycle accounting (plan §6.1; Goal amendment A1)

| Agent | Allowance | Used |
|---|---|---|
| Main agent | 9 | **5** |
| Subagent 1 (WP6 intelligence) | 2 | **0** |
| Subagent 2a (WP7a sql/rls/storage) | 2 | **1** |
| Subagent 2b (WP7b policy/hostname/mcp) | 2 | **1** |
| Subagents 3, 4, 5 (read-only review, cold read, spec trace) | 0 | **0** |
| **Total** | **15** | **7** |

**Reconciliation:** 5 + 0 + 1 + 1 = 7.

- No allowance was reallocated.
- The mandatory checkpoints at cycles 8 and 12 were not reached, because the goal finished at cycle 7.
- Routine runs that passed first time were not counted: the core and web first builds, generator tests, the bravo smoke, and the doc-fix re-run.

**Each counted cycle:**

| Cycle | Agent | Summary |
|---|---|---|
| 1 | Main, WP2 | The fixture schema failed on `ALTER FUNCTION … SET` with a custom setting (needs superuser on PG17). The audit trigger is now generated through `\gexec`. Retest passed |
| 2a-1 | Subagent 2a, WP7a | `TestRuntime_AuthoritySingleStore` catalog scan matched an index; restricted to relation kinds. Retest passed |
| 2b-1 | Subagent 2b, WP7b | The MCP HTTP smoke got 403 from the SDK's DNS-rebinding protection. The test now carries the virtual host in a header, with the protection kept on. Retest passed |
| 2 | Main, WP8/WP9 | First integrated run. The harness aborted on the first FAIL (`set -e`); S-PATHS false positive (`/home/nonroot`); non-literal check IDs. Fixed |
| 3 | Main, integration | S-FRONTMATTER mount name, shellcheck warnings, gitleaks on a test sentinel, ruff cache, the case-20 expectation. Fixed; run 3 green |
| 4 | Main, WP8 | T-PORTABILITY used `diff --exclude-dir` (invalid). Fixed; trials green |
| 5 | Main, review | Security-review fixes (M1, M2, L1–L7 and hardening). Retest `run.sh all` green (60/60) |

The full execution log, including each handoff report, is in §13.

## 9. Subagents used

| Agent | Responsibility | Outcome |
|---|---|---|
| 1 | WP6: `templates/intelligence/**` (FastAPI, model configuration, OpenRouter adapter contract, internal auth, schemas, pytest) | Complete; 0 cycles |
| 2a | WP7a: `validation/isolation/{rls,storage}`, plus defect authority over `sql/` (used for no changes) | Complete; 1 cycle |
| 2b | WP7b: `validation/isolation/{policy,hostname,mcp}` (`types.go` frozen) | Complete; 1 cycle |
| 3 | Read-only security review | 2 Medium, 7 Low findings, all fixed |
| 4 | Cold-read usability (headless session outside the repository) | Succeeded end to end; friction items fixed |
| 5 | Read-only spec-trace verifier of Appendix C | See §14 |

Every subagent worked on its own non-overlapping paths. I reviewed and integrated all delegated work, and I ran every final validation myself.

## 10. Architecture deviations

**None architectural.** The plan-level and tactical deviations, disclosed for review:

1. **Fixture schema ownership.** I wrote `sql/` as part of the WP2 contract freeze, where the plan gave it to 2a, so that 2a and 2b could work in parallel on one schema. This changes ownership only.
2. **Seed addition.** `bea@example.test` was added as a bravo-only identity so that Appendix A case 20 is asserted in its literal form. In bravo, ada is a plain member and by design sees only herself.
3. **Retention role at the template level.** `<db>_retention` is created by the template role script (plan §5), and the fixture adds a column-limited grant and RLS for it.
4. **Command output.** `validate` now prints `resolved\t<json>` instead of `valid\t<id>`, a usability fix from the cold-read trial. This is documented in `references/configuration.md`.
5. **Host Python used for editing.** During authoring I used host `python3` as a text-editing aid for plan and template files, before I stopped and switched to editor tools; subagents 2a and 2b each once ran an empty host `python3` heredoc by mistake. None of this is in any project or validation path: every build, test and check ran in containers, and T-DOCKER-ONLY passed with host runtimes shadowed.
6. **Harness-only command-line secrets.** These remain in the short-lived validation containers (`-e PGPASSWORD=…`, the busybox token). This is accepted as a validation-tooling residual. The consumer-facing templates were fixed (L5).

## 11. Unresolved issues

These are open items. Nothing is failing.

1. **R1 escalation.** MinIO is retained as specified and built from the archived final release. Replacing it remains a product-owner ADR decision; the risk is recorded in VENDORS.md and `references/security-operations.md`.
2. **Interpretations for Codex to confirm:**
   - Q2: identity inputs are required, with neutral samples provided;
   - Q3: `logout`;
   - Q4: Node rather than Bun;
   - Q5: intelligence port bound to loopback;
   - Q6: `actual_provider` is `unknown` unless the gateway reports it;
   - Q7: memberships are projections of RoleBinding;
   - Q8: a gateway named as provider is rejected.

   The implementation follows the plan's stated choices.
3. **Subagent 2b interpretation choices:**
   - a Group or BusinessEntity binding establishes Organization context;
   - role grants need `*.manage_members`, plus the no-escalation rule added in cycle 5;
   - scope arguments are rejected rather than ignored;
   - the catalog is emptied on logout or expiry.
4. **Intelligence.** A configured `embedding` capability returns `capability_not_supported`; this is the R11 extension point. There is also a Starlette TestClient deprecation warning, which is not failing.
5. **Accepted residuals:**
   - the owner/migration path can disable audit triggers (documented);
   - RLS does not enforce in-Organization grant authority; application policy does, by design.
6. **Image pins.** `alpine:3.22` and distroless are pinned by tag only (§6.3).

## 12. Recommended follow-up work

These are suggestions only and authorize no new scope.

- G001 consumes the Skill. Render into the platform root with the platform's own input values and reconcile conflicts. The platform values are `Beginify`/`foundation.test`, supplied at that point and never inside the Skill.
- Product-owner decision on the R1 local object-store ADR, and a Redis licence review (R5).
- STAGE items in `docs/foundation/FOLLOW-UP.md`: migration tool choice, MCP mount with identity, shadcn/ui with Base UI, scanning with pinned vulnerability databases, digest pinning for base images.

## 13. Execution log (append-only)

Cycle accounting follows plan §6.1. A cycle is counted when a test or validation run of implemented work fails materially and is then diagnosed, fixed and retested.

Allowance allocation (plan §6.2, r5):

| Agent | Allowance |
|---|---|
| Subagent 1 | 2 |
| Subagent 2a | 2 |
| Subagent 2b | 2 |
| Main agent | 9 |
| **Ceiling** | **15** |

| # | Agent | WP | Attempted | Failing IDs | Diagnosis / fix | Retest |
|---|---|---|---|---|---|---|
| 1 | Main | WP2 | Apply the frozen fixture schema (`sql/00–40`) to a disposable PostgreSQL 17 initialized by the template role script | `10_schema.sql` failed: "permission denied to set parameter app.retention_role" | A non-superuser cannot attach a custom setting to a function (`ALTER FUNCTION … SET`). The audit trigger function is now generated through psql `\gexec` with the retention role name embedded as a literal | Pass: the schema applies, runtime visibility is correct for ada and lin, a forged insert is blocked by RLS, no context gives 0 rows, a runtime audit UPDATE is denied |
| 2 | Main | WP8/WP9 | First integrated harness run: static, intelligence and isolation suites | S-GO (TestSkillPaths, TestSkillTrace); the harness aborted after the first FAIL | (a) Under `set -e`, `check` returned 1 and stopped the run; it now records the result and returns 0. (b) S-PATHS flagged the distroless home `/home/nonroot`; container homes are now allowlisted. (c) S-TRACE: package check IDs were built dynamically (`X-$pkg`); they are now literal | The cycle's defects are fixed and the run completes. New failures surfaced, see cycle 3 |
| 3 | Main | WP6/WP7a/WP8 integration | Retest of all three suites | S-FRONTMATTER, S-SHELL, S-SECRETS, I-LINT, X-RLS | (a) The Skill was mounted as `/skill`, so the name/directory check failed; it is now mounted under its own name. (b) Shellcheck warnings (unused variables) fixed; severity threshold set to warning; neutral mount paths. (c) gitleaks flagged the deliberately key-shaped sentinel in the intelligence privacy test; added inline `gitleaks:allow` with a justification. (d) The harness ran ruff without `--no-cache` in a read-only image. (e) The main agent's case-20 expectation was wrong: in bravo ada is a plain member and by design does not see other members; the B-side assertion was restored | Pass: run 3 green across all checks of static, intelligence and isolation |
| 4 | Main | WP8 | First trials-suite run | T-PORTABILITY (every other trial check passed) | `diff --exclude-dir` is not a GNU diff option; switched to `diff -x` | Pass: trials suite green |
| 5 | Main | WP10 (review fixes) | Fixes for the subagent-3 security review findings (M1, M2, L1–L7 and hardening notes) | Run 6 (`run.sh all`) is the retest | M1: no-escalation rule in `AuthorizeRoleGrant`, and role changes authorize against the old role. M2: `.env*` excluded from the web and core build contexts, plus a static check. L1: column-level audit INSERT (no `occurred_at`). L2: object key bound to its row (CHECK plus authorizer). L3: `members.manage` requires existing Organization membership, with uniform denial. L4: render binary keyed by source hash; report written with O_NOFOLLOW. L5: init scripts take secrets from the environment (`\getenv`, `MC_HOST_local`). L6: impossible deadline gives 422; streaming body cap. L7: `minio-init` runs non-root. Also: definer `search_path` adds `pg_temp`; audit comment corrected; duplicate custom domains rejected. New tests: `TestUnit_AdminCannotEscalateThroughRoleGrants`, `TestRuntime_ObjectKeyBoundToRow`, the backdating assertion, and two I-EXECUTE tests | See run 6 |

### Progress notes

- 2026-10-04: implementation started under plan r5 and Goal amendment A1.
- WP1 (generator and `render.sh`) authored. WP3 (core skeleton) authored; its first vet and test run passed, so no cycle was counted.
- WP2 frozen: isolation `go.mod` pre-seeded (pgx v5.11.0, MCP Go SDK v1.8.0, aws-sdk-go-v2 s3 v1.114.0), `internal/fixture` (seed IDs, guarded connection helpers, tenant transactions, labels, `TestGuard_ConnectionScope`) and `policy/types.go`.
  - **Tactical ownership adjustment:** the fixture SQL schema (`sql/00–40`) was written by the main agent as part of the WP2 contract freeze, so subagents 2a and 2b build on one schema in parallel. 2a may fix defects in it; 2b treats it as read-only. This changes ownership only, not scope.
  - The permission-matrix block in `references/tenancy-authorization.md` was frozen for 2b's doc-sync test.
- Delegation started. Subagents 1 (WP6 intelligence), 2a (WP7a sql/rls/storage) and 2b (WP7b policy/hostname/mcp) were spawned in parallel with written briefs. Each brief carries the §6.1 counting rules verbatim and an allowance of 2 cycles.
- WP4 (web skeleton): the first check-image run (lint + typecheck) and the first `next build` both passed, and every route is dynamic (`ƒ`). No cycle.
- WP5 (infra and docs): a smoke render of bravo through `render.sh` passed `compose config`. The documented start command brought all services healthy, and core `/readyz` and the web `<title>` reported "Harbor Desk". The check block passed (core-check, web-check, secret-scan). No cycle.
- **Subagent 1 handoff (WP6): 0 cycles, status complete.**
  - The `check` image passed 141 tests with `--network none`: I-CONFIG 36, I-GATEWAY 16, I-AUTH 17, I-EXECUTE 15, I-OBSERVABILITY 3, I-SCHEMA 52, I-LINT 2.
  - The runtime image was probed: `/healthz` 200; execute without a token 401; an unconfigured capability gives `error(capability_not_configured)` then `done`.
  - hadolint is clean.
  - Pins: `python:3.13.16-slim`, `ghcr.io/astral-sh/uv:0.12.23`, fastapi 0.142.2, uvicorn 0.54.0, jsonschema 4.26.0, httpx 0.28.1, pytest 9.1.1, ruff 0.16.10.
  - Unresolved items for review (nothing is failing):
    - a configured `embedding` capability returns `capability_not_supported` (R11 extension point);
    - extra fail-closed error codes (`deadline_exceeded`, `gateway_not_configured`, `gateway_timeout`/`gateway_error`), a 413 body limit, and docs/OpenAPI disabled;
    - content capture is a constructor flag, not an env variable;
    - `actual_provider` is read from the response `provider` field when present, else `unknown`;
    - a Starlette TestClient deprecation warning.
- **Subagent 2a handoff (WP7a): 1 cycle, status complete.**
  - 47 tests pass against the dev environment (28 in `rls`, 19 in `storage`). Every Appendix A case is mapped to a test (cases 1–24 plus audit attribution). No SQL changes; `go.mod` untouched.
  - **Its cycle 1:** `TestRuntime_AuthoritySingleStore`'s catalog scan matched the partial index `role_bindings_unique_grant` (relkind `i`). Fixed in the test by restricting the scan to relkinds r, p, v, m and f. Retest passed.
  - Unresolved items:
    - case 20 lacked a bravo-only identity in the seed, so 2a asserted the mirror form instead (actioned by the main agent; see the integration notes);
    - the case-24 evaluator portion belongs to 2b's store test;
    - the `TestMain` preflight opens raw identity-only connections;
    - the Real/S3 tests are named `TestS3_*`;
    - the Authorizer has a minimal read rule behind a `ReadCapability` interface;
    - 2a once ran an empty host `python3 -` heredoc by mistake; it executed nothing.
- **Subagent 2b handoff (WP7b): 1 cycle, status complete.**
  - Policy (13 tests), hostname (2) and MCP (15) all pass. The H6 suspended-catalog tests assert a structured `policy_denied` result that is distinct from the SDK unknown-tool error (-32602), plus a denial audit event.
  - Mutation checks on a throwaway copy proved the doc-sync and suspended-catalog tests fail when the behaviour breaks.
  - **Its cycle 1:** `TestUnit_MCPStreamableHTTPPerIdentity` got 403, because the SDK's DNS-rebinding protection rejected the virtual `Host` header on a loopback listener. Fixed in the test only: the virtual host now travels in an `X-Fixture-Host` test header and the SDK protection stays on. Retest passed (`-count=5`).
  - Unresolved items:
    - the doc-sync test needs `references/` mounted (actioned: the harness now mounts it and sets `ARCHETYPE_REFERENCES_DIR`);
    - interpretation choices: a Group or BusinessEntity binding establishes Organization context; role grants need `*.manage_members`; scope arguments are rejected rather than ignored; switch tools accept an ID or slug; the catalog is emptied on logout or expiry;
    - one run used `GOFLAGS=-mod=mod`, and `go.mod`/`go.sum` were verified unchanged;
    - 2b once ran an empty host `python3` heredoc; it executed nothing.
- **Cycle tally after all three handoffs: main 1, subagent 1: 0, subagent 2a: 1, subagent 2b: 1, total 3 of 15.**
- **Integration (main agent):**
  - Seed gained bravo-only identity `bea@example.test` (`IdentityBea`), so Appendix A case 20 is now asserted in its literal form: the B-only identity is invisible from A (2a issue 1).
  - The harness isolation runner now mounts `references/` for 2b's doc-sync test.
  - T-REPO-CLEAN takes `--repo-allow`, so the Skill carries no platform paths.
  - The IMAGES evidence now lists only the images that are actually referenced.
- **Subagent 3 (security review, read-only): 0 cycles.** Findings: no High; 2 Medium (M1 role-grant escalation by org_admin; M2 `.env*` not excluded from web/core build contexts); 7 Low (L1 audit backdating, L2 object key not bound to its row, L3 `members.manage` could add outside identities, L4 renderer cache race and report symlink, L5 secrets on command lines, L6 malformed deadline gives 500 and body cap applied after buffering, L7 `minio-init` runs as root); plus informational notes. All fixed in cycle 5, except harness-only command-line secrets (short-lived validation containers), which are recorded as a residual.
- **Cycle 5 retest:** run 6 (`run.sh all`) was green at 60/60, closing cycle 5. **Tally: main 5, subagent 1: 0, 2a: 1, 2b: 1, total 7 of 15.** The checkpoints at cycles 8 and 12 were not reached.
- **Subagent 4 (cold-read, headless `claude -p` outside the repository): 0 cycles.** It completed end to end (§6.4). Its friction items were fixed as documentation changes plus the `validate` resolved-configuration output. A re-run of static and generator was green first time (26 PASS), so no cycle was counted. The cold-read workspace, which held generated secrets, and the trial images were removed.
- **Final official run** (`run.sh all`, exit 0, 60/60 PASS) recorded in §6.
- **Subagent 5 (spec-trace verifier, read-only): 0 cycles.** Its gaps and corrections were applied as documentation and traceability changes (§14): new catalog ID AI-07, the SAFE-08 wording, and a STAGE-AI checklist item. A re-run of static and generator (S-TRACE, S-LINKS, S-LEAK, S-NAMING and the others) was green first time (26 PASS), so no cycle was counted. Trials were not re-run for this text-only change: the only rendered effect is the verbatim FOLLOW-UP checklist text, which the static suite covers (S-LEAK, S-LINKS, S-NAMING, S-SECRETS over the rendered variants).
- **Final cycle tally: main 5, subagent 1: 0, 2a: 1, 2b: 1, subagents 3/4/5: 0, total 7 of 15.**
- **Correction loop 1 (2026-10-04, §15):**
  - Fixed null `intelligence_enabled` acceptance and the per-datastore, per-port connectivity probing, adding T-BOUNDARIES-CONTROLS.
  - Run CL-1 (static + generator + trials): 44 PASS, 1 FAIL. The only failure was T-REPO-CLEAN, flagging Codex's own PROGRESS.md edit.
  - Run CL-2 (static, with that path disclosed in the allow-list): 17/17 PASS.
  - **Tally: main 6, total 8 of 15.** The mandatory cycle-8 checkpoint is recorded in §15.5.

## 14. Spec-trace verification (subagent 5)

Subagent 5 (read-only, 0 cycles) independently re-traced every spec requirement against this Result, the catalog, the follow-up checklist and the test sources.

**Confirmed accurate:**
- All 57 Go test names cited in Appendix C exist; the wildcards resolve (`TestRuntime_SignedURL*` matches 5, `TestS3_*` matches 6, `TestUnit_MCPSuspendedCatalog{…}` matches 4).
- The isolation inventory matches §6.2 (76 tests).
- The §6.1 table has 60 PASS rows, and every cited check ID is among them.
- The pytest suites exist (143 = 141 at handoff + 2 added in cycle 5).
- The catalog has 75 rows (76 after AI-07 below); all 15 STAGE sections exist.
- SKILL.md has 106 lines, 16 invariants and 10 steps.
- No platform root service directories exist.
- The cycle arithmetic is consistent (5 + 0 + 1 + 1 = 7 of 15).

**Gaps found and actions taken** (documentation and traceability only; no behaviour changed):

| # | Gap | Action |
|---|---|---|
| 1 | Web shell (Tailwind, shadcn/ui with Base UI) not traced | Added Appendix C row OPS-07: tokens and Tailwind are T; shadcn/Base UI is P, STAGE-WEB-SHELL → G002 |
| 2 | Migrations in containers not traced | Added row OPS-08/RLS-10, P (STAGE-BOOTSTRAP/STAGE-RLS) |
| 3 | AI row over-claimed Go-side streaming | Split it: new catalog ID **AI-07** (Go public streams, cancellation and deadlines, approvals; P, STAGE-AI → G034) and a matching STAGE-AI checklist item. The T row now covers the intelligence side only |
| 4 | Production follow-ups uncited | AUD-04, POL-06, RLS-10, STO-03, MCP-05, TEN-12 and CFG-03 added to their rows |
| 5 | SAFE-08 class | Re-labelled: T for paths and uncatalogued skills; the composition record is stated to be reviewed documentation, not a tested claim (the catalog text is updated too) |
| 6 | ADR rule untraced | Added a row (process; R1 escalation; AC7) |
| 7 | Stage map | G010 noted for production attack verification of the policy, RLS and storage stages; G038 annotated as a product-phase consumer of BillingProfiles |

**Factual corrections:**
- `T-CHECKS` replaced by the real IDs `T-CHECKS-ALPHA` and `T-CHECKS-BRAVO` throughout this Result. The catalog's prefix form is valid because S-TRACE matches it against `run.sh`.
- The file count is 161; `generator/go.mod` was carried by `06c8ca7`, whose other out-of-scope files are disclosed in §3.
- The T-BOUNDARIES-bravo message is annotated.
- **Test-count difference:** 2a's handoff counted 47 tests including subtest-level `--- PASS` lines (28 rls, 19 storage). The final top-level inventory is 44 across those packages (30 rls including the cycle-5 additions, 14 storage); §6.2 counts top-level tests plus named subtests from the harness's `tests.tsv`.

**Verdict after these actions:** every spec requirement is traced to T, F or P evidence. The catalog, checklist and appendix changes were re-validated by the static and generator suites (see the final note in §13).

## Appendix C — Spec requirement coverage index (final)

**Class:**
- **T:** template-enforced; verified by a harness check.
- **F:** fixture-exercised; verified by a named test.
- **P:** prescribed; owned by a foundation stage and mapped to this platform's goals.

Catalog IDs refer to `skills/project-archetype/references/requirements-catalog.md`.

**Foundation stages → platform goals:**

| Stage | Goal(s) |
|---|---|
| STAGE-BOOTSTRAP | G001 |
| STAGE-WEB-SHELL | G002 |
| STAGE-OBSERVABILITY | G003 |
| STAGE-IDENTITY | G004 |
| STAGE-TENANCY | G005/G006; BillingProfiles are consumed later by G038 (product phase) |
| STAGE-POLICY | G007; production attack verification in G010 |
| STAGE-RLS | G008; production attack verification in G010 |
| STAGE-STORAGE | G009; production attack verification in G010 |
| STAGE-AUDIT | G011 |
| STAGE-OPS | G012/G013 |
| STAGE-MCP | G014/G015 |
| STAGE-DOMAINS | G016/G017 |
| STAGE-ONBOARDING | G018/G019 |
| STAGE-LABELS | G020 |
| STAGE-AI | G033–G036 (G034 owns Go public streaming, cancellation and deadlines) |

| Spec area | Requirement (abridged) | Catalog | Class | Evidence (final run) / owner |
|---|---|---|---|---|
| Skill contract | Valid `SKILL.md` with metadata, activation, inputs, workflow, references, output and verification; invariants in the entry point | SAFE-07 | T | S-FRONTMATTER, S-LINKS, S-PATHS; T-DOCS-EXEC; cold-read trial |
| | Inputs (name, domain, database, CNAME, providers, port overrides, intelligence); required vs optional; validation before output | SAFE-01 | T | G-INVALID (20 invalid inputs plus 2 destination cases); `references/configuration.md` |
| | Identity via `PRODUCT_NAME`, `PLATFORM_BASE_DOMAIN`, `CUSTOM_DOMAIN_CNAME_TARGET` and `DATABASE_NAME`; neutral samples; no literals | CFG-01, CFG-02 | T | T-PORTABILITY, T-RENAME, S-LEAK; `examples/neutral.input.json` |
| | Manifest of files and follow-up requirements | SAFE-04 | T | G-OUTPUT-ID, G-DETERMINISM, G-REPEAT-CLEAN; `docs/foundation/FOLLOW-UP.md` |
| | Inspect the destination; no overwrite, deletion, instruction replacement, secrets or provisioning; collision behaviour; safe repeat | SAFE-02, SAFE-03, SAFE-05, SAFE-06 | T | G-SENTINEL, G-REPEAT-CLEAN, G-REPEAT-CONFLICT, G-REFUSE, G-REPORT, G-TOGGLE-RERENDER, S-SECRETS |
| | Commodity composition, with reuse, reasons, assumptions and fallbacks; no machine paths | SAFE-08 | T (paths) + documentation | S-PATHS verifies there are no machine paths and no required uncatalogued skills. The reuse, reasons, assumptions and fallbacks in `references/commodity-composition.md` are reviewed documentation, **not** a tested claim |
| Boundaries | Web shell: Tailwind + shadcn/ui with Base UI, semantic white-label tokens | OPS-07 | T/P | Semantic tokens and Tailwind are templated (T-CHECKS-ALPHA, T-CHECKS-BRAVO build and lint the web skeleton); shadcn/ui with Base UI is STAGE-WEB-SHELL → G002 |
| | Responsibilities of web, core, intelligence, contracts and infra; dependency direction | BND-02, BND-03 | T | S-DEPS, S-ENV-SCOPE, T-BOUNDARIES-alpha/bravo |
| | Non-redundant naming; packaging exceptions explained | BND-01 | T | S-NAMING; `references/service-boundaries.md` |
| | No placeholder modules; intelligence-disabled policy (omitted) | BND-01, BND-04 | T | S-NAMING (no empty directories), T-TOGGLE, G-TOGGLE-RERENDER |
| Docker | Containers only; Git and Docker host prerequisites; exceptions named | DKR-01 | T | T-DOCKER-ONLY; `references/docker-runtime.md` |
| | Migrations run in containers | OPS-08, RLS-10 | P | No migration tool is templated (the fixture schema is applied with `psql` in containers); STAGE-BOOTSTRAP → G001 and STAGE-RLS → G008 |
| | Compose services, health and readiness, volumes, start/check/stop; Redis ephemeral; S3 abstraction with MinIO | DKR-02, DKR-04, DKR-05 | T | T-DOCS-EXEC, T-CHECKS-ALPHA/BRAVO, T-OUTAGE, T-PERSIST |
| | +100 ports; configurable; collision validation; internal DNS unaffected; no unintended exposure | DKR-03 | T | G-INVALID (port collision and range), T-PORTS-CONFIG-*, T-PORTS-alpha/bravo (host reachability; refused on non-loopback), S-FAILCLOSED-ENV (CFG-03) |
| | Syntax of resources checked with containerized tools | DKR-06 | T | S-COMPOSE, S-SHELL, S-DOCKERFILE, S-FRONTMATTER (yq) |
| Identity and tenancy | Global identity across Organizations with differing roles; no bleed | TEN-04 | F | `TestUnit_NoCrossOrganizationBleed`, `TestUnit_MCPCatalogDiffersByOrganization` |
| | IdP interface, session/revocation, MFA/SSO; no custom auth | TEN-10 | P | STAGE-IDENTITY → G004 (MCP logout and expiry exercised in fixtures) |
| | Organization as boundary; `organization_id`; TenantContext not persisted | TEN-01 | F | `TestRuntime_RLSCoverage`, `TestUnit_TenantContextOnlyFromAuthenticatedIdentity` |
| | Group UUID and slug, `UNIQUE(organization_id, slug)`, reserved routes; slugs and hostnames grant nothing | TEN-02, TEN-11, TEN-12 | F | `TestRuntime_Slugs`, `TestRuntime_MoveRow`, `TestUnit_HostnameNormalizationAndSelection`, `TestUnit_HostnameSelectorGrantsNothing` |
| | BusinessEntity across many Groups; DB and service same-Organization enforcement | TEN-03 | F | `TestRuntime_CrossOrgAssociation`, `TestUnit_AssociationValidationSameOrganization` |
| | Memberships and direct grants (canonical shapes; roles derived from RoleBinding) | TEN-05 | F | `TestRuntime_AuthoritySingleStore`, `TestRuntime_CrossOrgMembershipBinding`, `TestRuntime_GroupScope`, `TestRuntime_IdentitiesAndMemberships` |
| | Group onboarding; unauthorized IDs cannot complete | TEN-06, TEN-07 | F/P | `TestUnit_OnboardingSelectionAndCompletion`; full flow STAGE-ONBOARDING → G018/G019 |
| | Billing-profile extensibility; label owner/target and privacy default | TEN-08, TEN-09 | P | STAGE-TENANCY → G006/G038; STAGE-LABELS → G020 |
| Policy and RLS | Shared RoleBinding; server-side derivation; no mirrored rows | POL-01, POL-06 | F/P | `TestUnit_MatrixMatchesReference`, `TestRuntime_RevocationSingleRowChange`; production policy engine STAGE-POLICY → G007/G010 |
| | Deny by default; Group-admin editing does not imply ownership; generic matrix | POL-02, POL-03, POL-05 | F | `TestUnit_GroupAdminEditsProfileButNotSensitive`, `TestUnit_DirectOwnerTransfersOwnership`, `TestUnit_PolicyToggleRemovesInheritedEditing`, `TestUnit_MemberRoleGrantEscalationDenied`, `TestUnit_AdminCannotEscalateThroughRoleGrants`, `TestUnit_SharedBusinessEntityGivesNoGroupVisibility` |
| | Server-resolved context; selectors are not authority; hostname still requires authorization | POL-04, TEN-11 | F | `TestUnit_ForgedSelectorsMatchNonexistent`, `TestUnit_MCPScopeArgumentsCannotChangeScope`, `TestUnit_HostnameSelectorGrantsNothing`; STAGE-DOMAINS → G016 |
| | RLS reads and writes; same-tenant FKs; global identity table | RLS-02, RLS-04 | F | `TestRuntime_ReadIsolation`, `TestRuntime_ForgedInsert`, `TestRuntime_ForeignUpdateDelete`, `TestRuntime_CrossOrgAssociation`, `TestRuntime_GroupScope`, `TestRuntime_IdentitiesAndMemberships` |
| | Role separation; no superuser or BYPASSRLS; owner bypass prevented; definer narrowly scoped; privileged paths documented | RLS-01, RLS-06, RLS-07, RLS-08, RLS-09 | T/F | X-SCHEMA; `TestRuntime_Precondition`, `TestOwner_ForcedRLS`, `TestRuntime_RoleSwitching`, `TestRuntime_DisableProtections`, `TestRuntime_RowSecurityOff`, `TestRuntime_ObjectCreation`, `TestRuntime_DefinerAllowlist`, `TestControl_BypassSeesForeign`; `references/database-rls.md` |
| | Per-transaction context; fail closed; pooled reset | RLS-03, RLS-05 | F | `TestRuntime_NoContext`, `TestRuntime_MalformedContext`, `TestRuntime_PooledReuse`, `TestControl_SessionSetConfigLeaks` |
| | Exact ObjectStore authorization; prefixes insufficient; credentials scoped and resolved on the server | STO-01, STO-02, STO-03 | F/P | `TestRuntime_SignedURL*`, `TestRuntime_ObjectKeyBoundToRow`, `TestS3_*` (against real MinIO); S-ENV-SCOPE; configurable production adapter STAGE-STORAGE → G009 |
| Audit and operations | AuditEvent fields and categories; controlled modification; retention | AUD-01, AUD-02, AUD-03, AUD-04 | F/P | `TestRuntime_AuditAttribution`, `TestRuntime_AuditPrivileges`, `TestOwner_AuditProtection`, `TestRetention_AuditRetention`, `TestUnit_MCPAuditEvents`, `TestRuntime_RevocationSingleRowChange`; completeness for every foundation mutation STAGE-AUDIT → G011 |
| | Secrets boundary, redacted logs, health and readiness, trace correlation | OPS-01 | T | Core template tests (T-CHECKS-ALPHA, T-CHECKS-BRAVO), T-OUTAGE, T-SECRET-CANARY |
| | Metrics, CI evidence | OPS-05 | P | STAGE-OBSERVABILITY → G003 |
| | Provisioning, scanning, encryption, backup/restore, retention/deletion, incident, vendor inventory | OPS-02, OPS-03, OPS-04, OPS-06 | T/F/P | secret-scan (T-CHECKS-ALPHA, T-CHECKS-BRAVO), X-BKP, VENDORS.md (T-PORTABILITY); remaining items STAGE-OPS → G012/G013 |
| MCP | Control-plane tools; switch flow; scope clearing; arguments cannot switch | MCP-01, MCP-02 | F | `TestUnit_MCPOrganizationSwitchClearsGroup`, `TestUnit_MCPGroupSwitchRequiresGrant`, `TestUnit_MCPListOrganizationsOnlyOwn`, `TestUnit_MCPLogoutInvalidates`, `TestUnit_MCPScopeArgumentsCannotChangeScope`, `TestUnit_MCPToolNamesValid` |
| | Context-filtered `tools/list`; per-execution revalidation; hidden-tool rejection; staleness; cross-Organization difference | MCP-02, MCP-03, MCP-04, MCP-05 | F | `TestUnit_MCPCatalogDiffersByOrganization`, `TestUnit_MCPHiddenToolFailsLikeUnknown`, `TestUnit_MCPApprovalRequired`, `TestUnit_MCPSuspendedCatalog{RevokedGrant,RoleDowngrade,PolicyToggle,SessionExpired}`, `TestUnit_MCPStreamableHTTPPerIdentity`; STAGE-MCP → G014/G015 |
| AI | `AI_MODEL_<CAP>` and `_FALLBACKS`; five capabilities; validation; samples; disabled means no credentials | AI-01, BND-04 | T | I-CONFIG, T-TOGGLE, S-ENV-SCOPE |
| | Resolver → Gateway → OpenRouter adapter; translation; no `AI_GATEWAY_DEFAULT`; key as secret | AI-02 | T | I-GATEWAY |
| | Go owns public streams, cancellation and deadlines, approvals and commit gates | AI-07 | P | Not templated in core; STAGE-AI → G034 (`checklists/foundation-follow-up.md` STAGE-AI) |
| | Envelope; typed events; no browser-to-intelligence path; internal authentication (intelligence side) | AI-03, AI-04 | T | I-AUTH, I-EXECUTE, I-SCHEMA, T-BOUNDARIES-alpha |
| | Authorized callbacks cannot bypass core policy; retrieval filtering; scoped memory; durable usage | AI-06 | P | STAGE-AI → G033/G035 |
| | Usage dimensions; observability privacy; vendors are not the ledger | AI-04, AI-05 | T/P | I-SCHEMA (no tenant or user dimensions in events), I-OBSERVABILITY; durable record → G035 |
| Security | Synthetic data, local credentials, redaction; no secrets in templates, logs, manifests or traces | SAFE-06, OPS-01 | T | S-SECRETS, T-SECRET-CANARY, I-OBSERVABILITY |
| | Development-only fixture identity adapter cannot enable a production bypass | — | F/P | Exists only in `_test.go` files that are never rendered (S-LEAK, structure); guard rule for STAGE-IDENTITY → G004 |
| | Restricted local exposure; authenticated internal contracts; production encryption documented | DKR-03, AI-03 | T/P | T-PORTS-*, I-AUTH, T-BOUNDARIES-alpha; `references/security-operations.md` |
| | Audit, storage authorization, RLS and context checks mandatory even with provider permissions | — | T | `SKILL.md` invariant 16 (S-FRONTMATTER section check) |
| | No product domain in references, templates, examples, generated resources, tool names or prompts | SAFE-07 | T | S-LEAK with the platform denylist (zero hits in the Skill and all four rendered variants) |
| | Major infrastructure deviations require an ADR | — | P (process) | Followed: MinIO is retained as specified, and a replacement is escalated as an ADR proposal, not adopted (§11 item 1; plan R1); AC7 |

## 15. Correction loop 1 (Codex CORRECTION LOOP, 2026-10-04)

**Scope:** only the two acceptance gaps Codex named. No other behaviour, template, fixture or document changed beyond what is listed below. Codex's edit to `management/v0/PROGRESS.md` (the CORRECTION LOOP record) is Codex-owned and was not changed by this work.

### 15.1 Gap 1: `intelligence_enabled: null` was accepted as false

- **Defect:** `generator/input.go` decoded `intelligence_enabled` into a `bool`. `json.Unmarshal("null")` succeeds there and leaves `false`, so null silently meant "disabled".
- **Fix:** decode into a `*bool`. Null or any non-boolean now produces `field "intelligence_enabled" must be a JSON boolean (true or false); null is not allowed`. `validate`, `plan` and `render` all exit 2 before any write, and `true` and `false` behave exactly as before.
- **Regression coverage added:**
  - **`TestIntelligenceFlagNullRejectedTrueFalseAccepted`** (`generator/render_test.go`):
    - null is rejected by `validate`, `plan` and `render` with exit 2 and the explanatory message;
    - a sentinel destination is left byte- and mtime-identical (zero writes);
    - `true` and `false` both parse correctly and `validate`, `plan` and `render` exit 0.
  - **New invalid fixture** `validation/inputs/invalid/null-intelligence-flag.json`, registered in `TestParseInvalidInputs`. G-INVALID now covers 21 invalid inputs through `render.sh render`.
  - **G-INVALID harness extension:** the null fixture is also run through `render.sh validate` and `render.sh plan` (exit 2; destination tree unchanged).

### 15.2 Gap 2: datastore connectivity check could miss an individually reachable endpoint

- **Defect:** in `validation/run.sh` (T-BOUNDARIES), the intelligence-side by-IP probe was `ns_tcp ip 5432 && ns_tcp ip 6379`. An IP reachable on only one of those ports, or on MinIO's 9000 (never probed), passed undetected.
- **Fix:**
  - New helper `datastore_port` (PostgreSQL 5432, Redis 6379, MinIO 9000).
  - New helper `forbidden_datastores_reachable PROBE DEST [SVC…]`, which probes each datastore **independently**, by service name and by **each** container IP, on its own port, and flags any single reachable endpoint.
  - Both the web and intelligence boundary checks now use it. T-BOUNDARIES also adds a positive control: core does reach each datastore, which proves the probe can detect reachability.
- **Negative controls added (new check T-BOUNDARIES-CONTROLS):**
  - **Real reachability.** A throwaway busybox container attached to the trial's `data` network is flagged for **PostgreSQL alone**, **Redis alone** and **MinIO alone**. Logged evidence:
    - `postgres by IP 172.23.0.3:5432`;
    - `redis by IP 172.23.0.2:6379`;
    - `minio by IP 172.23.0.4:9000`;
    - plus each by name.
  - **Independence.** For each datastore in turn, a stubbed probe reaches **only that datastore's IP on its own port**, and the full three-datastore check must still flag it and attribute it correctly. This is exactly the case the old AND condition missed.
- **Catalog:** BND-03 now cites T-BOUNDARIES-CONTROLS (S-TRACE verified).

### 15.3 Files changed in this correction loop

| File | Change |
|---|---|
| `skills/project-archetype/generator/input.go` | Null and non-boolean `intelligence_enabled` rejected |
| `skills/project-archetype/generator/input_test.go` | Expectation for the new invalid fixture |
| `skills/project-archetype/generator/render_test.go` | `TestIntelligenceFlagNullRejectedTrueFalseAccepted` |
| `skills/project-archetype/validation/inputs/invalid/null-intelligence-flag.json` | New invalid fixture |
| `skills/project-archetype/validation/run.sh` | G-INVALID: null through validate and plan. `datastore_port`, `forbidden_datastores_reachable`, rewritten `trials_boundaries`, new `trials_boundary_controls` (T-BOUNDARIES-CONTROLS) |
| `skills/project-archetype/references/requirements-catalog.md` | BND-03 verification cites T-BOUNDARIES-CONTROLS |
| `management/v0/results/G000-project-archetype-result.md` | This section, plus the status note |

### 15.4 Suites re-run and outcomes

- **Run CL-1:**
  ```
  validation/run.sh --denylist management/v0/results/G000-validation-denylist.txt \
    --repo-allow '^(skills/project-archetype/|management/v0/(goals|plans|results)/G000-)' static generator trials
  ```
  - **44 PASS, 1 FAIL.**
  - **Passed:**
    - static: S-GO (includes the new generator test), S-FRONTMATTER, S-LINKS, S-PATHS, S-TRACE, S-SHELL, S-DOCKERFILE, S-COMPOSE, S-FAILCLOSED-ENV, S-ENV-SCOPE, S-NAMING, S-DEPS, S-SECRETS, S-LEAK;
    - generator: G-INVALID (21 inputs plus null through validate and plan), G-SENTINEL, G-REPEAT-CONFLICT, G-REPEAT-CLEAN, G-DETERMINISM, G-OUTPUT-ID, G-REPORT, G-REFUSE, G-TOGGLE-RERENDER;
    - trials: T-PORTABILITY, T-DOCS-EXEC (plus start/stop for alpha and bravo), T-PORTS-CONFIG-alpha/bravo, T-PORTS-alpha/bravo, **T-BOUNDARIES-alpha**, **T-BOUNDARIES-CONTROLS**, **T-BOUNDARIES-bravo**, T-OUTAGE, T-CHECKS-ALPHA/BRAVO, T-PERSIST, T-RENAME, T-TOGGLE, T-DOCKER-ONLY, T-SECRET-CANARY.
  - **Failed:** **T-REPO-CLEAN**. The only unexpected path was `management/v0/PROGRESS.md`, Codex's own CORRECTION LOOP update made outside this session. No G000 work touched it.
- **Run CL-2:** the same command with the allow-list extended to `management/v0/PROGRESS\.md$`, `static` suite.
  - **17 PASS, 0 FAIL:** all 14 S-* checks plus T-DOCKER-ONLY, T-SECRET-CANARY and **T-REPO-CLEAN**.
- **Isolation suite:** not re-run. Neither correction touches `validation/isolation/**` or `validation/compose.yml`, and all datastore connectivity checks live in the trials suite, which was re-run in full.

### 15.5 Cycle usage (plan §6.1)

- **Correction-loop cycles: 1** (main agent). Both fixes passed on their first run.
  - CL-1 failed materially on T-REPO-CLEAN. The diagnosis was that the change was external (Codex's PROGRESS.md edit), not a defect in G000 work. CL-2 retested with the externally owned path disclosed in the allow-list.
  - That diagnose → retest loop is a doubtful case, and §6.1 counts doubtful cases, so it is counted.
- **Updated shared total: 8 of 15** (main 6, subagent 1: 0, subagent 2a: 1, subagent 2b: 1, subagents 3/4/5: 0). Reconciliation: 6 + 0 + 1 + 1 = 8.

**Mandatory checkpoint at cycle 8** (Goal amendment A1):

| Item | Status |
|---|---|
| Cycles per agent | Main 6 · S1 0 · S2a 1 · S2b 1 · read-only agents 0 |
| Suite status | static green (CL-2) · generator green (CL-1) · trials green (CL-1) · isolation green at the original final run (unaffected by the corrections) |
| Remaining work | None for these corrections; awaiting Codex re-review |
| Forecast | 7 cycles of headroom against the 15 ceiling; no risk of exhaustion. The checkpoint shows no stop condition |

### 15.6 Confirmation

- Only the two Codex-named gaps were addressed.
- No unrelated scope changed: no template, fixture schema, policy or MCP code, reference other than the BND-03 catalog row, Goal, Spec, Plan or PROGRESS.md was modified by this correction.
- G000 is not marked accepted, and G001 is untouched.
