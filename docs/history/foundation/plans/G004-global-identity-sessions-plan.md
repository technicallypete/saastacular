> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G004 — Global identity and sessions: implementation plan (r5, post-G048 mechanical reissue; WorkOS AuthKit, M2 assurance)

- **Plan revision:** r5, 2026-10-07 UTC. This is a **post-G048 mechanical reissue** of design-ACCEPTED r4 (r4 review (not included)).
  - **Design is unchanged.** Every r5 change maps to RC-1–RC-5, the CI-1 clarification of the repository reconciliation (not included), or the [tenant-redirect addendum](../specs/G004-tenant-redirect-scaling-addendum.md) (see the r5 record in §0).
  - Author: Claude Code (planning only, xhigh effort, no subagents). Reviewer: Codex. Major decisions: product owner.
- **Status: RECONCILED PLAN — submitted for Codex confirmation (E1 completion); NOT implementation-ready.**
  - **Execution blockers:**
    - Codex confirmation of this reissue;
    - escalation items E-1/E-2 (§15);
    - the access matrix (§14: WorkOS items unverified, A4 unmet, GitLab G4b pending E-2);
    - a fresh binding (E4);
    - named external-action authorization (E5);
    - RG (E3).
  - No open architecture decision remains. The owner resolved MFA (M2) in ADR-003's dated clarification.
  - **Nothing performed by planning:** nothing implemented, installed, built, migrated or run in Docker; no WorkOS account, key, callback or user was touched; no GitLab setting, secret or pipeline changed; nothing committed or pushed.
  - **Read-only observations:** the GitLab facts in §14 were observed read-only by the coordinator during preparation, not by this planner.
  - **Preparation update (2026-10-07):** Codex later applied G1, G2 and G4a through glab with the owner's authorization (§14.1). The issuer variable (G4b) is deliberately absent pending E-2. Neither Codex's setup nor the readback satisfies RG-4, E4, E5 or E3.
  - **Zero of seven B7 cycles consumed.**
- **History (preserved, linked):**
  - Plan r1 (not included): `4a525358ae3014e63dcaa88ea174061c9d7206afe74f53eea65fe543d03a21fb`.
  - Plan r2 (not included): `fe147cf004e85899791fa39d0d5dd5200e31fbb2f66fd1ecf8e427bbaa4563c4`. Re-verified before writing.
  - r1 Codex review (not included): five findings, all resolved in r2. Their r2 resolutions are carried forward here.
  - r2 Codex review (not included): owner M2 decision plus four bounded corrections, all resolved in r3.
  - Plan r3 (not included): `e574deeb34451db2eb2b8d555ab4ffa5526fdb54d8299723b6168d3429e1022b`. Exact reviewed bytes, archived and verified before writing r4.
  - r3 Codex review (not included): CORRECTION LOOP with four bounded corrections, resolved in r4.
  - Plan r4 (not included): `b347ad6a702144ff4cb51e86ddea1e435770950a9c7d2f89414fda565e12b26c`. Byte-exact copy (`cp`) of the design-accepted live r4, verified before and after copying.
  - r4 Codex review (not included): **ACCEPT** (design). Execution remains gated.
  - Repository reconciliation 2026-10-07 (not included): RC-1–RC-5 and the CI-1 clarification; authoritative for this reissue.
- **Planning provenance (r5):**
  - Fresh isolated checkout `<home>/.local/state/platform-evidence/g004-prep-7ec9c4d/repo`.
  - Detached at staging `7ec9c4d4ba8004bc7ba59c7b791480e7be2f50bd`, tree `4abae725a601eb9b46f6deae2a3ff471c1074c7b`, clean before the two r5 writes. G048 is accepted and integrated there.
  - The previous planning clone (`g004-planning-cc25a16`) was **not** used as an input.
  - Hashes below are this checkout's actual bytes. This is planning provenance, not an execution binding. The proposed execution binding is in §2.4.

