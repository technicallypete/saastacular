> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G058 selected-design/readiness addendum — Codex review

2026-10-18 UTC. **CORRECTION LOOP — one bounded recovery correction.**
Owner choices D-4r1/D-15(i)/D-16(c) remain selected, not reopened. Root and two
bounded read-only reviewers assessed exact
`2eba9764120eae4729f9fa0eb78af223be6f5a40`; no tests, builds, probes or edits.
Submitted history preserved, not accepted execution inputs:

- `plans/G058-selected-design-addendum-r0.md` (361lines), SHA-256
  `c251477029d90900bb97ed7b80c953834fb8b0a1ee03acc60d1bac86309c4ace`.
- `plans/G058-scope-allowlist-r4.txt`, SHA-256
  `a33960ca903eca88f04408bada3f12de1ff7992fe3c2b43d5d97c2ba81518bdc`.
  Executable45-pattern body matches r3; only header changed.

## Resolved technical choices

R-1: existing request_id carries an opaque operator reference with truthful
unknown actor attribution; no third personal-data column required. Reject NULL
p_by explicitly as well as malformed strings; SQL regex tests alone can yield
NULL rather than raise.
R-2: existing superuser operations channel, no new role/credential/pool, explicit
guarded INVOKER functions and no ordinary-role column privileges are coherent.
Superuser can bypass guards by definition; do not claim protection against a
malicious database superuser.
R-3: readiness CHECK and closed-existing/default-ready-new migration order are
coherent; request-access/invitations remain available and client member→fresh
login→G018 Group/profile remains intact. Applicant responses disclose no
readiness details.
R-4: pre-ledger migration can re-close readiness; explicitly require migration
before runtime verification/reopen in the existing runbook. recovery.sh does not
itself migrate; this is a supported procedure/test condition, not new automatic
recovery orchestration. Do not assert a persistent hosted environment exists
without a verified named target; apply the rule to any actual existing install.

## One required correction: later backfill after an already-ready checkpoint

A.4:221–223 assumes an attestation after checkpoint means restored columns are
NULL. A checkpoint can already contain a prior resolved readiness fact; later
backfill/re-attestation can disappear on restore while that older ready fact
survives. A.2's supported procedure omits mandatory G013 ledger publication and
A.4 makes it optional (if operator appended). The restored Organization could
therefore remain open without the later departure restriction, violating AC5.

Bind replay coverage for every supported operator backfill using the **existing**
G013 membership.remove ledger path, with explicit safe ordering/completion and
unknown/interrupted outcomes; alternatively conservatively close readiness until
that missing history is verified. Do not invent another ledger/framework or infer
restrictions from retention-deletable audit history. Correct the recovery table.
Extend existing ENR-16 only: checkpoint already ready → later backfill → restore;
the restriction must be replayed or open readiness must be closed before access.
Preserve all earlier replay/no-op states and fresh-install behavior.

Provide one compact correction/addendum linked to these exact archived bytes,
not a full Plan rewrite. Correct the NULL-reference check in the same delta.
No implementation or validation campaign authorized by this review. Full Plan
ACCEPT/import remains pending this correction. Even afterward G019 reviewed
integration/current-target binding, resources and SDK qualification must satisfy
release conditions; G018 qualification/final ACCEPT gates dependent integration.
B7/checkpoint4 remains unspent, no reset; no G020 planning. No further owner
decision is requested for this implementation of the already-selected safeguard.

## Correction01 and selected Plan package — provisional design ACCEPT, 2026-10-18 UTC

Root and a bounded read-only recovery reviewer accept correction01 at
`c1c7e5e1`, SHA-256
`2cecc812ef74eaec91c8d16cae13f78cb661d87e8bb7df20650f0307a8736d0b`
(exact imported bytes verified; do not infer execution from this planning
disposition). No tests, builds, probes or installs.

The ready-checkpoint/later-backfill gap is closed by mandatory append-first
existing G013 ledger coverage, idempotent backfill and attest-last procedure;
live insertion closes readiness atomically; replay materializes the restriction
or closes the existing Organization. cmd_replay's stopped-service and bound
restore/invalidation gates remain intact, updated/state semantics unchanged.
NULL reference checks and actual-environment wording are corrected. Owner
choices remain selected; client member→fresh login→G018 own Group/profile intact.

Import accepted exact Plan r3 as canonical Plan, selected-design addendum and
correction01 together; preserve all archived r0–r4 histories and reviews.
Plans/results scope inventories45 match SHA-256 `6bcfbdca413b1c6da4fb37d66a98bea496c2887273dcd86010c7a9b18dad5898`.
This is **ACCEPT of the complete selected provisional planning package**, not
application execution, provisioning, acceptance of G019 source or final G058.

Implementation clarifications (not new planning blockers):
- Emit one final ledger status per replay entry; preserve updated/state keys.
- ledger verify prints ledger/verify/ok, not the head sequence. Obtain the head
  from the verified existing ledger format; do not parse a nonexistent CLI field.
- At R0 pin the actual Docker-first migration command on the named closed target,
  before replay/runtime verification/reopen, preserving receipts and fence.
- Reconcile actual G0198-function baseline, other counts/exports/scope, current
  Goal/Spec/Plan inputs and fresh-fixture readiness. Retain the SDK recipient-
  proof qualification/native hold before affected code/qualification; no probe
  or executor substitution is authorized here.

**Execution release withheld:** G019 is unfinished/not integrated, current-target
R0/initial Result/readiness incomplete, SDK projection unqualified. Reassess
after reviewed predecessor integration under the existing overlap policy; final
predecessor ACCEPT still gates successor integration/final acceptance. Do not
start application implementation, tests, Docker or live actions merely because
this Plan is now imported. B7/checkpoint4 unspent; no G020 planning. No new owner
question, replan or duplicate validation campaign needed.


## Integrated-G019 readiness R0 — bounded correction, 2026-10-18 UTC

**CORRECTION LOOP, not full replan.** Reviewed readiness6b14b3c8 on integrated
staging36f31917 (G019 merge8f39d72f, reviewed0cddd0df). Exact readiness imported
SHA256 cd81d09898acdf500561b443793e0e97faa95b4c07a73042ce49b902ea7d15de.
Accepted r3/selected addendum/correction01 bytes remain unchanged. Owner choices
remain settled; same B7/checkpoint4 unspent. G019 qualification2960419227 is
RUNNING, not passed or accepted. Its pending state does not by itself prevent
released isolated overlap; a canceled run never satisfies required qualification.
Affected failures require safe dependency reconciliation, not automatic broad rerun.

D3 is a DTO reconciliation only if identity verifies the sealed AddressEvidence
issuer/subject against the exchanged identity (reject zero/unverified evidence) before constructing the merged
identity.VerifiedAddress via AttestVerifiedAddress. Retain provider-bound evidence,
canonical UUID mapping (VerifiedAddress.UserID must equal enrollment handle user),
private construction/call-site pins and no client proof.
No new grant/trust model beyond the already-selected runtime design. SDK boolean
observation time is the adapter observation from this exchange, not an older auth
timestamp or provider historical verification timestamp; qualify
actual pinned response projection before affected code, no emulator-as-provider proof.

**Required D11 correction:** reject readiness's fail-closed-only retirement option.
FinalizeRetirement disables the user but leaves the pending enrollment row unrevoked;
EnableUser within30m could revive it. Preserve accepted Plan1.3step5/D18/ENR19 terminal
revocation: same FinalizeRetirement transaction revokes every unrevoked enrollment
credential for that user, matching DisableUser. Keep audit/unknown-commit/rollback
and actual-scope semantics. No new public API, role, grant, pool or schema decision.
Extend existing ENR negative: retire -> re-enable -> old enrollment credential still
unusable/no admission; do not prove merely while-disabled refusal. Cover atomic
rollback using existing fixture/fault seam where affected; no whole retirement suite.
Claude supplies a compact linked readiness correction/addendum before implementation,
not a rewritten Plan or new owner question.

**Scope amendment01:** correct nonexistent docs/OPERATIONS.md to actual
foundation/OPERATIONS.md; add only core/internal/identity/retire(_test)?.go for the
required lifecycle correction. 45->46 executable patterns, plans/results exact
SHA256 d69470b5e334eda44745804d746d8be9d4b69d681ffaa3377cc5ccee4b03f4a2.
Prior exact45-pattern bytes preserved in plans/G058-scope-allowlist-before-readiness-r0.txt.
This inventory is scope authority, not application execution release.

Other R0 line/count/importer pins and unused first-time fixture rename are
mechanical within approved behavior; migration00016 verified free. Reuse unaffected
G019/G018 evidence; do not silently substitute weakened catalog/count assertions.

**Preparation-only release now:** one existing owner may perform proposed read-only
SDK inspection against pinned cached v10.5.0 offline image/module, no install/network/
provider call, then record file:line and actual type/JSON fields. Follow native
worktree guard's explicit plain-command instruction; actual classifier refusal
stops that operation for the supported permission route, no substitute executor.
Prepare initial Result with baseline/scope/input hashes, isolated checkout/owner,
actual resource readiness, budget and selected early-risk slice. No application/
migration/test implementation or test campaign yet. Missing projection stops the
first-time path for review; no invented email verification. Qualification and D11
addendum return for recorded supervisor release; full replan unnecessary.
G019 retains its staging repairs/cleanup in parallel. Final G019 ACCEPT gates G058
integration/final ACCEPT. No live action or G020 planning.


## Readiness correction01 ACCEPT and early-risk execution release — 2026-10-18 UTC

**ACCEPT planning/readiness correction; RELEASE G058 attempt1 selected early-risk
slice1a.** Root and bounded security reviewer accept exact00a39599 correction
5a950784717ecc099012ac5c83496a6261f4ee89c8f0603d399cefbf7fb77755
and initial Result9ab8aaa04e9419e92661aa1d6f590aa18ab67723f024602cf84e0f97c7a90cd3,
both imported byte-for-byte. Prior accepted r3/addenda/history remain unchanged;
no application/test implementation has occurred or passed. Owner D4r1/D15(i)/D16(c)
settled, current integrated G0198f39d72f reviewed0cddd0df and technical baseline
1055c93d (subsequent release publication management-only); G018 final accepted.
G019 staging2960419227 still pending: overlap starts do not claim qualification.

D3 sealed per-exchange evidence checks zero/unverified/mismatch/zero-time before
attestation, canonical VerifiedAddress.UserID equals handle, atomic claim/lift/
membership/audit preserved. D11 same retirement transaction revokes user-wide
pending credentials before gap read, shared with DisableUser/revokeAll; re-enable
never revives old credential. Existing fault seam at added UPDATE proves zero
partial user/session/enrollment/audit writes. Preserve existing status/revocation
unknown-outcome semantics. No new authority/API/pool/credential design.

**SDK projection qualified for pinned shape only.** Root independently inspected
cached v10.5.0 source via read-only g019-a1-gomod mount, network disabled, temporary
--rm container; ziphash h1:eClqevZs8AihL0kQOuX1Zzi2WT+8PYrYAuXkGszQdN4=
matches go.sum. User alias models.go1020, AuthenticateResponse.User6023,
Email string7099/email and EmailVerified bool7101/email_verified confirmed.
No provider call/test campaign. Adapter observedAt from current exchange clock,
not older AuthTime or claimed provider historical verification timestamp.

