> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G014 — Tenant-bound MCP connections and optional Group selection: Implementation Plan

Revision: r8, 2026-10-15 UTC. Status: **Plan with r7 provisional design ACCEPT; r8 R0 reconciliation, review-ready for
Codex re-review and execution release.** This is not an execution binding and not implementation. Tier:
architecture/security/public-contract (Goal Execution Policy).

**r8 revision note.** r8 is the R0 reconciliation and execution-readiness record (handoff db0ce620), after final
G013 ACCEPT (G013 review on staging `aae75c1`). It is narrow by design: only the gate table, the R0 record (§0.2),
pins confirmed against merged G013 code (§11.5), the finalized allow-list (§11.8) with its file-table row (§8), the
TestMCPRecovery wording, the sequence and the open questions change. There is no design change.
- **R0 found nothing material.** P-G013-1 to -8 hold against the merged G013 code; §0.2 cites file:line for each.
- The r7 replay-kind flag resolves without design change:
  - MCP sessions join the kind-agnostic restore invalidation;
  - no MCP replay kind is needed, because every post-checkpoint MCP restriction is either subsumed by that
    invalidation or already recorded by an existing ledger kind (`user.disable` for the RV-1 hold);
  - the wording lands in G014's already-planned SECURITY.md and RECOVERY.md edits. Codex is asked to confirm this
    reading (Q7).
- **Allow-list facts.** Three pre-existing pinned files must change with the version-10 and audit delta and are now
  listed:
  - `core/internal/audit/{audit,audit_test}.go` (the 29-action catalog pin);
  - `core/internal/tenancy/access/audit_integration_test.go` (every catalog action must be observed; was
    conditional);
  - `core/internal/tenancy/access/recovery_integration_test.go` (G013; pins version 9).
- **Gates.** QP-B passed and the RV-1 provider stage plus the QP-4 namespace are accepted (staging `33b6750`).
  OD-6 is settled ("Continue with workos test"). G013 has final ACCEPT. Execution release and the binding remain.
- The initial Result is drafted at `management/v0/results/G014-tenant-bound-mcp-connections-result.md`.
- Model and effort: Opus 5.5 (session model) at xhigh.

r7 is archived byte-identically as r7 (not included) (r7 file digest
`e1186f4cc6b5a07623ebdc03cb5d18bf862aac87bb3fea1385d93e6b72fcc4ab`, accepted as provisional design at `66a1b6f`). The
r6 to r0 archives are unchanged. Base and binding target: staging `ef0676a8ed830b776c8e46e37169a842377f1063`, which
contains the G013 ACCEPT commit `aae75c1`. The delta from `641e715` includes:
- the r7 review and Plan import;
- the QP-B authorization, readiness, result and acceptance;
- G013 attempt 4 and its final ACCEPT;
- workflow and governance updates;
- `b98cedc` (observer tooling only; no G014 input);
- `08b9629` and `ef0676a` (G013 evidence-folder approval records and PROGRESS; management only; the G013 review row is
  rebound).

There is no G014 code change. The branch was rebased with the redundant planning commits dropped, because staging
already holds r7 and r0 to r6 byte-exact.

**r7 revision note (retained).** r7 makes the two wording corrections from the
r6 review (not included) (CORRECTION LOOP, handoff 731b737d), with matching summary text.
1. **§4.7 step 0.** `DisableUser` is skipped **only** when U is already disabled. An active U is always disabled by
   this procedure, even when another operator hold is open; that other hold then forbids this procedure's release.
   C12 and the IT-MCP-14 coverage note match.
2. **§11.0B step 6.** No status code is assumed to prove that a refresh value was not consumed: a 5xx or any other
   non-definitive error may follow processing and rotation. Any ambiguous outcome leaves that family **unresolved**
   for the run; a later `invalid_grant` cannot qualify it, and there are no retries. The successful-issuance fail rule
   is unchanged.
- OD-6 stays pending; there is no research, probe or test. Model and effort: Opus 5.5 (session model) at xhigh.

r6 is archived byte-identically as r6 (not included) (r6 file digest
`34b7301c1e76eb3198f4a7e63204100e208738ddee3438be64026488c07c36dc`, reviewed at `010900b`). The r5 to r0 archives are
unchanged. Base: staging `641e715214406e358fcc94022e8728ba6d4358d5`. The delta from `d1dc108` includes:
- the r6 review;
- G013's merged implementation (MR into staging: migration 00009, RECOVERY.md, recovery tooling, harness, SECURITY.md,
  AUDIT, DATABASE-ISOLATION, FOLLOW-UP, MIGRATIONS, OBJECT-STORAGE and OPERATIONS changes);
- the G015 Goal and Spec definition, with its ROADMAP line;
- PROGRESS.

Rebound rows: SECURITY.md (its §6.2 G014 note is byte-identical) and ROADMAP. G013's merge is not final G013 ACCEPT;
P-G013 reconciliation stays at R0 (§0.2, with a new R0 flag).

**r6 revision note (retained).** r6 is a small correction for the
r5 review (not included) (CORRECTION LOOP, handoff 39c1b6a4). Owner option B stays the
intended global-identity model; there is no return to per-connection credential sealing.
1. **§4.7 runbook.**
   - A new step 0 records U's prior account state and any other open hold.
   - The procedure holds U only if U was active. Release enables U only to clear this procedure's own hold, after
     confirmed provider revocation, explicit operator clearance and no other open hold.
   - A previously disabled or otherwise held account is never enabled.
   - IT-MCP-14 gains one pre-disabled-user case, (d).
   - No new registry, schema or public operation.
2. **QP-B keeps the accepted QP-4 split** (§11.0B step 3).
   - The signed `sub` must equal the same-environment WorkOS User Management `id`, with the G004 source chain pinned.
   - The real principal lookup stays E2 implementation evidence.
   - Once the subject is known, U's unrelated authorized-application baseline is recorded (excluding the disposable
     client) and must be preserved (step 4).
3. **Definitive negative refresh** (§11.0B step 6).
   - Only a supported grant refusal (`invalid_grant`) is definitive. Outages, 429, 5xx, `invalid_client`, setup
     errors, resource refusals and unknown outcomes are inconclusive.
   - Any old-family issuance fails or stops the run. A later refusal of the rotated predecessor cannot launder it.
- **Narrowed:** RQ-5 to the client-ID form and the pre-registered public PKCE client type; RQ-7 to the documented
  targeted scope plus baseline preservation, with no multi-user campaign.
- OD-6 stays pending. B7 with the checkpoint after attempt 4 is unchanged.
- Model and effort: Opus 5.5 (session model) at xhigh.

r5 is archived byte-identically as r5 (not included) (r5 file digest
`f717ac0e2799b05cd09f95e27cf2b4463ed027a63bd0adf845445751c8fbd798`, reviewed at `85d6993`). The r4 to r0 archives are
unchanged. Base: staging `d1dc1080a3c190e6968c61d6d41e69ccad3efe10`. The delta from `fd21bdc` (`1f05b10`, `d1dc108`) is
the r5 review, G013 records and PROGRESS; it has no code, migration or harness change and no G014 input change.

**r5 revision note (retained).** r5 is a bounded revision for the owner's selection of **option B** ("Ok"; handoff
f3952b03), recorded in the provider qualification review (not included),
section "Owner selection B and revised acceptance gate — 2026-10-15 UTC". WorkOS stays the browser login and the MCP
token issuer.
The owner accepts broader same-user/client refresh-credential reach and requires supported compromised-client
revocation (Spec RV-1) before implementation. The earlier option-A recommendation is superseded. The assessment
history and the failed QP-3 are unchanged.
- Kept: exact issuer, audience, user, client and host checks; immutable Organization and live permissions; independent
  stored Group and lifecycle state per connection. A direct A-style bearer or session presented at B still fails.
- Not promised: independent provider refresh families per connection, or fresh WorkOS password entry or consent per
  connection. A separately issued token with B's exact audience is evaluated under B's own binding and live authority.
- Normal logout is **local**. It is not containment of a provider-side compromise; that is RV-1 (§4.7).
- RV-1 has two stages:
  - before execution, supported provider revocation of all of a user/client's families is qualified with provider-only
    flows (§11.0B);
  - during approved implementation, local pending and active invalidation, old-access refusal, replacements held on
    uncertainty and fresh local replacement are proved (§10.3 IT-MCP-14).
- **The RV-1 provider findings are in §11.0B of this Plan**, not a separate file: the documented operation, the exact
  unanswered questions and the proposed future probe scope.
- Model and effort: planned on Opus 5.5 (session model) at xhigh, matching the owner's Opus-planning rule.

r4 is archived byte-identically as r4 (not included) (r4 file digest
`8127eefe150b249e4da907fbb6e944aa08e9c9b347f33fa0488981d811527670`, reviewed at `d09eb92`). The r3, r2, r1 and r0
archives are unchanged. Base: staging `fd21bdcbae889613d766743998cf7eccb54a6991`. The G014-relevant delta from
`827070d` (`3f55952` … `fd21bdc`) is:
- the QP preparation, run and failure records (`ONEOFF-G014-resource-qualification.md` and the provider qualification
  review);
- the refresh-binding assessment r1 and its review;
- the owner selection B, with dated Goal, Spec (managed grants, RV-1, AC4/AC8/AC9), Architecture, DECISIONS, ADR-005
  and connection-contract amendments, and the planned SECURITY.md §6.2 note.
The same range also carries G013 records, model-routing and governance-policy edits, and `skills/sdlc-management`
template changes. None changes a P-G013 interface used here (§0.2), and there is no code, migration or harness change.

| Owner selection B item | r5 change | Where |
|---|---|---|
| WorkOS login and MCP issuance kept; broader same-user/client reach accepted | Approved direction recorded; token checks unchanged; activation and use by any same-user/client family disclosed; no per-connection family, password or consent promise | §2.1, §2A, §3.3, §3.4, §14 |
| Local state independent, credentials not | IT-MCP-4 and IT-MCP-8 restated: direct replay refused; stored Group and lifecycle state isolated; no credential-isolation claim | §10.3, §10.7 |
| Normal logout is local | Lifecycle table states logout is not compromise containment | §4.5 |
| RV-1 provider stage (before execution) | Documentation findings, unanswered questions and the proposed combined provider-only qualification (with the unreached QP-4 subject mapping) | §11.0B, §0.1, §12 |
| RV-1 local stage (during implementation) | Compromise procedure built from existing DisableUser, the provider operation and EnableUser, with no new framework; local cases as IT-MCP-14 in the existing B-MCP-IT suite | §4.7, §10.3, §10.7 |
| QP outcome | QP-1/QP-2 passed; QP-3 failed (kept as history, superseded as an acceptance requirement); QP-4 to QP-8 incomplete | §11.0 |

**r4 revision note (retained).** r4 is a concise reconciliation of the owner's OD-1, OD-3 and OD-4 answer ("I agree
with your recommendations, but it can be 30 minute for setup"; handoff de09f174), recorded in the
r3 review (not included). There is no design or research change.
- OD-1 is accepted as amended: pending setup 30 minutes (superseding 15), idle 30 days, absolute 180 days.
- OD-3 authorizes one bounded QP with the exact values now in §11.0. This revision does not run it.
- OD-4 is accepted: bounded PG-006 observations with escalation.
- The new Spec, contract (the neutral example from `dcb68a4`), PERFORMANCE.md and review-record hashes are bound.
r3 is archived byte-identically as r3 (not included) (r3 file digest
`89a5796ba985e5c6fa55e6281ee36ae1497b2307e3415eb09a7e146c9919c03a`, reviewed at `1ecf3ba`). The r2, r1 and r0 archives
are unchanged. Base: staging `827070d616cc2ae802bc6f579ee528654683b846`. The delta from `ee0f946` (`aae1101` …
`827070d`) is the r3 review with the owner dispositions, the neutral contract example, the Spec and PG-006 disposition
records, PROGRESS, and G013's implementation review. G013's review is G013's own matter; P-G013 interfaces stay
provisional until R0 (§0.2).

**r3 revision note (retained).** r3 is a bounded correction for the r2 review (not included)
(CORRECTION LOOP; review file digest `3fd851540e3986224d250c17cc0a1bb14b3d25515df5c7ba342d2fa45be0abe3`). Only the five
listed items and the review's settled decisions change; the approved UX, scope, twelve ACs and B7 are unchanged.
r2 is archived byte-identically as r2 (not included) (r2 file digest
`6de79b72c1c4f99f300bd965029d38cdd157b4831f9d83547e872a8a3ef7e7c8`, reviewed at `6aaccb8`); the r1 and r0 archives are
unchanged. Base: staging `ee0f946dbd826f32a889f498ef8bc013da2b9268` (delta from `3df04ea`: the r2 review and PROGRESS
only).

| r2 review item | r3 change | Where |
|---|---|---|
| 1. Anonymous metadata has no RLS read path | Protected-resource metadata is **stateless public protocol data**, built only from a registered host and a syntactically valid canonical UUID path. There is no database lookup and no existence or tenant signal. Protected calls still deny unknown or terminal connections; HTTP tests cover both | §3.2, §4.2, §4.5, §10.1, §10.3, §14 |
| 2. Logout-all snapshot semantics overstated | The accepted statement-snapshot cut-off of `revokeAll` (`identity/store.go:771-776`) is preserved: logout-all revokes MCP sessions visible at its snapshot, and a connection committed after it survives (disclosed). Disablement keeps its user-row serialization. The exact create/logout-all interleaving is tested; no redesign | §2.2, §4.5, §4.6, §10.2, §14 |
| 3. Renewal must not revive expiry | M5 is one conditional UPDATE that enforces session, user, `mcp` kind, unrevoked state, idle and absolute liveness and the throttle in the write itself. Zero rows is disambiguated: a live throttled no-op continues, while expired, revoked, unknown or wrong-kind refuses the request. Renewal runs before tool execution | §4.6, §10.2, §11.6 |
| 4. Bind own-connection reads to the Organization | SELECT is `user_id = U AND organization_id = O`. Same-binding revoke after membership loss still works (it needs no membership); a runtime-SQL cross-Organization read test is added; the validator keeps the existing impersonation (`act`) refusal | §3.3, §4.3, §5, §10.1, §10.3, §10.4 |
| 5. Disposable, bounded QP | QP-6 records token-shape facts only, not production validator behaviour (an implementation test). QP-9 is removed. The disposable client and two indicators, exact audience-only hosts, the reachable callback, deadlines, ownership and cleanup, and the original setting state are pinned. No environment-wide registration change | §0.1, §2.2, §10.6, §11.0, §12 |
| Settled by the review | OD-5, C1, C8 (host registry and fixture edge); C2–C5, C7, C9–C11 reasonable subject to evidence; QP client = one disposable pre-registered PKCE client plus the official SDK; logout-all includes MCP under its snapshot contract | §2.2, §2.3 |

Execution waits for (r8):
- Codex review of this R0 revision, including Q7 and the allow-list additions;
- Codex's recorded execution release with the Git-first binding (allow-list import, initial Result, target, budget
  and resource release).

Satisfied:
- the RV-1 provider stage and the QP-4 namespace (QP-B passed and accepted, `33b6750`);
- OD-6;
- final G013 ACCEPT (`aae75c1`);
- R0 (§0.2).

The QP-3 failure stays history.

**r2 revision note (retained).** r2 implements the owner's direction "Ok update with those details" (handoff
9cd9c0f5), the dated Goal, Spec, Architecture, ADR-005 and connection-contract amendments, and the live-access
clarification (handoff c145760a). **WorkOS AuthKit Connect is the MCP authorization server, and every MCP connection
has its own tenant-hosted resource URL** `https://{tenant_host}/mcp/c/{connection_uuid}`, created by an
authenticated, membership-checked setup in the web app. Core is the resource server and the local authority.

r1's embedded authorization server is **removed**: no authorization endpoint, consent page, authorization codes,
refresh issuance or core-issued bearer values. Archives are preserved byte-identically:
- r1 (not included) (r1 file digest
  `4ec75d026929f05add2fd94a78974d8f42bcc0f9c5ee6dc6c384c2a85518a22c`, reviewed at `ac7785d`);
- r0 (not included) (r0 file digest
  `f6d6fdef8ddc8ea5c539d2cd56be5cd697f21d5094ddf4c29a9bc25628e5d941`).

**r2 change map:**