| Governing input (at `7ec9c4d`) | SHA-256 |
|---|---|
| `AGENTS.md` / `CLAUDE.md` | `70b074a00b80de3f8e432a94fda5d1d7663c77cc40cf0ff0b66b774bad1534e8` / `678cc749730747f4a6956547d88389356be48c18cbedc2e8e432d0f1dafb3391` |
| `docs/ARCHITECTURE.md` / `docs/DECISIONS.md` | `403feae3773da49b5e2125e3b08f1b171b71b7c0a73c8d6ab48f280827d20b02` / `8a517ebf53f56dd554b32935df1eede3174d39b5e292493b46d0287e197cef79` |
| `docs/SECURITY.md` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` (empty file) |
| ADR-001 / ADR-002 / ADR-003 | `4fb60db722452042ffad8598370687f1d9e1154375a9cb6cc68f1b82aec70756` / `12329779e4677e49d826befbf9474be9f057f61957f1f4576b177051b5a0de0f` / `11937f24dcf10d53f0f555a21670d6ce1117c6923bf8d7c9096899993820ae4f` |
| `management/BUDGET_POLICY.md` | `b183179b4b34dbe29c024f25ed7cd133c0a51835b1db58381d4509e3bce8f437` |
| ROADMAP / PROGRESS | `33120c3054bb07fd94b824eb9cb1576550261587b6472d896fe0a477b5e9b5cb` / `8d6ec4f0e23a967551971d45b11a0be005735dea44e858ea852b75bb2b875129` |
| Goal r3 / Spec r3 | `1c365a42243636037454a1b1bacb9bb0e70428321261fd082afc0cf1eb1c90ee` / `e7144a7910c4947cf9455383d48fd84dd420985503ef19378df27270eb24f438` (Spec verified as instructed) |
| Tenant-redirect addendum | `cb1080142d1e0f77bec9540e4699c9d4741c2d648b32bfe533738aa0b43a15cb` |
| Plan archives r1 / r2 / r3 / r4 | `4a525358ae3014e63dcaa88ea174061c9d7206afe74f53eea65fe543d03a21fb` / `fe147cf004e85899791fa39d0d5dd5200e31fbb2f66fd1ecf8e427bbaa4563c4` / `e574deeb34451db2eb2b8d555ab4ffa5526fdb54d8299723b6168d3429e1022b` / `b347ad6a702144ff4cb51e86ddea1e435770950a9c7d2f89414fda565e12b26c` |
| Codex reviews r1 / r2 / r3 / r4 | `3954d2cb6b40b4a4375c2ff689fb73047988ca727a448dd72d2f7dc1b37f8fa6` / `7c9403e2d339e872fa9310ead75cf83cd83ceedbabc6792f3870e22a8573940b` / `7f31e6b65158cd505867bde3537df46351e7123dc29835dd3acab9bc3ee66cfc` / `79eb62c8f460be9afb22cbf0f6b8625ad72ff1498c79b83bdb382b88eab84d14` |
| Repository reconciliation | `606a8bf6b2965ecbcfdec7c6fd97166c90bd3364aa30dbb8cc552acbe7a09f89` |
| CI management-docs bypass record | `02cfe03e78d9d79c04c4655e2a396bd6d9adbc8bea34c5f672f83fa8070780fa` |
| G048 implementation review | `e138aa2a11268377777ee29d7e251ace060d14bd03a67eb89ad1e22ce543686c` |

---

## 0. Revision record and correction map

| Rev | Date | Outcome |
|---|---|---|
| r1 | 2026-10-06 | Provider-neutral OIDC/Keycloak recommendation. **ARCHITECTURE DECISION REQUIRED**: provider choice plus five findings. |
| r2 | 2026-10-07 | WorkOS contract, four evidence layers and CI-1. Proposed M1 MFA mapping. **ARCHITECTURE DECISION REQUIRED**: M1 rejected; owner chose M2; four corrections. |
| r3 | 2026-10-07 | Applies M2 (§3.4). Resolves corrections 1–4. Pins exact library/image versions with provenance (§3.7). Separates the read-only start gate from execution capability checks (§12). **CORRECTION LOOP**: four bounded pin, setup and contract corrections. |
| r4 | 2026-10-07 | The four r3-review corrections only (map below). **ACCEPT (design)**; execution gated. |
| **r5** | 2026-10-07 | **Post-G048 mechanical reissue. Design unchanged.** RC-1–RC-5, the CI-1 clarification and the tenant-redirect addendum, against `7ec9c4d` (map below). |

**r5 change record (2026-10-07).**
- Every change below is mechanical.
- No WorkOS/M2, UUID/session, PKCE/cookie/privacy, visual or dependency decision is reopened.
- Facts were re-derived from the actual files at `7ec9c4d`, not copied from summaries.
- Where a fact needs a choice the reconciliation did not settle, it is listed as an escalation item (§15 E-1, E-2) rather than redesigned.

| r5 section | Change | Source |
|---|---|---|
| Header | r5 status, history links (r4 archive, r4 review, reconciliation), fresh-checkout input hashes | RC baseline |
| §0 | This record | — |
| §2.1 | Observed state at `7ec9c4d` (post-G048, CI bypass, nine workflow rules, 211 controls) | RC-3, CI-1 |
| §2.2 | PENDING list replaced by reconciled RC-1–RC-5 facts; addendum boundary | RC-1–RC-5, addendum |
| §2.3 | E1 annotated: G048 part satisfied, RC part = this reissue (pending confirmation) | RC baseline |
| §2.4 (new subsection; numbering otherwise unchanged) | Proposed execution input and scope binding with `7ec9c4d` hashes | RC baseline / E4 input |
| §3.1 | One-line addendum boundary (exactly two explicit test origins; no wildcard, shared callback, cross-host transfer or tenant scope) | Addendum |
| §3.7 | Playwright image pin made exact (was ◆) | RC-2 |
| §5 | Account placement in the existing SiteHeader `action` region; Radix `Button` API; untouched files | RC-1, RC-4 |
| §8.2 | Browser projects and totals rule: six inherited plus two identity projects; inherited-assertion adaptation rule | RC-2 |
| §8.5 | Exact guard additions (B-WEB-BOUNDARY, B-NAMING, B-DEPS, customization register 20 → 23); re-derived counts; **CI-1 workflow rule 10** (protected qualification schedule), attribution and context pass-through; control count 211 → 224 | RC-3, RC-4, CI-1 |
| §9 | AC8 and AC11 rows: ◆ resolved; CI-1 admission evidence | RC-2, CI-1 |
| §10 | Inventory ◆ resolved; exact additions (`validation-tools.Dockerfile`, `infra/ci/validate.sh`, conditional inherited specs, register rows) | RC-1–RC-5, CI-1 |
| §12.2 / §12.3 | CAP-1 Playwright pin; S7 includes the CI-1 admission controls | RC-2, CI-1 |
| §14 | Owner-reported WorkOS setup (UNVERIFIED); read-only GitLab observations; status column | Coordinator addendum |
| §15 | Escalations E-1 (qualification-schedule attribution) and E-2 (issuer-variable sequencing); R-rows refreshed | CI-1, coordinator addendum |
| §16 | r5 local-file research record | — |
| Header, §14, §15 (r5 preparation update, 2026-10-07, after submission) | Status only, no design change. Codex applied GitLab G1, G2 and G4a with owner authorization via glab (setup record (not included), main checkout, SHA-256 `83a11c1e1a3a8c718b13dac661f6f5591a05be38d1318d006b0484e032e63842`; not yet committed). The coordinator read it back independently. The table observed at preparation is kept as history. The r5 bytes as first submitted (`9e2dd47603cb2f31d17654cf8e1503e60248919a4944874387063cd2458f526e`) are preserved outside the repository at `…/g004-prep-7ec9c4d/plan-r5-as-submitted-9e2dd476.md`. | GitLab setup record |

**r4 correction map (r3 review). Changed sections:**
- header and §0;
- §2.3 and §3.7;
- §8.3 and §8.5;
- §9 (AC10/AC11 rows);
- §12.1, §12.2 (CAP-1 row) and §12.3 (S0 and S7 rows);
- §14.1 (A5, A7) and §14.2 (item 4);
- §15 (R-1, R-2, R-3, R-4, R-5) and §16.

No other section changed.

| r3-review correction | Resolution |
|---|---|
| 1 — Test proxy image supports the actual runner | §3.7: the nginx pin is now the verified **OCI image index** `sha256:0985e772…7d94`, which contains linux/amd64 child `sha256:8f84ed99…7433`. The r3 digest `sha256:6b7a4f13…a64d` was a single **linux/386** OCI image manifest (not an index) and is withdrawn. Read-only registry evidence is in §3.7. No pull, build, emulation or other proxy. |
| 2 — Project staging variables scoped on all tiers | §8.5 G4/G6, CI-1 job, forwarding gate and controls. The three **project** variables carry exact environment scope `identity-staging` on every tier, and the job declares that environment on every tier. Unrelated `fast`/`gate` jobs therefore never receive them. The r3 "Free-tier residual" and its owner decision (old R-4) are **removed**. Paid protected environments are optional defence in depth only. |
| 3 — Credentials separated from public trust metadata | §8.3: strict secret classes (API key, bearer, code, verifier, factor and canary secrets) are registered with the existing scanners. Issuer and client ID are server-owned **public protocol metadata**, never registered as secrets, with named permitted channels. Cases P-NEG-1..3 (planted API-key leaks detected) and P-POS-1..2 (intended issuer/client protocol use accepted). No exclusions; no new scanner. |
| 4 — RG self-dependency removed | §2.3 / §12.1: one gate table. RG checks E1, E2, E4 and E5, performs the binding, access, pin and resource checks, and **establishes E3**. CAP placement is unchanged. |

**M2 changes from r2:**
- `mfa_by_provider_policy`, `IDENTITY_MFA_POLICY`, `IDENTITY_MFA_METHODS` and the `{Password}` method allow-list are **removed**. The A3 requirement to disable non-password methods to keep the mapping meaningful is removed with them.
- Retained restrictions, each independently justified:
  - impersonation refusal (vendor-admin acting as a user is not user authentication);
  - all signed identity, client and time checks;
  - PKCE and binding;
  - freshness when configured.

**Correction map (r2 review):**

| Correction | Resolution |
|---|---|
| 1 — PKCE fail-closed; namespace continuity; same-provider-session callback | §3.2 and CAP-4a: a real wrong-verifier exchange accepted by WorkOS means **STOP**, with no fallback. §3.6 / §6: single-row `identity_provider_binding(issuer, client_id)` guard. §4.2: identical-versus-additional `sid` effects table. Scenarios E-PKCE, I-NS-1..3, E-CB-1/2/9/10. |
| 2 — Coherent route, logout and traffic contracts | §7.3: form content type only on web form routes; JSON only on core mutations. §3.5 / §4.4: narrow `RevokeSession` wording; a fixed sequence, verified by CAP-4b with a stop on gap; no redirect-only fallback. §8.4: context-level listener plus an additional-page case; a separate driver for core probes. Scenarios E-CT-1..4, E-LO-5..8, E-NET-1..3, MQ-6. |
| 3 — GitLab secret boundary and profile resource budget | §8.5: named settings G1–G6 verified and recorded; forwarding gated on project, ref, protection, source, job and environment, never on a user-supplied profile; existing-mechanism profile budget. *r3's tier-dependent environment scoping and Free-tier residual are superseded by r4 correction 2: exact project-variable scope and the job environment on all tiers; controls C-SEC-1..10.* |
| 4 — Exact pins; read-only versus execution gates | §3.7 exact versions, `go.sum` hashes and image digests. §12: RG (read-only start gate) versus CAP-1..4 (execution capability checks, each with a stop placed after its prerequisites). A8 production terms are outside G004. V-IDs are reconciled into CAP-IDs. |

The r1-review resolutions in r2 are retained unchanged unless noted:
- server-stored per-session CSRF token;
- stable per-browser binding cookie;
- binding-matched consumption;
- repeat-logout contract;
- privacy classification;
- the email-reassignment real test;
- explicit catalog negatives.

## 1. Approach summary

1. **Core-owned identity.** Core owns:
   - canonical user UUIDs;
   - the exact (issuer, subject) principal mapping, inside **one bound provider namespace** (issuer and client);
   - PostgreSQL-only opaque sessions, revocation and local disable.

   There is no session cache, and Redis is unused for identity.
2. **WorkOS adapter.** `core/internal/identity/workos` runs the hosted AuthKit authorization-code flow with browser-bound one-use `state` and **mandatory PKCE S256**, then the core-side `authenticate` exchange. It validates the signed access token's:
   - exact issuer;
   - client-specific JWKS;
   - `client_id`;
   - `sub` equal to `user.id`;
   - `sid`;
   - `exp`/`iat`;
   - required `auth_time`.

   It fabricates no discovery, ID token, nonce, `aud`, `acr` or RP-logout parameters. Provider tokens are discarded. Only the non-bearer `sid` is kept.
3. **Assurance (M2).** Every session's MFA assurance is **`unknown`**. No configuration, method, test result, claim or browser input can change it.
   - The owned boundary exposes a fail-closed `RequireVerifiedMFA` guard. It denies every G004 session because no reviewed session-specific proof source exists.
   - There is no product operation, step-up route or factor UI.
   - Real provider MFA challenge success and cancellation, SSO reuse and signed `auth_time` forced reauthentication remain required evidence.
4. **No vendor authority.** WorkOS Organizations, roles, permissions, entitlements and feature flags confer nothing. Impersonation is refused.
5. **Evidence layers.** These are unchanged in structure from r2:
   - **L1** Go fixtures;
   - **L2** pinned official emulator plus real Chromium over test HTTPS;
   - **L3** real staging API in the protected CI-1 job;
   - **L4** manual hosted qualification with server-side observation.

   Fixtures and the emulator never substitute for real authentication, MFA, SSO or freshness evidence.
6. **CI.** The accepted CI is reused. The only increment is **CI-1**: one protected `identity-staging` job and a `staging` profile inside G004, per the r2 review. There is no separate CI Goal or framework.

## 2. Baseline and gates

### 2.1 Observed state

- **r5 facts at `7ec9c4d`.** Each was verified against the file named.
  - **Core:**
    - only `GET /healthz` and `GET /readyz` are mounted (`core/cmd/server/main.go`);
    - the migration set is solely `core/migrations/00001_baseline.sql` (version 1).
  - **Web:**
    - routes are exactly `/`, `/api/health` and `/api/ready`;
    - `cookies()`, `document.cookie` and `NEXT_PUBLIC_` are banned by B-WEB-BOUNDARY;
    - Radix Dialog/Slot shell (G048) with bundled fonts and blueprint assets.
  - **Configuration:** `.env.example` still holds the historical `IDENTITY_PROVIDER=oidc` placeholder and is not in the customization register.
  - **CI (`.gitlab-ci.yml`):**
    - **nine workflow rules**: open-MR feature-push suppression; MR and push changed-path admission with the management-Markdown omission (two `when: never` companions); default-branch schedule; web; api; final never;
    - two jobs, `fast` and `gate`; no CI/CD secret.
  - **CI controls:** **211** in `infra/ci/validate-controls.sh` (`CONTROLS`): A 25, B 34, C 8, D 23, E 22, F 77, G 22.
  - **Attribution (`infra/ci/validate.sh` `classify`):** refuses a schedule outside the default branch (control N12). API branch pipelines are attributed `branch-api` and select full (`select-profile.sh`).
  - **Security doc:** `docs/SECURITY.md` is empty.
- **Remote main.** Observed read-only during preparation at `952e0deda5570126d52c3a45b670e795dca4cece` (§14). This Plan does not assume its bytes; E4 binds a verified committed source.

### 2.2 Post-G048 reconciliation — RESOLVED in r5 (RC-1–RC-5; per the repository reconciliation)

| RC | Reconciled fact at `7ec9c4d` | r5 consequence |
|---|---|---|
| RC-1 | `SiteHeader` already accepts `action?: React.ReactNode` (`web/components/ui/site-header.tsx`). `AppShell` passes a flex-wrapped `action` div holding `AppearanceControl` and `ShellHelpDialog`. Radix-based `Button` (`web/components/ui/button.tsx`) supports `asChild`, variants `default/secondary/outline/muted/destructive/ghost/link`, sizes `sm/default/lg/icon` and `block`. | §5: one bounded, additive AppShell change places the Account link inside the existing `action` div. No new SiteHeader API; no Base UI assumption. `page.tsx` and `globals.css` untouched. Fonts, blueprint, scoped portals, tokens, first paint and hydration preserved. |
| RC-2 | Six Playwright projects (`web/playwright.config.ts`): `shell-light`, `shell-dark`, `tokens-light`, `tokens-dark`, `appearance`, `dev-hydration`. Shell `testMatch` = `(shell\|overlay\|accessibility\|boundary\|design)`. `@playwright/test` `1.63.0`. Browser image `mcr.microsoft.com/playwright:v1.63.0-noble@sha256:eff16c30e6f3f4af0a03fa4b706120d5e9b0891c344a27d64559aff5900a4a27` (`infra/docker/web-browser.Dockerfile`). Accepted inherited reports: root 133, disabled 133, rename 2 expected; zero skipped/unexpected/flaky. | §8.2: the six projects are inherited unchanged. `identity-light`/`identity-dark` are explicit separate projects on the disposable HTTPS environment. Shell `testMatch` additively gains `account`. New totals are derived from actually collected tests; 133/133/2 is scenario evidence, not a frozen total. Only genuinely affected inherited assertions are adapted. |
| RC-3 | Registry 71 rows (`required-checks.tsv`): lint 16, controls 3, provenance 5, contracts 6, unit 1, root 16, migrations 4, ops 7, rename 5, off 4, final 4. Profiles (`infra/ci/lib.sh`): fast 32, integration 64, full 70. Customization register 20 rows (`render-customizations.tsv`, including `infra/docker/web.Dockerfile`). CI: nine rules, 211 controls, API → full, management-Markdown bypass. | §8.5: re-derived counts (79 / 33 / 70 / 77 / 5). Register 20 → 23 (exact rows). CI-1 carried onto nine rules / 211 controls, adding exactly one workflow rule (10) and 13 controls (224). No regression to the 7-rule or 203-control snapshots; the bypass is kept. |
| RC-4 | Web `package.json`: **20** exact direct dependencies (10 runtime, 10 dev), including `@radix-ui/react-dialog 1.1.23`, `@radix-ui/react-slot 1.3.3`, `next 16.3.8`, `react`/`react-dom 19.3.0`, `@playwright/test 1.63.0`, `@axe-core/playwright 4.13.0`. B-NAMING allows bounded `web/public` and forbids `web/src` and `web/components/domain`. B-WEB-BOUNDARY route/import/header allowlists as quoted in §8.5. | G004 adds **no** web dependency. §8.5 lists the exact B-WEB-BOUNDARY and B-NAMING additions (no broad exemptions). |
| RC-5 | `docs/foundation/WEB-SHELL.md` carries the G048 source-lineage, token-role, theme, scoped-portal, font/asset, blueprint and deviation records. | WEB-SHELL.md gains only an "Account and sign-in (G004)" subsection (account control placement, BFF boundary, new browser projects). All G048 records are preserved. IDENTITY.md and the VENDORS.md WorkOS rows are added as planned. |

**Tenant-redirect addendum (referenced, not implemented).** G004 keeps exactly its two explicit HTTPS test origins, exact callback and sign-out allow-lists, host-only cookies, browser-bound state and PKCE. It adds **no** wildcard origin or redirect, shared callback, cross-host session transfer, tenant provisioning or tenant scope. Those requirements belong to G016 (and Organizations to G005) under their own future review.

### 2.3 Gate table (single source) — r4 correction 4; r5 status annotation

**r5 status:**
- **E1:** G048 ACCEPT and integration are **satisfied** (PROGRESS; staging `7ec9c4d`). The RC part is this reissue, pending Codex confirmation.
- **E2, E4, E5 and E3:** unmet.

Gates are listed in dependency order. **E3 is established by RG; it is not an input to RG.** No gate depends on itself or on a later gate. Execution-internal CAP checks are listed in §12.2, with unchanged placement.

| Gate | Requirement | Established by | Depends on |
|---|---|---|---|
| E1 | G048 ACCEPT and integration, then mechanical RC-1–RC-5 reconciliation in the reissued Plan | Codex | — |
| E2 | Codex ACCEPT of the reissued (post-E1) Plan | Codex | E1 |
| E4 | Fresh committed base and supervisor snapshot (`infra/validation/run-approved.sh` inputs), in a fresh isolated worktree | Codex | E2 |
| E5 | Named external actions authorized (§14.2) | Owner | E2 |
| **E3** | **RG passes** (§12.1): confirms that E1, E2, E4 and E5 are recorded, then performs the binding, resource, pin-provenance and access-matrix checks | **RG** (Claude, read-only, at S0) | E1, E2, E4, E5 |
| — | S1 onward, with CAP-1..4 at their §12.2 placements | Approved execution | E3 |

### 2.4 Proposed execution input and scope binding (r5; hashes at `7ec9c4d`; for Codex to confirm at E4)

These hashes come from the clean checkout. E4 re-binds a fresh committed base: any later change to a listed path requires a re-hash, and an unexplained difference stops at RG-1.

**Inputs (bound, not modified by G004):**
- the governing inputs in the header table;
- this r5 Plan (its hash as submitted);
- the r4 archive.

**Scope baseline — existing paths G004 modifies:**

| Path | SHA-256 at `7ec9c4d` | Change owner (RC / CI-1) |
|---|---|---|
| `.gitlab-ci.yml` | `9ab9378e67b6344f9a65e8a4b4a0c53e6adf5265a0088b2303cc3f3e15783c85` | CI-1 |
| `.env.example` | `314e1e9deff18d3c693eb7b7201870093b9c6964eacd84ccbfa94ee6809f927b` | §4.6 (registered) |
| `core/cmd/server/main.go` | `987c933560ef77ea9a2daa4a3bec4c9747ea69d578f3c9c9d8e6179d013551ea` | §4 |
| `core/go.mod` / `core/go.sum` | `b14ff8c01bbfd56e30f0ee688db2efe88ad1a1dbe0af366d7f84053e7cc15bb7` / `bea96f31ccb6b068b751c26f879608cb209768372d8029556b835f3d8a37937d` | §3.7 |
| `core/internal/config/config.go` | `f4822d416b428ed5a11dbc2c73858ee73801797ad74268219d1fe654e7fc8ca4` | §4.6 (registered, RC-3). The rendered `config_test.go` (`52f35c453b0191f1438d7704a978de5f3e432202a0908f9248e5738595926372`) stays **unchanged**; identity configuration tests go in a new file. |
| `core/internal/migrate/integration_test.go` / `runner_test.go` | `93cea9366a0818bdce2d66f288eb71c6fded4ed590ea68d9bc6bad4f4c78ed56` / `d1f070dffdb487903f0df17e1fade4cc14b15f7ccab13e8f616c8d2e92a604fa` | RC-3 (version 2) |
| `core/migrations/migrations_test.go` | `3c96c2321451e097a5ffeef23cac3c3bddfbe8e8f2ef0b92371e51083078bf25` | RC-3 |
| `web/components/layout/app-shell.tsx` | `ae24bf7518cf3bdeab481150fc534e4847cfc7cbc41a0f93b2c1b77daaa10b05` | RC-1 |
| `web/lib/request-context.ts` | `35fe370bc3d45bc39830bfbd4d2a480f3a3a870e1510079322b178b21fad251c` | §4.7 |
| `web/playwright.config.ts` | `b8b44afb597671c84f55079fb2136d152b135b416de2277ecbfc4c954cf04e37` | RC-2 |
| `web/tests/browser/shell.spec.ts` / `accessibility.spec.ts` / `design.spec.ts` (**conditional**: only if an assertion is genuinely affected by the Account control) | `7a9cdd9f097d71279553f99d0c183b60d4187cedce4fa9e853bd869ce4ca1b6d` / `f1f4a361401c1ecfc6766bf74aa2884b654afac34dfb396cfb39a4902ce37b6f` / `bd00b7059d0f22dec2a58968d8f03f2bbadd7244de55e0420140cf368b69e5f3` | RC-2 |
| `infra/validation/run.sh` / `lib.sh` | `476d7d67a0c0a240e25ab586f1a58c104c7d911fb6230f6423a6e292e9dc9dd9` / `044616d6e863765c3232d22ee766db6d258677236e6635edc068876ee34e7e3a` | RC-3, RC-4, CI-1 |
| `infra/validation/required-checks.tsv` / `render-customizations.tsv` | `2f86ac024750a966b32a6371e2c1f32a4fb689268e78af419f6c16f9ddf7e6a6` / `1e454dacc93b61b5462624e95ad918b81bf5e21029b11c5d064aba08ef6e62c9` | RC-3 |
| `infra/validation/tools/internal/logrecord/events.json` | `40e39b0e92621feb830aa0d128dd473b9fd15f3ed866b189bc3f903a8404a908` | §4.7 |
| `infra/docker/validation-tools.Dockerfile` | `86da5cc0a0136bc40d7b2838435d4a22bfd397891a109b922b1ceee1ef12cbd5` | RC-3: explicit `./cmd/...` build list gains `identitysetup` |
| `infra/ci/lib.sh` / `hosted-context.sh` / `validate.sh` / `validate-controls.sh` | `6d550a4af07b1e50356f485c9df37d36cb3e24f31fd770641284fb165c3c392f` / `7bb493408aa8bec0581fd83e8bfe8b3eaaf70b04d14ba6fa4648addbf364c157` / `0597b8552b5eb5f81ca1d0347fb0c6f22346d4f32a9c2ee5bc90f6f9ab9b3319` / `3c386d50793c08ddbbe6e602eab5eab134693ef8b038af1a79fe47106eb77aa9` | CI-1 (`validate.sh` per E-1) |
| `docs/foundation/MIGRATIONS.md` / `OPERATIONS.md` / `FOLLOW-UP.md` | `0f65dc2382f77f3bf2558fe12482276900a6e5ca0d64ed9f078505926279f657` / `d2622ad0f2d1b4937c85cea3e446c0130e09a56fb037a1c55dac611576f5050b` / `0373d3f84c5dbcbaeac181745ee46a13c99027f904bec254c442a5a4b67303f4` | RC-3, CI-1 |
| `docs/foundation/VENDORS.md` / `WEB-SHELL.md` | `6449303d06d168b15dfa47d9efe2659ea8db918edee7efb7ea57e626ad390207` / `d8911cccb110396f49d1e15602969aceb3b7ada0a36d5226401efe83c87d765b` | RC-5 (VENDORS registered) |

**Bound unchanged** (must stay byte-identical):

| Path | SHA-256 |
|---|---|
| `core/migrations/00001_baseline.sql` | `db18518a954290c71b607eb9413bd2d66b9d4bf78645ad3fb2d1057f2883c60f` |
| `web/app/page.tsx` | `03a939f60e2c982a868bddc62b29b6f906b88336e207b76e6cece557a5c1faf8` |
| `web/app/globals.css` | `513ab58838783106f55b26ad7508973a47020890e144c2bbe666ce5b61efcfdf` |
| `web/components/ui/site-header.tsx` | `4e1470f07db94f14a42cd0f1f11d85cf32de38fb7f630b3a66e68bbbc82dbda0` |
| `web/components/ui/button.tsx` | `e646c174fd69c246b6127b62b7cddc88d61588511f5d2d3bac84dc1f27685813` |
| `web/package.json` | `3e0c86502d40984959859d0703fad038ea0c4b196f7812451ab38b00630ff933` |
| `web/package-lock.json` | `8abae14eeb8514c01a7e2765cee15a86cdf1da7826a5426782cc7337a6446939` |
| `infra/docker/web-browser.Dockerfile` (unchanged unless the conditional browser-trust fallback of r2 §10 / r1 §7.2 is separately approved) | `26de9962a6690e616342a351f7fdd011d56b4810d76900c2fd2be6d613de68b1` |
| `infra/validation/run-approved.sh` (wrapper) | `620a3e84a68494f9eca9e1bc2ca92a5bf057ebb316c44a31eefd837b4c53c695` |
| `infra/ci/select-profile.sh` | `7d7cbb7b9f7ea094cc1cd00f0e4d97fe4bd2efca3b4bd08b0699fec5fc406ee6` |
| `infra/ci/profile-paths.tsv` (`core/` → runtime; `core/migrations/`, `infra/` → full already cover G004 paths) | `4bae75a0f89d72b5ded2b7d7fc602163afcbbf993170c93ca4bffa2adca32bc6` |

**New paths:** exactly the §10 inventory. The execution scope allow-list (the first execution write) is generated from §10 plus the table above, with anchored exact paths only.

## 3. WorkOS AuthKit contract

**Fact markers:**
- **[D]** documented publicly (§16);
- **[C]** confirmed by a named capability check during execution (§12.2), with a **stop** on mismatch;
- **[P]** a Plan decision.

### 3.1 Calls (core only; base URL configured)

| Purpose | Call |
|---|---|
| Authorize (browser redirect) | `GET {base}/user_management/authorize`: `client_id`, `redirect_uri` (configured callback for the exact matched origin), `response_type=code`, `provider=authkit`, `state`, `code_challenge`, `code_challenge_method=S256`, and `max_age` only when freshness is configured [D]. Never `organization_id`, `connection_id`, `screen_hint`, `login_hint` or `prompt` [P]. |
| Exchange | `POST {base}/user_management/authenticate`: `client_id`, `client_secret` (API key), `grant_type=authorization_code`, `code`, `code_verifier` [D]. SDK v10 `UserManagement().AuthenticateWithCode(ctx, params) (*AuthenticateResponse, error)` [D, signature confirmed by compile at S3]. |
| Keys | `GET {base}/sso/jwks/{client_id}` [D]. Built only from the configured base and client ID. |
| Provider session revoke | `POST {base}/user_management/sessions/revoke` `{session_id}` [D]. |
| Hosted browser logout | `GET {base}/user_management/sessions/logout?session_id=&return_to=` [D]. No confirmation or state field [D]. |

**r5 (addendum):** G004 uses exactly the two configured HTTPS test origins and their exact callback and sign-out pairs. It uses no wildcard redirect, shared callback, cross-host session transfer or tenant scope (§2.2).

### 3.2 Browser binding and PKCE (R3) — correction 1

- **Randomness:** `state` and the verifier each carry 256 bits from `crypto/rand`. Challenge = `BASE64URL(SHA256(verifier))`. The transaction stores `SHA256(state)` (PK), `SHA256(binding)`, the verifier (≤ 10 min, server-only), `client_id`, origin, return path, `max_age` and expiry.
- **Mandatory PKCE.**
  - The exchange always sends the transaction's own verifier. The application has no code path that exchanges a code without the verifier from the binding-matched transaction that issued its challenge.
  - That WorkOS rejects a **mismatched verifier while a client secret is also sent** is a provider property that public docs do not settle. It is verified by **CAP-4a** with a real code. **If WorkOS accepts the wrong verifier: STOP and return the capability gap.** There is no defense-in-depth reclassification and no switch to a different client-authentication mode without review.
- **Code injection is the threat PKCE closes here.** An attacker injects a victim's code into the attacker's own transaction, so it is exchanged with the attacker's verifier. L1/L2 prove the application consumes the attacker's transaction and issues **no session** when the provider rejects the exchange. CAP-4a proves the real provider rejects it.

### 3.3 Exchange and token validation contract — all must pass before mapping

| Check | Rule |
|---|---|
| Provenance | Only the `authenticate` response, from the configured base over TLS with the API key, for this transaction's code, verifier and client. No browser-supplied token or assertion is ever accepted. |
| Token kind | Only the response's `access_token` is verified. `refresh_token` and `oauth_tokens` are discarded unread. Any ID-token-like field is ignored. |
| Signature | JWS via the cached client JWKS. `alg` ∈ `IDENTITY_JWT_ALGORITHMS` (provisional `RS256`, confirmed CAP-3). `none`/HMAC are never accepted. Unknown `kid` → §3.6. |
| Issuer | `iss` must equal `IDENTITY_WORKOS_ISSUER` exactly. The configured value is copied from a decoded real staging token at CAP-3. Docs show `https://api.workos.com` with trailing-slash variance [D]. |
| Client | The `client_id` claim must equal the configured client [D]. No `aud` (none by default [D]); no JWT template relied on. |
| Subject | `sub` must equal the response `user.id`. Both match the `user_` prefix and the pinned charset, length 1–255 (CAP-3). |
| Session reference | `sid` is required and stored as `provider_session_id` (non-bearer). |
| Time | `exp` > now − 30 s; `iat` ≤ now + 30 s; `nbf` if present. `auth_time` is **required** [D] and ≤ now + 30 s. If `IDENTITY_MAX_AUTH_AGE` is set: now − `auth_time` ≤ max + 30 s, else `stale_authentication`. |
| Impersonation | `act` present, a non-null `impersonator`, or `authentication_method="Impersonation"` → `impersonation_refused`. This is the **only** use of `authentication_method`, and it is never persisted or mapped. |
| Vendor claims | `org_id`, `role(s)`, `permissions`, `entitlements`, `feature_flags` and `organization_id` are ignored: never persisted, logged or returned. |
| Attributes | `display_name` = trimmed `first_name` + `last_name` (≤ 200), else null. `email` and `email_verified` are dropped at the adapter boundary. |