**Scope amendment02:** add only exact core/internal/identity/workos/adapter_test.go
for real adapter projection verified/false/absent and observation-clock assertions;
fake-path-only evidence is insufficient for this changed trust boundary.
46->47 executable patterns, plans/results exact SHA256
f08ac956b0654d2216a6b83e2449fa5bfa8ba9940bf738cb1892ce5df446e3ff.
Prior46 preserved in plans/G058-scope-allowlist-before-amendment02.txt; prior45
archive unchanged. Scope is not permission for unrelated adapter behavior edits.

**Dispatch/release:** reuse single G058 implementation owner/branch from00a39599,
Opus5.5 high; recreate native isolated checkout when needed and verify actual path,
owner/HEAD/no conflicting writer before writes. Record current release/scope inputs
in initial Result without erasing its submitted snapshot. B7/checkpoint after4,
attempts0 -> start1 on first implementation; forecast4, no budget reset. No naming
blocker or new planner. Current Plan plus linked readiness correction controls.

Start slice1a in initial Result: migration00016/admission catalog, sealed identity
projection, first-time claim+lift and D11 revocation, core-first/no HTTP or UI yet.
Run task-bound build/cache preparation and fresh-fixture readiness within this
release using approved pinned inputs, shared mutable resources locked, isolated
read-only unit/static checks lock-free. Existing image presence is not task-bound
build evidence; rebuild changed images before tests, resolve only required pinned
resources for the actual batch. Record exact image/cache/container/volume ownership
and cleanup; don't prune unrelated images (minio:latest attribution is unproven).
No borrowing G019-owned cache as a mutable G058 resource. New evidence uses private
home platform-evidence/g058-a1. Q1/live targets not required or authorized.

Before first local batch, bind executable focused selectors to actual new test
names and required guarantees; do not treat ^TestEnrollment as permission for an
unexplained umbrella suite. Adapter tests + selected enrollment/retirement/fault/
claim/departure/catalog negatives first; unchanged G019 evidence reused. Existing
early auth/admission source/security review precedes UI composition under Goal
policy. Normal subsequent approved Plan sequencing continues under this bounded
release with review of actual early-risk evidence; no default full local profile,
extra hosted pipeline, duplicate control campaign or routine owner proceed prompt.
Stop affected chain on substantive failures; independent safe work may continue.

Safekeeping checkpoint publication/normal MR checks retain existing authority,
not merge ACCEPT. G019 owns parallel staging repairs/cleanup; reconcile actual
shared-source repair before affected G058 batch. Final G019 ACCEPT still gates
G058 integration/final ACCEPT. No live provider/DNS/TLS/deployment, unsupported
proof substitution, new scope/security-model change or G020 planning. Native
permission requests use exact-rule standing route, no alternate executor.


## Slice1a round1 and scope amendment03 — 2026-10-18 UTC

**Bounded correction/release to continue same attempt1, not Goal ACCEPT.** Reviewed
exact1ca50bcc technical diff and Result c5927cfd8e2fe7ae572f8618470ce18f91225a67446f1783c3c7ae41b9b367c6
(imported exact; prior snapshot preserved as prefix). Sealed provider-evidence
plumbing implemented; migration/claim+lift/D11 remain unwritten. Reported9/9 unit
PASS plus vet/format; private g058-a1 contains no retained test logs at review,
so counts are reported, not independently inspected runtime evidence. Record
compact actual commands/exits/counts and retain future focused logs; no rerun
of unchanged tests merely to recreate these artifacts.

Root plus bounded security reviewer dispositions:

- S1 approve exact access/integration_test.go version15->16 setup pin/comment.
- S2 approve exact invitations_integration_test.go update for owner-selected
  D15(i): current-schema runtime valid-recipient claim must succeed atomically,
  with correct membership/consumption, replacing obsolete runtime EXECUTE refusal.
  Retain all invalid principal/context/token/address/expiry/replay/authority
  negatives. No weakening to generic success; no extra version15 historical
  campaign needed solely to preserve an obsolete current-head assertion.
- F1 approve cw ownership for rls_invitation_pending, using existing invitation
  SELECT policy/privileges rather than new pr grant/policy. Preserve fixed SQL,
  boolean-only output, exact id/Organization predicates, search_path/PUBLIC revoke,
  reviewed EXECUTE/context bounds. Overall new DEFINER count remains8; owner
  distribution cw6->7/pr decreases1. Derive/pin actual catalog and policy counts;
  no broader reader or privilege grant. No new owner architecture choice required.
- F2 approve exact identity/enroll_integration_test.go (package identity with
  integration tag) for existing internal faultDB/newStore seam. Keep lock-free
  unit tests separate; do not export/widen NewStore or redesign production for tests.
- F3 accept implemented per-call minter-sink alternative as bounded internal
  contract deviation from ExchangeResult. Provider.Exchange return/other providers
  unchanged; identity-only constructor, synchronous validated adapter Mint, shared
  mutex/counter across copies, fresh sink per exchange and poisoned second mint
  preserve selected trust boundary. Add small assertions in existing unforgeability
  unit test for copied minter and concurrent two completed mints (join before
  readback), both must refuse; no race campaign or full provider suite required.
- F4 derive identity table/RLS catalog expectations from actual00016; retain current
  G019 invitations and all previous objects. Do not treat suspected stale pins as
  proven failures or silently remove security assertions.
- F5 reject temporary enrollment-refusal audit omission/misclassification. Move
  the exact already-approved reason/catalog and truthful refusal-recording wiring
  needed by slice1a into that slice before callback use. Keep existing NULL-context/
  scope/attribution/rollback/log-privacy rules, no identity_unavailable label for
  a known denial and no deferred mandatory audit. No new audit model/event family.

**Scope amendment03** adds only the S1/S2/F2 exact paths above:47->50 executable
patterns; plans/results SHA256
4f9dc632b9e228fa8a523e693f08bd9cdf33a8508c7b7e8735ceebf21e8b7b1f.
Prior47 preserved in plans/G058-scope-allowlist-before-amendment03.txt. Existing
technical scope otherwise unchanged. Capture these reconciliations in compact
Claude-owned Result/linked addendum; immutable accepted Plan bytes stay unchanged.

Resume single writer/core slice1a immediately after binding current scope/release,
no extra routine owner prompt/full replan. Same B7/checkpoint4 attempt1 in progress.
Use focused executable selectors from first batch: added minter units; actual
catalog/migration; runtime valid/invalid claim; D11 retirement/re-enable/fault;
changed audit refusal/atomicity. Retain accepted broader negative coverage; no
whole local umbrella/ops or new pipeline just for this scope record. Return
core source/evidence for early-risk security review before HTTP/UI composition.
G019 staging/cleanup remains predecessor-owned; G058 integration waits final G019
ACCEPT. No live provider action, new authority model or G020 planning.


## Slice1a round2 early-risk review and callback decisions — 2026-10-18 UTC

**CORRECTION LOOP for admission readiness; bounded core continuation released.**
Reviewed366caf83 partial1a source, root identity/host review and bounded SQL/security
plus audit/log reviewers. Exact Result imported SHA256
d5984bfc00e0ef50885d8f07ca31207109b7fdde59b4e5b8260f6883d58ef50c.
No G058 application integration/Goal ACCEPT or HTTP/UI composition release yet.

Accepted reviewed portions: sealed evidence/multiple-mint negatives; D11/D18
same-transaction bound/user-wide revocation and non-revival after re-enable;
partial00016 FORCE RLS/grants/narrow DEFINERs and trigger-bound durable departure;
atomic invitation claim+restriction lift+admitted proof with statement-fault
rollback. Root independently inspected hashed logs: minter6b69395d…514ff,
catalog0c0a29fe…a31d3, identity878e636f…31e2c, admission7235be92…11026,
corrected membershipf8f5ca84…22b8d4. Initial MEM10 failures are retained, not passes.
Request/decision/readiness setter/backfill/replay are still later scope; partial
migration object counts are not final expected counts. events_attribute differs
from00015 only by the four approved access-denial reasons at this checkpoint.

**Host decision:** access invokes the final guard inside its existing IDENTITY
inContextTx, before any invitation claim/admission write. Identity owns configured
resolver reuse: small CheckEnrollmentHost(ctx,q,h) in identity/enroll.go calls
existing s.cfg.Hosts.ResolveOrigin(ctx,q,h.Origin), where q is the passed tracked
admission transaction. access/enrollment.go may define a small identity-interface
assertion on s.id (existing trusted-service composition pattern); no new resolver,
constructor plumbing, pool or domains import. Nil/nonimplementing dependency fails
closed. Validate exact canonical/request/stored credential origin consistency,
Routing=true and same Organization; no routing/mismatch -> host_changed; resolver
error -> unavailable. No SessionHost/pool/pretransaction-only substitute. Reuse
G016 READ COMMITTED current-snapshot semantics; do not claim commit-time host locks.
Extend only existing method/call-site/test pins. Test same tx reaches resolver,
nonrouting/changed host prevents all claim/consumption/lift/binding/admitted writes.

**Audit/log decision:** add precisely three durable reasons to the closed catalog:
recipient_proof_unavailable, recipient_ineligible, session_invalid. The latter
maps enrollment_expired's missing/revoked/expired credential aggregate; do not
claim a disabled user or expired login transaction was proven. Current partial
reason62->65; final selected-design total65->68 when its three readiness reasons
also land. Operations unchanged63; derive each actual-phase count from source.
Reuse host_changed. Keep external error vocabulary unchanged.

Use one truthful phase-owned event. Consumed authentication/recipient-proof
failure stays auth.sign_in denied with the specific proof reason; normal T1/T3
authentication/provider failures retain existing ownership. T1 enrollment-policy/
intent denial uses planned enrollment.start/access.denied shape; T2 admission
uses access.denied enrollment.admit, NULL Organization/target, only an actually
established trusted subject. Narrow new reason/action combinations to these
reviewed enrollment producers; malformed inputs/unestablished identities must not
invent attribution. Known T2 denial must not additionally fall through the callback
signInFailure default to identity_unavailable or duplicate auth.sign_in. Genuine
infrastructure failures keep failed diagnostics, not fake policy denials.
Use actual semantic reasons in access.decision rather than write_conflict for
closed/restricted/host/proof/credential outcomes; extend existing log enum/pins.
Update00016 audit reason/shape/down restoration and Go catalog/parity; no new
framework/action family or weakening of old row shapes. Preserve independent
post-rollback denial recorder, append-failure diagnostics and atomic success audit.

**Remaining early-risk proof before callback completion acceptance:**
1. Host negatives above; exact closed audit/log mappings, one attributed denial,
   no duplicate sign-in; existing append-failure diagnostic plus catalog/down/log
   positive/negative controls. Reuse unaffected fixtures; small table extensions.
2. O-ENR-8 controlled claim/departure interleavings, both orders and in-flight.
   G019 claim locks invitation/user/bindings before enrollment body credential/Org;
   advertised body order is not whole-transaction lock order. Correct the compact
   design record. Existing class40 unavailable/aborted victim with complete rollback,
   no retry/partial grant, is an allowed abort outcome; do not promise deadlock-free
   success or change00015/G019 lock contracts speculatively. An unbounded hang or
   wrongly classified/partial result requires a bounded correction disposition.
3. One existing lostCommit seam proves T2 unknown outcome returns unavailable,
   and later guarded reconciliation reads durable state without duplicate grant/
   consumption/audit. Statement-fault ENR11 is not unknown-COMMIT evidence.
4. Admission followed by removal: old admitted_* is historical proof only. Existing
   SQL early already_member is not a claim of current membership. Prove no restored
   binding; callback/status must recheck current membership before presenting it
   as current or allowing ordinary sign-in, whose issuanceGuard remains authoritative.

