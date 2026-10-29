# ADR-008 — Explicit Group authority and owner-controlled administration

> **Editorial note (public export).** This historical decision record is retained for its reusable architecture and security reasoning. Links to private planning, review and evidence records that are not part of this repository are marked "(not included)"; references to "Architecture §N" or ARCHITECTURE.md point to the source project's architecture document, which is not included. Goal IDs, dates, approvals and superseded history are kept as written; product-specific passages are redacted where marked "[Editorial redaction: …]". Current behaviour is documented in the foundation guides under [`docs/foundation/`](../foundation/).

Date: 2026-10-19
Status: ACCEPTED design by product owner; implementation/evidence pending G060.

## Decision and authority

Owner agreed to explicit Group membership, Organization owners/admins reading
Groups and Business Profiles, and an Organization-owner-only configurable override
for ordinary Group management. Codex clarified three separate boundaries, default
off, no implicit content/ownership access; owner responded “Yeah, this should just
be the next spec ... goal and a plan ... before O5-5 gets finished”. Owner then
explicitly made G060 a SaaStacular prerequisite and waived its manual review.

- Same-Organization owners/admins retain read-only ordinary Group/profile metadata.
- Ordinary Group management defaults to actual Group owner/admin membership.
- Only an Organization owner may enable/disable an Organization-wide management
  override applying to Organization owners and admins. It defaults false for new
  and existing Organizations; current-authority checks and audited writes required.
- Override grants bounded ordinary Group/member/invitation/profile administration,
  never synthetic membership, self-escalation, Group ownership powers or content.
- Group data objects/storage/content always require an actual authorized Group role,
  including for Organization owners/admins. Metadata, profile sharing or Group
  selection cannot confer this. Organization-scoped resources remain distinct.
- Explicit invitation-only Group joining and atomic creator-owner bootstrap remain.
  Combined Organization+Group invitation admission is a separate follow-up.

## Transition and consequences

G060 Spec (not included) defines
finite actions, validation and live revocation. Supersedes the unconditional
Organization Group/profile-management and Group-object privilege portions of
Architecture §53.7/storage policy and G019 R1 when implemented; preserves other
ADR-004/ownership/cardinality/security controls. Existing implementation still has
broader authority until G060 integrates. This is not a claim that UI404 enforces it.

G058 closes on its original enrollment evidence/M1 and cleanup; its paused option B
UI correction is withdrawn from G058 delivery and reconsidered under G060. No old
patch/test pass is inferred. G060 must finish before G055 completion and G021/
SaaStacular; it has no dependency on unfinished G055. No G020 planning, new RC,
main promotion, export resume or native-permission bypass is authorized.