### 3.4 Assurance under M2 (R7/AC7; ADR-003 clarification)

- **Session MFA assurance is always `unknown`.** It is stored as `sessions.mfa_assurance` with `CHECK (mfa_assurance = 'unknown')`, so any future proof source needs a reviewed migration and Plan. It is returned as `"unknown"` and displayed as "Multi-factor status: not verified by this application".
- **Nothing can upgrade it:**
  - WorkOS environment MFA configuration;
  - `authentication_method`;
  - L3 or periodic qualification results;
  - token or response claims (`amr`, `acr`, `mfa`, or any custom claim);
  - browser parameters or headers;
  - configuration variables (none exist for MFA).
- **Fail-closed requirement.** `identity.RequireVerifiedMFA(SessionInfo) error` is the owned guard for any request or test profile that requires verified MFA. With no reviewed proof source it **always** returns `mfa_unverified` (HTTP 403 when wrapped by `identity.RequireVerifiedMFAHandler`).
  - No G004 route, configuration flag or test profile enables a verified-MFA requirement in production code.
  - Tests exercise the guard and the handler wrapper directly.
  - There is no step-up route, factor UI or product operation.
- **Evidence kept separate from session data.**
  - Real MFA challenge success and cancellation: MQ-1/MQ-2.
  - Provider MFA enforcement for the tested configuration: S-MFA.
  - SSO reuse: MQ-3.
  - Forced reauthentication with signed `auth_time`: MQ-5 plus U-TOK.

  Qualification records the *tested configuration and time*. It is not continuous attestation and is never copied into session rows.

### 3.5 Provider sign-out and revocation (R6) — correction 2

- **Narrow meaning.** A successful `RevokeSession(sid)` proves only that **WorkOS revoked that one session (`sid`)**. It does not prove provider-wide sign-out, other applications' sessions, or the clearing of the browser's hosted SSO state. UI, docs and evidence use exactly this wording.
- **Fixed sequence** (no automatic alternative):
  1. Commit the local revocation.
  2. `RevokeSession(sid)` (5 s).
  3. **Only if confirmed:** 303 to the WorkOS logout URL (`session_id=sid`, `return_to=<origin>/sign-in?logout_ref=<ref>`) so the hosted browser session is ended.
  4. On a revoke failure or timeout: no WorkOS redirect. 303 to `/sign-in?logout_ref=<ref>` with the recorded outcome `provider_revoke_failed`.
- **CAP-4b** must show, on real staging:
  - the logout URL accepts an already-revoked `sid` and returns to `return_to`;
  - the **next sign-in requires credentials** (MQ-6).

  **If the sequence cannot establish this: STOP and return the capability gap.** There is no redirect-only or reordered fallback without Codex review.
- **Deprovisioning:** remote deprovisioning is not claimed. Local disable is the trusted path.

### 3.6 Trust, network bounds and namespace continuity (R1/R2) — correction 1

- **API base:** default `https://api.workos.com`. Any non-default value is accepted only with `APP_ENV=test` (the emulator). Production refuses a non-default or non-`https` base.
- **Endpoints:** fixed relative paths only. No token, response, `Host` or forwarding header chooses an endpoint.
- **`http.Client`:**
  - no redirects;
  - deadlines: authenticate 8 s; JWKS, revoke and others 5 s; whole callback ≤ 10 s; web→core 10 s;
  - body limits: authenticate 64 KiB, JWKS 256 KiB.
- **JWKS (keyfunc/jwkset):**
  - refresh every 1 h;
  - unknown `kid` → at most one refresh per 30 s, de-duplicated;
  - unknown keys fail until a trusted refresh succeeds;
  - outages → `identity_unavailable`, never a signature failure.
- **Namespace binding.** Table `identity_provider_binding` holds one row: `provider='workos'` PK, `issuer`, `client_id`, `bound_at`.
  - **First mapping:** `INSERT … ON CONFLICT DO NOTHING`, then re-read in the same transaction.
  - **Every login start and callback:** the configured (`IDENTITY_WORKOS_ISSUER`, `IDENTITY_WORKOS_CLIENT_ID`) must equal the bound row. A WorkOS client ID identifies its environment [D], so a client change is an environment change.
  - **Startup:** a mismatch is logged once (`identity.namespace_mismatch`, no values).
  - **Mismatch effects** (issuer change, client/environment change, or both):
    - provider operations return 503 `identity_namespace_mismatch`;
    - **no mapping is minted, no session is issued and no transaction is consumed**;
    - existing local sessions are untouched (they remain valid local identities);
    - existing UUIDs are unchanged;
    - there is never email linking.
  - **Re-binding** (custom domain, environment migration) requires a separately reviewed migration. G004 provides none.

### 3.7 Exact pins and provenance (correction 4)

These were obtained by bounded read-only research on 2026-10-07.
- **RG** re-confirms them read-only: `sum.golang.org` and registry digest lookups.
- **CAP-1** reproduces them during execution.
- Any difference, or any extra module beyond this set and the existing `go.sum`, is a **STOP for review**.

| Module | Version | go.sum `h1:` (module / `go.mod`) | Notes |
|---|---|---|---|
| `github.com/workos/workos-go/v10` | **v10.5.0** (2026-09-18) | `eClqevZs8AihL0kQOuX1Zzi2WT+8PYrYAuXkGszQdN4=` / `/X1d193TBGdVfyKAdueguocR+a/V8ebuXdiKDoYQijg=` | MIT; generated client; `go 1.23`; `NewClient(apiKey, …ClientOption)`, `WithBaseURL(url)` [D] |
| `github.com/google/go-querystring` | v1.2.0 | `yhqkPbu2/OH+V9BfpCVPZkNmUXhb2gBxJArfhIxNtP0=` / `8IFJqpSRITyJ8QhQ13bmbeMBDfmeEJZD5A0egEOmkqU=` | BSD-3; required by workos-go |
| `github.com/google/uuid` | v1.6.0 | already in `core/go.sum` (lines 34–35; identical) | Required by workos-go |
| `github.com/stretchr/testify` | v1.12.1 | already in `core/go.sum` | workos-go module-graph requirement only; not linked into the server |
| `github.com/golang-jwt/jwt/v5` | **v5.3.1** (2026-01-28) | `kYf81DTWFe7t+1VvL7eS+jKFVWaUnK9cB1qbwn63YCY=` / `fxCRLWMO43lRc8nhHWY6LGqRcf+1gQWArsqaEUEa5bE=` | MIT |
| `github.com/MicahParks/keyfunc/v3` | **v3.8.2** (2026-09-04) | `eydEwk/pBAVrDIpmFfB/gkCcrp++xQ7YYXirrI2zlWE=` / `T4snFPe26GwMg45bBAdM5P6qWQyLxZHLwBhxR/9PnCs=` | `go 1.25.0`; requires jwkset v0.11.3, jwt v5.3.1, x/time v0.15.0; retractions predate v3.8.2 |
| `github.com/MicahParks/jwkset` | **v0.11.3** (2026-08-01) | `Phli4RdTDdIdLXZpuO7abkwZyzIk0RDTUPVVBHPRdkQ=` / `U2oRhRaLgDCLjtpGL2GseNKGmZtLs/3O7p+OZaL5vo0=` | — |
| `golang.org/x/time` | v0.15.0 | already in `core/go.sum` (lines 82–83) | — |

| Image (test-only) | Pin | Notes |
|---|---|---|
| WorkOS emulator | `ghcr.io/workos/emulate:0.14.0@sha256:767021483ed125614a1a489329477e60601a2b248eedc0ac74b1e571604319ed` | Release 2026-09-24; MIT. Digest as displayed for the tag (multi-arch index); RG confirms the index digest. Never in production Compose. |
| TLS test proxy (r4) | `nginx:1.30.5-alpine3.24@sha256:0985e772fb9f729e6fa0980da05fca5d9c468e870eed43071545afa9d2e27d94` | **OCI image index** (`application/vnd.oci.image.index.v1+json`), multi-platform. Its **linux/amd64** child is `sha256:8f84ed99befc3891b8f329c5c202785278a2cfb7c25107d57fb2a134a3117433` (OCI image manifest), which is the platform the qualified `saas-linux-medium-amd64` runner and amd64 developer hosts resolve. BSD-2. Test overlay only. |
| Reused unchanged | `golang:1.27.1-alpine`, `node:24-slim`, Playwright `mcr.microsoft.com/playwright:v1.63.0-noble@sha256:eff16c30e6f3f4af0a03fa4b706120d5e9b0891c344a27d64559aff5900a4a27` (r5: exact, RC-2), `pgvector/pgvector:0.8.7-pg17`, `redis:8.10.2-alpine`, `busybox:1.37.0`, `docker:29.8.2` CLI/DinD | Existing accepted pins. |

**Withdrawn r3 proxy pin.** r3 pinned `sha256:6b7a4f13aa1cb05946e90cfa70933418e51f99a9604305eec0779737077fa64d`. It is a single OCI image manifest (not an index) whose config is **linux/386**. It is unusable on the amd64 runner and must not be used. No emulation or alternative proxy is introduced.

**Read-only registry evidence for the proxy pin** (2026-10-07). Method: anonymous pull-scope token from `auth.docker.io`, then HTTP GET/HEAD against `registry-1.docker.io/v2/library/nginx`. Response bodies were hashed in memory only. Nothing was pulled, built, run or written to disk.