Continue same single G058 owner/core1a attempt1 after binding these decisions;
complete callback T1/T2 core wiring with these guards, then return focused actual
source/evidence for early-risk acceptance before HTTP/UI composition. Scope50
unchanged; existing allowed identity/enroll.go, access/enrollment.go,00016/audit/log
and tests suffice; report genuinely new exact file scope instead of silent edits.
No broad suite/default ops, repeated unaffected runtime proof, new hosted pipeline
or budget reset. Reconcile merged G019 repair ce7ace65 into affected store_test/
detector/harness at safe boundary before next batch (retain G058 version16 pins).
G019 gate2960555581 verified RUNNING on ce7ace65, not accepted. Predecessor owns
qualification/repairs/cleanup; final G019 ACCEPT gates G058 integration/final ACCEPT.
No live provider/DNS/TLS/deployment or G020 planning.


## Round3 bounded callback decisions and single-worker continuation — 2026-10-18 UTC

**CORRECTION LOOP at28ab7d7d; continue same core slice/attempt1/scope50.** Root and
bounded security review settle the four stopped decisions; no full replan:
1. Add enrollment.start as a distinct operation: current63->64, reasons remain65
   (final68 with selected readiness reasons). Correct prior claim it already existed;
   update exact Go/SQL/Down/shape/parity producers together.
2. Use planned enrollment/api composition: one facade delegates identity.Admitter,
   main imports enrollment/api only. Explicit startup binding before serving; nil
   fails closed. No direct access import/new pool/runtime registry/early HTTP enable.
3. Reconciliation must check current authoritative binding, not historical admitted_*.
   Gone/replaced stored binding -> existing enrollment_restricted, no new grant or
   rewritten proof. Extend existing removal/reconciliation row; fresh authorized
   enrollment required for reinstatement.
4. G058 maps issuer_authority_lost to closed not_found/target_not_found; preserve
   G019 elsewhere. Consumed replay already maps to already_consumed: no duplicate
   implementation. Invalid input remains input failure, actual class40/store failure
   stays unavailable, no blanket conflict->success or false membership.

Host guard structure accepted: existing identity resolver called on same tracked
transaction before claim/admit, no pool fallback/network; callback must still bind
exact stored/request origin. Tighten audit operation-level NULL Organization/Group/
resource for every enrollment.admit reason including target_not_found, equivalent
start producer bounds; extend one existing shape negative. Preserve phase ownership.

Resume one fresh Opus-high worker from28ab7d7d/current reviewed inputs, no nested
workers, full-history rereads or broad tests. Retain focused host, removed-member,
lost-COMMIT and O-ENR-8 proof; no claim these pending DB checks passed. Prior worker
safely checkpointed/pushed/retired; do not restart its completed work. Use compact
current Result/review sections, preserve required initial governance/Goal/Spec/Plan
reading, source/evidence and300k/350k thresholds. No repeated reports for stale
completion notices. Actual coordinator compaction remains unverified, not a fake
execution gate waiver. G019 staging/final ACCEPT gates G058 integration; no G020,
live actions, scope expansion or budget reset. HTTP/UI held for early-risk acceptance.


## Round4 partial callback source review and focused proof continuation — 2026-10-18 UTC

**CORRECTION LOOP / bounded continuation at `2f02b572`; not slice or Goal ACCEPT.**
Exact append-only Result imported (SHA-256
`700943e861d3b855829de7a8b936aef49067cc931cc1d5c718b6a17b3425f6f3`).
Root verified the round4 commit has no executable paths outside scope50 and the
three private unit-log hashes match its Result: identity9/9, access7/7 and
audit/migrations14/14. Vet/format success is implementer-reported. No round4 DB
or logrecord execution is claimed; no runtime early-risk acceptance follows from units.

The four prior decisions are implemented: enrollment.start operation64; planned
enrollment/api startup binding without HTTP registration; current-binding
reconciliation; closed issuer-authority refusal. NULL-scope audit narrowing is
present. T1 host refusal keeps existing auth.sign_in/host_changed ownership;
policy/intent refusals use one enrollment.start denial. A disabled subject may
not obtain usable admission/session authority: test actual T2/bootstrap refusal,
not merely existence of an inert credential. HTTP parsing/composition and request
behavior remain later-slice work, not falsely completed by these core DTOs.

**Bounded source correction:** T1 currently sets `Invitation=pending`, so an
explicit stored/forwarded expired, revoked or consumed invitation on an open
Organization silently becomes an open admission. Preserve stored invitation
intent separately (`StoredInvitation != ""`); keep the mode-or-pending T1 gate
unchanged, and let T2 claim return its closed refusal. This is an invitation-path
contract defect, not evidence that closed enrollment admits unauthorized users.
Extend the existing selector/fixture table for these stale-intent cases and one
no-invitation open positive; no new error vocabulary or policy change. Root and
bounded security review found no additional startup/T1-T2 authorization blocker.

Continue one fresh Opus-high writer from this checkpoint, same attempt1/B7/
checkpoint4/scope50, no nested delegation or restart of completed work. Finish
existing early-risk test rows, then the exact focused runtime selectors in the
round4 Result: host/denial atomicity, stale/replaced binding reconciliation,
unknown COMMIT, controlled claim/departure interleavings, T1 Login/callback and
audit/catalog/down/logrecord checks. Prepare only existing pinned jsonschema
cache under the shared lock; isolated readonly units remain lock-free. Reuse
unchanged passing evidence; no umbrella/ops or full local campaign. Report exact
commands/exits/counts/source and preserve failed evidence; security-boundary or
new scope/contract changes stop the affected chain. HTTP/UI stays held for the
resulting early-risk review. Enforce300k checkpoint/350k safe handoff; no budget
reset, live actions or G020 planning. G019 staging2960555581 still running at
this verification; predecessor qualification/final ACCEPT gates G058 integration.


## Round5 early-risk evidence and bounded core continuation — 2026-10-18 UTC

**CORRECTION LOOP — one test-proof correction; continued approved core execution
released.** Reviewed `afdc8d25` against round4; no new production authorization
blocker. Exact Result imported SHA-256
`76262fe02707829c6865e7a9d24293d8be42a68c747abf3ed0bd09d97f252f26`.
Historical Result text preserved (one latest-round index row added, then appended
round5). Root and bounded source reviewer checked selector preservation and
host/audit/lost-COMMIT/current/replaced-binding/stale-intent assertions. Root
independently matched all nine private log hashes and their outcomes to Result.

Local evidence accepted for these actual selectors: identity units9; enrollment
11 rows plus three ordering subcases; MEM10/11/14; identity9 top-level tests;
EnrollmentCatalog8 and MigrationLifecycle10; logrecord3 after cache preparation.
Keep the reused-database ENR-INTENT fixture failure and missing-cache setup failure
as failures; neither is a security mutation negative control. Final delta-count
fixture rerun passed; reuse unaffected attributable passes without a full rerun.
Observed O-ENR-8 departure-in-flight returned explicit enrollment_restricted, not
an unavailable branch. These are local checks, not hosted qualification.

**One narrow required correction:** O-ENR-8 currently accepts any CodeUnavailable
as a class40 victim. In that alternative require the existing decision reason
`aborted` (existing lastReason helper) or actual SQLSTATE40 attribution; retain all
pending-invitation, zero-binding and unadmitted rollback assertions. Transport,
cancellation or unknown COMMIT must not satisfy that branch. Rerun only the
O-ENR-8 row after correction. Describe timed noncompletion before barrier release
truthfully; it alone is not independently observed PostgreSQL lock-wait evidence.
No new concurrency framework, forced source mutation or whole-suite rerun.

**Dispositions:**
1. Accept preparation of existing tools go.sum-pinned x/text v0.14.0 as the minimal
   transitive closure of already approved jsonschema/v6. No application dependency,
   go.mod/go.sum change, new vendor or scope amendment is implied.
2. Accept O-ENR-8 under TestEnrollmentAdmission, retaining actual selector and count
   mapping; moving it solely to match a historical filename provides no guarantee.
3. Planned B-ENROLLMENT-CATALOG/IT mapping remains mandatory before hosted
   qualification/integration. Include the colocation and actual recursive subtest
   counts. Local success does not substitute for missing hosted selection. No new
   check framework or speculative separate full run.
4. The temporary source-revert negative control was natively refused and NOT run;
   preserve the refusal, do not retry/wrap/route around it or call the unrelated
   fixture failure a negative control. Existing end-to-end stale-intent negatives
   and open positive establish the regression; no extra permission remedy needed.

Resume one bounded writer: correct the ordering alternative, run that row, then
continue the next approved slice1 core/request/readiness work and selected-design
correction01 within scope50. Retain actual lock/claim chronology and closed
readiness, transactional revocation/decision guarantees; escalate genuinely new
schema/security/API/scope design. Source acceptance here is enough to continue
unaffected approved work; do not restart core1a or duplicate evidence. Complete
remaining core slice before HTTP/UI exposure under the accepted Plan sequence;
this does not authorize routes backed by unfinished core/readiness behavior.
Remaining required core proof and HTTP/harness composition return at the next
coherent handback, not another owner proceed prompt. Same attempt1/B7/checkpoint4,
300k checkpoint/350k safe transfer, one implementation writer/no nested delegates.
Reconcile current source and G019 repair at safe target boundary. No default
umbrella/ops/full local campaign, live provider/DNS/TLS, G020 planning or budget
reset. G019 final ACCEPT still gates G058 integration/final acceptance; current
G055 planning remains independent. No Goal/slice-wide final ACCEPT is claimed.

## Round6 partial policy/readiness source ACCEPT — 2026-10-18 UTC

Root and bounded independent security review ACCEPT partial source `d7cb93a9`
against round5 `afdc8d25`. Exact Result imported SHA
`45b8a851053042c7b7944f82742a640753b112306fb828602939e827ba207451`;
prior text preserved, one new latest-round index row plus appended round6.
O-ENR-8 now attributes the unavailable alternative to the admission's own aborted
reason; unrelated transport/cancellation/unknown-COMMIT failures cannot pass.
Observed database waiter is not claimed as an identified admission backend.
Readiness/backfill remain PUBLIC-revoked SECURITY INVOKER with explicit superuser
checks; runtime mode-only grant and owner guard remain. Organization locking,
close-on-backfill and mandatory atomic audit inserts match the reviewed design.
External ledger verification remains operator procedure, not a SQL guarantee.

Independently verified final log hashes: O-ENR8
`4ccef273010207fa20b824fb02605aa377b5b74d6a3e2b2f75cd1e757b94cbe9`,
migration/catalog `4a07524c3e0dfa35d823ec3b9fc81ddea873b729ea3fad630f0a702e76549569`,
final units `204d0b9f4b0622f6f63052db3c79130f346ff10f330f782f5bd4b32d053ed41c`;
no failing final rows. Earlier four failed migration runs remain recorded.
No supervisor tests rerun. Mechanical 68-reason/42-action reconciliation accepted
within reviewed catalog work; not permission to change authority/error contracts.

Continue with ONE FRESH high-effort worker from this checkpoint: policy facade
and its focused ENR proofs, then AT-INV/audit docs/count mapping; remaining request,
replay and hosted enrollment selection follow the approved core sequence. Reuse
unchanged core1a evidence. No HTTP/UI until required core readiness. Known-red
AT-INV/catalog/harness consumers must be fixed and evidenced before integration;
this partial ACCEPT does not waive them or imply final Goal acceptance.