| Input | Change | Where |
|---|---|---|
| Owner direction and amendments (issuer, URL, lifecycle) | WorkOS Connect issues tokens; per-connection tenant-hosted resource URL; web setup creates a pending connection; one atomic activation binds the client; ordinary refresh or reauthentication keeps the URL; terminal states need a replacement URL; recovery covers pending and active connections | §1, §3, §4, §5 |
| Spec R1 and R5 clarifications | Connect access tokens authenticate the subject only, after issuer, namespace and exact-audience checks; ordinary login/ID tokens are refused. Provider refresh is distinct from local authorization: core never trusts WorkOS refresh as authority and checks live local state on every protected call | §3.3, §4.5 |
| Live-access clarification | Each committed role, membership, session or selected-Group change applies to the next protected call, with no reconnect or token refresh. Every control-state mutation has a pinned transaction boundary (§4.6). No in-flight read cancellation or undoing of committed work. G015 keeps catalog invalidation and notification | §4.2, §4.6, §10, §16 |
| r1 review (resolved items kept) | Authority matrix, C9 lifecycle methods and the approved logout default are retained and adapted. r1's E4–E6 readers are not needed: r2 adds no new SECURITY DEFINER reader | §4.2, §5 |
| Alternative review and probe | `sid` is not used for identity (the probe showed it shared). The pre-acceptance qualification includes a **negative** old-family refresh against a new resource, concrete tenant-hosted audiences and the G004 subject mapping. First-token client binding is disclosed | §2A, §11.0, §14 |
| ACs | All twelve mapped, including the lifecycle and recovery evidence in AC10 | §10.7 |