| Request | Response | Body SHA-256 (= content address) | Finding |
|---|---|---|---|
| `HEAD manifests/1.30.5-alpine3.24` | `Content-Type: application/vnd.oci.image.index.v1+json`; `Docker-Content-Digest: sha256:0985e772…7d94` | — | The tag currently resolves to the pinned index |
| `GET manifests/sha256:0985e772…7d94` | 10,317 bytes | `0985e772fb9f729e6fa0980da05fca5d9c468e870eed43071545afa9d2e27d94` | Index includes `{digest: sha256:8f84ed99…7433, mediaType: application/vnd.oci.image.manifest.v1+json, platform: {architecture: amd64, os: linux}, size: 2493}`. It also has arm/v6, arm/v7, arm64 and attestation entries. |
| `GET manifests/sha256:8f84ed99…7433` | 2,493 bytes, OCI image manifest | `8f84ed99befc3891b8f329c5c202785278a2cfb7c25107d57fb2a134a3117433` | Config `sha256:43d9d8c1f8968f09df8c1aa6c136ecc617e62d64fe4c0b24c97de4eb210bd973` |
| `GET blobs/sha256:43d9d8c1…d973` | 12,296 bytes | `43d9d8c1f8968f09df8c1aa6c136ecc617e62d64fe4c0b24c97de4eb210bd973` | `"architecture":"amd64"`, `"os":"linux"` |
| `GET manifests/sha256:6b7a4f13…a64d` (withdrawn r3 pin) | OCI image manifest | `6b7a4f13aa1cb05946e90cfa70933418e51f99a9604305eec0779737077fa64d` | Config `sha256:15f71c3ef18844132d55b1f8cf6f4410508f3d66bc8ac859035d303bee11a784` (12,290 bytes): `"architecture":"386"`, `"os":"linux"`. Confirms the r3 review. |

**Checks at RG and CAP-1:**
- RG repeats the index lookup and amd64-membership/config check read-only.
- CAP-1 additionally requires the platform manifest resolved for the pinned index on the runner to equal `sha256:8f84ed99…7433`, and the config architecture to be `amd64`.
- Any difference → STOP.

- **Licences** for keyfunc and jwkset (expected Apache-2.0) and go-querystring are confirmed at RG and recorded in VENDORS.md.
- **No new npm dependency.** There is no Next auth installer or AuthKit Next SDK.
- **Provenance.**
  - The Go `h1:` values and the emulator digest were read via text extraction in r3. Codex then independently confirmed them against `sum.golang.org` and the emulator registry response (r3 review).
  - The proxy values above come from direct registry responses hashed locally.
  - RG re-reads every value before binding.

### 3.8 Emulator limits (official `SUPPORTED.md`, as read; re-read for 0.14.0 at CAP-2)

- **Supported:** code exchange with JWKS-signed tokens, sessions and revocation listing, seeded users and organizations, interactive login, YAML error hooks.
- **Not enforced or not confirmed:**
  - redirect URIs are not enforced;
  - TOTP codes are accepted without verifying the secret;
  - grant handling (PKCE) is hand-written;
  - logout and `max_age` are unconfirmed.

L2 therefore proves only **application** enforcement.

## 4. Core identity design

### 4.1 Layout and boundary

```text
core/internal/identity/
  identity.go service.go store.go credential.go allowlist.go logout.go assurance.go http.go
  workos/{adapter,token,jwks}.go          # sole importer of workos-go, jwt, keyfunc/jwkset
  identitytest/fake_workos.go             # test-only fake WorkOS API (authorize/authenticate/JWKS/revoke) + signer
```

```go
type Provider interface {
    AuthorizationURL(ctx context.Context, r AuthRequest) (string, error)
    Exchange(ctx context.Context, in Exchange) (VerifiedIdentity, error)
    RevokeProviderSession(ctx context.Context, sid string) error
    ProviderLogoutURL(sid, returnTo string) (string, error)
}
type VerifiedIdentity struct {
    Issuer, ClientID, Subject, ProviderSessionID, DisplayName string
    AuthTime                                                  time.Time
}   // no email, no method, no assurance field
```

### 4.2 Login, callback and continuity — correction 1 and r1 finding 2

- **Binding cookie** (unchanged from r2): one stable per-browser `__Host-login`. It is created when absent, `Max-Age=600` is refreshed on each start, and no callback clears it.
- **Start:** a live forwarded session → 303 `/account?status=already_signed_in` and no transaction. A namespace mismatch → 503.
- **Callback (in order):**
  1. **Parameters:** duplicate `code`, `state`, `error` or `error_description` → 400, nothing consumed.
  2. **Namespace check.**
  3. **Consumption:** `DELETE … WHERE state_hash AND binding_hash AND client_id … RETURNING` consumes only on a match.
  4. **`error=`:** consume, no session.
  5. **Exchange and validation** (§3.3).
  6. **Continuity:**

| Browser state at callback | Callback principal | Transaction | Local sessions | Mapping | Provider calls |
|---|---|---|---|---|---|
| Anonymous or dead credential | any (valid) | consumed | new session, cookie set | found or created | none |
| Live user A, recorded `sid` = S | user A, **`sid` = S** (SSO returned the same provider session) | consumed | **A's session unchanged**; outcome `already_signed_in` | unchanged | **none** (S is never revoked) |
| Live user A, recorded `sid` = S | user A, **`sid` = S′ ≠ S** (additional provider session) | consumed | A's old session revoked (`revoke_reason='superseded'`); **new** A session bound to S′; new cookie and CSRF token; outcome `session_rotated` | unchanged | **none** (neither S nor S′ revoked) |
| Live user A | user B ≠ A | consumed | **A's session unchanged**; outcome `identity_conflict` | **none created for B**; B is looked up read-only | `RevokeSession(S_B)` best effort; outcome logged only |
| Any | namespace mismatch | **not consumed** | unchanged | none | none |
| Any | wrong verifier rejected by provider (`code_rejected`) | consumed | unchanged; no session | none | none |

**Consumption outcomes** (unchanged from r2 §4.2):

| Callback case | Transaction effect |
|---|---|
| Unknown, replayed or already consumed `state` | none |
| Wrong or missing binding | not consumed |
| Expired | deleted |
| Parallel identical callbacks | exactly one consumes |

Rotation in the additional-`sid` case invalidates CSRF tokens rendered in other tabs. Those forms return 403 `csrf_failed` with no mutation until the tab is reloaded (documented).

### 4.3 Sessions and CSRF (unchanged from r2 §4.3)

- **Credential:** 256-bit; SHA-256 stored.
- **Lifetimes:** idle 30 min and absolute 8 h. Short values only with `APP_ENV=test`; production enforces [5 min, 24 h].
- **Renewal:** conditional, write at most every 60 s, never touches revocation or the absolute expiry.
- **Restart:** sessions survive a restart.
- **CSRF:** a raw per-session 32-byte `csrf_token` stored server-side.
  - It is returned only by core `GET /session` to web server components and rendered into hidden form fields.
  - It stays valid across reloads, core restarts and multiple tabs.
  - Constant-time compare against the caller's own session.
  - It never appears in JSON APIs, URLs, logs or evidence.

### 4.4 Logout outcomes — correction 2

**Rejection order** on web form routes: method (405) → `application/x-www-form-urlencoded` (415) → exact Origin (403) → session → CSRF (403, live sessions only). Core JSON mutations apply the §7.2 order.

| Action | Live session + valid CSRF | No live session |
|---|---|---|
| Local sign-out | Commit, then 303 `/sign-in?status=signed_out`. A commit failure → **503 page `signout_failed`** (cookie cleared, success never claimed). | Idempotent: cookie cleared, 303 `signed_out`, no state change. |
| Sign out everywhere | Revoke all of the user's sessions, 303 `signed_out_all` | 401 page |
| Sign out and revoke the WorkOS session | §3.5 sequence. Records `provider_revocation ∈ {confirmed, failed}` and a one-use `logout_ref` (256-bit, hashed, 10 min). | 401 page |
| Return `/sign-in?logout_ref=` | Shows the server-recorded outcome once: "Signed out of this application. The WorkOS session for this sign-in was revoked." / "…could not be revoked or confirmed." | Unknown or reused reference → generic "Signed out of this application"; no provider claim. Query `status` is never trusted for provider outcomes. |

**Outcome vocabulary (closed):** `local_revoked`, `provider_session_revoked`, `provider_revoke_failed`, `provider_logout_redirected`, `provider_logout_returned`. Only `provider_session_revoked` asserts a provider fact, and only for that one `sid`.

### 4.5 Disable and re-enable (unchanged)

- Trusted `DisableUser` and `EnableUser` service methods; no route.
- Disable revokes all of the user's sessions in one transaction.
- Re-enable cannot revive old sessions.
- Sign-in never changes `status`.

### 4.6 Configuration

| Variable | Service | Rule |
|---|---|---|
| `IDENTITY_PROVIDER` | core | `workos` (registered change from the rendered `oidc`). |
| `IDENTITY_WORKOS_CLIENT_ID`, `IDENTITY_WORKOS_ISSUER` | core | Unset → **unconfigured**: provider operations 503 `identity_unconfigured`; production refuses to start. Partial → startup `config_invalid` (names only). |
| One of `IDENTITY_WORKOS_API_KEY` / `IDENTITY_WORKOS_API_KEY_FILE` | core | Secret, in `SecretValues()`. |
| `IDENTITY_WORKOS_API_BASE_URL` | core | Default `https://api.workos.com`; other values only with `APP_ENV=test`. |
| `IDENTITY_PUBLIC_ORIGINS` | core | Exact `https` origins. Callback = `<origin>/auth/callback`; return = `<origin>/sign-in`. |
| `IDENTITY_SESSION_IDLE`, `IDENTITY_SESSION_ABSOLUTE`, `IDENTITY_JWT_ALGORITHMS`, `IDENTITY_MAX_AUTH_AGE` | core | §3.3 / §4.3 |
| `IDENTITY_COOKIE_PROFILE` | core, web | `https` or `loopback-http`; the latter is refused in production. |

- **Removed:** there is no MFA configuration variable (M2).
- **Secrets stay in core:** web and intelligence never receive the API key.

### 4.7 Logging

- **Events:** `auth.login.started`, `auth.callback.completed`, `auth.session.revoked`, `auth.provider.revocation`, `identity.namespace_mismatch`.
- **Fields:** closed `outcome`/`scope` only, plus `request_id`/`trace_id`.
- **Excluded** (as in r2 §4.7): no IDs, `sid`, names, emails, codes, state, verifier, binding, cookies, CSRF tokens, provider tokens, claims or WorkOS error bodies.

## 5. Web BFF and UI (R8; header placement reconciled in r5 — RC-1)

- **`web/lib/session.ts`** (server-only) is the only `cookies()` user. Cookies:
  - `__Host-sid` and `__Host-login`;
  - `HttpOnly; Secure; Path=/`, no `Domain`, `SameSite=Lax`.
- **`web/lib/identity-client.ts`** (server-only) sends exactly:
  - `outboundHeaders(ctx)`;
  - `Authorization: Session <credential>` when present;
  - `X-Login-Binding` (login and callback only);
  - `Content-Type: application/json` with the §7.2 bodies.
- **Pages:**
  - `/sign-in`: POST form, closed status map, `logout_ref` handling.
  - `/account`: display name, "Multi-factor status: not verified by this application", session expiry, and three CSRF-bearing forms.

  An accessible pending-submit client component handles busy state. Everything is server-rendered with `no-store`.
- **Header (r5, RC-1):** one "Account" link, placed **inside the existing `SiteHeader` `action` region**, in AppShell's current `action` div after `AppearanceControl` and `ShellHelpDialog`.
  - It renders `<Button asChild variant=… size="sm"><a href="/account">Account</a></Button>`, using the existing Radix `Button` API (`asChild`, variants, sizes, `block`). The exact variant follows the existing header-action styling at implementation. This is one bounded, additive AppShell edit.
  - **Unchanged:** `SiteHeader` (no new prop or API), its geometry and source lineage; configured identity, primary navigation, skip link, help and appearance controls, `main` and footer.
  - **Untouched:** `page.tsx` and `globals.css`.
  - **Preserved:** scoped portals, semantic tokens, bundled fonts, blueprint assets and Light/Dark/System first paint and hydration.
- **Forms (r5, RC-1/RC-4):** sign-in, sign-out, sign-out-everywhere and provider-sign-out submits use the existing `Button`. Navigation links use `Button asChild`. There is no new UI primitive, no other UI dependency and no Base UI API.

## 6. Schema and migration (`00002_identity.sql`; medium risk)

| Table | Definition | Runtime grants |
|---|---|---|
| `identity_provider_binding` | `provider text PK CHECK (provider='workos')`, `issuer text` (1–2048), `client_id text` (1–255), `bound_at` | SELECT, INSERT |
| `users` | `id uuid PK DEFAULT gen_random_uuid()`, `status ∈ {active, disabled}`, `display_name` ≤ 200, timestamps | SELECT, INSERT, UPDATE |
| `identity_principals` | `(issuer, subject)` PK, `user_id` FK RESTRICT, `UNIQUE(user_id)`, `subject LIKE 'user\_%'`, lengths | SELECT, INSERT, UPDATE |
| `sessions` | `credential_hash` UNIQUE (32); `csrf_token` (32, raw); `provider_session_id`; **`mfa_assurance text NOT NULL DEFAULT 'unknown' CHECK (mfa_assurance='unknown')`**; `auth_time`; idle ≤ absolute; `revoked_at`; `revoke_reason ∈ {logout, logout_all, provider_logout, disabled, superseded}`; `provider_revocation ∈ {confirmed, failed}` NULL; `provider_logout_ref_hash` UNIQUE NULL; ref expiry; `returned_at` | SELECT, INSERT, UPDATE |
| `login_transactions` | `state_hash` PK, `binding_hash`, `pkce_verifier`, `client_id`, origin, `return_path ∈ {/, /account}`, `max_age`, `created_at`, `expires_at ≤ created_at + 10 min`. No nonce. | SELECT, INSERT, DELETE |

- **No tenant scope:** no tenant, role or vendor-organization column. No RLS; G008 classifies these global tables.
- **No DELETE** on users, principals, sessions or the binding table.
- **Down:** drops all five tables (dev/test only).

**Explicit negatives**, carried from r2 §6.1, with additions:
- a second binding row is rejected;
- a non-`workos` provider value is rejected;
- runtime UPDATE or DELETE of the binding row is denied (no grant);
- `mfa_assurance` values other than `unknown` are rejected;
- concurrent first-ever mappings from 8 workers produce exactly one binding row.

## 7. Contracts

### 7.1 Browser-facing (web origin)

| Route | Layer | Outcomes |
|---|---|---|
| `GET /sign-in`, `GET /account` | page | 200 `no-store`. `/account` anonymous → 303 `/sign-in?return_to=/account`. |
| `POST /auth/login` | form mutation | 303 → WorkOS; 303 `/account?status=already_signed_in`; 403 origin; 405; **415 for non-form bodies**; unconfigured, unavailable or namespace mismatch → 303 `/sign-in?status=unavailable`. |
| `GET /auth/callback` | navigation | 303 → allow-listed path (`signed_in`, `session_rotated`); 303 `/account?status=already_signed_in\|identity_conflict`; 303 `/sign-in?status=canceled\|expired\|stale\|unavailable\|failed`; **400** for an ambiguous query. |
| `POST /auth/logout`, `/auth/logout-all`, `/auth/provider-logout` | form mutation | §4.4; 401, 403, 405, **415 non-form**, 503 as minimal `no-store` pages. |
| `GET /api/account` | JSON API | 200 `{id, display_name, mfa_assurance:"unknown", auth_time, session_expires_at}`; 401 `{"error":"unauthenticated"}` (independent of provider configuration); 503 `{"error":"unavailable"}` (store unavailable). Never carries CSRF. |

### 7.2 Core internal (`/internal/v1/auth/*`)