Reported 404k context exceeded the 350k safe-transfer threshold. Retire that writer
now; verify it is stopped and preserve its checkpoint/handback before cleanup.
Use the existing 300k checkpoint/350k handoff rule at the next safe boundary,
not another monitoring framework. No new substantial phase on that worker.
Attempt1/B7/checkpoint4/scope50 unchanged; G019 final ACCEPT still gates integration.
G055 implementation is independent; G020 and owner-held MR48 remain held.

## Round7 policy facade and audit parity source ACCEPT — 2026-10-18 UTC

Root and bounded independent read-only security review ACCEPT partial source
`98d2ef9d` versus round6 `d7cb93a9` (staging-only management/architecture drift
excluded). Exact Result SHA `4450ddd31794990a3ecf39f81cb377a335bdbeb121c54aaeb7251df65ee000ce`
imported; prior text preserved apart from appended latest-round index/history.
Policy read uses organization.read; write retains owner-only manage_policy,
Organization locking and same-transaction readiness read/check before open.
Invalid mode and narrow CHECK/error mapping retain closed handling. Audit capture
is mandatory/transactional; tests cover denials, no-write refusal, repeat no-audit
and re-attestation. AT-INV exercises all three new producers with exact attribution;
42-action/36-trigger parity matches 00016. Known-red round6 audit consumers now
have focused passing evidence; hosted qualification remains pending.

Independently verified log hashes: tenancy/access/audit units
`041147a763f300f4f4906755563ae5b9c463821bfa4d7befbdb78b8f5ca9d238`,
logrecord `43dd442b939696e3a7a7cecaa0def89ec1fb8142ec9558a59f8ee0821105584e`,
policy/audit DB `6385f7d91f67a62903a162d8aee0bdc2d92ac3426fef420aeb738662f5e998f4`,
logcheck `6bbd85960c8d751f7fbc5fa452305dff8801c850e0853f7f53b8af238ab5e273`;
no final failing test rows/logcheck failures. Earlier mechanical failures remain.
Local cached MinIO uses an older release than the pinned hosted source build:
accept evidence only with that explicit environment limitation, not equivalent
qualification of the pinned build. Required hosted jobs must use the pinned build;
no extra local image build/full campaign just to duplicate them.

Continue the already-running single round8 request-slice writer without restart
or duplicate dispatch. Requests/decision/revocation, replay materialization and
actual hosted enrollment selectors/counts remain required before integration;
HTTP/UI waits remaining core as approved. Same scope50/attempt1/B7/checkpoint4;
round7 early safe transfer (~160k reported) is appropriate for the next atomic
slice. No budget reset, integration or final Goal ACCEPT. G019 final ACCEPT,
G055/MR48 owner holds and G020 phase stop remain unchanged.

## Round8 design dispositions and c1/c2 split release — 2026-10-18 UTC

**Bounded design dispositions; parked SQL is NOT source ACCEPT.** Exact Result
`58695c7d` imported SHA `dc8b1209c741329b3fd5e22d56ddc8cd8d327e6234d6ad0b98041c86da6b1a12`.
Task source remains round7 plus documentation; parked local `7ebe584d` is a one-file
UNRUN SQL draft. No test, syntax, privilege or atomicity pass is inferred. Preserve
its branch/commit and record a scoped safekeeping checkpoint before relying on
worker cleanup; never discard it or re-create the whole draft from memory.

Root and bounded security review resolve the four questions within accepted r3:
1. Use the **13 explicitly named request columns** in §4.2, including authority_seq.
   Correct inconsistent15-column prose/counts; do not invent two new fields.
2. Keep enrollment.decided reason NULL per accepted catalog. Preserve exact action,
   request resource, applicant/decider and immutable decision metadata. Do not
   claim the event itself distinguishes approval from denial. A new reason/history
   schema would require its own reviewed contract, not silent expansion.
3. Keep expired pending requests occupying one_pending per §4.3 step6. Expiry still
   prevents consumption; repeat submission or later admin decision cannot refresh
   the lifetime or grant renewed authority. UI/status must communicate expiration,
   keep admin resolution possible and not promise automatic retry. This is an
   explicit usability limitation requiring admin resolution, not an expiry repair
   or new request lifecycle. Escalate a desired lifecycle change separately.
4. Keep applicant DisplayName nil when existing users RLS makes it unavailable.
   Show supported request/status/expiry identifiers only; do not infer names/email,
   introduce a global user lookup or widen profile reads for list convenience.

**Release one fresh writer for c1:** reuse/review parked SQL, pair it with minimal
migration/catalog/grant/helper/pin consumers, and run focused real PostgreSQL
catalog/updown/privilege and request-transition/mandatory-audit proof. Then submit
that coherent risky slice for source/evidence review. c2 follows with Go facades,
ENR/O-ENR ordering/authority/expiry proofs and actual AT-INV producers. This is a
bounded split of existing scope, not a new Plan or budget reset. c1 must not fake
producer success: declare any intermediate AT-INV45/45 mismatch until real c2
producers execute; do not lower/skip assertions or call the task fully green.
Reconcile changed consumer inventories in the same slice where technically
possible; distinguish SQL/catalog evidence from application behavioral proof.

Use compact exact handoff and relevant Plan sections, source delta and open
findings rather than another full-history read. Preserve prior green evidence,
remaining replay/hosted selectors and core-before-HTTP/UI order. Same
scope50/attempt1/B7/checkpoint4,300k checkpoint/350k safe transfer, one writer.
No unchanged full-suite restart, native workaround, new public lifecycle or
provider action. G019 final ACCEPT still gates G058 integration; G055 and MR48
merge holds until G058 final ACCEPT and G020 phase hold remain unchanged.

## Request c1 partial source ACCEPT and c2 release — 2026-10-18 UTC

Root and bounded read-only SQL/security review ACCEPT partial c1 source
`6a705da2`. Exact Result imported SHA
`230e1f36dce561b62aa3cfcbf53a2b0c628d6379dba6a2db19b2a827ab245f50`.
Request table, scoped grants/policies, state guard, authority watermark and audit
captures implement the reviewed request boundary. Admission checks live applicant
credential, exact subject/Organization/request, current mode/expiry/decider role,
newer restrictions and atomic membership/consumption/covered lift/proof.
Runtime cannot choose authority_seq or consumption fields. Mandatory append
failure rolls back the mutation; no exception-swallowing path accepted.

### Three dispositions

1. Accept guard equality of decided_by/revoked_by with ctx_user(): the acting
   manager is the recorded decider/revoker; no supplied third-party owner can
   preserve stale authority or misattribute the decision. Existing authority model,
   no new role or widened grant. Preserve negative tests and current-role checks.
2. Accept ON CONFLICT on the exact one_pending index, DO NOTHING then lookup.
   If the concurrent row cannot be read,40001 represents an aborted/transient
   attempt, not success. No internal retry introduced. c2 preserves existing
   class40/unknown-COMMIT distinction and adds the actual two-credential race proof;
   never map every unavailable outcome to safe retry or silently extend expiry.
3. Parked7ebe584d is a superseded draft, not qualification evidence. Keep its
   task-local branch/commit through current source safekeeping; it need not block
   c2. At final task cleanup verify supersession/retained unique work and remove
   finished task resources/branches under existing policy. No reimplementation
   of the draft, unnecessary force operation or extra publication pipeline.

**Issued-authority distinction:** accepted r3 and existing G019 invitation claim
recheck the issuing manager's current membership role, not active-user status.
Disabling a decider prevents their login/session execution but does not itself
invalidate an earlier approval while managerial binding remains. That approval
still requires expiry/revocation/current-role/restriction checks; applicant disable
is independently refused. A stronger issued-grant invalidation lifecycle would
need an explicit owner decision; none is silently added by this review.

Independently verified final logs: units
`5caa5443c0d98654b7d1204fea8b7ad85b5c6c7023f770ad4090dec813eae5fd`,
PG catalog/lifecycle `0b0d662e4bfd674f95210d9bf1cf67b33fb18b334df58b233bd9d74d675d7885`,
admission `1afd22144d1c2964e5790bd6e7136cb656b337b6aabda3861cac05e269af36b3`,
version16 populated stepdown
`fd7aae34e0a218b9b5c949fa3a69680c24a754892536d6c0bc6645b687912884`.
Audit/enrollment log `a74022f7a62592751b2ee3d73c6300adfe371d3a412f683fac9b1e733facbe96`
still FAILS AT-INV45/42 as declared; no full-green claim or relaxed assertion.
Earlier first-run failures retained; older local MinIO/pinned hosted distinction
remains. No supervisor tests rerun; static pins do not replace concurrency proof.

Release one fresh **c2** writer: approved request/decide/revoke/list/status/abandon
Go facades, errors/logrecord, actual AT-INV45/45 producers and matching wrapper,
ENR/O-ENR authority/expiry/race/rollback proofs. Reuse passed c1 catalog/migration
coverage unless changed inputs justify focused rerun. Return a coherent behavioral
handback before HTTP/UI. Replay materialization and hosted enrollment selector
mapping remain required; no default broad local campaign, new product lifecycle,
provider action or quota bypass. Same scope50/attempt1/B7/checkpoint4; exact source
binding at safe boundary,300k checkpoint/350k safe transfer. G019 final ACCEPT
still gates G058 integration; GitLab quota block and G055/MR48/G020 holds remain.

## Request c2 partial source ACCEPT; replay/harness release — 2026-10-18 UTC

Root and bounded read-only security review ACCEPT partial c2 source86a22c95.
Exact Result imported SHA
`e66e78dbfcad3d500ed46e4aa7537d8fe5f5fc6a81086af6ad5b4eac70482d38`.
Own-credential applicant operations retain exact subject/Organization binding;
manager list/decide/revoke use organization.manage_members and the Organization
write lock before request mutation. No supplied watermark or authority expansion.
Failure results are empty; transaction handling preserves class40 aborted versus
unknown COMMIT and does not retry internally. Actual race tests pin observed
waits/outcomes plus exact row/event counts, not only test-name presence.

### Three dispositions and one minor correction

1. Accept pending/denied/already_member as unrecorded status conflicts with
   write_conflict; do not invent a denial reason/catalog producer for ordinary
   status reconciliation. Preserve recorded real refusal/authority cases.
2. EnrollmentStatus admitted is **historical T2 commitment**, never current
   membership or session authority. It may direct the browser to ordinary fresh
   sign-in/current guards only; HTTP/UI must not use it to restore removed/replaced
   binding access. Dead-credential read need not append an independent denial.
   Abandonment revokes the credential, not the applicant's durable pending request.
3. Accept request_on_arrival callback wiring with the later HTTP/identity slice,
   as a pending requirement, not implemented behavior or permission to expose
   incomplete core. Include automatic request creation and closed-policy/error/
   replay/host checks in the eventual end-to-end mapping.
Minor correction: current abandonment SQL may revoke an expired but unrevoked
own row. Align comments with this safe revocation behavior; do not add a new
expiry predicate/design just to match inaccurate no-op prose. Idempotent repeats
and no authority widening remain; no full rerun needed for comment-only correction.

Independently verified final logs: request11rows/6ordering lines
`42ec4347ad07d642c8807da922727745fc87ed6b71e360b968303a74e505090f`,
audit9/9+admission11/11+policy2/2
`6e148c74f30dc30a499d49e5cf668d541b53b5d18917c165ac7f9f5be7311925`,
units `0da074c0f23c1133d6202fa4b612508da9e1af3b76fdf1d1288bfad080297c5f`,
changed authority row rerun
`1bc572f79fa7930ad7bc448a96be0bfc1a2ef1957b3550aa0f762e269cb75683`.
AT-INV45/45 is now observed from real producers; the earlier45/42 failure is
retained history. First-run PostgreSQL log was not captured; keep that evidence
limitation explicit. Other declared logs/scans and older cached MinIO limitation
remain attributed; no complete pinned hosted proof claimed. No supervisor tests.

