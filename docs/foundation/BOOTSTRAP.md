# Bootstrap record

This repository's platform skeleton was originally rendered by a foundation generator Skill (formerly named `project-archetype`; version `0.1.0`). A small project-owned layer was then added. This page separates the two, so later changes stay attributable.

## Provenance

- **Generator retired.** The generator Skill, its output manifest and its render-customization register are retired and are not part of this distribution. The committed repository is the maintained template; nothing is regenerated, and the validation harness has no renderer checks.
- **Identity values.** The product name, domains and database name live only in `.env.example` and the `.env` created from it.
- **Intelligence-disabled variant.** The harness derives it from the repository with the reviewed opt-out edits in [`infra/validation/intelligence-off.tsv`](../../infra/validation/intelligence-off.tsv).

## What was inherited, added and deferred

| | Paths |
|---|---|
| **Rendered, unchanged** | `web/`; `intelligence/`; `contracts/`; `infra/compose/local.yml`; `infra/docker/{core,web,intelligence,minio}.Dockerfile*`; `infra/postgres/`; `infra/minio/`; `infra/security/`; `.env.example`; `.gitignore`; `core/` (except the files below); `docs/foundation/{GETTING-STARTED,VENDORS}.md` |
| **Rendered, deliberately customized** | `core/go.mod` and `core/go.sum` (migration and MCP dependencies); `docs/foundation/FOLLOW-UP.md` (bootstrap items labelled). |
| **Added by the bootstrap** | Migration runner, image and overlay (`core/cmd/migrate`, `core/internal/migrate`, `core/migrations`, `infra/docker/migrate.Dockerfile*`, `infra/compose/migrate.yml`); dormant MCP construction (`core/internal/mcp`); `infra/local/configure-env.sh`; project validation (`infra/validation/`); `README.md`; this page; `MIGRATIONS.md` |
| **Deferred to later foundation stages** | Web shell and design-system initialization; CI, metrics and contract validation; identity and sessions, together with mounting MCP behind authentication; Organization, Group and BusinessEntity schema; shared policy and row-level-security policies; storage authorization; audit; MCP context switching and dynamic tools; domains; onboarding; hot reload. See [FOLLOW-UP.md](FOLLOW-UP.md). |

The archetype's isolation fixtures and development authenticators stayed inside the generator Skill as contract demonstrations and were retired with it. They are not, and must not become, platform runtime code.

## Dormant MCP boundary

`core/internal/mcp.NewServer` builds an official MCP Go SDK (`v1.8.0`) server with an empty tool catalog. Its in-process tests run the `initialize` handshake and `tools/list` over the SDK's in-memory transport.

- **Not called.** Nothing in the core server calls it, so the server binary does not even link the SDK.
- **No exposure.** No transport is mounted: core serves only `/healthz` and `/readyz`, and MCP-style requests return 404.
- **Later stages.** These add authentication first, then build one server per authenticated session context (never a shared, process-global catalog), and mount the transport only together with that authentication.

## Local data safety

- `infra/local/configure-env.sh` never overwrites an existing `.env`.
- The validation harness refuses to touch an existing root `.env` or existing Compose resources for this project. Destructive tests use uniquely named disposable copies and Compose projects. It removes only resources it created.

## Validation

`infra/validation/run.sh` validates the actual repository and disposable copies of it, with host language runtimes shadowed. The suites are:

| Suite | Covers |
|---|---|
| `static` | Shell, Dockerfile and Compose checks; environment scoping; identity-literal confinement; pinned dependencies; MCP source boundary |
| `provenance` | Protected files and unchanged skill roots against the base commit (no renderer checks) |
| `root` | Documented commands, default ports, loopback binding, service boundaries, MCP HTTP probes, checks |
| `migrations` | Integration tests, documented migration commands, catalog and privileges |
| `variants` | Outage and recovery, persistence, rename, database rename, port overrides, the intelligence-disabled variant |

```sh
infra/validation/run.sh --base <commit> --allow-file <scope-allow-list> [--expect-hash PATH=SHA256]... [--exception PATH]... [--keep] [--report-dir DIR] static|provenance|root|migrations|variants|all
```

The harness is a developer check that needs a Git repository: in a copy without Git metadata, initialize a fresh local repository and commit the tree first, then pass that commit as `--base`. `--base` is the commit to compare protected files and the change scope against. `--allow-file` lists the exact paths a change may touch. `--expect-hash` names a file that legitimately differed from `--base` and must still match the given sha256. `--exception` names an exact path changed by its owner outside this change; it is reported, not hidden.