- **Mutations** accept **`application/json` only** (anything else → 415 `unsupported_media_type`), with a body of at most 16 KiB.
- **Rejection order:** 415 → 400 → 403 origin → 401 → 403 `csrf_failed`.
- **Ignored headers:** forged `X-User-Id`, role and MFA headers are ignored.
- **Reachability:** the same rules apply on the loopback host port.

| Route | Input | Success | Errors |
|---|---|---|---|
| `POST /login` | `{origin, return_to?}` + optional credential and binding | `{authorization_url, binding?}` | `origin_rejected` 403, `already_authenticated` 409, `identity_namespace_mismatch` 503, `identity_unconfigured`/`identity_unavailable` 503 |
| `POST /callback` | `{origin, raw_query}` + binding + optional credential | `{outcome: signed_in\|session_rotated, session, return_to}` or `{outcome: already_signed_in}` | `ambiguous_parameter` 400, `transaction_invalid` 401, `code_rejected` 401, `identity_conflict` 409, `user_disabled` 403, `impersonation_refused` 403, `stale_authentication` 403, `provider_error` 502, `provider_timeout` 504, `identity_namespace_mismatch`/`identity_unconfigured`/`identity_unavailable` 503 |
| `GET /session` | credential | `{user_id, display_name, mfa_assurance, auth_time, expires_at, csrf_token}` | `unauthenticated` 401, `identity_unavailable` 503 |
| `POST /logout` | `{scope: current\|all\|provider, csrf_token, origin}` + credential | `{outcome, provider_logout_url?}` | §4.4 |
| `POST /provider-logout/return` | `{logout_ref}` | `{outcome}` from the server record | Generic outcome (200) for unknown or reused references |

- **No route for the guard.** No G004 route uses `RequireVerifiedMFAHandler`. Its 403 `mfa_unverified` contract is tested on the wrapper directly.

### 7.3 Layer contract

- **Session resolution** (`/api/account`, `/session`, local logout and logout-all) needs only PostgreSQL. It is identical whether WorkOS is configured or not.
- **Provider operations** need WorkOS configuration and a namespace match.
- **Content types:** the browser-form layer is form-encoded only, and the core JSON layer is JSON only. They never cross. Web converts form fields into the JSON contract.

## 8. Tests and evidence

### 8.1 Layers (unchanged roles; see r2 §8.1)

| Layer | Proves | Never proves |
|---|---|---|
| L1 fixtures | Validation, PKCE, binding, continuity, namespace, logout logic, assurance guard, configuration, log privacy | Provider behaviour |
| L2 emulator | Our HTTPS cookie, CSRF, origin, cache, callback, session and logout behaviour; mapping; replicas, restart and outages; accessibility; network boundaries | Real authentication, MFA, SSO or freshness |
| L3 staging API | Real token, issuer, client, subject, `sid`, `auth_time` and JWKS contract; MFA enforcement **of the tested configuration**; email reassignment; `RevokeSession` | Hosted UI flows; session MFA |
| L4 manual hosted | Real hosted sign-in, callback, MFA success and cancel, SSO reuse, forced reauthentication, logout sequence, two-origin UUID | — |

Unavailable L3/L4 access is **UNMET**. No session injection or bypass login exists.

### 8.2 Scenario catalogue (expected state outcomes)

**L1 (Go unit, offline)**
- **U-CFG:** missing, partial or malformed configuration; production refuses `loopback-http`, emulator/non-default base, http and test lifetimes.
- **U-TOK:**
  - tamper, `alg:none`/HS256, unlisted algorithm;
  - wrong `iss` (including the trailing-slash variant), wrong `client_id`, `sub` ≠ `user.id`;
  - missing `sid`;
  - **missing, forged (future) or stale `auth_time` → rejected**;
  - expired, future `iat`, leeway boundaries;
  - impersonation (`act`, `impersonator`, method) → `impersonation_refused`;
  - a refresh token offered as an access token.
- **U-JWKS:** rotation, unknown `kid`, rate limit, attacker key, oversize, timeout, redirect.
- **U-PARAM:** parameter handling cases.
- **U-RET:** return-path cases.
- **U-AUTHZ:** vendor claims are not persisted.
- **U-HTTP:** §7.2 table, including **E-CT core cases**:
  - `application/x-www-form-urlencoded`, `text/plain` and `multipart` to core mutations → 415 with no state change;
  - valid JSON → 2xx.
- **U-LOG:** log privacy cases.

**L1/L2 assurance (R7/AC7 deterministic set)**

| ID | Case | Expected |
|---|---|---|
| A-1 | Valid ordinary sign-in whose token and response carry **no MFA metadata** | Session issued; `mfa_assurance='unknown'`; `/api/account` shows `unknown` |
| A-2 | Same, with `authentication_method="Password"` and the WorkOS fixture reporting "environment MFA required" | `unknown` (policy and method cannot upgrade) |
| A-3 | Forged labels: `amr:["mfa","otp"]`, `acr:"mfa"`, `mfa:true` claims; `?mfa=verified` query; `X-MFA-Verified: true` header | `unknown`; no column or field accepts another value |
| A-4 | **Policy drift**: the fixture or emulator issues tokens for password-only (MFA "disabled") versus MFA-required | Identical session rows (`unknown`); no false verified claim |
| A-5 | `RequireVerifiedMFA` on a valid `unknown` session; the handler wrapper on a live session | `mfa_unverified` / 403; no state change |
| A-6 | Attempt to insert or update `mfa_assurance='verified'` as runtime | CHECK violation |
| A-7 | Cancellation: `error=access_denied` (emulator error hook) after start | Transaction consumed; **no session**; `canceled` |
| A-8 | Invalid signed identity or time (U-TOK set) | Rejected; no session; transaction consumed |

**L2 integration (real PostgreSQL; B-IDENTITY-IT)**
- **Carried from r2:** I-MAP-CONC, I-ATOMIC, I-LIFE, I-REPLICA, I-DISABLE, I-ALL, I-TXN, I-OUTAGE, and the §6 concurrency set.
- **New namespace cases:**

| ID | Case | Expected |
|---|---|---|
| I-NS-1 | Same issuer, **changed client** | Login and callback 503 `identity_namespace_mismatch`; no mapping, session or consumption; existing sessions valid |
| I-NS-2 | Changed issuer, same client | Same as I-NS-1 |
| I-NS-3 | First binding race with 8 workers | One binding row |

**Browser projects and totals (r5, RC-2):**

| Project(s) | Base URL / environment | Specs | Status in r5 |
|---|---|---|---|
| `shell-light`, `shell-dark` | Ordinary root stack (`WEB_BASE_URL`) | Existing `(shell\|overlay\|accessibility\|boundary\|design)` **plus additive `account`** | Inherited; `account.spec.ts` covers the anonymous/unconfigured account and sign-in surface, a11y and tab order in both appearances |
| `tokens-light`, `tokens-dark` | Fixture build `127.0.0.1:3001` | `tokens` | Inherited unchanged |
| `appearance` | Root stack | `appearance` (excluding `@dev-hydration`) | Inherited unchanged |
| `dev-hydration` | `next dev` `127.0.0.1:3003` | `appearance` `@dev-hydration` | Inherited unchanged |
| **`identity-light`, `identity-dark`** (new, explicit) | Separate disposable HTTPS identity environment (`https://org-a.localhost`, emulator) | `identity.spec.ts` only | New. Not matched by any inherited project; inherited projects never run against the identity environment. |

- **Pins:** Playwright `1.63.0` and the image pin in §3.7. Fixture public-asset copy, fontconfig fallback and `pwuser` are unchanged.
- **Totals:** the accepted 133 (root) / 133 (disabled) / 2 (rename) are **scenario evidence**, not frozen totals. After G004, totals are **derived from the actually collected tests** in each report, with zero skipped, unexpected or flaky. `BROWSER_MATRIX_FULL` gains the `account` scenario IDs for the shell projects, and a separate identity matrix covers the identity projects (B-IDENTITY-E2E).
- **Inherited assertions:** adapted **only** where the Account control genuinely changes them, for example S-TAB tab order or header action-region contents, with the reason recorded per assertion in the Result. Accessibility, scope, first-paint, font, blueprint and fidelity guarantees are retained. No visual expectation is changed to conceal an unrelated regression.

**L2 browser E2E (emulator; `identity-light`/`identity-dark`; B-IDENTITY-E2E)**
- **Carried from r2 §8.2:** E-LOGIN, E-ORIGINS, E-ATTR, E-FIX, E-CB-1..8, E-CSRF-1..5, E-LO-1..4, E-CACHE, E-A11Y.
- **Changed or added:**

| ID | Steps | Expected |
|---|---|---|
| E-CB-9 | Live A (`sid` S); new tab completes a callback for A where the emulator returns **the same `sid` S** | A's session unchanged; `already_signed_in`; zero revoke calls (emulator request log) |
| E-CB-10 | Live A (S); callback for A with a **new `sid` S′** | Old session revoked `superseded`; new session on S′; other-tab form → 403 until reload; zero revoke calls |
| E-CB-1 (restated) | Pending B; browser signs in as A; B completes | A unchanged; no B mapping; `RevokeSession(S_B)` called once |
| E-PKCE | Attacker transaction (own verifier) receives the victim's code (emulator if CAP-2 confirms PKCE verification, else the L1 fake) | Provider rejection → `code_rejected`; attacker transaction consumed; **no session** for anyone |
| E-CT-1/2 | Valid form POST to web routes; JSON, `text/plain` and multipart to web form routes | 303 success; 415 with no state change |
| E-LO-5 | Local commit failure (PostgreSQL paused) | 503 `signout_failed`; never `signed_out` |
| E-LO-6 | Provider sign-out with revoke failing (error hook) | Local revoked; **no WorkOS redirect**; return page shows "could not be revoked or confirmed" |
| E-LO-7 | Return with a forged or reused `logout_ref` | Generic text; no provider claim |
| E-LO-8 | Revoke succeeds; the emulator does not implement the logout URL (if CAP-2 confirms) | Redirect issued and recorded as `provider_logout_redirected` only; the real browser-session outcome is proven only by MQ-6 |
| E-NET-1 | Context-level `context.on('request')` and `context.on('page')` capture during the full flow | Only the app origins under test, the emulator origin, and `attacker.localhost` in CSRF tests |
| E-NET-2 | Open an **additional page** (`context.newPage()`) and a `window.open` popup from the account page | Their requests are captured and subject to the same allow-list; an injected request to `core:8180` from a page fails the assertion (negative control) |
| E-NET-3 | Direct core probes | Run only in a **separate driver** (`playwright.request.newContext()`, no shared storage) or harness `httpprobe`; never in the browser context |

**L2 probes and config:** B-IDENTITY-PROBES and B-IDENTITY-CONFIG as in r2, plus production refusal of any MFA-labelled configuration attempt (unknown variables are ignored, and the guard is unaffected).

**L3 staging (tag `workos_staging`; B-IDENTITY-STAGING)**

| ID | Expected |
|---|---|
| S-JWKS, S-ATTR, S-TAMP, S-CODE, S-REV, S-CLEAN | As in r2. |
| S-MFA | A per-run user with an API-enrolled TOTP factor authenticates with password only → MFA-challenge error class and **no tokens**. With TOTP → tokens, validated by the production validator. The mapped session's `mfa_assurance='unknown'`. Evidence records "tested configuration: MFA required, observed at <UTC>" separately from session data. |

### 8.3 Data classification — r4 correction 3 (refines r2 §8.3; existing scanners only)

**Reused scanners, unchanged:**
- the run's out-of-tree secret registry (0600), consumed by B-SECRET-CANARY;
- logcheck;
- the evidence publication scan;
- in the hosted staging job, the `hosted-context.sh` credential-proof pattern file.

**Policy:**
- Only the **strict secret classes** below are registered.
- Public protocol metadata is never registered, so its intended use needs **no exclusion**. Every scanner still scans every file it scans today.

| Class | Members | Registered as strict secret? | Permitted channels | Forbidden (asserted) |
|---|---|---|---|---|
| **S1 Provider credential** | WorkOS API key. Sources: `IDENTITY_WORKOS_API_KEY`/`_FILE`; staging `IDENTITY_STAGING_WORKOS_API_KEY`. L2 uses a per-run generated value where the emulator accepts a configured key (CAP-2); the emulator's public default `sk_test_default` is a published test constant, not a secret. | **Yes**, plus the CI credential-proof pattern file | core process; core→WorkOS `Authorization`; the 0600 secret file or env of core and the staging test container | Logs, URLs, DB, web/intelligence env, bundles, images, evidence, artifacts |
| **S2 Application bearer values** | Session credentials, login-binding values, `logout_ref` values observed in tests (all revoked or expired at teardown) | **Yes** | `Set-Cookie`/`Cookie`; web→core headers; `logout_ref` only in its one-use return URL | Logs, evidence, artifacts; saved network evidence keeps a counted SHA-256-prefix placeholder |
| **S3 CSRF tokens** | Per-session synchronizer values | **Yes** (logs and published evidence) | Own account HTML hidden fields | JSON APIs, URLs, logs; saved bodies pass the counted sanitizer first |
| **S4 Protocol secrets** | Authorization codes, `state`, PKCE verifiers; provider access/refresh tokens held by L1/L3 tests | **Yes** | Core memory; the WorkOS-issued callback URL (`code`, `state`); core→WorkOS exchange | Logs, DB (except verifier and hashes), evidence; saved request lists replace values with counted SHA-256 prefixes |
| **S5 Factor and user secrets** | L3 per-run TOTP secrets and passwords; CAP-4a's in-memory code. MQ operator factors never enter project processes. | **Yes** | L3 test process memory; WorkOS API | Everywhere else |
| **S6 Canaries** | `return_to`, `error_description`, hostile display-name and email canaries | **Yes** (display canary: see the restricted row) | Fixed UI messages only | Logs, UI (except own escaped display), evidence |
| **Public protocol metadata** | Configured issuer (e.g. `https://api.workos.com`), WorkOS client ID, API base and JWKS URLs, public origins and callback/return URIs | **No** (configured trust and protocol inputs, not bearer secrets) | Core config (server-owned); core→WorkOS requests; **client ID inside the 303 `Location` to the WorkOS authorize URL** (intended protocol use, then WorkOS-hosted pages); sanitized provenance (issuer value; client-ID fingerprint = `client_` + SHA-256/16) | Web or intelligence env and bundles; delivered HTML/JS/CSS bodies; intelligence. Asserted by the bundle/boundary checks, not the secret scanner. |
| **Restricted references** | WorkOS `sid`, WorkOS user ID (`sub`), application user UUID, own display name | No (not credentials) | `sid`: DB and core; the WorkOS logout `Location`. `sub`: DB and core. UUID and display name: own `/api/account` and own account HTML. | Logs (U-LOG / B-IDENTITY-PRIVACY assertions); other users' responses; evidence except correlatable prefixes |

**Required cases (implemented during approved execution; inside the existing B-IDENTITY-PRIVACY, B-SECRET-CANARY and C-SEC checks):**

| ID | Case | Expected |
|---|---|---|
| P-NEG-1 | Plant the registered API-key value into a scratch copy of would-be-published evidence | Existing scanner **FAILs**; publication refused |
| P-NEG-2 | Seeded disposable copy places the API-key value into web env or a delivered bundle | Boundary/privacy check **FAILs** |
| P-NEG-3 | Planted core log line containing the API-key value (scanned log copy) | logcheck/secret scan **FAILs** |
| P-POS-1 | Ordinary sign-in: the `/auth/login` `Location`, the sanitized network evidence and the provenance record carry the configured client ID, WorkOS host and issuer | Scans **PASS** (metadata not registered; no exclusion used); S2/S4 values in the same evidence are replaced by counted placeholders |
| P-POS-2 | Delivered HTML/JS/CSS bodies of all identity pages | Contain neither client ID, issuer nor API base → **PASS**. Negative control: a fixture body containing the client ID fails the bundle-boundary assertion. |