Release one fresh writer for **(d) replay materialization/readiness close + ENR-16**
and **(e) actual B-ENROLLMENT-CATALOG/IT selection/required-check mapping** under
accepted correction01. Preserve append-first ledger provenance, conservative
post-replay re-close/restore behavior and current state/membership authority.
Inventory actual recursive selectors/counts for TestEnrollment, Admission,
Requests and Catalog; retain public log privacy and failures. Update only approved
consumers/fixtures; no CI trigger/runner redesign, broad default ops campaign or
quota bypass. Return source and focused replay/selector evidence at a coherent
handback. HTTP/UI follows only after core/replay/harness readiness, accepted
Plan/current binding and its own required tests. Reuse unchanged c1/c2 proofs.
Scope50/attempt1/B7/checkpoint4 unchanged;300k checkpoint/350k safe transfer.
G019 final ACCEPT still gates integration; CI quota and G055/MR48/G020 holds remain.

## Replay/hosted-selector partial review; re-attestation correction — 2026-10-18 UTC

Root and bounded read-only recovery/security review inspected370e47f7. Exact
Result imported SHA3984c01453eb99afa03a895e20c2b4f708d45612ac7b8dfe496ab752a093a003.
Replay delta partial source ACCEPT: original restriction/update/state predicates
unchanged; Organization lock protects chronology; same-transaction ledger insertion
or conservative readiness close; repeated replay converges. Pre16 path remains
closed through mandatory migration. No new restore orchestrator or live restore.

Hosted selection/registry mapping ACCEPT on reviewed source: new catalog/IT IDs
in required registry and MinIO-failure list, exact catalog selector and combined
access selector, recursive counts19 and2/14/15 reproduced from actual logs.
Root independently hash-verified catalog22b3c4ae…18b5, real-psql parity1f367dd8…f768,
IT21035853…e9f7. These are real-PG emulated restored-state proofs plus a real-psql
spot comparison, **not** a full recovery.sh restore/replay-drill pass. Required
hosted B-RECOVERY-DRILL and new B-ENROLLMENT rows remain pending. Counted three
psql statements instead of the Plan's two is acceptable for safe variable quoting
and the unchanged output contract; ENR-16 colocation in catalog is accepted with
its actual hosted selection. Preserve outstanding final attack coverage.

**CORRECTION LOOP on re-attestation before HTTP execution release.** Accepted
correction01 C4 promises explicit re-attestation after replay reconstructs history
from a pre-attestation checkpoint, but C1's literal backfill-only evidence gate
makes that impossible when all rows are recorded_via=replay. Reconcile that one
predicate: operator_backfilled may use at least one **same-Organization departed
row with recorded_via IN(backfill,replay)**. This restores the promised path, not
a new enrollment mode or automatic readiness. Preserve row provenance/seq/times,
superuser-only operator channel, l<seq> reference grammar, Organization locking,
append-first/operator ledger verification and mandatory audited explicit setter.
No inference from audit history, no accepting another Organization's rows, no
provenance rewrite and no automatic open. Do not recommend operator_verified_absent
when known departures exist; that would be an untruthful attestation workaround.

Extend ENR-16f (and existing setter negative fixtures) minimally: replay-only
history allows explicit re-attestation, readiness stays closed before it, the
replayed restriction stays intact and still denies that departed subject after
an authorized open, while empty/foreign-only evidence returns no_backfill. Update
exact setter body/caller catalog pins and recovery guidance for this reconciliation.
Run only affected setter/replay/catalog proof, not whole core/ops/drill by default;
reuse unchanged c1/c2 and selector evidence. Submit the small corrected source and
proof for review before HTTP execution. Unaffected parser planning/inspection may
proceed under the accepted Plan; do not expose routes backed by unresolved readiness.

One fresh writer for this correction; same scope50/attempt1/B7/checkpoint4,
300k checkpoint/350k safe handoff. Required core/callback/HTTP/UI/final attack and
hosted evidence remain. G019 current staging running, G055/MR48 owner merge holds
and G020 phase stop unchanged. No new owner decision or migration number needed.


## R12 re-attestation ACCEPT; HTTP execution release — 2026-10-18 UTC

**ACCEPT for the bounded correction at 2db012c9 (source cc15b995).** Root and
bounded read-only security reviewer inspected the migration/test delta against
370e47f7. Exact Result imported: SHA-256
f56ed3e531e54aafecce4e2bc00a55f988bc3c201309c712c318ed23b0f8fd38.
Same-Organization departed backfill/replay evidence now permits the explicit
operator_backfilled setter. Superuser channel, Organization lock, reference
validation, audited transition and operator ledger procedure remain unchanged.
No automatic readiness/open or restriction/provenance rewrite. Extended ENR-16f
proves closed-before-attestation, exactly one readiness audit row, preserved
restriction, departed-user refusal after authorized open and eligible-user control;
empty/foreign/trigger-only evidence remains insufficient.

Root independently hash-verified final catalog log
bce621c7bedca71b7dac087ab6fa5dcf02ee72be51cf33ca329da2726eee951e
and static-unit log
775c84314a8ae62c597affd06c445f337a8abacc4279e12951da3ec3fa1faeec.
Plain Go output confirms catalog 19 recursive subtests and static units 70 top-level passes (2,066 including subtests),
no final failure. First catalog failure (runtime-role test reset) remains in the
Result and retained log; corrected test reset is appropriate. Evidence remains
local PostgreSQL restored-state emulation, not hosted restore-drill qualification.
Count-only credential-term scan is weaker than exact-value canary scanning;
native refusal preserved, no bypass. Unchanged earlier proofs reused.

**Release the approved HTTP/identity callback slice** under accepted Plan r3 and
selected-design corrections: enrollment API nine method/path pairs on eight paths,
modified identity routes and request_on_arrival wiring, with focused ENR-H tests,
credential/cookie/host/unknown-commit negatives and applicable registration parity.
Use the existing single owner or one fresh writer at a safe handback; reconcile
management-only staging inputs without repeating unchanged technical evidence.
Keep status admitted historical, never a membership/session grant; confirmed
admission leads to ordinary fresh sign-in/current membership guards. Preserve
minimal applicant disclosure and same-user/Organization credential binding.
No new public contract/design, migration scope, broad default ops run or native
permission workaround is authorized. Submit a coherent HTTP source/proof handback
before the next UI slice; unchanged core need not be re-reviewed in full.

Scope50/attempt1/B7/checkpoint4 unchanged; checkpoint at300k and safe transfer
by350k. Final attack ENR17/21/22, affected UI/HTTP tests, pinned hosted enrollment
and recovery qualification remain required. G019 final ACCEPT gates integration;
G055 and MR48 merge holds until G058 final ACCEPT and G020 phase stop remain.
This is source correction ACCEPT and bounded execution release, not Goal ACCEPT.


## R13 HTTP partial source ACCEPT; web slice release — 2026-10-18 UTC

**ACCEPT for reviewed HTTP/identity source at 07256a5e**, with runtime qualification
still pending. Root plus bounded read-only security and HTTP/harness reviewers
inspected the delta from 2db012c9. Exact Result imported SHA-256
8daabb59c6e635cf145f49e47b825b60593ffec7f826cceb4bfd6ca269e8e1f4.
No blocking security/source finding: enrollment credentials remain distinct from
sessions; exact-origin/current-Organization binding, mutation CSRF, manager
session/context guards and same-transaction host checks remain. Unknown COMMIT
is unavailable. Historical admitted status and current membership remain separate;
ordinary fresh sign-in alone grants a current session. request_on_arrival submits
an own request, never membership; invitation intent retains precedence.

**Explicit technical reconciliation of Plan summaries (immutable history retained):**
- Enumerated public/applicant/manager set plus revoke is ten method/path pairs on
  nine paths. The nine/eight summary omitted the already-enumerated public route.
  Accept CORE_ENROLLMENT_ROUTES ten and the nine completion-schema paths.
- EnrollmentCredentialInfo exposes the resolved handle, own CSRF token and expiry;
  status includes own csrf_token for BFF forms, no-store and without CORS. This
  completes the planned CSRF boundary, not a session or provider assertion.
- POST request returns 200 for both newly-created and existing pending requests:
  the accepted helper does not expose creation-vs-existing. Do not invent a
  guessed 201 or alter durable SQL solely to distinguish them. BFF must accept200.
- Callback202 credential is {value,max_age}, consistent with the existing secret
  projection and bounded cookie lifetime; never mix it into session credentials.
- Root/unserved enrollment subpaths return closed JSON404; wrong methods on
  defined paths return405. This expressly reconciles r0's root-status prose;
  align runtime B-TENANCY-HTTP assertions to this actual closed route contract.
Already-used x/time/rate is not a new dependency/vendor. No migration change.

Root independently matched logs: units37bc856a0ddb894f93deb36078afa5e5dd1e19b41b644804fac440fc8ec05e19
(144 top-level,2150 recursive passes), logrecord6a8c81f0f30d6a37035cca0a64675632cd9ca1f323f53637ad839c66f07365e0
(10), DB0f6e7c6ed1967d4847dacf7506e3433b8a990a833018c7847f957c27ea6be444
(ENR-1/ENR-INTENT selected rows, both pass), access log2082834d…680c and static
scriptcf2b0f22…5a9a. No final FAIL in those test logs. Reported vet/shellcheck and
canary negative control retained in Result, not independently rerun here.

**Release approved web/BFF/UI slice**, using one writer, current accepted Plan,
service instructions and the repository Vestive skill explicitly loaded from its
current path. MR48 remains held: do not assume the relocation is integrated.
Implement enrollment-client's ten exports, cookie/join intent/status and manager
surfaces with the reviewed contracts and existing minimal disclosure. Preserve
historical-status/fresh-sign-in semantics and truthful expired/pending states.

The same slice must complete actual successful HTTP route-to-facade/serializer
proofs, including applicant status's current-member and csrf_token response,
request/admit/abandon and authorized manager policy/list/decide/revoke, plus relevant
wrong-authority negatives. Existing forced-transaction-failure HTTP units and
serializer helpers do NOT establish those positives. Extend existing tests or
focused started-server integration; do not create a second harness. Add the
planned B-TENANCY-HTTP positives/root closure pins and focused E-ENR browser rows.
Run only affected selectors on the required shared stack under its existing lock;
no default full/ops campaign. Preserve unchanged core/replay proofs. Submit
coherent source and attributable evidence; no advance claim of final coverage.

Final ENR17/21/22 attack coverage, M1 where required, pinned hosted enrollment,
B-MCP-SOURCE and recovery qualification remain. Scope50/attempt1/B7/checkpoint4
unchanged;300k checkpoint/350k safe handoff. G019 pipeline2960883415 independently
still running at review; its final ACCEPT gates integration. G055/MR48 owner
merge holds until G058 final ACCEPT and G020 phase stop remain unchanged.


## Scope amendment04; Organization invitation entry — 2026-10-18 UTC

Decision on R14 checkpoint a52ee651: **ACCEPT the narrow scope extension**, not
acceptance of its unfinished web slice. G058 explicitly owns first-time
Organization invitation enrollment. Existing membership-client invitationIssuedPage
still generates /invitations/<id>#token and says Organization web acceptance is
unavailable. That G019 transitional presentation must now point Organization
recipients to the G058 entry /join/invitations/<id>#token. Keep Group invitations
on /invitations/<id>#token with their existing member-only acceptance flow.
Do not infer the scope from the recipient/account status or create a new admission
path. Core verified-recipient/current-policy checks remain authoritative.