- **Bound inputs (r8 base staging `ef0676a8ed830b776c8e46e37169a842377f1063`, SHA-256; rows marked r8 are new or
  changed since r7, rows marked r7 since r6, the r6 row since r5, rows marked r5 since r4, rows marked r4 since r3;
  every other row is unchanged at the r8 base):**

  | Input | SHA-256 |
  |---|---|
  | Plan r7 review (not included) (provisional design ACCEPT; r8) | `dd5e824aef7e592ec1beaea3d3dfaa2f4f945d8133c617f75552fd03ac727fbf` |
  | Revocation qualification record (not included) (QP-B result; r8) | `9cd7709c3eade3fce4118496d4b91045345fda8e0cd26bd7311f402bbe44be79` |
  | [G013 implementation review](../results/G013-implementation-codex-review.md) (final G013 ACCEPT plus later evidence-folder approval records; r8) | `51658846ca380fa635b9c63da49e0f21bbd56322e77a46df186445042081ae0d` |
  | [G013 Result](../results/G013-recovery-retention-incident-evidence-result.md) (gate 77 / controls 441; r8) | `e6921e3688c047d195947d6c89a59b11c40184bdc718d6711795d4ce6d96a0a8` |
  | [RECOVERY.md](../../../foundation/RECOVERY.md) (G013 delivered; R0; r8) | `12670f744937ea98e62d44ca3177f88d22d6d86f3446f3512497e0ba4a7b405e` |
  | [invalidate.sql](../../../../infra/recovery/sql/invalidate.sql) (G013 restore invalidation; R0; r8) | `f198f41d0a15b0d4444b153a3db638deb90b9911d4bb0ee70444d41d9b5b2f52` |
  | Plan r6 review (not included) (r7) | `5c9d62c36c4b116da065cf6a3b8260eef0af72ce8e6876f7e92e0cde707d7d0a` |
  | Plan r5 review (not included) (r6) | `6433e6e83f7477541cfe5bf1fa4580c2bf2e8714b36eed6aceec8c124540884b` |
  | Provider qualification review (not included) with owner selection B and the QP-B / RV-1 provider-stage acceptance (r8) | `d4bea7bd794f2c669138e859d998d5fe0a377402d555012f7947358926cd42e4` |
  | Refresh-binding assessment review (not included) (r5) | `27bdb8b4cf1e2991bfa3aed1321a7792e3b07c0eb6255f3e216a4aa52e8a12e7` |
  | Refresh-binding assessment r1 (not included) (r5; option A superseded) | `15754605334ef494e07e5627cedd17aecbb05632ac52d6befefb9e73ceb6c33a` |
  | Resource qualification record (not included) (r5; QP runs) | `76489c2edee40bb506b6bcfcf199585d011a4fbdd3e9c3a5a5f4ed6166096c3b` |
  | [DECISIONS](../../../DECISIONS.md) (ADR-005 owner amendment line; r5) | `d9a90800ff43b16d53a3bef8e2db9027ae14ee9a42a494354d849d00d1235f5e` |
  | Plan r3 review (not included) with the dated owner dispositions (r4) | `0b3d975d57cff3b66798e2b02ad9cc7e004242ef83ebe488ec5b8cc019b885ce` |
  | Neutral contract example record (not included) (r4) | `294271a6b9542ded990064c8b6f71d93d1986938816154780ebfa5c8460cb6b2` |
  | Plan r2 review (not included) | `3fd851540e3986224d250c17cc0a1bb14b3d25515df5c7ba342d2fa45be0abe3` |
  | [Goal](../goals/G014-tenant-bound-mcp-connections.md) (owner UX/route amendment and owner selection B; r5) | `e35a9e5dbf503c22b81a07653c17841d354818159fab6a05947268c4888fc255` |
  | [Work Spec](../specs/G014-tenant-bound-mcp-connections-spec.md) (earlier amendments plus managed grants, RV-1 and AC4/AC8/AC9; r5) | `c5bc60c7c4bfc3ae92a0e63711e41a57eea6d375b6b78139bf2599eb68021a44` |
  | Architecture (not included) (§3 MCP connection setup and lifecycle; managed MCP issuance; r5) | `e6993bcc1c4324f867434c65216ec5d6caad084065da0da65c6ca07324a67011` |
  | [ADR-005](../../../adr/ADR-005-tenant-bound-mcp-connections.md) (owner UX addendum; owner decision on WorkOS grants; r5) | `db31ffd657d168796773bd21c8257085e24daca0d4947200facdf9ffac573b79` |
  | [Connection contract](../../../../contracts/mcp/connection-context.md) (earlier amendments plus managed grants; r5) | `23edbe6404694fa15866929775d224f9456a43f6d6f6df0e070faa0228ba2140` |
  | Alternative review (not included) (with owner UX and live-change sections) | `e3287536c884d6f9f9e62b80cd7a0ac7b08193884d8329774c0c1b8c25727fc5` |
  | ROADMAP (not included) (G015 line; r8) | `15cd15c2c9f713cc22afcfb6cf1102cc607ce9cb237b0e727ab9f03823ed4407` |
  | Plan r1 review (not included) (probe and assessment history) | `d8f3a9c2f464876aaaa31d8d1ac04080adc5c5ada787d87c52b1cc32e38d6898` |
  | WorkOS probe record (not included) | `f2e751fd5d1b471205e5521bf3f50b3817da7d089367d6f452a66233c463f568` |
  | Plan r0 review (not included) | `58bda575f8761ccb58bc28cdf74d4b1a32ad37ec219c78b488804b3f50222b85` |
  | Definition review (not included) | `7fdc90c6d5d1eddf681434b03b91fe2ec19cf8bfbc4519ef8c7d2f153658f35c` |
  | [PERFORMANCE.md](../PERFORMANCE.md) (PG-006, with the owner disposition; r4) | `53d6b7f208f3530d797179dee92a613c9d6e243b079a8b224a95c60b0f95e180` |
  | [ADR-003](../../../adr/ADR-003-workos-identity-provider.md) | `537522dca62e57cd291ffdad581ea000a1b04c45b378d5f505af061fdb952a62` |
  | [ADR-006](../../../adr/ADR-006-audit-schema-boundary.md) | `dc7eda774dc06c8ca1eafc52e723772d4e4c8f32d3141187759b9c47cadabc2a` |
  | [docs/SECURITY.md](../../../SECURITY.md) (accepted G012, the planned §6.2 G014 note unchanged, and G013's delivered sections; r7) | `66e80909c686de57700c754443bbf12e287fb6850548e18130e9d9d7f4b04229` |
  | [G013 Plan r3](G013-recovery-retention-incident-evidence-plan.md) (provisional recovery inputs) | `4c936221bf680a1d870c26ecc648fd6413d525ee5a77160effc366549c8184ab` |
  | G013 binding record (not included) (with amendment 01) | `dc85100f104d6819cfb56a601751b25b560e2b7b064eac4c08b126861633b77f` |

  Everything else is bound Git-first by path at the base commit.
- **Baseline (r8):** rebased onto `ef0676a`, which contains the G013 ACCEPT commit `aae75c1`; the delta from r7's
  base `641e715` is in the r8 revision note. R0 against the merged G013 code is recorded in §0.2.
  - Retained baselines: r7 `641e715` (its delta included G013's merged implementation); r6 `d1dc108`; r5 `fd21bdc`
    (delta in the r5 note); r4 `827070d` (delta in the r4 note).
  - Retained r2 baseline: rebased onto `3df04ea`. The delta from r1's base `3bbadd6` is management and documentation
    only (`f1b7a08` … `3df04ea`):
  - the r1 review and owner dispositions;
  - the WorkOS probe record and cleanup;
  - the alternative review;
  - the G013 F4 and final-batch records;
  - the Goal, Spec, Architecture, ADR-005, contract and ROADMAP amendments.
  There is no code, migration or harness change. G013 has no implementation on staging yet; every G013 interface stays
  provisional (§0.2). G013's F4 owner record is G013's own matter and implies nothing here.
- **Planner:** task G014, role plan; native worktree `.claude/worktrees/agent-<id>`; branch `g014-plan`.
  Effort xhigh as instructed; the runtime setting is not verifiable from inside the worker (unknown). The implementation
  default is high.
- **Method (r2–r6):** read-only file and Git inspection; r6 fetched nothing new.
  - Provider facts come from the earlier documented assessment (public WorkOS docs, 2026-10-15) and Codex's probe
    record; no re-fetch was needed.
  - No subagent, probe, provider change, credential, install, build, scan, test, Docker command or harness run.
  - r5 re-read public WorkOS documentation for RV-1 only (§11.0B lists URLs and access times). There was no provider
    call, experiment, support message or settings change.

## 0. Gates and predecessor reconciliation

### 0.1 Gates

| Gate | Status at r8 |
|---|---|
| Owner direction: WorkOS per-connection URLs, tenant-hosted UX, lifecycle | **approved** as UX/contract direction (Goal/Spec/ADR/contract amendments); not Plan ACCEPT |
| Owner selection B: WorkOS login and MCP issuance with broader same-user and same-client reach | **approved** (2026-10-15; dated Goal, Spec, ADR-005, Architecture and contract amendments; Codex documentation ACCEPT); not Plan ACCEPT |
| Pre-acceptance provider qualification QP (§11.0) | **run** under OD-3 (2026-10-15): QP-1/QP-2 passed, **QP-3 failed** (kept as history; superseded as an acceptance requirement by owner selection B), QP-4 to QP-8 incomplete; cleanup verified |
| RV-1 provider stage and remaining subject mapping (§11.0B) | **passed and accepted** (QP-B under OD-6; staging `33b6750`): all 20 old-family checks returned a definitive `invalid_grant`, R4 worked, and the namespace proof holds. Limited to the actual user, pre-registered public client and configuration; no global revocation SLA or JWT invalidation is claimed. App E2, local containment and holds stay implementation evidence |
| Owner decisions (§2.2) | all answered: OD-1 (as amended), OD-3, OD-4, option B and OD-6 ("Continue with workos test"); OD-5 and OD-2 for QP settled by the r2 review. The OT-1 third-party client pin (OD-2) applies only if a third-party client is claimed |
| Codex ADR for WorkOS Connect per-connection issuance and the subject-namespace rule | ADR-005 owner amendment recorded; QP-4 namespace evidence accepted (`33b6750`); Codex-owned |
| G013 final ACCEPT | **done** (`aae75c1`; G013 review bound). R0 is recorded in §0.2 with nothing material |
| Codex Plan ACCEPT | r7 provisional design ACCEPT (`66a1b6f`); r8 R0 revision awaits Codex review |
| Git-first binding and execution release (allow-list §11.8, initial Result, target `ef0676a`, B7 with the checkpoint after attempt 4, resources) | ready for Codex; the release is not yet recorded |
| G015+ and G016 | not authorized. No catalog invalidation or notification work and no hostname, DNS, TLS or provisioning work |

### 0.2 G013 interfaces (P-G013-n) and the R0 reconciliation (R0 done in r8)

- **P-G013-1:** G013 owns migration 00009; G014 uses the next free number, **00010**, re-checked at R0.
- **P-G013-2:** RECOVERY.md's identity data class gains the MCP connection relation (§5.5).
- **P-G013-3 (r2 simplification):** `invalidate-auth` expires every unrevoked `app.sessions` row (G013 Plan r3
  :471-476). Every MCP connection, pending or active, owns one `mcp` session row (§4.1), so restore invalidation covers
  both **with no change to G013's SQL**. Core stores no MCP bearer, code or refresh material, so nothing else needs
  deleting. R0 confirms the final G013 statement still expires every unrevoked session.
- **P-G013-4:** no restore-ledger kind is needed. A post-checkpoint logout or Group change cannot resurrect anything,
  because restore invalidates every connection and tokens never re-create records.
  - **R0 flag (r7):** G013's delivered SECURITY.md note says future MCP sessions or connection bindings "must join that
    invalidation and the replay kinds". R0 reconciles P-G013-3 and P-G013-4 against the merged G013 code and the
    final G013 ACCEPT. r7 makes no design change.
- **P-G013-5, -6, -7, -8 (unchanged from r1):** version-9 reconciliation pins, required-check totals, SECURITY.md
  sections and the canonical `auditACL` helper are the base for G014's version-10 changes.

**R0 record (r8).** Checked against staging `b98cedc`, which contains the G013 ACCEPT commit `aae75c1`. It still holds
at the binding target `ef0676a`, because the delta is management only. Nothing is
material. The P-G013 items above are now confirmed facts, not provisional assumptions.

| Item | Finding | Evidence (file:line) | Material |
|---|---|---|---|
| P-G013-1 | The last migration is 00009, so the next free number is **00010**. | `core/migrations/00009_retention_maintenance.sql`; set pinned at `core/migrations/migrations_test.go:44-49` | no |
| P-G013-2 | RECOVERY.md's identity data-class row exists; G014 adds `mcp_connections` to it. | `docs/foundation/RECOVERY.md:22` | no |
| P-G013-3 | `invalidate-auth` expires every unrevoked `app.sessions` row with no kind filter, then requires zero live sessions. MCP pending and active connections are covered with no G013 SQL change. | `infra/recovery/sql/invalidate.sql:12-16`, `:23-31`; `infra/recovery/recovery.sh:713-725` | no |
| P-G013-4 and the r7 flag | The ledger records restrictions only: `object.delete`, `user.disable` and `policy.disable` by sync; `membership.remove`, `membership.downgrade` and `profile.unlink` by append. No MCP replay kind is needed (details below the table). G014's planned SECURITY.md §9 and RECOVERY.md edits record this. | `docs/SECURITY.md:322-323`; `docs/foundation/RECOVERY.md:153-161`, `:166-174` | no; Codex to confirm the reading (Q7) |
| P-G013-5 | 00010 must reconcile these version-9 pins (detailed below the table). `recovery_integration_test.go` was missing from the r7 allow-list and is added in §11.8. | see below | no (allow-list fact) |
| P-G013-6 | The totals equal §11.5's "After G013" column: 114 rows, migrations 31, ops 13; profiles fast / integration / full 35 / 74 / 107; VA7 77; VU1 35/77/108; VU9 35/43/74. G013's staging gate: 77 PASS and B-CI-CONTROLS 441/441. | `infra/validation/required-checks.tsv:4`; `docs/foundation/OPERATIONS.md:90`, `:92`, `:93`, `:131-132`; `infra/ci/validate-controls.sh:1112`, `:1115`, `:1206-1207`, `:1948`; G013 Result `:8` | no |
| P-G013-7 | SECURITY.md §1 to §9 are present. §6.2 (planned G014 note) is unchanged, and the §9 G014 handoff line is present. Its digest is in the bound inputs above. | `docs/SECURITY.md:227`, `:307`, `:322-323` | no |
| P-G013-8 | The `auditACL` helper is at `core/internal/tenancy/access/audit_integration_test.go:208`. The same test requires every catalog action to be observed (`:1211-1221`), and the Go catalog pins 23 trigger, 6 core and 29 actions (`core/internal/audit/audit_test.go:71`, `core/internal/audit/audit.go:22`). G014's two new trigger actions therefore require changes to all three files (§11.8). | as stated | no (allow-list fact) |
| Recovery tooling at version 10 | The catalog and data fingerprints are computed generically. Restore only checks that the backup schema version is at most the checkout's maximum. 00010 needs no recovery-tool change. | `infra/recovery/sql/fingerprint-catalog.sql:9`; `infra/recovery/lib.sh:131`; `infra/recovery/recovery.sh:548` | no |

**Why no MCP replay kind is needed (P-G013-4):**
- Post-checkpoint MCP logout, revoke and Group changes are subsumed by invalidation: every restored connection is
  terminal.
- The RV-1 hold is `DisableUser`, which the existing `user.disable` kind already records.
- Provider revocation is external state, so a database restore does not undo it.
- Post-checkpoint connections are absent after a restore, and tokens never re-create them.
- The note "must join that invalidation and the replay kinds" is therefore met by joining invalidation plus the
  existing kinds.
- Disclosure: replaying `user.disable` after a released RV-1 hold re-disables the user, which is fail-closed; the
  operator re-enables the user after the restore.

**Version-9 pins that 00010 must reconcile (P-G013-5):**
- `core/migrations/migrations_test.go:44-49`, `:82`;
- `core/internal/migrate/runner_test.go:122-125`;
- `core/internal/migrate/integration_test.go:446-448`, `:594`, `:600-602`, `:656`, `:667`, `:727`;
- `core/internal/migrate/retention_integration_test.go:23`, `:126-135`, `:253`, `:258-310`;
- `core/internal/tenancy/integration_test.go:116-132` (the down ladder from 9);
- `core/internal/tenancy/access/integration_test.go:202-204`;
- `core/internal/tenancy/access/recovery_integration_test.go:88` (G013; pins version 9);
- 00009's relation-count precondition (41) at `core/migrations/00009_retention_maintenance.sql:327-330`, which 00010's
  preconditions copy.

**Accepted assumptions (unchanged):**
- A-1: migrations end at version 8 on staging.
- A-2: the 29-action audit catalog and M1–M5 semantics.
- A-3: G008 shapes, readers E1–E3 and their policies.
- A-4: G007 request scope, binding check and matrix, plus AUTHORIZATION's guarded-read snapshot and guarded-write
  linearization (`docs/foundation/AUTHORIZATION.md:183-216`).
- A-5: dormant MCP and the current B-MCP-SOURCE.
- A-6: Go SDK v1.8.0 is already direct; no `go.mod` change expected (S0).
- A-7: the braces exception expires 2026-10-29 (an external blocker if it lapses).

## 1. Implementation approach

```text
Member (browser, G004 session) ── https://{tenant_host}/connections/mcp ──> web (setup page, thin adapter)
                                                                  └─> core /internal/v1/mcp/connections (create/list/revoke)
MCP client ── https://{tenant_host}/mcp/c/{uuid} (POST) ──────────> core (resource server: verify, bind, authorize)
     │     └─ /.well-known/oauth-protected-resource/mcp/c/{uuid} ─> core (names the WorkOS issuer)
     └──── WorkOS AuthKit (Connect): client registration, hosted sign-in, consent, tokens, refresh
```

- **Authorization server:** WorkOS AuthKit Connect, in the same WorkOS environment as G004 sign-in. Core never issues,
  stores or refreshes MCP bearer values. WorkOS may keep refreshing after a local revocation; core refuses locally
  (Spec R5 clarification).
- **Connection = server-side record.** It holds:
  - a random canonical UUID minted by core at setup;
  - the canonical user and immutable Organization UUID;
  - the exact tenant host and resource URL;
  - the client bound at activation;
  - an optional selected Group, a context version and lifecycle state;
  - one `mcp` session row, which carries liveness, revocation and recovery.
  The URL and UUID are selectors only, never authority.
- **Per call:** verify the Connect access token (issuer, signature, exact audience equal to the stored resource, time,
  subject, client). Map the subject to the canonical user only through the existing G004 principal. Then read the
  connection under row-level security and apply live authority through the access facade in a fresh request scope.
  Nothing is cached.
- **The five tools, Group selection ordering and the authority matrix** are carried from r1 and adapted (§3.5, §4.2,
  §4.4).

**Rejected alternatives (recorded):**
- *Embedded issuer (r1):* superseded by the owner's direction; removed. It is not a fallback.
- *A shared tenant URL with `sid` as connection identity:* the probe showed two authorizations sharing one `sid`, so
  the key collides.
- *Literal per-connection provider registration at runtime:* it would need core to call the WorkOS management API with
  the environment-wide key on every setup. That is a runtime provider mutation and wider credential use; rejected.
- *Standalone Connect:* its documented completion exposes no resource, and its external-ID user writes conflict with
  existing G004 WorkOS users (assessment Q3).
- *Deriving the tenant from Host or forwarded headers, or from the slug:* forbidden. Hosts come only from an explicit
  trusted registry (§3.1).
- *A Redis authority cache, or `.env.example` changes:* unchanged reasons from r1.

## 2. Decisions

### 2.1 Approved direction (not pending)

- **URL:** `https://{organization_slug}.${PLATFORM_BASE_DOMAIN}/mcp/c/{UUID}`. Host, resource, canonical user/client and
  immutable Organization must agree. A different Host cannot reuse an otherwise valid credential.
- **Connections:** one connection per user, Organization and AI client is the default UX, not a limit; multiple
  connections are supported. Agency staff and client users use their existing roles; there are no MCP-only grants.
- **Same URL, separate URLs:** one URL is one logical connection, so reusing it shares Group selection and logout.
  Separate URLs isolate **stored** Group and lifecycle state, not provider credentials (owner selection B). Group
  switching happens inside one connection.
- **Renewal:** ordinary refresh or reauthentication of a still-valid connection keeps the URL and binding. Logout,
  revocation, recovery invalidation, missing state and expired pending setup are terminal and need a new UUID and URL.
- **Spec clarifications:** R1 (dedicated Connect tokens, not login/ID tokens; no email linking) and R5 (provider
  issuance is distinct from local authorization), both recorded in the Spec amendment.
- **Live access:** authority applies to subsequent protected calls with no reconnect or refresh.
- **Logout:** connection-local, approved in the definition. The r1 review accepted the C9 lifecycle-method distinction.
  Under owner selection B, logout ends only the local connection and is **not** containment of a provider-side
  compromise (RV-1, §4.7).
- **Owner selection B (2026-10-15):**
  - WorkOS is the browser login and the MCP token issuer.
  - The owner accepts that one user/client refresh credential may obtain tokens for that user's other
    already-authorized resources. One same-host case was observed; cross-host, unconsented and future replacement
    reach are unqualified exposures, not provider guarantees.
  - No independent provider refresh family, unique provider authorization, fresh WorkOS password entry or fresh consent
    is promised per connection. A new local connection still needs authenticated browser setup (M1) and current
    membership; a token never creates a connection.
  - Compromised-client containment is RV-1: supported provider revocation of all the user/client's families, qualified
    before execution, plus local invalidation proved during implementation.

### 2.2 Owner decisions

**Owner dispositions (r4; dated record in the bound r3 review).** The owner answered "I agree with your
recommendations, but it can be 30 minute for setup".

**OD-1 — Connection lifetime policy: accepted as amended.**
- Pending setup expires **30 minutes** after creation (supersedes the proposed 15).
- An active connection has idle expiry 30 days, renewed by protected use (M5, at most once per hour), and absolute
  expiry 180 days.
- **Disclosed consequence:** idle or absolute expiry is terminal, so the member must create a new connection and paste
  a **new URL** into the client. Routine token refresh or reauthentication of a still-valid connection keeps its URL.

**OD-3 — QP authorization: authorized** for exactly one bounded WorkOS staging run under §11.0, with the values pinned
there:
- one disposable public PKCE client and two path-wildcard, audience-only Resource Indicators for `org-a.localhost` and
  `org-b.localhost`;
- a temporary protected HTTPS callback;
- setup idle cap 30 minutes, and an active phase of at most 30 minutes after the owner's explicit Start;
- deadlines and process identities recorded before the ready link is shared;
- no automated second run;
- scoped cleanup with independent readback and a compact receipt.
It authorizes no environment-wide setting change (client metadata documents, dynamic registration, JWT template, login
or MFA policy), and no change to existing clients, callbacks or users. Execution-time settings (OT and production)
are separate later authorizations. This Plan revision does not run QP.

**OD-6 — RV-1 provider qualification authorization (open, r5).** Requested only after Codex reviews §11.0B. It would
authorize one bounded, provider-only run with the records, steps, deadlines and cleanup pinned there, combined with
the unreached QP-4 subject mapping. It is not authorized by OD-3 or by the earlier probes.

**OD-4 — PG-006 disposition: accepted.** Bounded per-request query-plan and lock observations (§11.6), with unexpected
scans, fan-out, contention or cost escalated. There is no benchmark, SLO or production performance acceptance, no
speculative index or cache, and no weakened live-authority check. Prior PG-005 acceptance does not transfer.

**Settled by the r2 review (not reopened):**

- **OD-5 / C1 / C8:** an explicit trusted host registry (`MCP_TENANT_HOSTS_FILE`, §3.1), each host bound to one
  Organization UUID with its origin in `IDENTITY_PUBLIC_ORIGINS`, and fixture edge routes on `org-a.localhost` and
  `org-b.localhost`. Configured-host readiness and the G016 provisioning boundary are retained: real
  `*.${PLATFORM_BASE_DOMAIN}` hosts, TLS, edge routing and identity-origin scaling are prerequisites, not built here.
- **OD-2 for QP:** one disposable pre-registered public PKCE client plus the official Go SDK for automated evidence,
  with no environment-wide client-registration enablement. Before any third-party compatibility claim (OT-1), the
  client, its version, its registration mode and the operator test must be pinned. That pin is an execution-time item
  and does not hold QP.
- **Logout-all and disable:** logout-all includes MCP sessions under its existing statement-snapshot contract (§4.6).
  Account disable keeps its existing user-row serialization and stays effective; re-enable revives none. Excluding MCP
  is not the default.

### 2.3 Codex decisions

r3 status: the r2 review settled C1 and C8 (with OD-5). It found C2, C3, C4, C5, C7, C9, C10 and C11 reasonable, subject
to the corrected design and actual evidence; that is not blanket permission for new schema or grants. C6 remains: one
Goal. C7's metadata is now stateless (§3.2).

| # | Decision | Recommended default | Alternative |
|---|---|---|---|
| C1 | Host registry format | TSV `host[:port]` → Organization UUID; exact lowercase match; one host per Organization; https only in production; read once at start | database registry (later, G016) |
| C2 | Audit mapping | §6.1: trigger-owned `mcp.connection_established` on activation and `mcp.context_changed`; resource type `mcp_connection`; termination by existing `session.revoked`; `access.denied` operation `mcp.request` with reason `resource_mismatch` | extra Organization-scoped creation or termination events |
| C3 | Setup surface | One web page `/connections/mcp` on the tenant host: create, list and revoke **own** connections in that Organization; no product screens | create only (revoke via the tool) |
| C4 | Reusable-guidance depth | as r1 (reference and catalog edits, fixture comments, template deferred) | as r1 |
| C5 | Admission | per-connection token bucket 20/s, burst 40; global in-flight 64; 10 s deadline | none |
| C6 | Goal size | one Goal, fallback split §13 | split now |
| C7 | Metadata scopes | `scopes_supported` in protected-resource metadata pinned at QP to the minimal Connect set that yields refresh (expected `offline_access`); scopes never confer authority | omit |
| C8 | Edge routing on tenant hosts | the edge routes only `/mcp/c/` and `/.well-known/oauth-protected-resource/mcp/c/` to core; everything else to web; `/internal/*` never | second listener |
| C9 | Lifecycle methods | accepted in the r1 review: `initialize`, `notifications/initialized` and `ping` need a verified token and binding but no tenant check | strict |
| C10 | Test trust anchor | `MCP_CONNECT_JWKS_FILE`, accepted only with `APP_ENV=test`, so E2E can use a per-run signer generated by the existing `identitysetup`; refused in production and covered by a guard test | a test JWKS HTTP service |
| C11 | E1 kind filter | keep: E1 resolves only `browser` sessions (defence in depth) | rely on the discarded credential |
| C12 (r5–r7) | RV-1 local procedure | record the prior account state; always run existing trusted `DisableUser` as this procedure's hold when the user is active, even if another hold is open, and skip it only when the user is already disabled; run the qualified provider operation; call `EnableUser` only to clear this procedure's own hold, never while another hold is open or for a previously disabled account (§4.7); no new table, route, tool or registry | a narrow per-user/client hold record (new schema; not proposed) |

## 2A. Provider facts relied on (documented or probed; nothing else assumed)

Sources: the documented assessment and its corrected revision (public WorkOS documentation read 2026-10-15
11:45–12:30 UTC: `https://workos.com/docs/authkit/mcp`, `.../authkit/connect/token-claims`, `.../authkit/connect/oauth`,
`.../authkit/connect/oauth/verifying-tokens`, `.../reference/workos-connect/token/refresh-token-grant`,
`.../reference/authkit/user/authorized-application/list`; MCP authorization 2025-11-25) and the probe record.

| Fact | Basis | Used for |
|---|---|---|
| AuthKit is an OAuth authorization server for MCP; tokens verify against the environment key set with the AuthKit issuer | docs | verification (§3.3) |
| `aud` equals the requested resource; path wildcards issue the concrete URI | docs; **QP-1 passed** for two hosts (r5) | per-connection audience |
| Refreshed tokens keep their original audience on ordinary refresh | docs (default-audience note); probe (literal resource); **QP-2 passed** (r5) | renewal keeps binding |
| **A refresh retargeted to another resource already authorized by the same user and client returned that other audience** | **QP-3 failed** (r5; one same-host case); cross-host, unconsented and replacement cases untested | accepted exposure under owner selection B; never relied on as a guarantee either way |
| Consent is per user and application; later requests skip consent unless scopes or organization access change | docs (r5 re-read, §11.0B) | no fresh consent promised per connection |
| No Connect token-revocation endpoint and no event for authorized-application deletion; deleting a user's authorized application is the documented per-user × application operation, with its token effect undocumented | docs (r5, §11.0B) | RV-1 provider stage |
| Two same user/client/resource authorizations share `sid`; `sid` stable on refresh | probe | `sid` never used for identity |
| Refresh values rotated on the observed refresh | probe | no reuse detection inferred (QP-5) |
| `auth_time` present and stable on refresh; meaning unqualified | probe | not used |
| Client ID metadata documents and dynamic registration are dashboard options | docs | OD-2 |
| MCP clients send `resource` in authorization and token requests; servers validate audience; public clients' refresh must rotate | MCP 2025-11-25 | wire contract, QP |
| Connect `sub` is the WorkOS user | docs; **QP-4 (r6 split): the signed `sub` equals the same-environment WorkOS User Management `id`, plus the pinned G004 source chain (§11.0B step 3)**; the real principal lookup is E2 implementation evidence | subject mapping |

## 3. Wire contract (R1, R2, R3)

### 3.1 Hosts, URL and registry

- **Resource URL** R = `https://{host}/mcp/c/{uuid}`.
  - `host` is an entry of the trusted registry (`MCP_TENANT_HOSTS_FILE`, exact lowercase `host[:port]` → Organization
    UUID, one-to-one).
  - `uuid` is the canonical lowercase random UUID minted by core at setup.
  - The raw path must equal `/mcp/c/{uuid}` exactly. No trailing slash, case folding, percent-encoding or extra
    segments; a query string is 400. Anything else is 404 with no fallback.
- **Host checks:** the request's Host must equal the connection's stored host, and the registry must still map that
  host to the stored Organization. Host and forwarded headers can only cause refusal, never select a tenant, issuer or
  audience.
- **The registry also lists the setup origin** `https://{host}`, which must be one of `IDENTITY_PUBLIC_ORIGINS`. The
  slug inside a host is cosmetic; Organization identity comes only from the registry and the record.

### 3.2 Endpoints and discovery

| Endpoint | Method | Behaviour |
|---|---|---|
| `https://{host}/mcp/c/{uuid}` | POST | MCP streamable HTTP (SDK, stateless, JSON responses). Checks per the §4.2 matrix. No token: 401 with `WWW-Authenticate: Bearer resource_metadata="https://{host}/.well-known/oauth-protected-resource/mcp/c/{uuid}"` (and `scope` per C7). Invalid token, unknown or terminal connection, or host/audience/client/subject mismatch: the same 401 with `error="invalid_token"`. Malformed Authorization: 400. Body over 64 KiB: 413. Origin not allowed: 403. Rate limit: 429. Verifier, key set or store unavailable: 503, never anonymous |
| same | GET, DELETE | 405 |
| `https://{host}/.well-known/oauth-protected-resource/mcp/c/{uuid}` | GET | **Stateless public protocol metadata (r3):** `{resource: R, authorization_servers: [MCP_CONNECT_ISSUER], bearer_methods_supported: ["header"], scopes_supported: (C7)}`, built only from a registered host and a syntactically valid canonical UUID path. There is no database lookup and no tenant name, and it carries no promise that the UUID exists or is active. Unregistered or malformed host, or a malformed path: 404. Spec R1's permitted public metadata; never authorization |
| `https://{host}/connections/mcp` (web) | GET | Setup page: sign-in-first when anonymous; generic "not available" for non-members; otherwise the member's own connections in this Organization (URL, status, bound client ID, created and activated times) and a Create button |
| `/connections/mcp/create`, `/connections/mcp/revoke` (web) | POST | Existing form-route order (method, form encoding, exact Origin, then core checks session and CSRF) |
| `/internal/v1/mcp/connections` (core, internal) | GET, POST | List own, or create pending, for `{origin}`; credential only from `Authorization: Session`; CSRF on POST |
| `/internal/v1/mcp/connections/revoke` (core, internal) | POST | Revoke one **own** connection `{origin, connection_id, csrf_token}`; connection-local |

The authorization-server metadata, consent and token endpoints are WorkOS's. Core exposes none.

### 3.3 Token verification (R1 clarification)

Implemented in `core/internal/identity/workos/connect.go` (provider JWT libraries stay in that package;
B-IDENTITY-GUARDS), reusing the bounded key-set source (`workos/jwks.go:51-109`) and the validator pattern
(`workos/token.go:51-70`).

- **Keys:** from `${MCP_CONNECT_ISSUER}/oauth2/jwks`, or `MCP_CONNECT_JWKS_FILE` with `APP_ENV=test` only (C10).
  Algorithms come from an allow-list (default RS256). An unknown `kid` triggers one rate-limited refresh. A key set that
  is unreachable with no matching cached key gives 503; a reachable set without the key gives 401.
- **Claims:**
  - `iss` must equal `MCP_CONNECT_ISSUER` exactly;
  - `aud` must be exactly one value, equal string-for-string to the connection's stored resource;
  - `exp` in the future and `iat`/`nbf` not ahead, each within 30 s leeway;
  - `sub` and `client_id` present;
  - **an `act` (impersonation) claim is refused**, exactly as the existing G004 validator refuses it
    (`workos/token.go:24-25`, `:81`). The token namespace change does not weaken this identity boundary.
  `sid`, `org_id`, `auth_time` and `scope` are ignored for identity and authority.
- **Refused token types:** ID tokens (audience is a client ID), AuthKit session tokens (G004 issuer, no resource
  audience), internal values and browser cookies all fail the issuer or audience checks. QP-6 records only the
  token-shape facts that support this; the validator's actual refusals are implementation unit tests
  (U-MCP-CONNECT-TOKEN).
- **Owner selection B (r5):** none of these checks changes. A's unchanged access token presented at B fails the exact
  audience check, and A's browser session or any other bearer fails the issuer or type checks. A separately issued
  token whose `aud` is exactly B's resource is a different credential: it is evaluated against B's stored user, client,
  host, Organization, liveness and current permissions, even if the provider refreshed it with a credential that also
  serves A. No generic environment-client or wildcard audience is accepted.
- **Subject mapping (namespace rule for the ADR):** `sub` maps to the canonical user only through an existing
  principal row (configured G004 issuer, `sub`) via the E2 reader (`00006_row_level_security.sql:319`,
  `identity/store.go:631`). There is no user creation and no email linking (ADR-003 :18). No row gives 401.
  `MCP_CONNECT_ISSUER` must belong to the same WorkOS environment as the G004 binding (QP-4).

### 3.4 Setup, activation and renewal

- **Create:** core checks:
  - the origin is a registry host origin and in `IDENTITY_PUBLIC_ORIGINS`;
  - the browser session (E1) and CSRF;
  - `Resolve(registry Organization)` and membership.
  It mints `uuid` and, in one TENANT-ORG transaction (§4.6 M1), inserts the `mcp` session row and the connection row
  (`pending`, `pending_expires_at = now + MCP_PENDING_TTL`). Row-level security proves membership again. It returns R.
- **Activation:** the first protected request carrying a token that passes §3.3 for a pending, unexpired connection
  whose stored user equals the mapped user. One TENANT-ORG transaction (§4.6 M2) sets `active`, the `client_id` and
  `activated_at` only if the row is still pending and unexpired (a single conditional UPDATE: one winner). It emits
  `mcp.connection_established`.
  - **Under owner selection B (r5):** the activating token may come from any provider family of the same user and
    client that can obtain this resource's audience. Activation proves the exact audience, mapped user, client and
    current membership; it does **not** prove that the presenting family was uniquely or freshly authorized for this
    connection. The authenticated browser setup (M1) remains the local bootstrap.
- **After activation:**
  - a token with a different `client_id` is refused;
  - any other family of the same client and user that presents this URL's exact audience (reauthentication on another
    device, a re-run of the client's authorization, or a refresh credential that also serves another connection) is
    accepted as the same connection under its binding and live authority (approved same-URL semantics; owner selection
    B);
  - the first-token client-binding race is disclosed: the setup page shows the bound client so the member can revoke.
- **Renewal:** WorkOS refreshes; core only verifies. A protected use renews the connection's idle expiry when renewal
  is due, before the tool runs (§4.6 M5).
  Credential expiry is not connection revocation.

### 3.5 Tool contract (unchanged from r1 except where noted)

The five tools keep r1 §3.5 exactly:
- names, input and output schemas with `additionalProperties: false`;
- errors `context_denied`, `group_not_selectable`, `invalid_input`, `unavailable`;
- the opaque Organization-bound cursor, selectability policy, idempotent no-ops, selected-Group-loss behaviour and the
  dotted-name compatibility caveat.

`logout` now revokes the connection's `mcp` session (terminal; the URL can never be reused). Repeat calls get 401.
Retired switching names are absent.

### 3.6 Transport bounds and configuration

- **Bounds:** as r1 §3.6 (64 KiB body, 16 KiB internal, 10 s deadline, C5, no wildcard CORS, SDK protections unless
  S0 shows a conflict).
- **Configuration (read once at start; documented disabled state when all unset):**
  - `MCP_CONNECT_ISSUER` (exact https issuer; http only with `APP_ENV=test`);
  - `MCP_TENANT_HOSTS_FILE`;
  - `MCP_ALLOWED_ORIGINS` (optional);
  - `MCP_PENDING_TTL` (2–60 min, default 30; OD-1);
  - `MCP_CONNECTION_IDLE` (1–90 d, default 30);
  - `MCP_CONNECTION_ABSOLUTE` (1–730 d, default 180);
  - `MCP_CONNECT_JWKS_FILE` (test only).
- **Startup refusals** (each a `core.exit config_invalid` naming variables only): partial configuration; MCP without
  identity; a host origin missing from `IDENTITY_PUBLIC_ORIGINS`; registry errors.
- No secret is introduced; `.env.example` is unchanged.

## 4. Binding, lifecycle and live authority (R2, R4, R5)

### 4.1 Data model (detail in §5)

- `app.sessions.kind` (`browser` | `mcp`). An MCP session row has no issued bearer (a discarded-random credential hash,
  a never-returned CSRF value). Idle expiry is the connection idle expiry and absolute expiry the connection lifetime.
  E1 serves `browser` only.
- `app.mcp_connections`:
  - immutable: `id` (the UUID), `session_id`, `user_id`, `organization_id`, `host`, `resource`;
  - `state` (`pending` | `active`);
  - `pending_expires_at`;
  - `client_id` (set once at activation);
  - `activated_at`;
  - `selected_group_id`, `context_version`;
  - timestamps.
  No token, hash or refresh value is stored.

### 4.2 Authority matrix (r1 matrix adapted to Connect)

"Binding" means all of the following hold:
- a §3.3-verified token whose `aud` equals the stored resource;
- the mapped user equals the stored user;
- the path UUID equals `id`;
- the request Host equals the stored host, and the registry maps it to the stored Organization;
- for an active connection, `client_id` equals the stored client;
- the session is live and the user active.

| Operation | Binding | Membership | Selected-Group validity | Failure |
|---|---|---|---|---|
| Protected-resource metadata (r3: stateless) | none: registered host and a valid canonical UUID path only; no database lookup; no existence signal | — | — | 404 only for an unregistered or malformed host or path |
| `initialize`, `notifications/initialized`, `ping` (C9) | yes (active) | — | — | 401; static protocol content only |
| `tools/list` | yes (active) | yes | yes | JSON-RPC `context_denied`; no catalog |
| `get_current_user`, `get_current_context`, `list_groups` | yes (active) | yes | yes | tool `context_denied` |
| `switch_group` with a Group | yes (active) | yes, locked | current selection valid (if any) and target selectable, locked | `context_denied` / `group_not_selectable`; unchanged |
| `switch_group` with `null` (explicit clear: Spec exception) | yes (active) | yes, locked | not required | `context_denied`; unchanged |
| `logout` (self-termination: Spec exception) | yes (active) | not required | not required | repeat 401 |
| Activation (first token, pending) | yes (pending, unexpired; client not yet bound) | yes (RLS) | — (starts Organization-only) | 401; stays pending until expiry |
| Setup create/list/revoke (web → core) | browser session + CSRF (mutations) + registry origin | yes (create, list); revoke needs ownership only | — | sign-in-first, "not available", or 401/403 |
| Provider refresh at WorkOS | not a core operation (R5) | — | — | the next protected call applies every check |
| Verifier, key set or store unavailable | — | — | — | 503, never anonymous or cached |

Unknown, terminal (revoked, expired pending, expired idle or absolute) and missing connections fail every protected row
with 401. The stateless metadata row deliberately says nothing about them. The only exceptions are the Spec's explicit
clear and exact-binding self-termination.

### 4.3 Request flow

1. Exact path and registry host check, else 404.
2. Verify the token (§3.3, no database).
3. E2 maps the subject.
4. Open one transaction in the registry Organization's TENANT-ORG shape and read the connection by `id`. Row-level
   security admits only the mapped user's rows in that Organization (r3, item 4). Check the binding, session liveness
   and user status. Then, according to
   the matrix: activation (M2), a read-only tool snapshot, or a guarded write.
5. `access.BeginVerifiedRequest` creates a fresh per-call scope from the identity-minted reference (session UUID,
   user). Selectors come from the record only. `end` runs after the call.

### 4.4 Selection ordering

As r1 §4.4: a READ COMMITTED guarded write with lock order users → sessions → organizations → Organization binding →
Group bindings and Group rows (current and target) → connection row. O-MCP-1..4 tested with inert hooks.

### 4.5 Lifecycle and live access (Spec R4, R5 and clarifications)

| Event | Effect |
|---|---|
| Token expiry; client refresh at WorkOS (same `aud`, user, client) | accepted; URL, binding and selection unchanged |
| Reauthentication by the same user and client for the same URL (new family) | accepted as the same connection; URL unchanged |
| Same URL from a different client after activation | 401; create a new connection |
| Pending setup expired (OD-1: 30 min) | terminal; replacement URL |
| `logout` tool or web revoke | terminal (`mcp` session revoked `logout`); WorkOS refreshes still verify but are refused locally (R5). **Local only (owner selection B):** it does not revoke the provider grant, and the provider family may still obtain tokens for the user's other already-authorized connections; it is not compromise containment |
| Suspected compromised client (RV-1) | the separately authorized procedure of §4.7: hold and local invalidation, qualified provider revocation for the user and client, release only after confirmation |
| Logout-all (settled; existing snapshot contract) | every MCP session of the user visible at the `revokeAll` statement snapshot becomes terminal; a connection whose creation (M1) commits after that snapshot survives, is disclosed, and stays listed and revocable (§4.6) |
| Account disable | every MCP session of the user becomes terminal under the existing user-row serialization; re-enable revives none |
| Idle or absolute expiry (OD-1: 30 d idle, 180 d absolute) | terminal; a new connection and URL are needed |
| Role downgrade | next protected call uses the current role; only the affected capability is lost (no reconnect or refresh) |
| Organization membership loss | every protected call and clear `context_denied`; `logout` still works; restoring membership makes the same connection usable again (it was not revoked) |
| Selected-Group access lost | normal calls fail closed; explicit clear and logout only; never an auto-select or auto-clear |
| Host reassigned or removed from the registry | registry/record mismatch, 401; replacement on the current host |
| G013 restore | every session expired (pending and active connections dead); post-checkpoint connections absent, so unknown and 401; tokens never re-create records |
| Unknown UUID on a registered host | stateless metadata (no existence signal); every protected call 401 |

**Not promised:** cancellation of an already authorized in-flight read, or undoing committed effects (Spec
clarification). G015 owns `tools/list` invalidation, change notifications and broader concurrency.

### 4.6 Control-state mutation boundaries (live-access clarification)

Each state-changing control operation commits only inside a transaction that re-reads and locks the authority it relies
on, following AUTHORIZATION's guarded-write order. None is authorized solely by an earlier adapter check.

| # | Mutation | Transaction and serialization point |
|---|---|---|
| M1 | Create pending connection | TENANT-ORG READ COMMITTED: user `FOR SHARE` → browser session `FOR SHARE` → Organization binding `FOR SHARE` → INSERT `mcp` session and connection (row-level security requires a non-null `rls_org_role()`). A concurrent membership removal, a revocation of that browser session, or a disable (users `FOR UPDATE`) serializes on those locks. Logout-all does **not** serialize all new rows (see below) |
| M2 | Activation | TENANT-ORG: user `FOR SHARE` → the connection's `mcp` session `FOR SHARE` (live) → Organization binding `FOR SHARE` → connection row `FOR NO KEY UPDATE` → conditional UPDATE `pending→active` (one winner) |
| M3 | Select or clear Group | r1 §4.4 guarded write (TENANT-GROUP for select, TENANT-ORG for clear); connection row last |
| M4 | Logout tool or web revoke | the existing `revokeSession` IDENTITY transaction (`identity/store.go:754`); revocation priority (M5 audit path); no membership needed |
| M5 | Idle renewal (r3) | One IDENTITY transaction **before tool execution**, at most one renewal per hour per connection. As in the browser precedent, it runs only when the binding read shows renewal due, so most calls write nothing. The single conditional UPDATE enforces everything in the write itself (below), so concurrent due calls renew once. It never reads-then-extends, never touches revocation and never exceeds absolute expiry |
| — | Read-only tools and lifecycle methods | REPEATABLE READ READ ONLY snapshot; a read authorized before a concurrent revocation may complete on its snapshot; the next call sees the revocation |

**M5 conditional renewal (r3, item 3).** The UPDATE sets
`idle_expires_at = LEAST(now + MCP_CONNECTION_IDLE, absolute_expires_at)` and `renewed_at = now` on `app.sessions`,
WHERE all of the following hold:
- `id` = the connection's session and `user_id` = the mapped user;
- `kind = 'mcp'` and `revoked_at IS NULL`;
- `idle_expires_at > now` and `absolute_expires_at > now`;
- `renewed_at <= now - 1 hour` (the throttle);
- `EXISTS` an active `app.users` row for that user.

It returns the new expiry.
- **One row:** renewed, so continue.
- **Zero rows:** a second read in the same transaction applies the same liveness predicates without the throttle.
  - Live: a **throttled no-op**, so continue.
  - Otherwise (expired, revoked, unknown session, wrong kind or inactive user): **terminal**, so the request is refused
    with 401 and the tool does not run.
- An expiry or revocation committed between the binding read and this write can never be extended, because the write
  itself re-checks it. The browser renewal at `identity/store.go:445-447` is the accepted precedent.

**Revocation writers (r3, item 2).**
- Disable, browser-session revocation and membership removal conflict with locks M1–M3 hold, so they either commit
  first (M1–M3 then observe them and refuse) or wait.
- **Logout-all keeps the accepted `revokeAll` statement-snapshot cut-off** (`identity/store.go:771-776`: "a session
  committed after this statement's snapshot is not affected"). It revokes every MCP session visible at its snapshot.
- If M1 holds the browser session `FOR SHARE` while logout-all's UPDATE waits on that row, the MCP session M1 inserts
  commits after the snapshot and survives. That connection was authorized by a session live at M1's lock time; it is
  listed on the setup page and is revocable.
- This is disclosed (E14) and tested (IT-MCP-ID-5). No global-revocation redesign or new lock scope is proposed. A
  stronger guarantee would be a separate escalation, not a claim about the existing code.
- Policy changes keep their existing statements and locks.

### 4.7 Compromised-client containment (RV-1, r5; r6 correction 1)

Normal logout is local. Containing a leaked or compromised user/client refresh credential is a separate, explicitly
authorized operator procedure: a bounded runbook, not a route, tool or registry. It reuses existing trusted functions
(Codex decision C12).

0. **Record prior state (r6).** Before any step, the operator reads U's existing account state (`app.users.status`
   and `disabled_at`, through the F-6 operator entry point or a read-only operator query). The operator also checks,
   by operator coordination, whether another disable or hold for U is unresolved. Both go into the operator record.
   - If U is **active**, this procedure always establishes its **own** hold in step 1, **even when another hold is
     open** (r7). A pre-existing hold then forbids this procedure from releasing U (step 4).
   - If U is **already disabled**, and only then, step 1's call is skipped: the existing disable already revokes
     sessions and blocks sign-in, setup and activation. This procedure establishes no hold of its own and must never
     release that state.
1. **Hold and local invalidation** (whenever step 0 found U active, whether or not another hold is open). Run
   `identity.Service.DisableUser(U)` (existing trusted method; store `identity/store.go:860`). Its G011
   revocation-priority audit records `user.disabled` and one `session.revoked` per session, or takes the gap path.
   - It revokes every session of U, so every pending and active MCP connection of U becomes terminal, and their
     already-issued access tokens fail locally.
   - It blocks sign-in, setup (M1 needs a live browser session) and activation (M2 needs an active user). That is the
     replacement hold.
2. **Provider revocation.** An authorized operator runs the operation qualified in §11.0B: delete U's authorized Connect
   application for client K (`DELETE /user_management/users/{user_id}/authorized_applications/{application_id}`, using
   the environment management credential), then reads back U's authorized applications.
   - WorkOS documents no event for this operation, so the operator record carries the operation, scope, time and
     outcome. No identifier values are written to the repository.
3. **Uncertainty.** A failed, non-2xx, timed-out or unconfirmed provider step leaves containment **unresolved**. U
   stays disabled (replacement held), and the step is retried or escalated. U is never re-enabled on uncertainty.
4. **Release (r6).** Run `EnableUser(U)` (existing; it records `user.enabled` and revives nothing) only when **all**
   of these hold:
   - this procedure established the hold in step 1;
   - provider revocation is confirmed;
   - the trusted operator explicitly clears this procedure's hold in the operator record;
   - no other disable or hold for U is unresolved.

   Otherwise U stays as found. A previously disabled account, or one under another hold, is **never** enabled by this
   procedure. After a release, U creates fresh connections through authenticated browser setup, with new UUIDs.
   There is no new hold registry, schema or public operation: the prior state lives in the existing user record and
   the operator record.

- **Disclosed side effects:**
  - every browser session and every MCP connection of U ends, across all clients and Organizations;
  - U cannot sign in while held;
  - the provider step, if qualification confirms it, ends every family of U with client K, including families serving
    connections U meant to keep. User × client is the documented granularity.
- **Unavoidable issued-token lifetime:**
  - access tokens already issued stay cryptographically valid until `exp` (the documented example `expires_in` is
    3600 s);
  - they are refused locally because their connections are terminal;
  - replacements have new resources, so old audiences never match;
  - whether old refresh values can still mint tokens is exactly what the §11.0B provider qualification must show.
- **Entry point:** `DisableUser` and `EnableUser` are trusted, in-process functions with no route. SECURITY.md already
  records an operator entry point for `DisableUser` as a production prerequisite (F-6). G014 adds none; the runbook
  depends on F-6.
- **Not added:** no MCP tool, global-logout tool, revocation registry, scheduler, hold table or identity platform.
  Logout-all (`revokeAll`) is not enough on its own, because it holds nothing.
- **Local proof** during approved implementation: IT-MCP-14 (§10.3), bound to the accepted provider facts.

## 5. Schema: migration `00010_mcp_connections.sql` (flagged; number provisional)

One goose transaction; additive; reversible in development and test; preconditions and postconditions copy the exact
version-9 and version-10 inventories.

- **`app.sessions.kind`:**
  - `text NOT NULL DEFAULT 'browser' CHECK (kind IN ('browser','mcp'))`;
  - `sessions_insert` replaced: browser rows unchanged; `mcp` rows only in TENANT-ORG with a non-null `rls_org_role()`;
  - E1 replaced with a `kind = 'browser'` filter through the 00006 SET-role chain; the auth reader gains SELECT on
    `kind`.
- **`app.mcp_connections`:**
  - `id uuid PRIMARY KEY` (core-minted; no default);
  - `session_id uuid NOT NULL UNIQUE` (FK sessions);
  - `user_id`, `organization_id` (FKs);
  - `host` (1–253);
  - `resource` (1–2048, UNIQUE);
  - `state` CHECK (`pending`, `active`);
  - `pending_expires_at`;
  - `client_id` (NULL until active; 1–2048, since a client metadata document's ID is an https URL);
  - `activated_at`;
  - `selected_group_id uuid NULL`;
  - `context_version bigint NOT NULL DEFAULT 0`;
  - `created_at`;
  - CHECK that `state = 'active'` exactly when `client_id` and `activated_at` are set.
  - No DELETE path and no foreign key on `selected_group_id` (rechecked per call, never auto-cleared).
- **Row-level security** (FORCE; policies `TO runtime, owner`):
  - SELECT `user_id = U AND organization_id = O` (r3, item 4).
    - Own rows of the context Organization only. A tenant-shaped transaction for Organization A cannot read the same
      user's Organization B connection, and the IDENTITY shape (O empty) reads none.
    - It requires no `rls_org_role()`, so the exact-owner revoke (logout tool, web revoke) and the binding read for
      lifecycle methods still work after membership loss. The registry Organization supplies O.
    - A token for another user's connection reads nothing.
  - INSERT `user_id = U AND organization_id = O AND R IS NOT NULL AND state = 'pending'`;
  - UPDATE `USING (user_id = U AND organization_id = O AND R IS NOT NULL)`.
- **Guard trigger** (BEFORE UPDATE, SECURITY INVOKER):
  - the immutable columns stay unchanged;
  - `state` may only go from `pending` to `active`, once, setting `client_id` and `activated_at` from NULL, while
    `pending_expires_at` is in the future;
  - selection changes need `context_version = OLD + 1` with `selected_group_id` equal to the context Group
    (TENANT-GROUP) or NULL (TENANT-ORG);
  - any other change raises.
- **Grants (runtime):** column SELECT/INSERT as used; UPDATE of `state`, `client_id`, `activated_at`,
  `selected_group_id`, `context_version`; nothing else.
- **No new SECURITY DEFINER reader** (r1's E4–E6 removed): the connection is read under row-level security after the
  subject maps through the existing E2.
- **Audit delta:**
  - `events_attribute` replaced to add `mcp.connection_established` (trigger on UPDATE OF `state`,
    pending → active), `mcp.context_changed` (trigger on UPDATE OF `selected_group_id` when changed), resource type
    `mcp_connection`, and `access.denied` operation `mcp.request` with reason `resource_mismatch`;
  - two capture triggers;
  - no new revoke reason (refresh is WorkOS's).
- **Risk:** brief ACCESS EXCLUSIVE on `app.sessions` (column, policy); quiescent deploy order provision → migrate →
  code. Down refuses while any `mcp` session or connection row exists, then restores the exact version-9 catalog.
  00001–00009 stay byte-identical.

## 6. Audit, privacy and logging (R6)

### 6.1 Events (C2)

| Event | Producer | Scope |
|---|---|---|
| `session.issued` (existing) | trigger on the `mcp` session insert at setup (M1, strict) | global, subject |
| `mcp.connection_established` (new) | trigger on activation (M2, strict: an audit fault refuses activation) | Organization, subject, resource `mcp_connection` |
| `mcp.context_changed` (new) | trigger on select or clear (M3, strict) | Organization (and Group for select) |
| `session.revoked` (existing) | logout, web revoke, logout-all, disable (revocation priority; gap path) | global, subject |
| `access.denied` (existing) | facade refusals at setup, selection and protected calls | Organization when its gate passed |
| `access.denied` operation `mcp.request` / `resource_mismatch` (new) | facade independent recorder, when a verified, mapped subject presents a token at a host, connection or audience it is not bound to | global, actor = mapped user |

Not durable (operational logs only): pre-verification failures, unknown connections, expired pending setup, rate
limiting. There is no network or provider I/O inside an audit-holding transaction (key-set fetches happen before any
transaction).

RV-1 (r5) adds no audit action. The local procedure (§4.7) uses the existing `user.disabled`, `session.revoked` and
`user.enabled` events through the G011 revocation-priority path. WorkOS documents no event for the provider step, so
the operator record attributes it, as SECURITY.md already does for service-credential rotation.

### 6.2 Privacy

- No access value, refresh value, code, state or verifier ever reaches core logs, URLs, Results or artifacts. Core
  never receives refresh values.
- The connection URL is a selector shown to its owner. It is not secret, but it is not logged with user identifiers.
- New operational events (`mcp.request`, `mcp.connection`, `mcp.setup`) are catalogued in
  `infra/validation/tools/internal/logrecord/events.json`. Observed test values join the run's secret registry.

## 7. Repository transition (R6, AC11)

Carried from r1 §7:
- the B-MCP-SOURCE successor (positive SDK, mcp, access and tenancy linkage; exact SDK schema closure; one handler call
  site; retained negatives; main.go digest and routes);
- B-MCP-HTTP wording;
- the reusable-guidance edits (C4) with the B-SKILL-UNCHANGED reviewed delta.

r2 change: the main.go route set has no authorization or token endpoints. It mounts the `/mcp/c/{uuid}` handler, its
protected-resource metadata route and the internal connections routes, only when MCP is configured.

## 8. Files and modules

| Path | Change |
|---|---|
| `core/migrations/00010_mcp_connections.sql` | **new** (§5) |
| `core/internal/identity/workos/connect.go` (+ test) | **new**: Connect token validator on the existing key-set source |
| `core/internal/identity/mcp.go`, `mcp_store.go` (+ tests) | **new**: verified-session reference, create pending, activation, idle renewal, revoke own; the TENANT-ORG initializer use |
| `core/internal/identity/{identity,store}.go` | small: kind constant, initializer helper |
| `core/internal/identity/identitytest/fake_workos.go` | test-only: Connect issuer mode (key set at `/oauth2/jwks`, mints Connect-shaped tokens) |
| `core/internal/tenancy/mcp.go` (+ `store.go`) | **new**: selectable-Group page, connection-context read, selection UPDATE |
| `core/internal/audit/audit.go`, `audit_test.go` (r8, R0) | small: the two trigger actions `mcp.connection_established` and `mcp.context_changed`, resource type `mcp_connection`, operation `mcp.request` and reason `resource_mismatch` (§6.1); catalog pins 23/6/29 become 25/6/31 |
| `core/internal/tenancy/access/mcp.go` (+ `{access,context,errors,log}.go`) | **new**: `BeginVerifiedRequest`, `GetConnectionContext`, `ListSelectableGroups`, `SelectConnectionGroup`, inert ordering hooks |
| `core/internal/mcp/*.go` | registry, routes, metadata, middleware, tools, internal connections handlers, errors, handler; unit/integration/e2e tests |
| `core/internal/config/mcp.go`, `mcp_config_test.go` | **new**: `LoadMCP` (§3.6) |
| `core/cmd/server/main.go` | mount when configured (reviewed digest) |
| `core/internal/migrate/mcp_integration_test.go` | **new**: CT-MCP, MG-MCP |
| `web/app/connections/mcp/page.tsx`, `web/app/connections/mcp/{create,revoke}/route.ts`, `web/lib/mcp-client.ts` | **new** setup page and routes (Vestive skill; existing components) |
| `web/tests/unit/mcp-connections.test.mjs`, `web/tests/browser/mcp.spec.ts`, `web/playwright.config.ts` | **new** tests; `mcp-light`/`mcp-dark` projects |
| `infra/compose/identity.yml`, `infra/identity/{proxy.conf,emulator-seed.yaml,mcp-hosts.tsv,mcp-seed.sql}` | MCP environment for core and replica (test trust anchor); `/mcp/c/` and metadata paths on org-a/org-b routed to core; synthetic user `dave`; registry; tenancy seed |
| `infra/validation/tools/cmd/identitysetup/*` | per-run Connect test signer and key set in the identity volume |
| `infra/validation/{run.sh,lib.sh,required-checks.tsv,render-customizations.tsv}`, `infra/ci/{validate-controls.sh,profile-paths.tsv}`, `infra/validation/tools/internal/logrecord/events.json` | wiring, pins, counts; `foundation dir core/internal/mcp/` |
| version-10 reconciliation | as r1 §11.3, against the R0 pin list (§0.2), including G013's `recovery_integration_test.go` and the AUDIT-INV coverage in `access/audit_integration_test.go` (r8) |
| `contracts/mcp/control-plane.md`, `docs/foundation/MCP.md` | **new** wire contract and operating guide (setup, lifecycle, R5 semantics, host prerequisites, successor mapping) |
| `docs/SECURITY.md`, `docs/foundation/{IDENTITY,AUTHORIZATION,DATABASE-ISOLATION,AUDIT,MIGRATIONS,OPERATIONS,FOLLOW-UP,RECOVERY}.md`, `docs/foundation/VENDORS.md` | records; VENDORS adds the WorkOS Connect feature row (existing vendor); SECURITY.md §6.2 moves from planned to implemented with the accepted RV-1 facts; OPERATIONS.md gains the bounded RV-1 compromised-client runbook (§4.7) (r5) |
| `skills/saas-foundation/...` | as r1 §7.2 |

**Unchanged:** `.env.example`, `infra/compose/local.yml`, images, `core/go.mod`/`go.sum` (S0), `web/package*.json`,
`.gitlab-ci.yml`, G013 SQL (P-G013-3), ARCHITECTURE, ADRs, the contract and every Codex-owned file.

## 9. Interfaces and contracts summary

- **Public:** §3.1–§3.5 (per-connection URL, metadata, setup page, tools), documented in `contracts/mcp/control-plane.md`.
- **Internal:** the connections endpoints (§3.2).
- **Go surfaces:**
  - workos: `ConnectValidator` returning a provider-free verified claim set;
  - identity: `VerifiedSession`, `MCPConnection` (unexported fields), `CreateMCPConnection`, `ActivateMCPConnection`,
    `RenewMCPConnection`, `RevokeMCPConnection`, `ListMCPConnections`;
  - access: `BeginVerifiedRequest`, `GetConnectionContext`, `ListSelectableGroups`, `SelectConnectionGroup`.
  All are pinned by the boundary tests.
- **Database:** version 10 (§5); no new role and no new definer reader.

## 10. Tests and acceptance-criteria mapping

### 10.1 Unit (core check image)

- **U-MCP-CONNECT-TOKEN:** algorithm allow-list; exact issuer; single exact audience; time leeways; missing `sub` or
  `client_id`; `kid` rotation and refresh; key set unreachable vs missing key; ID-token and session-token shapes refused
  (the production refusals QP-6 does not test); an `act` claim refused, with the existing G004 impersonation cases kept
  unchanged (r3, item 4).
- **U-MCP-HOST:** registry grammar, one-to-one mapping, Host comparison, forwarded headers ignored.
- **U-MCP-URL:** exact path matrix.
- **U-MCP-META:** exact `WWW-Authenticate`, and exact stateless metadata for any valid canonical UUID on a registered
  host, built with no store dependency (a store stub that fails on any call); 404 for an unregistered or malformed host
  or path (r3, item 1).
- **U-MCP-TOOLS, U-MCP-ERR, U-MCP-BOUNDARY:** tools, closed error codes, import boundary.
- **U-MCP-CONFIG:** refusals, including C10 refused in production.
- **Updated:** U-ACCESS-BOUNDARY, U-BOUNDARY, U-RLS-BOUNDARY, U-CTX-MCP.

### 10.2 Identity integration (B-IDENTITY-IT, `ops`)

- **IT-MCP-ID-1:** create requires TENANT-ORG membership (row-level security proof) and a live browser session.
- **IT-MCP-ID-2:** activation is atomic under concurrent first tokens (one winner); client binding; expired pending
  refused; a wrong subject refused.
- **IT-MCP-ID-3:** E1 ignores `mcp` rows. The subject maps only through an existing principal; no user creation or email
  linking.
- **IT-MCP-ID-4:** logout, web revoke, logout-all and disable revoke MCP sessions; re-enable revives none; the
  audit-fault revocation path. **M5 renewal cases (r3, item 3)**, added to this existing test with no new concurrency
  suite:
  - due and live: one row, idle expiry extended but capped at absolute expiry;
  - inside the throttle and live: zero rows, then a live no-op, so the call continues with expiry unchanged;
  - zero rows, then terminal and 401, with expiry unchanged, for each of: idle expired, absolute expired, revoked,
    unknown session, a browser-kind session ID, another user's session, and a disabled user;
  - stale read: the renewal runs with a binding read taken before a committed revocation or expiry (sequential, with
    no new interleaving machinery). It extends nothing and refuses the call.
- **IT-MCP-ID-5 (r3, item 2):** the one create/logout-all interleaving, in this existing suite with the existing inert
  hook.
  - M1 is held after its browser-session `FOR SHARE` while logout-all for the same user waits on that row. After M1
    commits, logout-all completes and revokes the browser session and every MCP session visible at its snapshot (a
    connection created before it is terminal).
  - The connection M1 committed survives, as disclosed. It is listed on the setup page, and its revocation is terminal.
  - The same hold with a disable instead waits on the user row. The disable then revokes the new connection's session
    too (the existing user-row serialization).

### 10.3 MCP and facade integration (B-MCP-IT, `migrations`)

Real HTTP (`httptest`, production constructor), the official SDK client, the real facade and runtime role, and the
identitytest Connect issuer mode as the test signer.

**Fixture:**
- Organizations A (Groups A1, A2) and B (B1);
- **agency staff** `ownerA` (org_owner A) and `adminB` (org_admin B);
- **client users** `memberA1` (org_member A + group_member A1) and `cross` (A1 and B1 member);
- one non-member;
- registry hosts for A and B.

| Test | Covers |
|---|---|
| IT-MCP-1 | discovery and transport through the actual HTTP path (r3, item 1): per-connection 401 header; unauthenticated metadata identical in shape for a pending, an active, an unknown and a revoked UUID on a registered host, and the same metadata while the store is unavailable (no lookup), when protected calls give 503; 404 for an unregistered host or malformed path; then protected calls for the unknown and revoked UUIDs return 401 and run nothing; transport bounds; store/key-set outage 503 on protected calls |
| IT-MCP-2 | host and route: unknown host or path 404; Host differing from the stored host 401; registry reassigned (A's host to B) 401; forwarded headers ignored |
| IT-MCP-3 | setup and activation: member succeeds; non-member refused at setup and activation; wrong subject; expired pending; second client after activation refused; token-type confusion refused |
| IT-MCP-4 | binding: A's token at B's host or connection 401; C1's audience on C2's path 401; a token whose audience is a revoked connection's resource can never activate or use another connection (audience mismatch). Under owner selection B (r5), a separately issued exact-audience token for C2 (same user and client, signed by the test issuer) is accepted only under C2's binding and live authority, and never changes C1. No claim about which provider family minted it |
| IT-MCP-5 | identity and context output; fresh scope per call |
| IT-MCP-6 | listing: staff see all Groups of their Organization, client users only their own; bounds; no fan-out |
| IT-MCP-7 | selection, O-MCP-1..4, refusal cases |
| IT-MCP-8 | **same URL vs separate URLs:** two token families (same user and client) on one URL share selection and logout; two URLs keep isolated **stored** Group and lifecycle state (a call targeting C1 never changes C2's selection or terminates C2); browser logout and restart leave state. No credential-isolation claim (owner selection B) |
| IT-MCP-9 | **lifecycle:** refresh and reauthentication on the same URL keep binding and selection; terminal states refuse (logout, web revoke, idle/absolute expiry, expired pending, missing); a token the provider still treats as valid is refused after local revocation (R5); re-enable revives none |
| IT-MCP-10 | audit (AUDIT-MCP exact events; strict faults on activation and selection), privacy scan, and **TestMCPRecovery** (r8: the actual `infra/recovery/sql/invalidate.sql` at the bound base): apply that invalidation, then pending and active connections refuse, a deleted (post-checkpoint) connection refuses, and a fresh setup works |
| IT-MCP-11 | authority matrix: membership loss, selected-Group loss, verifier/key-set unavailable, across every row |
| IT-MCP-14 | **RV-1 local containment (r5), in the existing suite with existing fixtures:** user `memberA1` holds one pending and two active connections; another user holds one. (a) Trusted `DisableUser`: every pending and active connection of `memberA1` becomes terminal; pre-issued exact-audience access tokens from the test issuer get 401 and run nothing; the other user's connection is unaffected. (b) Hold: while disabled, setup is refused and activation of any pending connection is refused even with a valid exact-audience token. This is the replacement hold that stays in place on provider uncertainty; no code path re-enables automatically. (c) Release: after `EnableUser`, the old connections stay terminal; a fresh authenticated setup with a new UUID activates and works; old tokens still get 401 there (audience mismatch). (d) **Pre-disabled user (r6):** for a user already disabled before the procedure, following §4.7 (prior state recorded as disabled; no own hold; no release) leaves the user disabled with `disabled_at` unchanged. The user's connections stay terminal, setup and activation stay refused, and no `user.enabled` event is recorded. **Coverage (r7):** (a) also stands for an active user with another open operator hold, who is still disabled by the procedure; (b)'s absence of any automatic enable path keeps that user disabled; the rule that the other hold forbids release is an operator-record condition, reviewed with the runbook rather than tested in the database. The provider stage is qualified separately (§11.0B) and bound by reference |
| IT-MCP-13 | **Organization-bound reads (r3, item 4):** through runtime SQL as the runtime role, the TENANT-ORG shape for A with user `cross` reads A's connection but not `cross`'s B connection (zero rows, not an error); the IDENTITY shape reads none; after `cross` loses A membership, the same-binding `logout` still revokes A's connection while every other protected call is denied |
| IT-MCP-12 | **live access changes (clarification, only behaviour not already in IT-MCP-6/7/9/11):** after activation, without a new token — (a) downgrade `ownerA` to org_member: the next `list_groups` returns only own Groups; (b) remove `memberA1`'s A1 binding while A1 is selected: normal calls fail closed, clear succeeds, no auto-clear; (c) a role change between two calls of one connection is visible to the second call; (d) a write held at the M3 hook while a membership removal waits commits first, after which the next call is denied; existing interleaving hooks reused. No in-flight read cancellation asserted |
| E-MCP-1..3 | §11.6 |

### 10.4 Catalog and migration (B-MCP-CATALOG, `migrations`)

- **CT-MCP-1:** the exact version-10 delta.
- **CT-MCP-NEG** detectors, each yielding its code with the catalog unchanged:
  - runtime DELETE or UPDATE of an immutable column;
  - an INSERT or UPDATE policy without a non-null `rls_org_role()`;
  - a SELECT policy without the `organization_id` predicate (r3, item 4);
  - `sessions_insert` admitting browser rows in a tenant shape;
  - E1 without the kind filter;
  - a missing guard trigger;
  - an activation transition allowed twice;
  - a definer reader added on the table.
- **MG-MCP-1..5:** as r1.

### 10.5 End to end (B-MCP-E2E, `ops`, identity project) and browser scenarios

**Identity-run order:** B-IDENTITY-IT → MCP seed (emulator lookup of `dave`; superuser SQL from `mcp-seed.sql` with the
registry UUIDs) → B-IDENTITY-E2E (with `mcp-light`/`mcp-dark`) → B-IDENTITY-PROBES → **B-MCP-E2E** →
B-IDENTITY-PRIVACY → B-IDENTITY-CONFIG.

**Browser (`mcp.spec.ts`, axe):**
- M-NOAUTH: sign-in-first page;
- M-NONMEMBER: `carol`, generic page;
- M-SETUP: `dave` creates a connection on `org-a.localhost`, sees the URL and pending status;
- M-LIST: list and revoke own;
- M-A11Y;
- M-HANDOFF: browser sessions and the created URLs.

**B-MCP-E2E** (Go, tag `mcpe2e`, core check image in web's namespace; run CA; per-run Connect test signer from
`identitysetup`):
- E2E-MCP-1: activation and all five tools through the TLS edge at `https://org-a.localhost:543/mcp/c/{uuid}`;
- E2E-MCP-2: second connection on A and one on B (`org-b.localhost`) keep isolated stored state (no credential-isolation
  claim); a same-URL second family shares
  state;
- E2E-MCP-3: cross-host and cross-connection replay 401;
- E2E-MCP-4: the edge exposes only the MCP paths on tenant hosts; internal paths 404;
- E2E-MCP-5: web revoke and logout-all are terminal; browser logout (current) leaves connections;
- E2E-MCP-6: real core image refusals (C10 in production, http issuer, host origin missing from identity origins,
  malformed registry);
- E2E-MCP-7: observed values registered for scans.

### 10.6 Manual testing and operator feedback

Required only if OD-2 keeps a third-party client claim.
- **Readiness:** the named client is verified before execution. OT runs after focused b1–b3 pass and before final
  qualification, on a reachable HTTPS test tenant host explicitly configured in the registry, with execution-time
  staging settings that are authorized separately (not by OD-3).
- **OT-1** (real WorkOS staging Connect): setup in the web app → paste the URL → client discovery, sign-in and consent →
  five tools → refresh after token expiry with an unchanged URL → web revoke → refused → a new connection works.
- **Evidence:** client name and version, negotiated revision, per-step results, no values.
- Findings are fixed before final CI; a relevant change repeats OT-1.

### 10.7 Acceptance criteria

| AC | Evidence |
|---|---|
| AC1 | IT-MCP-1, U-MCP-TOOLS/META, E2E-MCP-1, OT-1 if claimed, QP-1/QP-2 (real Connect audience and refresh facts) |
| AC2 | U-MCP-HOST/URL, IT-MCP-2, IT-MCP-13 (Organization-bound reads), E2E-MCP-3/4 (host, Organization and audience mismatch; registry reassignment) |
| AC3 | IT-MCP-3, IT-MCP-ID-1/2/3, M-NONMEMBER, M-SETUP, E2E-MCP-1, QP-4 namespace (signed `sub` equals the WorkOS User `id`, §11.0B step 3) with the E2 principal lookup in IT-MCP-ID-3; staff and client fixtures |
| AC4 | IT-MCP-4 (direct A-token or session replay at B refused; a separately issued exact B-audience token evaluated only under B's binding), E2E-MCP-3. QP-3 is no longer AC4 evidence: it stays a recorded failure, superseded as a requirement by owner selection B |
| AC5 | IT-MCP-5, U-CTX-MCP, IT-MCP-11, IT-MCP-12(c) |
| AC6 | IT-MCP-6 (staff vs exact client-Group roles), E-MCP-1, IT-MCP-11, IT-MCP-12(a) |
| AC7 | IT-MCP-7 with O-MCP-1..4, IT-MCP-11, IT-MCP-12(b) (no auto-clear; explicit clear only) |
| AC8 | IT-MCP-8 (same URL shares; separate URLs keep isolated stored state; no credential-family claim), E2E-MCP-2/5 |
| AC9 | IT-MCP-9 (renewal and reauthentication keep URL; terminal refusal; R5 local refusal), IT-MCP-11, IT-MCP-12(d), IT-MCP-13 (logout after membership loss), IT-MCP-ID-4 (M5 cases), IT-MCP-ID-5 (logout-all snapshot), E2E-MCP-5; **RV-1:** the §11.0B provider qualification record (before execution) and IT-MCP-14 (local stage) |
| AC10 | CT-MCP/MG-MCP; IT-MCP-10 (audit events and faults; recovery invalidation of pending and active connections; deleted-state refusal; fresh setup); IT-MCP-ID-4 (revocation-priority audit); privacy and canary scans; §6.1 inventory in AUDIT.md |
| AC11 | B-MCP-SOURCE successor and negatives, U-ACCESS-BOUNDARY, identity and health regressions, §7 guidance edits with the B-SKILL-UNCHANGED reviewed delta; r1's embedded-issuer design is not implemented |
| AC12 | QP record (§11.0, including the failed QP-3 history), the RV-1 provider qualification record and receipt (§11.0B), OT-1 if claimed, focused and final path-selected runs, security review, B7 ledger with checkpoint 4, failed/reused evidence, cleanup receipt, Codex final ACCEPT and the G015 handoff; G021 closed |

## 11. Validation

### 11.0 Pre-acceptance provider qualification (QP; authorized under OD-3 on 2026-10-15; one bounded run)

**Outcome (r5; from the provider qualification review, bound above).** The run is complete, and cleanup was
independently verified.
- QP-1 **passed**: three code flows, each with its exact concrete audience.
- QP-2 **passed**: each ordinary refresh kept audience, subject and client, and every refresh rotated.
- QP-3 **failed** on its first negative case: a C1-family refresh requesting C2 (already authorized by the same user and
  client) returned a validly signed token whose audience was another concrete test resource. It stays a recorded
  failure and is never relabelled as passing. Owner selection B supersedes it as an acceptance requirement.
- QP-4 and QP-5 were **not reached**; QP-6 to QP-8 are partial.
- QP-4 subject mapping moves to the combined run in §11.0B.

The text below is the contract the run was authorized under, kept unchanged.

Required by the Spec amendment before Plan acceptance. It is not an execution attempt and not a CI ID. Nothing in this
section runs during planning, and this revision does not run it. OD-3 is granted with the exact values below; the run
is prepared only when separately dispatched, and its active phase waits for the owner's explicit Start. It reuses the
completed probe's pattern (`../results/ONEOFF-G014-workos-capability-probe.md`): a disposable listener outside the
repository, owned records, private raw values and verified cleanup. It adds no new probe framework, no production
validator implementation, and no G016 DNS or TLS provisioning.

**Pinned scope (r3, item 5; exact OD-3 values, r4).**

| Element | Pin |
|---|---|
| Disposable provider records | Exactly three, in the WorkOS staging environment, created for QP and deleted after it: **R-CLIENT**, one pre-registered public PKCE (S256) client with no client authentication value and exactly one redirect URI, the temporary callback; **R-IND-1** and **R-IND-2**, one path-wildcard Resource Indicator each, `https://org-a.localhost/mcp/c/*` and `https://org-b.localhost/mcp/c/*`. Nothing else is created |
| Not changed | Environment-wide registration settings (client ID metadata documents, dynamic registration), JWT templates, login, MFA, existing clients, callbacks, resources and users. QP does not depend on the environment-wide registration settings: it records their observed state only and assumes no value |
| Hosts H1, H2 | H1 = `org-a.localhost`, H2 = `org-b.localhost` (the exact OD-3 values). In QP they are **audience-only**: nothing is served there for QP, and no DNS, TLS, edge or future MCP server is needed or created. If WorkOS refuses an indicator for either host, QP stops and reports it |
| Resources | Three concrete URLs, each with a fresh random UUID generated at run time: C1 and C2 on H1 and C3 on H2. They are not created in any core database (no core runs) |
| Client and flows | R-CLIENT drives authorization code with PKCE, then refresh, from the disposable listener, which verifies signatures against the issuer key set as the completed probe did. The official Go SDK stays the automated client for implementation evidence (OD-2). QP needs no MCP server or SDK session, because its hosts are audience-only. OT (§10.6) later needs the actual reachable server |
| Callback transport | A temporary publicly reachable HTTPS callback, terminated by the same kind of temporary account-less tunnel the completed probe used, forwarding only to the isolated loopback listener. The entry is private and session-protected, state and S256 are bound, and callback and request logging are suppressed. HTTPS reachability is checked before the link is shared. No account, DNS, router or service installation |
| Operator | The owner signs in through a real browser on an explicit Start action. Human authentication and MFA are never bypassed |
| Deadlines | Setup idle cap **30 min**; the active operator phase is at most **30 min after the owner's explicit Start**. The exact deadlines and the owned process identities are recorded before the ready link is shared. Setup readiness is not operator readiness: no operator timer starts before Start. A one-shot owned-process deadline stops only the recorded listener and tunnel processes and runs cleanup. Transport retries (tunnel propagation, bot protection) stay inside that window. There is no automated second run after expiry or failure; another run needs a new authorization |
| Ownership record | Private and outside the repository until cleanup is verified: owned record identifiers, creation times, process identities, deadline and the baseline below. It is then reduced to a sanitized receipt |
| Original state | Before any creation, record by read-only API or dashboard observation: the lists of Connect applications and Resource Indicators, and the environment-wide registration settings as observed |
| Cleanup | Delete exactly the owned records by identifier; stop the listener and tunnel; remove token, state and verifier material and the temporary executable. **Verify** by independent readback: the lists equal the recorded baseline (owned identifiers absent, prior entries unchanged) and the observed settings equal the original state. Delete responses are not proof. Produce a compact sanitized cleanup receipt |
| Recorded values | Sanitized facts only: claim names, equalities and presence. No raw token, code, state, verifier, subject or identifier value enters the repository or chat |

**Checks.**

| # | Check | Pass condition |
|---|---|---|
| QP-1 | Authorize C1, C2 (H1) and C3 (H2) as separate families | each token's `aud` is exactly its concrete resource URL |
| QP-2 | Refresh each family | `aud`, `sub` and `client_id` unchanged; the rotation behaviour is recorded |
| QP-3 | **Negative:** refresh family C1 asking for C2's resource, then for C3's resource | refused, or the issued token keeps C1's audience; it never acquires the new audience |
| QP-4 | Subject mapping | the Connect `sub` equals the subject of the operator's existing G004 principal row; no email used; equality recorded, never the value |
| QP-5 | Reuse an already rotated refresh value | behaviour recorded; nothing inferred as reuse detection |
| QP-6 | **Token-shape facts only (r3)** | sanitized: claim-name sets; `aud` form (one value or a list) and whether it equals the resource or the client ID, for the access token and any ID token of the same flow; whether the Connect issuer string equals the configured G004 issuer. Recorded only. The production validator's refusals are implementation unit tests (U-MCP-CONNECT-TOKEN), not QP |
| QP-7 | Issuer, key-set path and algorithms; metadata `scopes_supported` minimum (C7) | values pinned for configuration |
| QP-8 | Per-host indicator acceptance and limits observed while creating R-IND-1 and R-IND-2 | recorded; any undocumented limit becomes an escalation |

QP-9 (combined host-and-path wildcard exploration) is **removed** (r3): the design uses per-host path indicators. A
failure of QP-1, QP-3 or QP-4 stops acceptance and returns to Codex and the owner. It is not a reason to fall back to
another issuer. Any departure from the pinned scope (another record, another host, a setting change) stops QP for a new
authorization.

### 11.0B RV-1 provider stage and remaining provider evidence (r5; documentation only)

This section holds the **RV-1 provider-qualification findings**. It is needed before implementation release (Spec
RV-1, provider gate). Status: public documentation reviewed; nothing was run, called or changed. A live qualification
needs the owner's OD-6 after Codex review.

**Sources (public WorkOS documentation, fetched 2026-10-15 UTC)**

| Ref | URL | Fetched |
|---|---|---|
| R1 | https://workos.com/docs/reference/authkit/user/authorized-application/delete.md | 19:35–19:36 |
| R2 | https://workos.com/docs/reference/authkit/user/authorized-application/list | 19:10–19:11, 19:35 |
| R3 | https://workos.com/docs/reference/authkit/session/revoke.md | 19:35–19:36 |
| R4 | https://workos.com/docs/reference/workos-connect/introspection.md | 19:35–19:36 |
| R5 | https://workos.com/docs/reference/workos-connect/applications/delete.md | 19:36 |
| R6 | https://workos.com/docs/authkit/connect.md | 19:36 |
| R7 | https://workos.com/docs/events.md | 19:36 |
| R8 | https://workos.com/docs/llms.txt (reference index) | 19:11–19:12, 19:36 |
| R9 | https://workos.com/docs/reference/workos-connect/token/refresh-token-grant | 19:08 |
| R10 | https://workos.com/docs/authkit/connect/oauth | 19:09–19:10 |
| R11 | https://workos.com/docs/reference/authkit/oauth-resource.md | 19:12 |
| R12 | https://workos.com/docs/reference/workos-connect/metadata/oauth-authorization-server.md | 19:12–19:13 |

**The supported provider operation (documented; containment effect unqualified)**
- `DELETE /user_management/users/{user_id}/authorized_applications/{application_id}`: "Delete an existing Authorized
  Connect Application" (R1).
  - `application_id` is "The ID or client ID of the application".
  - On success it "Returns an empty response".
- Readback: `GET /user_management/users/{user_id}/authorized_applications`, "Get a list of all Connect applications that
  the user has authorized" (R2).
- Granularity is one user × one Connect application. That matches the documented consent model: requests "for the same
  user and application skip the consent page" (R10).
- It is a management API operation run by an authorized operator with the environment management credential.
- **Documented gaps:**
  - its effect on outstanding refresh and access tokens;
  - propagation;
  - errors and idempotency;
  - there is no event for it (R7);
  - there is no Connect token-revocation endpoint (R8, R12);
  - the refresh grant documents no `resource` parameter (R9).

**Other documented operations (not proposed; listed with their side effects)**
- Deleting the whole Connect application (R5, the application `DELETE` operation) removes the client for every user,
  and its effect on issued values is undocumented. It is too broad for one user and only a last resort for a
  compromised client application itself.
- Deleting a Resource Indicator removes "any application consents granted against it" (R11). That is environment-wide,
  with an undocumented token effect.
- `POST /user_management/sessions/revoke` (R3) revokes one AuthKit user session. No link to Connect grants is
  documented, so it is not Connect revocation.
- User deletion is permanent and would destroy the G004 identity. It is not acceptable.
- Introspection (R4) accepts access or refresh tokens but requires a client ID and client secret. The public PKCE client
  has no secret, so introspection is not a documented check for it.

**Exact unanswered questions** (each must be answered by qualification or by a supported WorkOS answer before release)
- RQ-1: After deleting U's authorized application for client K, is every refresh value previously issued to U for K
  (all families) unusable, both for ordinary refresh and for refresh requesting another resource?
- RQ-2: Is the effect complete when the success response returns, or eventual? If eventual, within what bound?
- RQ-3: After a fresh provider authorization by U with K, do the pre-deletion refresh values stay unusable (no revival
  by re-consent)?
- RQ-4: Already-issued access tokens are locally verified JWTs, so they stay valid until `exp`. What is the Connect
  access-token lifetime? Documented examples show `expires_in` 3600 and an `exp` 300 s after `iat`.
- RQ-5 (narrowed, r6): does the selected identifier form (the **client ID**, which the runbook knows from the
  connection's bound `client_id`) work for the selected client type (a pre-registered public PKCE client)? The
  application-ID form, dynamic registration and metadata-document clients are later qualification, not release
  blockers.
- RQ-6: What does deletion return when no authorization exists, or on repeat? Does absence from the R2 list reliably
  confirm completion?
- RQ-7 (narrowed, r6): does the documented targeted operation (one user × one application) leave U's unrelated
  authorized applications unchanged against the recorded baseline? There is no multi-user campaign.
- RQ-8: Is any provider-side audit trail available? If not, the operator record is the attribution.

**Proposed minimal future qualification (QP-B; needs OD-6; one bounded provider-only run; not authorized here)**

It combines RV-1 with the unreached QP-4, as the Spec asks related observations to be combined. It needs no core
database, application UI or implementation.
- **Records:**
  - one disposable pre-registered public PKCE client with one callback;
  - one path-wildcard indicator on `org-a.localhost` (audience-only);
  - four fresh concrete resources R1–R4 under it.

  Cross-host behaviour is not needed for RV-1, and QP-1 already covered two hosts.
- **Operator:** the owner signs in through a real browser three times (R1, R2, and R4 after revocation), after an
  explicit Start. SSO may skip the password and consent.
- **Steps:**
  1. Baseline readback: the application and indicator lists and the advertised metadata.
  2. Authorize R1 (family F1) and R2 (family F2), and record:
     - the exact audience and `client_id`;
     - one ordinary refresh each, with rotation recorded. This also shows the client and request form work, so a
       later `invalid_client` or setup error is never revocation evidence.

     Keep only the current values, in memory.
  3. **QP-4, the accepted split (r6 correction 2).**
     - Compare the signed Connect `sub` with the WorkOS User Management `id` in the same environment: a read-only
       lookup of the user by the signed subject must return a user whose `id` equals it. Record equality only, never
       the value. No email lookup, identity creation or application database.
     - Pin the accepted G004 source chain:
       - the adapter requires the response user ID (`workos/adapter.go:112`) and validates the AuthKit token against
         it (`:115`);
       - the validator requires `sub` to equal that user ID (`workos/token.go:72`);
       - principals are keyed by (issuer, subject) with a `user_` subject (`00002_identity.sql:43-50`);
       - the E2 reader maps exactly (issuer, subject) (`identity/store.go:629`).
     - This proves the namespace only. The real existing-principal lookup and missing or foreign refusal stay E2
       implementation evidence (IT-MCP-ID-3, IT-MCP-3).
  4. **Baseline of unrelated authorizations (r6).** Now that the subject is known, read U's authorized applications
     and record the sanitized baseline **excluding** the disposable client. It must be unchanged after the deletion
     (RQ-7).
  5. Run the deletion once for (U, K), using the client-ID form (RQ-5). Record status and time, then read back: K's
     entry must be absent and the step-4 baseline unchanged.
  6. **Negative refresh (r6 correction 3).** At T+60 s after the deletion's success response, and again at T+5 min
     (inside the cap), try each pre-deletion current refresh value:
     - an ordinary refresh;
     - a refresh requesting the other family's resource;
     - a refresh requesting R3, which was never authorized.

     Classify every response by HTTP status and OAuth error code, never recording values:
     - **Definitive:** a supported grant refusal, meaning the token endpoint returns the OAuth `invalid_grant` error
       (RFC 6749 §5.2). Only this counts as revocation evidence for that attempt.
     - **Ambiguous, so the family is unresolved (r7):** every other outcome. That includes timeouts, transport
       failures, lost responses, 429, 5xx, `invalid_client`, request or setup errors, a resource refusal such as
       `invalid_target` (it refuses the resource, not the grant) and any unknown outcome.
       - Refresh processing may rotate the value before an error is returned. No status code is assumed to prove
         that the value was not consumed.
       - Only a documented, definitive pre-processing refusal could justify treating the value as unconsumed, and
         WorkOS documents none (RQ-6).
       - After any ambiguous outcome, that family's qualification is **unresolved** for the run. A later
         `invalid_grant` for that value cannot qualify it, and retries cannot erase the ambiguity. There are no
         retries and no retry framework.
       - Ambiguous outcomes are recorded and are never proof of revocation.
     - **Fail:** **any** successful token issuance from an old family. It stops qualification and leaves RV-1 failed
       or unresolved. A later refusal of that value's now-rotated predecessor cannot erase it.
  7. Freshly authorize R4 as U with K; it must work with the exact audience. Then repeat step 6 with the old values,
     including a request for R4, under the same classification.
  8. Record the `exp` of tokens issued before deletion (no values), as the unavoidable lifetime. Record the real
     propagation observation (definitive at T+60 s or not) without claiming an immediate guarantee.
  9. Cleanup: delete the owned client and indicator only, read back the baseline lists, the U baseline and the
     metadata, stop the listener and tunnel, remove material and write a compact receipt.
- **Pass and fail:**
  - Pass: every old-family attempt in steps 6 and 7 ends in a definitive `invalid_grant` (RQ-1, RQ-3; RQ-2 recorded),
    R4 works, the U baseline is preserved (RQ-7), and QP-4 shows equality.
  - Unresolved: any family with an ambiguous outcome in steps 6 or 7 (r7). That is not a pass and returns to Codex and
    the owner.
  - Fail: any token issued from an old family. That stops the run and returns to the owner; no other issuer or model is
    selected automatically.
- **Pins:** the same transport, deadlines (30-minute setup idle cap, 30 minutes after Start), ownership record, original
  state and verified cleanup as §11.0. One run, with no automated second run. No environment-wide setting change, no
  existing record touched, and raw values never stored.
- QP-7 values (issuer, key-set path, algorithms, the C7 `scopes_supported` minimum) are recorded passively in the same
  run. QP-5, QP-6 and QP-8 leftovers are not needed for RV-1.

### 11.1 Before every batch

As r1 §11.1: fetch and the exact-delta rule; the README secret-scan line; the Docker Desktop mount check; the
coordinator lock; B-PROTECTED with `--expect-hash` for SECURITY.md; the standard `run.sh` form with the G014 allow-list.

### 11.2 Batches

| Batch | Content | Wall (estimate, not a quota) |
|---|---|---|
| S0 | vendored SDK v1.8.0 contract (stateless semantics, context propagation, protections, negotiated revisions, schema closure); no repository change | 15–25 min |
| b1 | core check (gofmt, vet incl. tags, unit), web check, `run.sh lint lint-base lint-web lint-core provenance` | ~15 min |
| b2 | B-MCP-CATALOG, B-MCP-IT, reconciled inherited ITs, identity MCP ITs (focus stack) | 5–10 min per run |
| b3 | identity project: seed, `mcp.spec.ts`, B-MCP-E2E driver, privacy subset | 20–30 min per run |
| S-review | delegate 1 | — |
| OT-1 | if required | 30–45 min |
| final | reconcile, `run.sh all unit` (ci-controls auto) | 75–95 min |

Targeted reruns follow failures only. One fixture and setup are shared; there is no new campaign, oracle or benchmark.

### 11.3–11.4 Reconciliation and resources

As r1 (version-10 pins via `subsetFS`, grep inventory, conditional amendments; existing projects and port sets; fixed
fixture paths; 0600 private inputs).

### 11.5 Wiring and count pins ("After G013" confirmed at R0, r8; §0.2 P-G013-6)

New IDs: B-MCP-CATALOG and B-MCP-IT (`migrations`), B-MCP-E2E (`ops`). Changed:
- B-MCP-SOURCE, B-MCP-HTTP wording;
- B-WEB-BOUNDARY (`WEB_ROUTES` + `connections/mcp/page.tsx`, `connections/mcp/create/route.ts`,
  `connections/mcp/revoke/route.ts`; request-context allow-list);
- B-SKILL-UNCHANGED delta;
- B-IDENTITY-E2E matrix row;
- B-IDENTITY-GUARDS (C10 guard);
- `APPLICABLE` (`core/internal/config/mcp*.go`, `web/app/connections`, `web/lib/mcp-client.ts`, `docs/foundation/MCP.md`);
- events catalogue; version-10 pins.

| Pin | After G013 | After G014 |
|---|---|---|
| rows / migrations / ops | 114 / 31 / 13 | 117 / 33 / 14 |
| profiles fast / integration / full | 35 / 74 / 107 | 35 / 75 / 110 |
| full gate remainder (VA7) | 77 | 80 |
| VU1 full split / VU9 integration split | 35/77/108 / 35/43/74 | 35/80/111 / 35/44/75 |

### 11.6 Performance (PG-006)

OD-4 accepted (owner, 2026-10-15; PERFORMANCE.md PG-006 disposition bound): the bounded observations below, with
escalation; no benchmark, SLO, speculative index or cache.

| Operation | Work |
|---|---|
| any protected request | token verification (CPU, cached key set) + E2 (1 autocommit) + 1 TENANT-ORG transaction (connection row by primary key, session and user liveness, facts) |
| lifecycle methods | the same, without facts |
| `list_groups`, `get_current_*` | one read-only transaction (≤ 9 statements) |
| `switch_group` | one guarded write (≤ 12 statements + trigger) |
| create / activation | one write each |
| renewal | only when due: ≤ 1 successful IDENTITY write per hour per connection; a concurrent due call adds one zero-row UPDATE and one read |

Observations in B-MCP-IT on the fixture plus 30 extra Groups:
- E-MCP-1: EXPLAIN of both listing forms;
- E-MCP-2: EXPLAIN of the connection read;
- E-MCP-3: `pg_locks` during a held M2 or M3.
Statement counts are asserted constant in Group count. Unexpected fan-out, scans or locks print `E-MCP-TRIGGER` and
return to review. There is no WorkOS call per request (only rate-limited key-set refreshes), no new index beyond
constraints, no cache and no SLO.

### 11.7 CI selection and hosted gates

`core/migrations/`, `core/internal/{tenancy,config,mcp}/`, `infra/`, `skills/`, `contracts/` and
`web/playwright.config.ts` select **full** (+ **ci-controls**). MR fast is automatic; the optional gate is not requested;
the post-merge staging full gate is required (success or owner-canceled; failures fail).

### 11.8 B-SCOPE allow-list (final at R0, r8; imported by Codex at binding)

```text
# Create
^core/migrations/00010_mcp_connections\.sql$
^core/internal/mcp/[a-z0-9_]+\.go$
^core/internal/identity/mcp[a-z0-9_]*\.go$
^core/internal/identity/workos/connect(_test)?\.go$
^core/internal/tenancy/mcp[a-z0-9_]*\.go$
^core/internal/tenancy/access/mcp[a-z0-9_]*\.go$
^core/internal/config/mcp(_config_test)?\.go$
^core/internal/migrate/mcp_integration_test\.go$
^web/app/connections/mcp/page\.tsx$
^web/app/connections/mcp/(create|revoke)/route\.ts$
^web/lib/mcp-client\.ts$
^web/tests/unit/mcp-connections\.test\.mjs$
^web/tests/browser/mcp\.spec\.ts$
^infra/identity/mcp-(hosts\.tsv|seed\.sql)$
^contracts/mcp/control-plane\.md$
^docs/foundation/MCP\.md$
^management/v0/results/G014-(scope-allowlist\.txt|tenant-bound-mcp-connections-result\.md)$
# Modify: production and test helpers
^core/cmd/server/main\.go$
^core/internal/identity/(identity|store)\.go$
^core/internal/identity/identitytest/fake_workos\.go$
^core/internal/tenancy/store\.go$
^core/internal/tenancy/access/(access|context|errors|log)\.go$
^core/internal/audit/audit(_test)?\.go$
# Modify: boundary and version-10 reconciliation tests
^core/migrations/migrations_test\.go$
^core/internal/migrate/(integration|audit_integration|retention_integration|runner)_test\.go$
^core/internal/tenancy/(integration|boundary)_test\.go$
^core/internal/tenancy/access/(integration|boundary|audit_integration|recovery_integration)_test\.go$
# Modify: web, harness, fixtures, tools, catalogs and counts
^web/playwright\.config\.ts$
^infra/compose/identity\.yml$
^infra/identity/(proxy\.conf|emulator-seed\.yaml)$
^infra/validation/tools/cmd/identitysetup/(main|main_test)\.go$
^infra/validation/(run\.sh|lib\.sh|required-checks\.tsv|render-customizations\.tsv)$
^infra/validation/tools/internal/logrecord/events\.json$
^infra/ci/(validate-controls\.sh|profile-paths\.tsv)$
# Modify: docs
^docs/SECURITY\.md$
^docs/foundation/(IDENTITY|AUTHORIZATION|DATABASE-ISOLATION|AUDIT|MIGRATIONS|OPERATIONS|FOLLOW-UP|RECOVERY|VENDORS)\.md$
# Modify: reusable guidance (C4)
^skills/saas-foundation/references/(mcp-control-plane|requirements-catalog)\.md$
^skills/saas-foundation/validation/isolation/mcp/(tools|session|harness_test|mcp_test)\.go$
```

**Conditional** (only with a Codex-reviewed amendment):
- `core/internal/identity/workos/{jwks,token}.go` if the key-set source needs a shared constructor;
- attack-suite version pins;
- `web/tests/browser/support/*`;
- any reconciliation hit outside this list.

## 12. Delegation and sequence

- **Delegation:** one implementer (high). Delegate 1 (required, read-only security review after b2/b3) covers token
  verification and namespace mapping, binding agreement, setup/activation atomicity and the client race, the RLS and
  guard trigger, live-access boundaries M1–M5, the audit, the edge, logs and the setup page. Delegate 2 (optional,
  read-only) covers S0 and the QP record. There are no delegate edits. B7 is shared.
- **Sequence:**
  - **Done:** OD-3 → QP run (QP-3 failed) → assessment → owner selection B → dated Spec, ADR and contract amendments.
  - **Pre-acceptance (r5):** Codex review of this revision and §11.0B → OD-6 → QP-B, one bounded provider-only run
    with verified cleanup, separately dispatched, with the active phase on the owner's explicit Start → RV-1
    provider stage and QP-4 evidence → Plan ACCEPT.
  - **Execution:**
    1. R0 (done in r8, §0.2), then binding and execution release.
    2. S0, then 00010 and CT/MG.
    3. Identity (setup, activation, renewal, revoke), then the workos Connect validator.
    4. Tenancy and facade operations and hooks; then config, `core/internal/mcp` and main.go; then B-MCP-IT.
    5. The web setup page; the identity overlay, `identitysetup` signer, seed, `mcp.spec.ts` and B-MCP-E2E; then the
       harness.
    6. Docs, contract, guidance, SECURITY.md and RECOVERY.md.
    7. b1–b3, S-review, OT-1, reconcile, final.
    8. Result, safekeeping pushes, Codex review, MR.

## 13. Budget, forecast and scope

- **Budget:** provisional B7, checkpoint after attempt 4; planning, assessments and QP spend no attempt.
- **Forecast:** 2 attempts likely, 3 plausible (a QP surprise returns to review before any attempt).
- **Size:** about 7–9k lines, smaller than r1 because there is no authorization server, code or refresh issuance.
- **Split:** none required. Fallback split if needed:
  - G014a: migration, identity, Connect validator, facade, transport, B-MCP-CATALOG/IT;
  - G014b: web setup, identity overlay, E2E, OT, guidance, closeout.

## 14. Escalation items and risks

| # | Item | Class |
|---|---|---|
| E1 | WorkOS AuthKit Connect as the MCP authorization server, with per-connection resource indicators (existing vendor, new product feature and settings) | vendor/security architecture: owner direction given; Codex ADR + QP |
| E2 | Public contract: per-connection URL, protected-resource metadata, setup page and internal API, tools | public API/MCP contract |
| E3 | Migration 00010 (sessions kind, E1 replacement, `sessions_insert`, connection table, guard trigger, audit delta) | migration; G004/G008/G011 catalog change |
| E4 | Identity contract: `mcp` session kind; logout-all includes MCP under its existing snapshot contract and disable keeps its user-row serialization (settled, §4.6); Connect-to-G004 subject namespace rule (no email linking) | identity contract |
| E5 | Access facade surface and import boundary; server links SDK, access and tenancy | authorization boundary |
| E6 | Audit catalog 29 → 31; resource type; `access.denied` operation/reason | audit contract |
| E7 | Validation scope: +3 IDs, changed pins, identity overlay, `identitysetup` signer, events, one `profile-paths.tsv` row | validation/CI mapping |
| E8 | PG-006: OD-4 accepted (bounded observations with escalation) | performance policy |
| E9 | Reusable guidance (C4) | foundation guidance |
| E10 | **Host prerequisites:** real tenant hosts with TLS, edge routing of `/mcp/c/` and identity-origin entries per host are G016/deployment prerequisites; G014 uses explicitly configured test hosts only | G016 boundary / deployment readiness |
| E11 | G013 reconciliation: RECOVERY.md only; no G013 SQL change (P-G013-3) | cross-Goal |
| E12 | Dotted tool names | compatibility risk |
| E13 | **Provider behaviours still unverified (r5):** subject equality (QP-4), reuse (QP-5), token shapes, indicator limits, and the RV-1 effect of authorized-application deletion on every old family (RQ-1 to RQ-8). QP-1 and QP-2 passed; QP-3 failed | qualification; RV-1 is a release blocker |
| E14 | **Accepted semantics to disclose:** **owner selection B: one user/client refresh credential may obtain tokens for the user's other already-authorized connections (one same-host case observed; cross-host, unconsented and future-replacement reach unqualified), so separate URLs isolate stored state but not credentials, and normal logout is local, not compromise containment;** same-URL merging; first-token client binding; WorkOS refresh continuing after local revocation (R5); a connection whose creation commits after a concurrent logout-all's statement snapshot survives it, stays listed and is revocable (existing `revokeAll` contract, IT-MCP-ID-5); public metadata answers for any valid UUID on a registered host and promises nothing about existence | security disclosure |
| E15 | Test-only trust anchor `MCP_CONNECT_JWKS_FILE` (C10) | test configuration |
| E17 (r5) | **RV-1 containment side effects:** the procedure disables the user (all sessions and MCP connections end; sign-in held) and removes every family of that user with that client; it depends on the F-6 operator entry point | security disclosure; production prerequisite |
| E16 | **None:** new vendor, platform, service, image, database role, dependency (S0), Redis, product schema, Organization switching, G015 catalog work, G016 provisioning, embedded issuer | explicit none |

**Risks:**
- R-a: SDK behaviour (S0).
- R-b: the 2026-07-28 revision (claimed only if met).
- R-c: the named client's registration support (verified before execution).
- R-d: identity-project fragility on hosted DinD.
- R-e: version-10 interactions.
- R-f: scanner-sensitive text.
- R-g: the braces exception lapse.
- R-h: G013 interface changes (R0).
- R-i: WorkOS behaviour changes between QP and execution (record the configuration and versions; repeat the affected QP
  items only on a relevant change).

## 15. Open questions

- Q1 (owner): **OD-6**, authorization of the QP-B run in §11.0B, requested only after Codex reviews it. OD-1 (as
  amended), OD-3, OD-4 and option B are answered; OD-5 and OD-2 for QP are settled.
- Q2 (Codex): the ADR text, including the subject-namespace rule; confirm the r3 corrections to C7 (stateless metadata)
  and the five review items.
- Q3 (Codex): accept the `dave` emulator-seed addition for member setup evidence.
- Q4 (Codex): keep the identity MCP ITs in B-IDENTITY-IT (`ops`).
- Q5 (Codex, r5): accept C12, the DisableUser-based RV-1 local procedure (§4.7), and its side effects, rather than a
  new narrow hold record.
- Q6 (Codex, r5): accept the §11.0B unanswered-question list and QP-B scope as the RV-1 provider-stage contract.
  (Answered: QP-B accepted at `33b6750`.)
- Q7 (Codex, r8): confirm that G013's note "must join that invalidation and the replay kinds" is met by the
  kind-agnostic invalidation plus the existing `user.disable` and membership kinds, with no new MCP replay kind
  (§0.2). G014's planned SECURITY.md §9 and RECOVERY.md edits would record it. If Codex reads the note as requiring
  a new kind, that is a bounded escalation, not part of r8.
- Q8 (Codex, r8): accept the three allow-list additions (`core/internal/audit/{audit,audit_test}.go`,
  `access/audit_integration_test.go` and `access/recovery_integration_test.go`) as R0 facts.

## 16. Closeout and handoffs

- **Cleanup:** bounded, as before.
- **G015 receives:**
  - connection identity (UUID and `mcp` session), the immutable binding and `context_version`;
  - the selection ordering and M1–M5 boundaries;
  - the per-call scope pattern and the static five-tool catalog;
  - the stateless-transport note (notifications need a reviewed stream or re-list contract);
  - the live-access semantics (next protected call, no reconnect).
  G015 owns `tools/list` invalidation and notification and broader concurrency.
- **G016 receives:** the host registry contract, tenant-host prerequisites (TLS, edge routes, identity origins per host)
  and slug-rename replacement semantics.
- **G013:** the MCP data class.
- **G021:** verifies the reusable contract; it stays closed.

## Revision history

- r0, 2026-10-15 UTC: initial provisional Plan on staging `1b15499` (embedded issuer proposal). Read-only.
- r1, 2026-10-15 UTC: documentation-only revision for the r0 review (WorkOS comparison, reader policies, authority
  matrix, consolidated decisions); staging `3bbadd6`.
- r2, 2026-10-15 UTC: bounded revision for the owner's per-connection direction (handoff 9cd9c0f5) and the
  live-access clarification (handoff c145760a), on staging `3df04ea`.
  - Replaced the embedded issuer with WorkOS Connect per-connection tenant-hosted resource URLs.
  - Pinned setup, activation, renewal and terminal lifecycle, R1/R5 semantics and M1–M5 mutation boundaries.
  - Simplified the schema: no new definer readers, codes or refresh material.
  - Recovery now needs no G013 SQL change.
  - Added the QP pre-acceptance qualification (negative retarget, concrete audiences, subject mapping) and IT-MCP-12.
  - Mapped all twelve ACs.
  r1 and r0 are archived byte-identically. Read-only inspection only; no subagent, probe, provider change, credential,
  install, build, scan, test, Docker command or harness run.
- r3, 2026-10-15 UTC: bounded correction for the r2 review (CORRECTION LOOP, handoff 37443b0a), on staging `ee0f946`.
  - Item 1: protected-resource metadata is stateless public protocol data from a registered host and a valid UUID path;
    no lookup or existence signal; protected calls still deny unknown or terminal connections (IT-MCP-1, U-MCP-META).
  - Item 2: logout-all keeps the accepted `revokeAll` statement-snapshot cut-off; the M1/logout-all survivor is
    disclosed (E14) and tested (IT-MCP-ID-5); disable keeps its user-row serialization.
  - Item 3: M5 is one conditional UPDATE that checks session, user, kind, revocation, idle and absolute liveness and the
    throttle; zero rows is split into a live no-op and a terminal refusal (IT-MCP-ID-4 cases).
  - Item 4: connection SELECT is bound to the Organization; exact-owner revoke and `logout` after membership loss stay
    reachable; the `act` refusal is preserved (IT-MCP-13, U-MCP-CONNECT-TOKEN, CT-MCP-NEG).
  - Item 5: QP pinned to three disposable records, two audience-only hosts, a temporary callback, deadlines, ownership,
    original state and verified cleanup; QP-6 records token-shape facts only; QP-9 removed; no environment-wide change.
  - Recorded the settled OD-5, C1, C8 and OD-2-for-QP decisions; OD-1, OD-3 and OD-4 stay open.
  r2 is archived byte-identically; r1 and r0 are unchanged. Read-only inspection only; no subagent, probe, provider
  change, credential, install, build, scan, test, Docker command or harness run.
- r4, 2026-10-15 UTC: concise reconciliation of the owner's OD-1, OD-3 and OD-4 answer (handoff de09f174), on staging
  `827070d`.
  - OD-1 accepted: 30-minute pending setup (the `MCP_PENDING_TTL` default), 30-day idle and 180-day absolute expiry.
  - OD-3's exact QP values are pinned in §11.0: `org-a.localhost` and `org-b.localhost`, 30-minute setup and active
    caps, no automated second run.
  - OD-4 accepted.
  - New Spec, contract, PERFORMANCE.md and review-record hashes bound.
  No design or research change. r3 is archived byte-identically; r2, r1 and r0 are unchanged. Read-only inspection
  only; no subagent, probe, provider call, credential, install, build, scan, test, Docker command or harness run.
- r5, 2026-10-15 UTC: bounded revision for the owner's selection of option B (handoff f3952b03), on staging `fd21bdc`.
  - WorkOS login and MCP issuance are kept, with broader same-user/client credential reach accepted.
  - Exact token checks, the immutable Organization, live permissions and independent stored state are kept. There is
    no promise of independent refresh families, fresh password entry or fresh consent per connection. Normal logout is
    local.
  - The QP outcome is recorded: QP-1/QP-2 passed, QP-3 failed (history kept; superseded as an acceptance requirement).
  - RV-1:
    - §4.7 adds a local procedure from existing `DisableUser`, the provider operation and `EnableUser`;
    - IT-MCP-14 adds the local cases;
    - §11.0B records the provider findings, the exact unanswered questions and the proposed QP-B scope.
  - New Spec, Goal, Architecture, DECISIONS, ADR-005, contract, SECURITY.md and review-record hashes are bound.
  r4 is archived byte-identically; r3, r2, r1 and r0 are unchanged. Public documentation was re-read only for RV-1.
  There was no subagent, probe, provider call, credential use, install, build, scan, test, Docker command or harness
  run.
- r6, 2026-10-15 UTC: small correction for the r5 review (handoff 39c1b6a4), on staging `d1dc108`.
  - §4.7 records the prior account state, holds only an active account and releases only its own explicitly cleared
    hold; it never enables a previously disabled or otherwise held account. IT-MCP-14 adds case (d).
  - QP-B keeps the accepted QP-4 split: the signed `sub` equals the WorkOS User `id`, with the G004 source chain
    pinned; the principal lookup stays E2 evidence. U's unrelated authorization baseline is recorded and preserved.
  - Negative refresh requires a definitive `invalid_grant`. Other outcomes are inconclusive. Any old-family issuance
    fails the run and cannot be laundered.
  - RQ-5 and RQ-7 are narrowed. The r5 review hash is bound.
  r5 is archived byte-identically; r4 to r0 are unchanged. No research, probe, provider call, settings change, test
  or implementation.
- r7, 2026-10-15 UTC: two wording corrections for the r6 review (handoff 731b737d), on staging `641e715`.
  - §4.7 skips `DisableUser` only for an already disabled account. An active account is disabled even with another
    open hold, and that hold forbids release. C12 and the IT-MCP-14 coverage note match.
  - §11.0B step 6: no status code proves non-consumption. Any ambiguous outcome leaves the family unresolved, with no
    retries; a later `invalid_grant` cannot qualify it.
  - The r6 review, SECURITY.md (with §6.2 unchanged) and ROADMAP hashes are rebound. An R0 flag records G013's
    delivered MCP note.
  r6 is archived byte-identically; r5 to r0 are unchanged. No research, probe, provider call, test or implementation.
- r8, 2026-10-15 UTC: R0 reconciliation and execution-readiness record (handoff db0ce620) on staging `ef0676a`, which
  contains the G013 ACCEPT commit `aae75c1`.
  - The redundant planning commits were dropped in the rebase, because staging holds r7 and r0 to r6 byte-exact.
  - §0.2 R0 record: P-G013-1 to -8 and the recovery-tooling check hold; nothing is material. The replay-kind flag
    resolves with no new kind (Q7).
  - §11.5's "After G013" pins are confirmed.
  - §11.8 is final. It adds the audit catalog files and G013's `recovery_integration_test.go`, and
    `access/audit_integration_test.go` moves from conditional to listed (Q8). §8 gains the audit row.
  - TestMCPRecovery now names the actual `invalidate.sql`.
  - The gate table, the sequence and the bound inputs are updated: QP-B accepted, OD-6 settled, G013 ACCEPT.
  - The initial Result is drafted.
  r7 is archived byte-identically; r6 to r0 are unchanged. Inspection only: no implementation, harness, Docker,
  probe or subagent.