`mfa_assurance` is non-sensitive.

### 8.4 Network boundaries — correction 2

- **Page traffic** is captured at **browser-context level**: `context.on('request')` plus `context.on('page')` for pages and popups added later (E-NET-1/2).
- **Driver traffic** uses a separate `APIRequestContext` without shared storage, or harness `httpprobe`. Credentials pass through 0600 files under `$OBS_IN` for the restart/outage phases and are deleted after use.
- **Driver requests** are outside the page assertion by construction, never by filtering.

### 8.5 Harness, CI-1 secret boundary and counts — correction 3

**Suites (unchanged from r2):**
- B-IDENTITY-GUARDS (`lint`);
- B-IDENTITY-CATALOG (`migrations`);
- B-IDENTITY-IT, -E2E, -PROBES, -PRIVACY, -CONFIG (`ops`, in a separate disposable project after `variants_ops`);
- B-IDENTITY-STAGING (`staging` suite, `staging` profile only).

Existing-guard reconciliation follows r1 §7.3 with WorkOS names substituted. **r5 (RC-3/RC-4)** makes it exact against the current guards in `infra/validation/run.sh` at `7ec9c4d`. These are additions only; no existing rule is relaxed and no broad exemption is added.

| Guard | Current allowlist (`7ec9c4d`) | G004 exact addition |
|---|---|---|
| B-WEB-BOUNDARY routes | `api/health/route.ts api/ready/route.ts page.tsx` | +8 (2 pages + 6 routes): `account/page.tsx`, `sign-in/page.tsx`, `api/account/route.ts`, `auth/login/route.ts`, `auth/callback/route.ts`, `auth/logout/route.ts`, `auth/logout-all/route.ts`, `auth/provider-logout/route.ts` (r1 §7.3's "7" was a miscount) |
| B-WEB-BOUNDARY server-only modules | `lib/config.ts`, `lib/core-client.ts` | + `lib/session.ts`, `lib/identity-client.ts` |
| B-WEB-BOUNDARY `cookies()` | Banned in all web sources (with `document.cookie`, `NEXT_PUBLIC_`) | The single exception `lib/session.ts`; the other bans are unchanged |
| B-WEB-BOUNDARY request-context importers | `app/api/(health\|ready)/route.ts`, `lib/core-client.ts`, `lib/request-context.ts` | + the six auth/account route files, the two pages, `lib/identity-client.ts` |
| B-WEB-BOUNDARY outbound headers | `core-client.ts` sends exactly `outboundHeaders(ctx)` | Unchanged. New rule: `identity-client.ts` sends exactly `outboundHeaders(ctx)` + `Authorization` + `X-Login-Binding` + `Content-Type` |
| B-WEB-BOUNDARY client/component code | No `@/lib/config`, `@/lib/core-client`, `server-only`, `process.env` | Unchanged; also covers `web/components/account/*` |
| B-NAMING | Forbids `web/src`, `web/components/domain` and redundant service paths; bounded `web/public` allowed | **No change**: `core/internal/identity/`, `web/components/account/` and `infra/identity/` are responsibility-named and outside the forbidden set; G004 adds no `web/public` asset |
| B-DEPS (web) | 20 exact direct dependencies (10 runtime, 10 dev), lockfile/optional-dev controls | **No change**; G004 adds no web dependency (asserted) |
| B-DEPS (core / validation tools) | goose and MCP SDK required; tools module = exactly `santhosh-tekuri/jsonschema/v6` | Core direct requirements add the §3.7 modules. The tools module is unchanged (`identitysetup` is stdlib-only). |
| B-MCP-SOURCE | `main.go` digest; routes `GET /healthz`, `GET /readyz` | Reviewed G004 digest. Routes + exactly 5 core auth patterns (§7.2: `POST /login`, `POST /callback`, `GET /session`, `POST /logout`, `POST /provider-logout/return`; r1's "four" predates the return route). No `identitytest` in any binary. |
| B-MIGRATE-SET / -CLI / -CATALOG / -IT; root/off/rename migrate expectations | `{00001_baseline.sql}`, version 1 | `{00001, 00002}`, version 2; the catalog includes the §6 objects; the IT subtest count is re-derived |
| Customization register (B-RENDER-PLAN/REPEAT, B-ENV-EXAMPLE) | **20** rows (`render-customizations.tsv`) | **20 → 23**. Append rendered paths (per `.archetype` manifest) `.env.example`, `core/internal/config/config.go`, `docs/foundation/VENDORS.md`. Update reasons on `core/go.mod`, `core/go.sum`, `core/cmd/server/main.go`, `docs/foundation/FOLLOW-UP.md`. Identity configuration tests go in a new `core/internal/config/identity_config_test.go`, so rendered `config_test.go` stays unmodified. B-ENV-EXAMPLE accepts exactly the reviewed identity block. |
| B-TOGGLE-OFF-ADDITIONS (`APPLICABLE`) | Current list | + the new identity paths (§10) |
| `infra/docker/validation-tools.Dockerfile` | Explicit `go build ./cmd/contractcheck ./cmd/httpprobe ./cmd/logcheck` and runtime `COPY` | + `./cmd/identitysetup` in both; hadolint target set unchanged |

**GitLab settings verified read-only at RG and recorded without values (owner provides; A5) — r4 correction 2:**

| ID | Setting | Required state (every tier unless noted) |
|---|---|---|
| G1 | Settings → CI/CD → Variables → "Allow merge request pipelines to access protected variables and runners" (GitLab ≥ 18.1) | **Disabled** |
| G2 | Dedicated protected branch `identity-qualification` | Push and merge Maintainers only; no force-push; no unprotected wildcard |
| G3 | `ci_pipeline_variables_minimum_override_role` | `no_one_allowed` (G003 API-verified; re-read) |
| G4 | **Project** CI/CD variables:<ul><li>`IDENTITY_STAGING_WORKOS_API_KEY` — S1 secret; masked, and masked-and-hidden where offered;</li><li>`IDENTITY_STAGING_WORKOS_CLIENT_ID` and `IDENTITY_STAGING_WORKOS_ISSUER` — public metadata; not masked.</li></ul> | **Protected** and **environment scope exactly `identity-staging`** (no wildcard). [Project variables](https://docs.gitlab.com/ci/variables/#for-a-project) support [environment scoping](https://docs.gitlab.com/ci/environments/#limit-the-environment-scope-of-a-cicd-variable) on **Free**. Group-variable scoping, which is paid, is not used. |
| G5 | Project tier | Recorded only. It no longer changes the boundary. |
| G6 | *Optional, only if Premium/Ultimate is already available:* protected environment `identity-staging` (Allowed to deploy: Maintainers) | Defence in depth only. Not required; no purchase. The boundary never relies on deployment-permission semantics of non-default environment actions. |

- **The boundary on every tier** is G1 + G2 + G3 + G4, plus the job's `environment: identity-staging` declaration, plus the forwarding gate below.
- **Unrelated jobs are excluded.** `fast` and `gate` declare no environment, so GitLab does not provide the `identity-staging`-scoped variables to them. The r3 "Free-tier residual" no longer exists.
- **Rules use predefined variables only.** Per the GitLab guidance, no environment-scoped variable is used in `rules`.
- **Developers cannot create project variables.** This prevents predefined-variable shadowing by non-maintainers.

**CI-1 changes (r5: carried onto the current nine-rule, 211-control CI; the bypass is kept):**
- **`.gitlab-ci.yml` workflow (CI-1 clarification).** Exactly one rule is added, placed directly after the existing default-branch schedule rule:

  ```yaml
  - if: $CI_PIPELINE_SOURCE == "schedule" && $CI_COMMIT_BRANCH == "identity-qualification" && $CI_COMMIT_REF_PROTECTED == "true"
  ```

  - **Nine rules become ten.** All nine existing rules are byte-identical:
    - open-MR feature-push suppression;
    - MR and push changed-path admission with the management-Markdown omission and their `when: never` companions;
    - default-branch schedule; web; api; final `when: never`.
  - **Admitted / refused:**
    - an unprotected `identity-qualification` schedule and any other non-default-branch schedule still fall through to `when: never`;
    - web and api branch pipelines remain admitted as explicit full requests.
- **Consequence for `fast`/`gate` in an admitted qualification schedule** (**escalation E-1**, §15).
  - `fast` has no rules and `gate`'s `schedule` rule matches, so both run.
  - `infra/ci/validate.sh` `classify` currently refuses any schedule outside the default branch (N12).
  - **Proposed narrow fix:**
    - `classify` accepts a schedule when `CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH`, **or** when `CI_COMMIT_BRANCH == "identity-qualification"` **and** `CI_COMMIT_REF_PROTECTED == "true"`. Both are attributed `branch-schedule` (full), exactly like the default-branch schedule. Every other schedule is refused as today.
    - `hosted-context.sh` adds `CI_COMMIT_REF_PROTECTED` (non-secret, predefined) to the context's allow-listed environment. It is not forwarded today.
    - No credential flows to `fast`/`gate`: environment-scoped variables are absent, and the C-SEC-7 detector fails closed if they appear.
- **`.gitlab-ci.yml`:** job `identity-staging`:
  - `stage: gate`, `needs: []`;
  - `rules: if $CI_COMMIT_BRANCH == "identity-qualification" && $CI_COMMIT_REF_PROTECTED == "true" && ($CI_PIPELINE_SOURCE == "web" || $CI_PIPELINE_SOURCE == "schedule")`;
  - **`environment: {name: identity-staging}` on all tiers** (default action);
  - the same pinned CLI/DinD, medium tag, variables pattern and artifacts policy;
  - `timeout: 45m`, `RUNNER_SCRIPT_TIMEOUT: 40m`, `RUNNER_AFTER_SCRIPT_TIMEOUT: 4m`;
  - script `sh infra/ci/hosted-context.sh -- --profile staging --publish-dir ci-evidence`.
- **`hosted-context.sh` forwarding gate.** It forwards the API key and the two metadata values (as one 0600 file to the context, never env or argv) **only if all hold**:
  - `CI_PROJECT_ID` = approved project (<gitlab-project-id>) and `CI_PROJECT_PATH` = approved path;
  - `CI_COMMIT_BRANCH` = `identity-qualification` and `CI_COMMIT_REF_PROTECTED` = `true`;
  - `CI_PIPELINE_SOURCE` ∈ {`web`, `schedule`}, and `CI_MERGE_REQUEST_IID` unset. **`api`, `push` and `merge_request_event` are never accepted** (r5: the API branch source added after G003 gains no credential access);
  - `CI_JOB_NAME` = `identity-staging`;
  - **`CI_ENVIRONMENT_NAME` = `identity-staging` (all tiers)**;
  - requested profile = `staging`.

  Further rules:
  - A `staging` profile requested in any other context is refused (exit code for profile misuse).
  - In the approved context, a missing API key or metadata value → **FAIL (UNMET)**, never skip.
  - **Only the API key** is added to the credential-proof pattern file and the strict secret registry (§8.3 S1). The client ID and issuer are public metadata, recorded as sanitized provenance.
  - **Scope-misconfiguration detector.** In every other profile (`fast`, `integration`, `full`, `gate`), the presence of any `IDENTITY_STAGING_WORKOS_*` variable in the job environment makes the adapter **fail closed** before starting a context. Nothing is forwarded.
- **`infra/ci/lib.sh`** (existing budget mechanisms):
  - `CI_PROFILES+=" staging"`, `CI_SUITES_staging="staging"`;
  - `CI_RESERVE_staging=300`, `CI_CAP_staging=1800`, `CI_MINBUDGET_staging=900`;
  - `CI_JOBFLOOR_staging=2700`, with `ci_time_ok` extended to that job key;
  - `CI_CPUS_staging=2`, `CI_MEMGIB_staging=4`;
  - `CI_DISKGIB_staging=15`. The hosted 35 GiB floor stays limited to integration and full.
  - The provisional 15 GiB floor is re-derived from the first authorized run's measured peak with the accepted `max(15, ceil(1.5×peak+4))` rule; any change requires review.
- **Harness and controls.** `run.sh` gains the `staging` suite. `static_ci_config` (B-CI-CONFIG) is updated exactly:
  - top-level keys `fast gate identity-staging stages workflow`;
  - the **ten**-rule workflow list;
  - pinned `identity-staging` keys, rules, environment, image, services, tag, variables, timeouts, script, after-script and artifacts;
  - `staging` added to the known suites;
  - the `fast` and `gate` definitions and the full = `all` rule unchanged.

  `validate-controls.sh` adds the controls below. They are implemented only during approved execution.

| Control | Case | Expected |
|---|---|---|
| C-SEC-1 | MR pipeline (`merge_request_event`, MR IID set) | No forward |
| C-SEC-2 | Unprotected or other ref | No forward |
| C-SEC-3 | `push` or `api` source | No forward |
| C-SEC-4 | Wrong `CI_JOB_NAME` | No forward |
| C-SEC-5 | **Wrong or missing `CI_ENVIRONMENT_NAME`** (none, `identity-staging-x`, `production`) | No forward |
| C-SEC-6 | Profile spoof: `--profile staging` from the `fast`/`gate` job context | Refused |
| C-SEC-7 | **Unrelated job**: `fast`/`gate` stub context with an `IDENTITY_STAGING_WORKOS_*` variable present (simulated scope misconfiguration) | Fail closed; nothing forwarded |
| C-SEC-8 | **Approved context** (approved project, protected `identity-qualification`, `web`/`schedule`, job and environment `identity-staging`) | API key + client ID + issuer forwarded via one 0600 file; only the API key registered as strict secret |
| C-SEC-9 | Approved context with a missing key or metadata value | FAIL (UNMET) |
| C-SEC-10 | Planted API key in a would-be-published artifact (P-NEG-1, CI level) versus intended client ID/issuer in sanitized provenance (P-POS-1) | Publication refused versus accepted |

These are stub-based like the existing groups A–G. They are **implemented only during approved execution** (S7). The `fast`/`gate` jobs, selection, merge gating and `all` = full are unchanged. The recorded settings evidence (G1–G5, and G6 if used) lists variable keys and flags (protected, masked/hidden, environment scope), never values.

**CI-1 admission validation (r5; minimum list from the reconciliation; existing controls and the CI-definition guard only; during approved execution, no preapproval campaign):**

| Case | Mechanism | Expected |
|---|---|---|
| Qualification **web** pipeline | B-CI-CONFIG exact rules (web rule unchanged); `identity-staging` job rules; C-SEC-8 | Admitted; `fast`/`gate` full; `identity-staging` runs with forwarding |
| **Protected** qualification **schedule** | B-CI-CONFIG rule 10; new group-F attribution control `F-schedule-qualification` (`branch-schedule` → full) | Admitted; `fast`/`gate` full; `identity-staging` runs |
| **Unprotected** qualification schedule | Rule 10 condition (static); new group-B attribution negative `N12q` (`CI_COMMIT_REF_PROTECTED` ≠ `true` → "schedule outside the default branch") | Refused at workflow; attribution refuses as a second layer |
| Unrelated non-default schedule | Existing N12 (unchanged), plus a new `F-schedule-other` selection negative | Refused |
| Default-branch schedule; web/API full requests | Existing `F-schedule`, `F-web-intent`, `F-api-intent` (unchanged) | Still valid, full |
| Management-Markdown-only MR/push | Existing workflow rules and group-F management controls (unchanged) | Still omitted (no pipeline) |
| Secret isolation and missing inputs | C-SEC-1..10 (unchanged semantics; C-SEC-3 explicitly includes `api`) | Exact isolation; missing → FAIL |

**Control and registry accounting** (expected; control identities, suites and completeness rules are authoritative if an approved change alters totals):

| Count | Current (`7ec9c4d`) | After G004 | Source |
|---|---|---|---|
| Workflow rules | 9 | **10** | `.gitlab-ci.yml`; B-CI-CONFIG pin in `infra/validation/run.sh` |
| CI jobs | 2 (`fast`, `gate`) | 3 (+ `identity-staging`) | `.gitlab-ci.yml` |
| CI controls | **211** (A 25, B 34, C 8, D 23, E 22, F 77, G 22) | **224** (B 35 [+`N12q`], F 79 [+`F-schedule-qualification`, `F-schedule-other`], G 32 [+C-SEC-1..10]) | `infra/ci/validate-controls.sh` `CONTROLS` |
| Registry rows | **71** (lint 16, controls 3, provenance 5, contracts 6, unit 1, root 16, migrations 4, ops 7, rename 5, off 4, final 4) | **79** (+1 lint, +1 migrations, +5 ops, +1 staging) | `infra/validation/required-checks.tsv` |
| Profile required IDs | fast **32**, integration **64**, full **70** | fast **33**, integration **70**, full **77**, staging **5** (1 + 4 final) | `infra/ci/lib.sh` `CI_SUITES_*` × registry suites |
| Customization register | **20** | **23** | `infra/validation/render-customizations.tsv` |
| Web direct dependencies | **20** (10 runtime, 10 dev) | **20** (unchanged) | `web/package.json` |
| Browser projects | 6 | 8 (+`identity-light`, `identity-dark`) | `web/playwright.config.ts` |
| Inherited browser evidence | root 133, disabled 133, rename 2 | Derived from collected tests (§8.2) | G048 accepted reports |

Accounting notes:
- B-UNIT-CHECKS is in the unit suite only (fast and integration); B-CHECKS-ROOT is the full equivalent.
- The four `final` IDs count in every profile.
- L4 is Result evidence, not a registry row.
- Required-ID rows land with their recording implementation.

### 8.6 L4 manual hosted qualification (MQ)

The setup is as in r2 §8.6:
- operator workstation, exact submitted commit;
- proxy at `127.0.0.1:543`;
- `https://org-{a,b}.localhost:543`;
- staging configuration from an out-of-repo 0600 file;
- run CA trusted only in a dedicated browser profile;
- the observer runs as `go test -tags workos_manual -json`.

Radar and bot protection are never bypassed.

| Step | Operator | Server-observed pass condition |
|---|---|---|
| MQ-0 | Preflight | Tree, configuration and CA hashes; binding row absent or matching |
| MQ-1 | Sign in on `org-a` with password + TOTP challenge | Session; **`mfa_assurance='unknown'`**; recent signed `auth_time` |
| MQ-2 | Start as user B; cancel at TOTP | **No B session**; transaction consumed or unconsumed per the §4.2 table |
| MQ-3 | Local sign-out, then sign in again (no prompt) | Same UUID; unchanged `auth_time` (SSO reuse proven server-side); identical-versus-new `sid` recorded and the §4.2 effect observed |
| MQ-4 | Sign in on `org-b` | Same UUID |
| MQ-5 | Wait > `IDENTITY_MAX_AUTH_AGE`, then sign in | Prompt attested; `auth_time` advanced; validator accepted it only because it is fresh |
| MQ-6 | Sign out and revoke the WorkOS session | Local revoked; `provider_session_revoked`; redirect returned; **next sign-in requires credentials** (attested, plus `auth_time` advanced) |
| MQ-7 | Disable via the observer, sign in, re-enable, sign in | `user_disabled`, then a new session; old sessions stay revoked |

### 8.7 Commands

- Final run: `bash infra/validation/run-approved.sh … all`.
- CI profiles: `bash infra/ci/validate.sh --profile integration|full`.
- Hosted only: `--profile staging`.
- IDENTITY.md blocks:
  - `identity:start|migrate|browser|stop` (emulator, offline, executed verbatim by the harness);
  - `identity:staging-local` (L3 and CAP-3 on the operator workstation with an out-of-repo secret file);
  - `identity:manual` (MQ and CAP-4).

## 9. AC mapping (Spec r3 AC1–AC12)

| AC | Evidence |
|---|---|
| AC1 | ADR-003; §3.7 exact pins (RG/CAP-1); U-CFG/U-TOK/U-JWKS; S-JWKS/S-TAMP; CAP-3; MQ-1; B-IDENTITY-CONFIG; rotation via U-JWKS (real rotation recorded only if observed). |
| AC2 | I-MAP-CONC, **I-NS-1..3**, §6 negatives; E-ORIGINS + MQ-4; E-ATTR + S-ATTR (real email reassignment); forged UUID/header probes; no tenant objects. |
| AC3 | MQ-1; **CAP-4a (real wrong verifier rejected, else STOP)**; **E-PKCE**; U-PARAM/U-TOK/U-RET; E-CB-1..10 with the §4.2 effect tables; S-CODE. |
| AC4 | I-LIFE, E-FIX, I-REPLICA, E-LO-3, B-IDENTITY-PROBES, E-LOGIN storage inspection. |
| AC5 | E-LOGIN cookie attributes on two HTTPS hostnames; E-CSRF-1..5; E-CT-1/2; E-CACHE; production `loopback-http` refusal; no CORS. |
| AC6 | E-LO-1..8; I-DISABLE; I-ALL; revoke-versus-renew race; **MQ-6 + CAP-4b** (narrow `sid` revocation plus next sign-in requires credentials, else STOP); MQ-7. |
| AC7 | **Real:** MQ-1 (MFA challenge success), MQ-2 (cancel → no session), MQ-3 (SSO reuse, same UUID), MQ-5 (forced reauthentication, signed `auth_time`), S-MFA (enforcement of the tested configuration, recorded separately). **Deterministic:** A-1..A-8 (unknown assurance, no upgrade by policy, method or forged labels, policy drift, verified-MFA requirement denied, cancellation → no session, invalid identity/time rejected). IDENTITY.md delimits federation, logout and the unavailable session-proof capability. |
| AC8 | §7 contracts; U-HTTP + E-NET-3 driver probes; E-A11Y; inherited scenarios on the six reconciled projects (r5 §8.2: totals derived from collected tests; only genuinely affected assertions adapted, each with a recorded reason); `account.spec` in the shell projects; **E-NET-1/2 context-level capture**; hostile display name; ambiguous callback. |
| AC9 | B-IDENTITY-CATALOG + reconciled migration checks; §6 negatives; baseline byte-identical. |
| AC10 | B-IDENTITY-PRIVACY with the §8.3 classes: strict S1–S6 registered; public issuer/client metadata unregistered with named channels; **P-NEG-1..3** (planted API-key leaks detected) and **P-POS-1..2** (intended metadata use accepted; absent from bundles). B-SECRET-CANARY including the staging API key. No emulator or test code in production images. |
| AC11 | Full wrapper run; IDENTITY.md blocks; hosted full + integration; protected `identity-staging` run (scoped project variables plus the job environment) on the same tree; **C-SEC-1..10**; r5 CI-1 admission evidence (§8.5: ten-rule B-CI-CONFIG pin, `F-schedule-qualification`, `F-schedule-other`, `N12q`; unchanged bypass, web and API controls); recorded G1–G5 (and optional G6) settings; MQ on the same tree. |
| AC12 | Hashes, wrapper pre/post checks, cleanup (including WorkOS per-run users), CAP results, cycle-4 checkpoint, Result, Codex ACCEPT. |

## 10. File inventory (reconciled in r5; ◆ resolved — exact baseline hashes in §2.4)

**r5 exact deltas (RC-1–RC-5, CI-1):**
- **Modified:** `infra/docker/validation-tools.Dockerfile` (adds `identitysetup` to the explicit build list) and `infra/ci/validate.sh` (qualification-schedule attribution, E-1).
- **Rendered and registered:** `core/internal/config/config.go`, `docs/foundation/VENDORS.md` and `.env.example` (register 20 → 23).
- **New:** `core/internal/config/identity_config_test.go`. Rendered `config_test.go` is unchanged.
- **Conditional:** `web/tests/browser/{shell,accessibility,design}.spec.ts`, only for assertions genuinely affected by the Account control (RC-2).
- **Header change:** `app-shell.tsx` gains one additive Account link in the existing `action` region (RC-1). `site-header.tsx`, `button.tsx`, `page.tsx` and `globals.css` are unchanged and bound.
- **Playwright config:** `playwright.config.ts` adds `account` to the shell `testMatch` and the explicit `identity-light`/`identity-dark` projects.

Changes from r2 §10:
- **Added:** `core/internal/identity/assurance.go` (+ test).
- **Migration:** `00002_identity.sql` gains `identity_provider_binding` and `mfa_assurance`.
- **Removed:** MFA configuration fields from `config.go`.
- **CI-1:** `infra/ci/lib.sh` gains the `staging` profile and budget.
- **Docs:** IDENTITY.md (narrow revocation wording, M2 assurance statement, CAP results) and VENDORS.md (WorkOS, emulator and nginx test-proxy rows with licences).

All other paths are as in r2 §10:
- core identity, WorkOS adapter, fixtures, integration, staging and manual tests; `main.go`; config; `go.mod`/`go.sum`;
- web `session`/`identity-client`, auth routes, `/api/account`, pages and components;
- `request-context`; `app-shell` (RC-1); unit and browser tests; `playwright.config` (RC-2);
- `infra/compose/identity.yml`; `infra/identity/*`; the `identitysetup` tool;
- validation `run.sh`/`lib.sh`/`required-checks`/`render-customizations` (RC-3); `events.json`;
- CI-1: `.gitlab-ci.yml`, `infra/ci/lib.sh`, `hosted-context.sh`, `validate.sh` (E-1), `validate-controls.sh`;
- `.env.example`; docs; scope allow-list; Result.

**Untouched:** intelligence, contracts, MCP, `infra/compose/{local,migrate}.yml`, the wrapper, skills, `.archetype`, the baseline migration, governance, ADRs, Goal/Spec, ROADMAP/PROGRESS, SECURITY.md.

## 11. Delegation

- **Planning:** no subagent.
- **Execution:** one owner for all shared files. An optional E2E-spec subagent owns `web/tests/browser/{identity,account}.spec.ts` and `support/sanitize.ts`; its cycles count toward B7.
- **Read-only reviewers** (zero cycles): session/provider security after S5; CI-1 secret boundary and evidence before S8.

## 12. Gates and sequence — correction 4

### 12.1 RG — read-only start gate (no Docker, builds or provider calls; S0; consumes no cycle) — r4 correction 4

RG's **inputs are E1, E2, E4 and E5** (§2.3). Its **output is E3**. It never checks E3 itself.

| RG step | Check | Stop on |
|---|---|---|
| RG-0 | E1, E2, E4 and E5 are each recorded by their owner (G048 ACCEPT/integration and RC reissue; Plan ACCEPT; committed base and snapshot; named-action authorizations) | Any missing |
| RG-1 Binding | Wrapper `run-approved.sh` check-only against the E4 snapshot; HEAD equals the bound base; clean status; root `.env` absent | Mismatch |
| RG-2 Resources | Owned ports and Docker resources free (existing harness checks, read-only inventory) | Conflict |
| RG-3 Pin provenance | §3.7 re-read: `sum.golang.org` for every module; registry lookups for the emulator index and the **nginx index `sha256:0985e772…7d94` with linux/amd64 child `sha256:8f84ed99…7433` and amd64 config**; licences recorded | Any difference |
| RG-4 Access matrix | A1–A7 confirmed. GitLab G1–G5 (G6 if used) read and recorded without values, including **G4's exact `identity-staging` scope**. Staging redirect URIs registered (A3; `.localhost` accepted, else the A6 `.identity.test` alternative recorded). *(r5: the issuer-variable timing within G4 is escalation E-2, §15.)* | Any absent or wrong |
| **RG pass** | **Establishes E3**; S1 may begin | — |

### 12.2 Capability checks during approved execution (each placed after its prerequisites; a failure **stops** the dependent work)

| CAP | When | Checks (replaces r2 V-IDs) | Stop condition |
|---|---|---|---|
| CAP-1 | S1, after the allow-list | `go mod download` / `go mod verify` reproduce the §3.7 `h1:` values with no extra modules. Images pulled by the pinned digests match. The nginx index resolves on the runner to linux/amd64 child `sha256:8f84ed99…7433` with an `amd64` config. The browser image is unchanged at `playwright:v1.63.0-noble@sha256:eff16c30…4a27` (r5, RC-2). | Any difference |
| CAP-2 | S3, after the adapter and fake exist | Emulator 0.14.0 matrix (old V-11): interactive login, exchange shape, JWKS/`alg`/`iss`, PKCE-mismatch behaviour, same-`sid` / new-`sid` control, duplicate-email seeding, error hooks, logout URL | A needed L2 feature is missing and not already covered by L1 → stop and report (no silent re-layering) |
| CAP-3 | S3, real staging API (operator workstation via `identity:staging-local`; A1–A4) | Old V-1 exact `iss`; V-2 JWKS/`alg`; V-3 `sub`/`sid`/`auth_time`; V-10 SDK calls and `WithBaseURL`; S-MFA observation | A §3.3 guarantee cannot be validated |
| CAP-4a | After S5 (real hosted; A3/A4/A6) | **Wrong verifier with secret:** the observer creates its own authorize URL (own `state` and verifier), and the operator signs in. The web callback rejects it as `transaction_invalid` **without exchange**. The operator pastes the code (in memory only) into the observer, which exchanges with a **different** verifier. | The provider returns tokens → **STOP** (capability gap). Any token obtained is revoked immediately. |
| CAP-4b | After S5 | Old V-6 `max_age`/`auth_time` advance; V-7/V-8/V-9: §3.5 sequence, logout URL after revoke, return reached, **next sign-in requires credentials**; V-12 redirect acceptance observed | Any property fails → **STOP** (no redirect-only fallback) |

### 12.3 Sequence

| Step | Work | Gate |
|---|---|---|
| S0 | RG (§12.1); inputs E1/E2/E4/E5 | Pass → **E3** |
| S1 | Scope allow-list; CAP-1 | Pass |
| S2 | Migration (binding table, `mfa_assurance`), store, B-IDENTITY-IT/CATALOG | Green |
| S3 | Configuration, WorkOS adapter, assurance guard, L1, fake; CAP-2, CAP-3 | Green and CAP pass |
| S4 | Core routes, `main.go`, events schema | Static green |
| S5 | Web BFF, pages and guards; security review; CAP-4a, CAP-4b | Green and CAP pass |
| S6 | Overlay, emulator E2E, probes (`ops`) | Green |
| — | **Cycle-4 checkpoint** (expected around S5/S6) | Assess remaining CI-1, hosted, MQ and staging risk explicitly, even though that work is later |
| S7 | CI-1: workflow rule 10, `identity-staging` job (with `environment: identity-staging`), `staging` profile, qualification-schedule attribution (E-1), controls C-SEC-1..10 plus the r5 admission controls (§8.5); authorized protected run | B-IDENTITY-STAGING, B-CI-CONFIG and B-CI-CONTROLS green |
| S8 | Docs, register, `APPLICABLE`; full wrapper run | All required IDs PASS |
| S9 | Hosted full + integration + `identity-staging` on the final tree; final MQ-0..7 on the same tree; Result | Return to Codex |

**Outside G004 (removed as a gate):** production terms, data processing, residency and subprocessors (former A8). These are rollout prerequisites, not staging-implementation prerequisites.

## 13. Budget, reforecast and stops

**Reforecast: 6 of 7 cycles, one reserve. Risk: high but bounded.**

| Cycle | Work | Note |
|---|---|---|
| C1 | Schema/store, namespace binding and concurrency | — |
| C2 | WorkOS adapter and token contract against the fake; CAP-3 real contract | — |
| C3 | Emulator HTTPS E2E environment | — |
| C4 | Web, guard and variant reconciliation full run; CAP-4 | — |
| C5 | CI-1 forwarding gate and controls; protected run | G003 history shows CI is costly |
| C6 | Hosted and MQ corrections | — |

M2 removed the M1 mapping work, while correction 3 added controls. The net forecast is unchanged.

**Checkpoint after cycle 4:**
- record CAP outcomes;
- assess CI-1, hosted and MQ risk;
- if C3 or CI-1 is not converging, stop with a bounded proposal and do not spend the reserve on environment work.

**Stop and escalate** (in addition to the Goal):
- any CAP stop condition;
- token retention, sealed sessions, a Next auth SDK, injected sessions, vendor-claim authority, email linking or tenant schema;
- any MFA upgrade path;
- a CI change beyond CI-1;
- a G048 conflict;
- modules or images outside §3.7;
- WorkOS paid-feature need;
- Radar or rate limits blocking L3/L4;
- B7 exhaustion or thrashing.

**No silent skips.**

## 14. Access matrix and external actions (none satisfied; none performed by planning)

### 14.1 Access matrix

**r5 status column.**
- **Owner-reported / UNVERIFIED:** stated by the owner (PROGRESS "Setup update, 2026-10-07"; coordinator handoff), not verified. The reconciliation's earlier "not set up" is a superseded historical observation.
- **Observed at preparation, 2026-10-07:** read-only GitLab facts gathered by the coordinator during preparation and recorded here **without values**. This planner made no GitLab or provider call.
- **Not RG verification:** none of these statuses satisfies RG-4. **A1–A7, G1–G5, E2, E3, E4 and E5 are NOT satisfied.**
- **No values recorded:** no secret value, client ID or file content is recorded anywhere in this Plan.

| ID | Owner | Requirement | Needed by | r5 status |
|---|---|---|---|---|
| A1 | Owner | WorkOS team and a staging environment (free); no production environment or data | RG, CAP-3 | Owner-reported: WorkOS staging account and project exist. **UNVERIFIED.** |
| A2 | Owner | Staging client ID and a staging API key scoped for user management (users, factors, sessions) and authentication; storage location defined | RG, CAP-3, CI-1 | Owner-reported: the API key and client ID are held in a private out-of-repository file. **UNVERIFIED**; the file was not read; no values recorded. |
| A3 | Owner | AuthKit staging: email + password enabled (other methods may stay at defaults; tests use password); **MFA Required** (qualification environment); self sign-up disabled. Redirects `https://org-{a,b}.localhost:543/auth/callback`; sign-out return `https://org-{a,b}.localhost:543/sign-in`. | RG, CAP-4, MQ | Owner-reported: email/password enabled, MFA required, sign-up disabled, two callback/sign-out pairs configured. **UNVERIFIED** (exact URIs not verified). |
| A4 | Owner | Synthetic identities on an owner-controlled email domain: persistent `mq-a`/`mq-b` (verified; TOTP enrolled by the operator); L3 creates and deletes per-run users | CAP-3/4, L3, MQ | Not reported. **Unmet / UNVERIFIED.** |
| A5 | Owner | GitLab G1–G5 (§8.5): MR protected-resource access disabled; protected `identity-qualification`; override role `no_one_allowed`; the three **project** variables protected with exact environment scope `identity-staging` (API key masked/hidden); tier recorded. Optional G6 only if already available. Settings recorded as keys and flags without values. | RG, S7 | **Applied by Codex (owner-authorized) and read back independently, 2026-10-07:** G1, G2, G3, G4a and G5 match §8.5; G4b (issuer) is absent pending E-2. See the table after setup below. Not RG verification; RG-4 re-reads. |
| A6 | Owner/operator | MQ workstation: Docker, dedicated browser profile trusting only the run CA, authenticator app, out-of-repo 0600 staging file. If `.localhost` redirects are refused: `org-{a,b}.identity.test` hosts entries plus matching redirects | CAP-4, MQ | **UNVERIFIED.** |
| A7 | Owner | Paid features: **none expected** (AuthKit free to 1M MAU; MFA included; staging free). The GitLab boundary needs only Free features (protected, environment-scoped project variables). G6 is optional, with no purchase. | RG | **UNVERIFIED** planning assumption. GitLab Free observed (G5) is sufficient for the designed boundary. |

**GitLab settings observed at preparation, 2026-10-07** (read-only, by the coordinator; flags only, no values):

| Setting | Observed | Status |
|---|---|---|
| G1 | `protect_merge_request_pipelines = true`. Per GitLab MR !188008, `true` marks protected-to-protected MR pipelines as protected, i.e. the UI checkbox "Allow merge request pipelines to access protected variables and runners" is **ENABLED** | **UNMET**: the owner must disable it and confirm in the UI |
| G2 | Protected branches: `main`, `staging` only (Maintainers push and merge; no force push). No `identity-qualification` branch | **UNMET** |
| G3 | `ci_pipeline_variables_minimum_override_role = no_one_allowed` | **Met** as observed (RG-4 re-reads it) |
| G4 | 0 project and 0 group CI/CD variables | **UNMET** (see also §15 E-2 on issuer sequencing) |
| G5 | Project <gitlab-project-id> on the **Free** plan | Recorded; G6 not applicable unless the tier changes, with no purchase |
| Other | No environments; no pipeline schedules. Remote `main` at `952e0deda5570126d52c3a45b670e795dca4cece` | Recorded |
| Owner-reported, not G1 | "Pipelines must succeed" (successful-pipeline merge enforcement) reported disabled. It is a different setting from G1 (reconciliation). | Owner-reported; outside G4's boundary; not a substitute for any G-row |

**GitLab settings after Codex setup, 2026-10-07.** This table supersedes the preparation observation above for G1, G2 and G4; that table is kept as history.
- **Applied** by Codex via glab with the owner's authorization (setup record (not included), `83a11c1e…3842`).
- **Read back** independently by the coordinator, read-only. GraphQL requested metadata only, never variable values.

| Setting | Read back | Status |
|---|---|---|
| G1 | `protect_merge_request_pipelines = false`, i.e. the UI checkbox "Allow merge request pipelines to access protected variables and runners" is disabled | **Matches §8.5** |
| G2 | Protected `identity-qualification`: push Maintainers, merge Maintainers, force push false. It exists on the remote at `7ec9c4d4ba8004bc7ba59c7b791480e7be2f50bd`. Exact-name protections only (`main`, `staging`, `identity-qualification`); no wildcard. | **Matches §8.5** |
| G3 | `ci_pipeline_variables_minimum_override_role = no_one_allowed`; `restrict_user_defined_variables = true` | **Matches §8.5** |
| G4a | Exactly two project variables:<ul><li>`IDENTITY_STAGING_WORKOS_API_KEY`: env var, protected, scope `identity-staging`, masked, hidden, raw;</li><li>`IDENTITY_STAGING_WORKOS_CLIENT_ID`: env var, protected, scope `identity-staging`, unmasked, raw.</li></ul>0 group variables. | **Matches §8.5** for the key and client ID |
| G4b | `IDENTITY_STAGING_WORKOS_ISSUER` absent (deliberate) | **Pending E-2** (created after the verified issuer contract) |
| G5 | Free | Recorded |
| Pipelines on `identity-qualification` | Only 2923741057 (`push`, **skipped** via `ci.skip`); no qualification or full run | Recorded |

### 14.2 Named external actions (separate authorization each)

1. Create the staging environment, client and key.
2. Configure AuthKit, redirects and sign-out return.
3. Create MQ users and enroll their TOTP.
4. GitLab:
   - G1 setting;
   - G2 protected `identity-qualification` branch;
   - G4 project variables (protected, exact `identity-staging` environment scope). The issuer variable's sequencing is escalation E-2 in §15 (proposed split into G4a/G4b);
   - *Status 2026-10-07:* G1, G2 and G4a were performed by Codex with the owner's authorization (setup record). G4b remains pending E-2. Optional G6 was not used (Free).
   - optional G6 only if already available.
5. Push the evidence branch and run the existing manual full and integration pipelines.
6. Push or update `identity-qualification` and run the protected `identity-staging` pipeline(s), including per-run staging user creation and deletion.
7. CAP-3/CAP-4 and MQ sessions against real staging, including CAP-4a code handling.
8. Post-acceptance key rotation and revocation hygiene.

## 15. Remaining items for Codex review

| ID | Item | Status |
|---|---|---|
| R-1 | r5 mechanical reissue confirmation (design accepted at r4) | For Codex confirmation |
| R-2 | E1: G048 ACCEPT and integration **satisfied**; the RC-1–RC-5 part is this r5, pending confirmation | Pending confirmation |
| R-3 | Access matrix A1–A7 and actions 1–8 (E5; RG-4).<ul><li>Owner-reported WorkOS items remain UNVERIFIED.</li><li>A4 (MQ users and TOTP) is unmet.</li><li>A6 and A7 are unverified.</li><li>GitLab G1, G2, G3 and G4a now match §8.5 (applied by Codex, read back 2026-10-07).</li><li>G4b is pending E-2 (§14.1).</li></ul> | Execution blocker (owner) |
| **E-1** | **Qualification-schedule attribution (CI-1).** Admitting the protected `identity-qualification` schedule (workflow rule 10) also runs `fast` (no rules) and `gate` (`schedule` rule). `infra/ci/validate.sh` `classify` refuses schedules outside the default branch (N12), and `hosted-context.sh` does not forward `CI_COMMIT_REF_PROTECTED`, so both jobs would fail attribution. The reconciliation's clarification named workflow admission only. **Option (a), recommended and written into §8.5:** narrowly extend `classify` to accept that schedule only when `CI_COMMIT_REF_PROTECTED == "true"` (→ `branch-schedule`, full), and add the predefined `CI_COMMIT_REF_PROTECTED` to the context allow-list. **Option (b):** exclude `fast`/`gate` from qualification schedules via job rules. That changes the pinned `fast`/`gate` definitions and leaves the scheduled tree without its own validation. No credential path changes under either option. | **Escalation — Codex confirm** |
| **E-2** | **Issuer-variable sequencing (coordinator addendum).** §3.3 copies the exact issuer from a decoded real staging token at CAP-3 (S3). RG-4 (before S1) expects G1–G5 present, and G4 includes `IDENTITY_STAGING_WORKOS_ISSUER`, so that variable cannot hold a verified value before RG. **Option (a), recommended because it changes no security contract:** RG-4 verifies G4 for the API key and client ID only, and records the issuer variable as "created after CAP-3, before S7"; named action 4 splits into G4a (key and client ID) and G4b (issuer, after CAP-3). **Option (b):** a separately authorized read-only issuer verification before RG. Not applied in the body beyond cross-references in §12.1 and §14.2. *Status 2026-10-07:* G4a is applied, and the issuer variable is intentionally absent. That matches option (a) in practice but does not decide it. | **Escalation — Codex decide** |
| R-4 | *Removed in r4:* there is no tier-dependent residual or owner trust-model decision. Scoped project variables plus the job environment apply on all tiers (§8.5). | — |
| R-5 | §3.7 pins: Go hashes and the emulator digest were independently confirmed by Codex; the nginx index and amd64 child come from direct registry responses (§3.7). RG-3 re-confirms all before binding. | Execution start gate |
| R-6 | Provider facts left to execution (CAP-2/3/4), each with an explicit stop: wrong-verifier rejection; logout-after-revoke and next-sign-in prompt; exact `iss`; `alg`; `sub`/`sid`/`auth_time` format; `max_age`; `.localhost` redirect acceptance; SDK v10 field shapes | Execution capability gates |
| R-7 | Review pins: lifetimes 30 min / 8 h; leeway 30 s; `SameSite=Lax`; return allow-list `{/, /account}`; impersonation refusal; `UNIQUE(user_id)`; same-user additional-`sid` rotation; `*.localhost` origins; email dropped | Confirm at review |

## 16. Research record (read-only, 2026-10-07)

**WorkOS (as listed in r2 §16):**
- code exchange; authorization URL/PKCE; session tokens/JWKS;
- the verify-tokens article; MFA (environment-required, TOTP, SSO excluded); reauthentication / `auth_time`;
- logout; session revoke; testing; emulator and `SUPPORTED.md`; pricing.

**r3 additions:**

| Topic | Sources |
|---|---|
| Emulator | [releases (0.14.0, 2026-09-24)](https://github.com/workos/emulate/releases), [GHCR package](https://github.com/workos/emulate/pkgs/container/emulate) |
| Go modules | [workos-go v10 versions](https://pkg.go.dev/github.com/workos/workos-go/v10?tab=versions) and [`go.mod` at v10.5.0](https://github.com/workos/workos-go/blob/v10.5.0/go.mod); [golang-jwt v5 versions](https://pkg.go.dev/github.com/golang-jwt/jwt/v5?tab=versions); [keyfunc v3 versions](https://pkg.go.dev/github.com/MicahParks/keyfunc/v3?tab=versions) and [`go.mod` v3.8.2](https://github.com/MicahParks/keyfunc/blob/v3.8.2/go.mod); [jwkset versions](https://pkg.go.dev/github.com/MicahParks/jwkset?tab=versions); `sum.golang.org` lookups for each pinned module |
| Images (r4) | Direct read-only Docker Hub registry API responses for `library/nginx`: tag `1.30.5-alpine3.24`, the index, the amd64 child manifest and config, and the withdrawn 386 manifest and config (§3.7 evidence table). Page text: [nginx tags](https://hub.docker.com/_/nginx/tags). |
| GitLab (r4-corrected) | [MR access to protected variables and runners](https://docs.gitlab.com/ci/pipelines/merge_request_pipelines/#control-access-to-protected-variables-and-runners); [project variables (all tiers)](https://docs.gitlab.com/ci/variables/#for-a-project); [limit the environment scope of a CI/CD variable](https://docs.gitlab.com/ci/environments/#limit-the-environment-scope-of-a-cicd-variable) (exact project-variable scope; do not use scoped variables in `rules`); [protected environments (Premium/Ultimate; optional G6 only)](https://docs.gitlab.com/ci/environments/protected_environments/). r3's statement that environment scope is Premium/Ultimate applied to **group** variables only and is corrected. |
| Spec-cited | RFC 9700, RFC 7636, RFC 8725; OWASP Session and CSRF cheat sheets |

**r5 (local files only).** No network, provider, GitLab or `glab` access; no `.env` or staging secret file was read. Facts were derived by read-only inspection of the clean checkout at `7ec9c4d`:

| Area | Files read |
|---|---|
| Reconciliation and review inputs | the repository reconciliation, tenant-redirect addendum, r4 review, CI management-docs bypass record and G048 implementation review |
| CI | `.gitlab-ci.yml`; `infra/ci/{lib.sh,select-profile.sh,validate.sh,hosted-context.sh,validate-controls.sh}`. Control counts come from the `CONTROLS` array. The one in-process `--list` invocation only prints that array and exits, with no Docker, test or network use. |
| Harness | `infra/validation/{run.sh,lib.sh,required-checks.tsv,render-customizations.tsv}`; `infra/docker/{validation-tools,web-browser}.Dockerfile`; the `.archetype` manifest (rendered-path list) |
| Web | `web/{package.json,playwright.config.ts}`; `web/components/{layout/app-shell.tsx,ui/site-header.tsx,ui/button.tsx}` |
| Core and docs | `core/cmd/server/main.go`; `core/migrations/`; `docs/foundation/{WEB-SHELL,VENDORS}.md`; PROGRESS and ROADMAP |

The GitLab facts in §14.1 were observed read-only by the coordinator during preparation and relayed here; they are not this planner's observations.