Scope50→52 adds exactly web/lib/membership-client.ts and
web/tests/unit/membership-forms.test.mjs. Only Organization link selection,
obsolete unavailable-copy replacement and directly corresponding tests authorized;
no general membership-client refactor, G019 lifecycle/authorization change or
changes to existing invitation pages. Preserve one-time token in fragment only,
no query/server Location/log/title leak, exact trusted origin, no-store response.
Extend the existing issued-page test to distinguish Organization/Group paths and
retain single token occurrence/no Location/privacy assertions; exercise the
Organization link through the already-required enrollment browser flow. Focused
membership-forms and affected enrollment tests only, no broad rerun for scope.

Prior scope50 bytes preserved in plans/G058-scope-allowlist-before-amendment04.txt;
current scope52 SHA-256 dc791b4b6bfbf454a9a5eec686ac5c4d50f894872ceaf19121f339965e77ee23.
Same running continuation writer at its next safe boundary; reconcile scope/hash
under the existing procedure, do not restart it or duplicate ownership. Same
attempt1/B7/checkpoint4 and all release/dependency/merge/hosted holds. R14 partial
source is a checkpoint, not a final reviewed submission; unrelated web work
continues under the existing R13 release.


## Frozen web checkpoint preliminary correction — 2026-10-18 UTC

Read-only source review of71388d12 versus07256a5e (R14/R15 PARTIAL checkpoint,
not final submission); Result SHA aeb0f221b054e6fd01bf3f225a8b9b09026e27203f878a1b05d2e85d2ac318e1.
Root plus bounded security reviewer found no credential/session or authority
bypass, but two bounded presentation/cleanup corrections before source acceptance:

1. auth/callback early refusals (MAX_QUERY, resolver/nonrouting) can return before
   clearing a present one-use join intent. Complete tenant-side cleanup on refusal
   paths where tenant classification is established; arrange classification safely
   where needed and preserve the shared auth relay's absolute no-cookie boundary.
   Unknown/unclassifiable host is not permission to read/write auth-host cookies.
   Add focused early-tenant-refusal and shared-auth no-cookie tests; no token in
   Location/query/log and no change to provider relay/session authority.
2. components/enrollment/request-list.tsx hides both decision actions for expired
   pending requests. The accepted one_pending contract deliberately retains them
   until admin decision; removing both leaves no supported UI resolution. Keep
   an explicit existing Deny action for expired pending rows with truthful copy
   (expired, deny to resolve); do not imply approval renews validity, add automatic
   expiration, renew credentials or change SQL. Add focused expired-pending UI
   proof; expired approved state remains non-consumable.

Same current r16 writer handles these at its next safe code boundary; do not
interrupt an immutable running batch or duplicate ownership. Reconcile tested
source/delta and run only affected tests. Independent runtime qualification may
continue under the R13 release. All runtime/browser and final attack/M1/hosted
proof remains pending, no final/source ACCEPT implied by this preliminary review.
Scope52/attempt1/B7/checkpoint4 unchanged, G019 and other merge/phase holds remain.


Bounded UI/browser/harness reviewer completed the same preliminary checkpoint:
no additional concrete blocker. Shared fixture sequencing is mandatory: prior
identity/nonmember checks before enrollment, E-ENR-4 before E-ENR-2, one browser
worker. Light/dark subjects are intentionally separate. Runtime/accessibility
proof is still unexecuted; source inspection is not its substitute.


## Scope amendment05; consumed invitation parser parity — 2026-10-18 UTC

**ACCEPT bounded correction scope** from R16d0093ab8. Root verified core00015
claim UPDATE assigns intended_user_id=consumer while retaining the original
address; CHECK invitations_consumed_is_intended binds the consumer. The existing
web parser unconditionally rejects both fields, so an authentic consumed address
Organization invitation makes the administrator's list malformed. E-ENR-3 reveals
this existing G019 consumer mismatch; no SQL/API/authority change is needed.

Extend the existing membership-client.ts authorization beyond amendment04's link
scope only for this parser predicate: permit both recipient fields exactly when
scope=organization, status=consumed and intended_user_id=consumed_by (validated
nonzero canonical IDs). Retain strict shapes, byte bounds and all other recipient
validation. Group dual-fields, pending/revoked/expired dual-fields and mismatched
consumer remain malformed; no broadened recipient disclosure or admission grant.
Add exact web/tests/unit/membership-client.test.mjs (scope52→53); extend a relevant
existing case with consumed-positive and adjacent negatives. Do not weaken the
browser assertion or hide the malformed list. Current worker/fresh continuation
under existing handoff policy; same attempt1/B7/checkpoint4, no G019 budget reset.

Focused unit tests and E-ENR-3 both themes only; reuse unaffected67/67 identity,
8/8 fixture-impact and8/10 enrollment successes with exact source/delta provenance.
No full browser matrix/ops repeat. Existing two source fixes and route probes
still need final submitted evidence review. Remaining ENR17/21/22/M1/hosted proof
is not waived. G019 final ACCEPT dependency is clear; G055/MR48/G020 holds remain.
Prior inventory archived; scope53 SHA-256 ce8ff6b6f9bf59456300761509eb4d66f462d52d7bc0ea194bb0411f9833954e.


## R16 runtime slice ACCEPT; final proof release — 2026-10-19 UTC

**ACCEPT for reviewed runtime/web slice at d5d928ad**, not pre-integration or
final Goal acceptance. Root and bounded security/evidence reviewers checked the
71388d12→d5d928ad source/correction delta and attributable runtime proof. Exact
Result imported SHA cdab9f3e71d3e15dfffb8d4c9380b7562eac0400a2ac3e4df3886c198374e81f.
Both prior review findings close: classified tenant callback refusals clear join
intent while auth/unclassified paths touch no cookie; expired pending keeps Deny,
never renewal. Consumed Organization invitation parser exception matches existing
SQL exactly and retains strict other-scope/status/consumer negatives. No API,
authority or SQL change. First-time invitation path reaches admission, fresh
sign-in and a correctly rendered administrator invitation list.

Independent evidence hashes/counts reproduced:
- root e1ab244e…2b62e:19 technical PASS, including86 tenancy HTTP probes. Scope/
  aggregate failures caused by contemporaneous management-only staging delta
  remain recorded; reconciled scope passes later, not a fictitious full green root.
- h1 2f618cf7…a76d and h2 aab30538…2849 pass their18 required IDs including
  B-UNIT-WEB and exact scope53. Passing wrapper retains no unit count: do not
  claim an invented precise final unit count.
- identity report b5f1b305…a23f:67/67, fixture impact c435c4a3…b84d:8/8.
- enrollment f877f8f0…6045:original8 pass/2 fail; corrected targeted E-ENR-3 report
  33880f89…4682:2/2. Combined ten-case coverage with parser-only delta and
  attributable reuse, not one uninterrupted ten-case green run. Earlier failure
  and registry-proxy non-execution retained. Logs/scans/stack cleanup as Result;
  no duplicate runtime or whole-browser repeat required for this acceptance.

**Release remaining approved final security proof and qualification preparation**
with one bounded current/fresh writer: implement/complete ENR17/21/22 using existing
attack/oracle tests and approved Plan selectors, plus only any actual remaining
AC coverage gap. Run focused selected rows and changed producer/catalog consumers;
reuse unchanged core/replay/browser proof. Keep evidence and candidate attribution,
record all failures and stop affected chain on real boundary failure. No default
full/ops or new control framework. Scope53/attempt1/B7/checkpoint4 unchanged.

Prepare the concrete owner M1 preview/checklist at a safe resource boundary and
request that required feedback once ready; do not assume a waiver or silently
replace owner review with browser automation. Existing release permits independent
proof while waiting. DraftMR50 automatic checks must be observed to terminal;
no optional duplicate full MR job merely to announce progress. Final source/
security review, applicable M1, current-head CI/Duo and required post-merge hosted
enrollment/recovery qualification still gate pre-integration/final acceptance.
G019 dependency is ACCEPTED; G055/MR48 merge holds until G058 final ACCEPT,
G020 phase stop and G021 barrier remain. No next-task dispatch implied.


## R17 security proof ACCEPT; M1 preview setup already released — 2026-10-19 UTC

**ACCEPT for bounded final security proof at4ec314b8** (test-only delta), not Goal
acceptance. Exact Result36303a063b192f7fc127c952050c4678e280992d95e702ba1609e4f612dfd2bb
imported. Root and bounded security reviewer inspected actual assertions and logs:
forged NULL-context applicant audit with tenant attribution refuses; enrollment
credentials never resolve as sessions; expired/revoked/disabled resolution remains
closed; EMPTY reads/by-ID deletes see no rows but reviewed unqualified expiry
sweep deletes exactly expired rows; helpers cannot inject restriction rows; owner
cannot rewrite/delete them; removal/self-leave/denial rollback atomically without
restriction insertion and positive controls each produce one ledger row.

Coverage colocation in existing catalog/admission/request/RLS/identity tests is
accepted instead of adding attack-oracle rows: these execute the actual boundaries
and are selected by existing hosted checks. No security assertion waived to avoid
count-pin changes. ENR17/21/22 map is now source/evidence complete within this
slice; existing producer, replay and oracle coverage remains required/reused.
The new assertion fixes match real guard-trigger refusal and SELECT-policy sweep
semantics; original failed attempts remain documented, not passing. No production,
SQL, contract or oracle change; no repeat browser/replay run justified.

Root hash-verified final logs in r16/enr: catalog5ee865a4…ae29 (19 subtests),
access90e5dd3e…785e (14/15), RLSf4c37506…68f0 (10), identity78c36673…3943 (8),
answer-shapes83b26a38…dd0f and API package7c716c7c…69a7 (10 top-level). Final
reports contain no FAIL; bounded reviewer independently verified six final logs.

M1 preview setup was already authorized in R16 release; coordinator may start the
prepared disposable stack now without another owner setup question. Exact
candidate/modes/CA/smoke must be verified before reporting ready. Owner performs
manual checklist and supplies feedback, or an explicit waiver is required; neither
agent fabricates that outcome. No credentials in chat. Preview lifetime/lock and
resource ownership stay under existing local harness policy, no hosted provider/
DNS action. Source test SHA changes need no runtime retest for unchanged UI.

MR50 head4ec314b8/pipeline2961150120 independently running at review. Remaining
before pre-integration disposition: real M1 outcome, current-head CI/Duo and a
concise AC evidence reconciliation identifying actual missing coverage, not a new
blanket suite. Required post-merge hosted qualification/cleanup still gate final
Goal ACCEPT. Same scope53/attempt1/B7/checkpoint4; G019 dependency ACCEPTED,
G055/MR48/G020/G021 holds unchanged.


## Owner M1 deferred until morning — 2026-10-19 UTC

Owner: “I will check m1 in the morning.” Manual feedback deferred to the morning
of2026-10-19 America/New_York; this supersedes the earlier midnight completion
target for M1-dependent integration/acceptance. M1 is pending, not waived/passed.
Keep MR50 unmerged and auto-merge unarmed until actual feedback and other gates
clear. Already-authorized independent current CI/source/evidence reconciliation
may finish; this is not a stop-all or new-task authorization.

