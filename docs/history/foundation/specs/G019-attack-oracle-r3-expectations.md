> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G019 attack oracle r3: source-derived expectations

2026-10-18 UTC. Bounded completion of accepted G019 Spec R1–R4 and Plan§6.2
TestIsolationAttacks extension, interpreted with accepted r2/r3/r4 addenda.
Independent source-only reviewer read no implementation, executable tests, Results
or runtime output. These expectations do not imply observed execution. Preserve
existing oracle history and independent literal expectations.

## Required G019 extension and mapped proof

Use existing two-Organization fixtures and paired legitimate controls. Add bounded
attack oracle/driver rows for direct binding writes, invitation tampering and
Organization consumer forgery without valid token/proof. Existing attributable
RLS/CT-MEM/MEM coverage may supply repeated invariant proof where explicitly mapped;
do not create duplicate suites for the same boundary.

| Attack | Expected outcome and state |
| --- | --- |
| Arbitrary ordinary runtime RoleBinding INSERT | 42501, no new binding; bootstrap/claim grants no general INSERT authority |
| Admin/member owner transition; Organization manager touching Group owner without that Group owner role | Denied, original role and owner count unchanged |
| Ineligible owner promotion, sole-owner removal, dependent-binding Organization departure | Named eligibility/owner/dependency refusal where guard reached, no partial write, at least one assigned owner |
| Binding tenant/subject/scope/source change; foreign tenant or selected sibling target | Denied, original bindings/immutable keys unchanged |
| Ordinary runtime reads invitation token_digest or updates consumption/recipient columns | 42501 even for issuer/intended recipient; digest not returned, invitation pending and no membership effect |
| Invitation intent/digest/expiry tampering, owner-role issue, consumed-intent revocation | Grant/immutability/CHECK refusal, original intent/state unchanged |
| Ordinary runtime calls Organization claim, even with correct token and chosen subject | 42501 EXECUTE denial; Organization claim owner-only in G019, no consumption or membership |
| Authorized synthetic Organization consumer with wrong/missing token, kind or subject/address/principal mismatch | Closed not_found or pre-SQL invalid-evidence refusal, no admission/consumption |
| Group token replay/wrong subject/tenant/kind/context; recipient/issuer loses eligibility | Closed refusal with applicable named cause; no false consumption/grant |
| Consumed-token reuse after removal; arbitrary invitation setting or forged request recipient | No restored membership or new authority |

Pair refusal cases with same-scope authorized administration, eligible owner
promotion and valid intended claims; blanket denial does not prove correctness.
Raw self-admin demotion follows accepted DB semantics; facade self_target remains
separate. Selected-Group sibling mutations are forbidden under SpecR1, as already
clarified in G019-rls-oracle-r4-expectations.md; Plan shorthand grants no exception.

## Attribution and atomicity

Failed attacks preserve membership/invitation state and create no success capture.
Facade denials carry truthful access.denied where recordable; raw SQL failures do
not invent facade audit. Successful consumption commits membership and correctly
attributed invitation.consumed together, preserving equal/higher roles. Audit
failure/class40 abort rolls back ordinary mutations. Duplicate consumption has
one effect. No token/digest leakage into answers, logs or audit. Direct claim-writer
audit append is denied; legitimate nested capture is allowed, never based on depth
alone. Map previously demonstrated atomicity/concurrency evidence where sufficient.