Preview at4ec314b8 independently returned200 on org-a/b/c with its public CA.
No owner manual findings yet. Coordinator should stop the disposable local M1
stack at a safe boundary, release its harness lock and preserve the exact
candidate, checklist, seeds and necessary preview setup references. No need to
leave idle containers overnight or ask another setup approval. Recreate/smoke it
under existing authority when owner returns; do not schedule a new session,
reminder or automatic morning launch. Preserve other active resources and report
native cleanup refusals once. G055/MR48/G020/G021 holds unchanged.


## Owner sequencing override — merge before M1, 2026-10-19 UTC

Owner: “Can we merge and then do m1 review?” Approve conditional staging integration
before manual M1, once current source review, dependency, CI and Duo gates clear.
M1 moves to a **post-merge final-acceptance gate**, not waived/passed. Supersedes
the prior M1-before-merge hold only. Manual review remains scheduled by owner for
the morning; final G058 Goal ACCEPT requires its real outcome, required hosted
qualification and cleanup. G055/MR48 remain held until that final ACCEPT; G020
phase stop/G021 barrier unchanged. No public deployment or native bypass.


**Pre-integration source ACCEPT at4ec314b8**, conditional on exact candidate/current
input reconciliation and CI/Duo disposition. Prior source/security/Plan mapping
and R13/R16/R17 proofs cover the automated local requirements: core admission,
restrictions/rollback/audit/credential negatives; HTTP/cookie/callback/current
membership separation; affected fixture and enrollment browser paths; runtime
registration; recovery emulation/operator re-attestation. Reuse attributable
unchanged proof. The Result's generic “final local affected coverage” line is not
itself an order for another broad suite: identify a concrete uncovered criterion
before any additional run. M1 is explicitly deferred as above. Hosted qualification
still follows integration; local emulator evidence does not claim G057 live proof.

Make draftMR50 ready, reconcile current target by the accepted procedure, request
Duo and observe required checks. Avoid changing tested technical bytes solely for
management updates; record exact irrelevant-delta/rebinding where permitted, and
rerun only actually affected proofs if a source change is necessary. Merge only
when review gates are clear, using the verified exact source SHA and keeping
normal temporary-branch deletion. Current pipeline2961150120 on4ec314b8 was still
running at this decision, no failure inferred. One merge owner; Codex may perform
the requested GitLab merge once clear, coordinator must not race another merge.
No extra full MR pipeline; automatic required staging qualification follows merge.
Keep M1 preview stopped overnight if idle and recreate the integrated candidate
for morning feedback. No need another owner proceed prompt for conditional merge.


## Conditional integration; M1 post-merge — 2026-10-19 UTC

Owner-approved sequencing executed: MR50 merged at00:43:51UTC, exact source
4ec314b8725f14b7cd6913554e8c3553dfeee18d → staging
27ee7c205791ab7d5f4c0e4a216c5108a78472ca. MR checks2961150120 succeeded on that
source; Duo reviewer state approved, no findings in reviewed portion (21% diff
limitation recorded, independent prior source/security/evidence reviews cover
remaining scope). No review-state override used for MR50. Root verified no
conflict, all discussions clear, source SHA unchanged and target delta from
813d2296 to3eadd855 limited to six management/Goal/Spec/Result files. Owner M1-order
supplement explicitly reconciled; technical test inputs unchanged. No redundant
rebase/CI/local campaign solely for management prose. Merge retained current
staging authorization/Goal/Spec supplements. Remote temporary source branch absence
verified; no long-lived branch deletion or squash.

Required automatic staging pipeline2961169069 independently running on exact
27ee7c20, source push. Coordinator watches terminal result and diagnoses actual
first failures using the complete repair inventory, no duplicate run. Final Goal
ACCEPT remains held for required hosted disposition, owner morning M1 outcome and
mandatory closeout. M1 pending/not waived; recreate smoke-verified preview from
the integrated candidate when owner returns. Prior disposable preview shutdown
and preserved setup permit that without more setup approval. G055/MR48 cannot
merge before G058 final ACCEPT; G020 phase stop/G021 barrier unchanged.
This conditional staging merge does not promote G058 to main or move pre.4.


## First staging failure; verified complete inventory — 2026-10-19 UTC

Root independently verified pipeline2961169069 failed on exact merge27ee7c20:
checks17270568376 SUCCESS, integration17270568377 script_failure. Downloaded
results SHA25b92770ef6cbd1985f7b2c566ddf197d4b0d33a52a27d7e28d83b42c962d292:
integration87 PASS/3 FAIL/6 INFO (pipeline-wide counts are distinct).

| Required row | Actual failure | Disposition before any repair |
|---|---|---|
| B-OBJ-IT | Object10/10 and service297/0,RLS85/0 pass; attack142/1 fails F2-12: actual42501 invitation_context vs expected42501 | Read-only trace confirms old oracle claims runtime EXECUTE denial, but accepted D-15(i)/00016 explicitly grants runtime EXECUTE; guard refusal remains. Verify exact error/effects and reconcile oracle to accepted contract, not privilege/model weakening. |
| B-CI-CONTROLS | VA7 summary required_count90 vs88; VU1 fast/gate/union35/90/121 vs35/88/119; J8 remainder90 vs88; other438 controls pass, G3 passes | Stale exact-count consumers after two required enrollment checks; inventory every affected pin and prove corrections with focused controls. |
| B-REQUIRED-CHECKS | Derived from the two rows above | Clear only after underlying attributable proof, never separately waive. |

Evidence inspected from coordinator's g058-stg1-art/ci-evidence diagnostics,
including obj-it-attack.log and ci-controls.log. No wrong-allow demonstrated by
this oracle error, but denial-shape reconciliation is not yet accepted as a fix.
Do not rerun unchanged full job. Claude retains owned diagnosis/proposal and must
include every failed component and all affected pins in the existing Result.
infra/ci/validate-controls.sh is outside scope53: request exact amendment before
editing it; existing attack oracle path is scoped but security expectation changes
still require review. Focused pre-merge proof or evidenced existing MR qualification
where genuinely unavailable; required post-merge gate remains. M1 pending morning,
no final Goal ACCEPT or downstream merge release. Prior successes/failures retained.

## Local qualification placement override — 2026-10-19 UTC

Owner decision (not included) supersedes pending hosted full qualification. Failed staging evidence remains failed. Final repaired candidate needs local full PASS plus fast hosted checks, source/Duo review and owner M1. Correction: validate-controls.sh was already scoped; my prior out-of-scope statement was wrong. Current repair ownership and budgets remain.

## Hosted CI restored; timing correction released — 2026-10-19 UTC

Owner rescinded local-only override. Current decision and preserved history (not included). Restore previous hosted configuration; amendment07 scope54 adds bounded JUnit/validate correction, replaces removal delta. Existing repair ownership and budgets unchanged. Correct timing aggregation without removing actual failure cases or erasing the failed run. Focused existing JUnit/control proofs required before source acceptance; normal G058 qualification and owner M1 remain.

## Staging repair1 source/evidence ACCEPT — 2026-10-19 UTC

**ACCEPT for pre-integration repair/MR readiness**, source `8851d4a1fa5aae2479d64ceff7c9530f748146e0`, base `baac320d`. This is not final G058 ACCEPT. Independent bounded reviewer and supervisor found no blocker. Scope54: six changed paths, zero outside. Exact Result imported SHA256 `77d603c241cf92cc6c0785776ae4c7b2e5dd040676764414581f4ae9d25cd141`.

Complete failed-run inventory checked against retained staging2961169069/job17270568377 artifacts: F2-12 retains exact42501 invitation_context refusal, membership/invitation effects unchanged; VA7/VU1/J8 now90/121; required aggregate and phase failures are derived, not silently waived. JUnit emits one explicitly named validation-elapsed timing case, no overlapping suite/phase case times; original failure identities/counts, diagnostic suite breakdown, XML shape and fail-closed handling remain. Duration is validation-start through report construction, excluding later publication overhead; native GitLab job duration remains authoritative. Hosted CI YAML/selection/README are byte-identical to target; local-only change remains rescinded.

Independently inspected matching retained evidence at `~/.local/state/platform-evidence/g058-a1/repair1/`:

| Proof | Outcome | SHA256 |
|---|---|---|
| run1/attack.log | 142 cases/0 disagreements, 22 subtests PASS | f3791700c81b6974fa526fc99aa8b944a1b64662e331b0ad88a9ac1e29f76a48 |
| ci-controls/ci-controls.log | 441/441 PASS | 4448213ce7a9f0d9d68691b8ea89c74eb6ccddeea860a161b90be59a404fe6a5 |
| ctl-junit2.log | 10/10 PASS, successful/failed overlap fixtures | 92524a6756e04132cbe8b7fc7177227c9d404691aa0c813592e61e4d5e10eea3 |

Core proof unchanged from808b0a66; CI/foundation code unchanged between tested41720275 and b9febbb6. Narrow wrapper scope FAIL from target drift and secret-canary UNMET from omitted lint remain explicit, not an overall green-run claim. Current diff independently scope-valid; ordinary MR checks still required.

Claude may reconcile the management-only target advance, push the same repair branch and open its MR to staging now. Do not split timing fix or add another MR. Retain current-head fast/Duo dispositions before enabling standing auto-merge; required post-merge hosted qualification and morning owner M1 still gate final G058 ACCEPT. No unrelated full local run or unchanged hosted retry is added. Existing budgets, downstream G055/MR48 holds and cleanup apply.

## MR52 Duo correction delta ACCEPT — 2026-10-19 UTC

**ACCEPT** current source `6c77b4f6d161ba229cdea801b42dc2032dcdc1b3` for the two bounded corrections. Independent reviewer and supervisor inspected delta from8851d4a1: F2-12 r4 provenance accurately names the accepted D-15(i)/migration basis and repair-author deviation; J1/J5 remove a runner-speed bound against fabricated220s timings while retaining the single timing-case shape. Deterministic J4b still proves non-additivity. No application behavior or oracle effects changed. Exact updated Result imported SHA256 `fa5ae99bc4a5769e798574e0985d2bc05b4aa41d6f7eccbcddbb8c198ed5c82f`.

Verified `repair1/ctl-junit3.log` SHA256 `5ac0009e7cfc5a4f6ccf7c28571761ae982349130cca563508da2c95ff82c722`:10/10 PASS. Reuse prior441-control and142-case evidence for unchanged code. MR52 current checks2961442501 running at inspection. Codex requested Duo re-review in note4039299507; API commands_changes confirms reviewer21826781 re-requested. The earlier Claude refusal concerned a complex computed shell command under worktree isolation, not an owner decision to withhold review; supported plain request used without settings/guard changes.

Both original Duo comments are corrected in source; current Duo re-review and fast pipeline remain merge gates. Standing feature auto-merge may proceed once those current-source gates clear, after reconciling any management-only target advance without an unchanged test campaign. Hosted post-merge qualification and owner M1 still gate final G058 ACCEPT; no role switch, split MR or additional local full run.

## Repair MR52 integrated; qualification running — 2026-10-19 UTC

Verified current-source fast pipeline2961442501 SUCCESS and Duo approved note4039303739 (review note4039303727, both corrections accepted, no comments). Exact source6c77b4f6 remained mergeable with no technical target delta. Codex merged under standing feature authority: MR52 merged `f7a28581d8e2b21b9a7bb7b9f5990bf7d8548704` at2026-10-19T04:31:45.469Z; source branch deletion independently verified. Native pending AskUserQuestion was inspected by attach; Codex selected “Leave it to Codex” to prevent a competing merge, then detached without stopping Claude. This was a Codex response under existing authority, not new owner-typed approval.

Automatic staging pipeline2961594265 is RUNNING on exact mergef7a28581. Claude retains its existing qualification/first-failure/cleanup ownership. No extra local full run, manual duplicate pipeline, final G058 ACCEPT or downstream release is implied. Owner M1 remains pending.


## RC1 owner-disposed limiter follow-up — 2026-10-19

Independent focused source review confirms LOW map-overflow limiter reset issue (http.go469–470/560–564), no admission/authz/RLS bypass. Owner selects keepingRC1auto-merge and tracking the fix as a prereleasefollow-up. Bounded backlog (not included) preserves required depleted-bucket/churn regression and scope review. No fix/test PASS claimed, no new G055/RC blocker, no G058 finalACCEPT inferred. Existing ownerM1 gate remains.


## Owner M1 feedback imported; focused Group-admin 404 triage held — 2026-10-19

Owner feedback record (not included) imported byte-exact
SHA256ca8225adfe29b0e07641d9eb6cbbe5d98b80fef88151fa3604f0653c49b90327.
Preview7fafd1f4 includes accepted limiter correction. Owner reports “everything
works”; coordinator did not observe each checklist step, so accept as owner-
reported manual feedback, not independently executed per-step proof. Wrong-
account invitation observation did not consume grace's invitation as another
subject per reported DB checks. Owner confirmed bob sees Group membership.
Preview shut down at owner instruction, with reported zero task resources/free
lock. Source's “18:5xZ” is an invalid placeholder: exact stop time not retained,
not an asserted timestamp. Coordinator must clarify in the next append-only
handback. Existing earlier hosted qualification remains independently accepted;
no new broad run follows M1 reporting alone.

Bounded independent read-only review finds concrete suspected UI mismatch:
GroupMembersPage first reads Group with org-admin authority, but membersView
requires selfMembership; core SelfMembership intentionally returns404 if the
caller has no Group binding. Thus a tenant admin can have Group read/manage
authority yet the members page refuses it. Existing E-MEM-5 explicitly expects
this404 for an Organization owner with no Group binding; accepted UI/test history
must be reconciled against G019 R1 and policy, not silently called intended or
changed merely to make a test pass. E-MEM-3 retains a Group-member binding and
does not resolve the missing-binding case.

Keep final G058 ACCEPT/G055 merge held pending this bounded triage. Release only
the existing coordinator's focused reproduction/diagnosis and correction proposal:
same-Organization admin with no Group binding, Organization-only context and exact
known Group members URL; compare Group read, self-membership, member-list responses
and page outcome. Capture exact source/context/URL/status, existing oracle rationale
and proposed minimal UI/test paths. Use existing disposable fixture tooling, no
owner browser retest, broad/full run, new persistent coordinator or implementation
yet. Preserve ordinary nonmember/foreign/selected-Group refusal and Group-owner-
only powers. Do not fabricate a Group binding or broaden SelfMembership semantics.

Separate new scope: combined Organization+Group invitations by Group owner/admin,
always invitation-only Group joining, is not completed by current G019/G058 and
is not an M1 failure. No new Goal/Spec/implementation release for that request here.
Styled uniform not-found UI may be a bounded later presentation follow-up; do not
use it to conceal an admin-access bug or reveal Group existence.


## M1 Group-admin 404 option B — ACCEPT and bounded correction release, 2026-10-19

Exact Claude diagnosis/proposal (not included)
SHA256a3c4572921ce5f0e6406f7f6afafbbf34da0b07d003f28f83f4c72cc944997d8
imported unchanged. Independent read-only review and supervisor inspection choose
**B** as UI conformance to existing G019 R1/core policy, not new Organization
authority. Option A would narrow the Spec's existing ordinary administration
contract and is not selected. Earlier accepted E-MEM-5 404 expectation is explicitly
superseded for same-Organization owner/admin ordinary Group administration only;
plain nonmember, absent/foreign/selected-other-Group restrictions remain.

The code chain plus original E-MEM-5 PASS in RC2 is sufficient pre-correction
evidence; no extra full/browser run merely to reproduce known old behavior.
Release one existing-coordinator-owned web-root implementation writer at reconciled
current staging. Read applicable web governance/design instructions; no new design
or layout. Amendment08 adds exact members-view path and allows this bounded change
to existing membership.spec.ts beyond its earlier nav-pin restriction. Writable
implementation paths only:
- web/components/membership/members-view.tsx
- web/tests/browser/membership.spec.ts
- management/v0/results/G058-generic-organization-enrollment-result.md (append only)

Mandatory implementation conditions:
1. Preserve the current status=left branch before any fallback. Fallback only for
   Group scope and self-membership not_found, never Organization scope or other
   self-read errors. Probe core-authorized listMembers with the same validated
   pagination cursor and reuse its successful response; no duplicate first-page
   read or fabricated canManage/self role/binding.
2. Allowed fallback renders existing ordinary administration without self card,
   Leave, ownership promotion/demotion/transfer controls or owner candidate read.
   All subsequent reads/actions retain current core authorization; no core,
   SelfMembership, capability, schema, seed or invitation-authority change.
3. Actual existence/authorization denial retains the fixed404 privacy result.
   Backend unavailability, transport/configuration or malformed responses retain
   existing availability/error handling; do not mask service failure as404.
4. Amend E-MEM-5 owner/no-binding oracle and cover admin/no-binding positive,
   ordinary nonmember/absent/foreign/selected-context negatives where applicable,
   retaining no-owner-control assertions. Preserve Group-owner/admin/member and
   stale-action/leave behavior. No weakening of existing true security negatives.

Validation: focused E-MEM-3 and E-MEM-5 in membership-light/dark (including added
safety subcases), relevant existing unit/type/lint checks and normal feature MR
fast/Codex/Duo. Record actual command scope, tested commit, counts/exits/logs and
cleanup. Reuse unrelated RC2 full evidence/unchanged backend proofs. This bounded
UI-only correction is exempt from another local full-suite run; NOT PERFORMED,
not PASS. If core/authority/data or shared-context change is needed, stop for
review rather than extending the exemption/scope. No new RC/main/finaltag.

Use remaining G058 ledger/budget, no reset/duplicate writer; report existing
remaining allowance and stop if exhausted. Usual context/scope/native stops hold.
Submit exact diff/Result for source acceptance, then current MR gates, merged
delta and actual cleanup before final G058 ACCEPT. Existing M1 enrollment feedback
stands; no broad owner retest of unchanged flows. G055 merge remains held until
G058 final ACCEPT. Combined Organization+Group invite feature remains separate,
unreleased scope; no G020 planning or OS001 resume here.

Timestamp clarification: proposal's18:42:01Z is stop-log mtime/observed completion
evidence, not independent proof of the exact stop instant. Preserve raw report/
proposal and append that qualification; do not invent a more precise event time.


## Group authority design reconsideration — execution HOLD, 2026-10-19

Owner asks whether Organization administrators should require explicit Group
membership/invitation rather than automatic Group member-management authority,
and whether an Organization-owner setting is appropriate. This is a design
question, not selection of a replacement security model. Suspend the preceding
option B release pending the owner's decision. Existing core authority remains
as implemented; a UI denial alone does not remove server-side authority.

Verified pinned-coordinator handoff receipt; coordinator reports the writer
paused safely at pushed checkpoint 5f94a705 on g058-group-admin-404: no MR,
no tests executed, no Result change, no containers, lock free; worktree retained.
This is coordinator-reported state, not an independent runtime audit. Preserve
work, evidence and remaining budget; no new implementation release or reset.
G058 final acceptance and G055 merge remain held.

## Owner separates enrollment closeout from new Group policy — 2026-10-19

Owner explicitly directs G058 acceptance/closeout against enrollment requirements;
the paused UI correction is folded into the new G060 Goal/Spec/Plan. Supersede
optionB's execution release; no patch merge or passing tests claimed. Current
source/manual/qualification and cleanup are reviewed for final original-scope
acceptance separately. Record broader current backend behavior as outstanding G060
work, not enforced by the existing UI404. G055 integration/completion additionally
waits for G060 final ACCEPT; closing G058 does not release it prematurely.


## Enrollment source/M1/qualification ACCEPT; final cleanup handback pending — 2026-10-19

Bounded independent read-only review and supervisor inspection find no remaining
original-scope enrollment blocker under the owner's closeout direction. Prior
R16/R17/MR52 source/security corrections stand; imported M1 is accepted as owner-
reported feedback, not independently witnessed per-step proof. The RC2 full
qualification independently accepted in ONEOFF-rc2-promotion-20261019.md supplies
121 distinct required checks PASS, no FAIL/UNMET. Supervisor compared core/web
from784be0d7 to current staging56749602: no content difference. No extra tests
or full run are justified solely to close the original enrollment task.

**ACCEPT original enrollment source, manual feedback and qualification.** This is
not yet final Goal closure: await the compact append-only Result and actual
resource cleanup/retention receipt. Preserve original failures, the imprecise
preview stop-time qualification and the explicit untested G060 transfer of
5f94a705. No claim that pending Group policy/UI is implemented. G055 remains
held for G060 and its own remaining gates after G058 closes.


## FINAL ACCEPT — original enrollment and closeout, 2026-10-19

**ACCEPT G058, complete.** Owner explicitly moves the Group authority/UI change
to G060 and closes G058 against its original enrollment requirements. Source,
security corrections, complete qualification and owner-reported M1 are accepted
as recorded above; no new code/full test campaign or per-step manual proof is
claimed. Original failed/canceled/UNMET outcomes remain historical.

Imported exact append-only Result from closeout37edda2a, SHA256
8fc5511f3ac2124184bb421360b3d8e9b126a5cc4424e0fcd9a5738876610eb9, and exact
coordinator cleanup receipt (not included), SHA256
ae107f71322734d34441277cdf1b0138db0233fff4a9b64ab23352f87a0afcf5. Supervisor
verified deleted worktrees absent, original planning/repair remote branches absent,
only transferred g058-group-admin-404 remains remotely, and no G058-named container
or image was present. Shared Go module volume retained as shared cache, not a
task-private leftover. No blanket host/builder-cache prune performed.

After importing, supervisor verified a small prerequisite-bound bundle of37edda2a
and deleted g058-closeout local/remote; absence rechecked. This supersedes the raw
receipt's pending-import retention, without rewriting its original bytes. Previous
11-ref history bundle independently passes git bundle verify. Retained private
history bundle is16,865,510 bytes, separate from35,969 logical bytes of toolkit/M1
files (coordinator du80KiB allocation); do not describe the entire retained
footprint as80KB. Unique history release condition: G060 reuse/history reconciliation
at final ACCEPT or explicit owner disposal. Tooling/M1 provenance and untested
5f94a705 branch/worktree transfer to G060, with no G058 runtime resources held.

Closeout Result's RC1 limiter row describes its historical deferred disposition:
the limiter follow-up is now separately FINAL ACCEPT, MR56/3eec8f29, and included
in the manual preview/RC2 source. It is not an unresolved enrollment blocker.
No fix/test result is retroactively attributed to the original deferral.

Current backend Group authority remains broader until G060; UI404 is not proof of
content isolation. G060 final ACCEPT now gates G055 completion and G021/SaaStacular.
G058 closure does not release G055 merge, G020 planning, OS001, an RC or final tag.
PROGRESS closeout assessment: retain one concise current G058 entry and acceptance
link; historical anchors remain needed by active G060/G055 reconciliation, so
further archival is deferred to the next safe G060/EPIC-007 closeout checkpoint.
EPIC-007 remains open for G060; no epic completion inferred.
