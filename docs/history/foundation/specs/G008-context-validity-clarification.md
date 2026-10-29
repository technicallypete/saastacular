> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G008 Spec clarification — Context validity and global self identity

2026-10-13 UTC. Applies to G008 Spec r3 R3/R4/R6 and AC3–AC5/AC8.
This is a clarification of the accepted G007/G008 boundary, not a new capability,
authority path, identity bootstrap or mutation permission.

## Tenant-selection validity precedes capability checks

After structural context validation, a selected Organization requires current
Organization membership. If a Group is selected, it must belong to that
Organization and be selectable under G007 Resolve: Organization owner/admin, or
a binding in that exact Group. An unknown, foreign or unselectable selection is
not a valid TenantContext. It denies every tenant-owned row and projection,
including Organization rows and the subject's own Organization membership/binding
rows. A capability such as organization.read cannot override invalid selection.
The capability-specific selected-Group restriction applies only after resolution
has established a valid context. Existing Organization-only access and independent
Organization owner/admin paths remain available in valid contexts.

## Structural rejection versus global self liveness

Malformed, zero, missing required fields, mixed illegal shapes and partial context
initialization retain their required protected-access rejection. A structurally
valid shape whose tenant selectors fail data-dependent resolution denies tenant
and shared-scope identity access, but does not conceal the established actor's
permitted global self identity/liveness row. Invalid tenant selection does not
turn a live authenticated actor into an unauthenticated actor; resolution retains
the accepted context_denied outcome. No foreign/co-member identity, tenant data,
authentication material or mutation authority follows from this self-read.

This distinction clarifies R3's broad no-data wording together with R6's global
self-read requirement. Existing self-column limits, authentication/lifecycle
contracts, subject establishment, enabled/session checks, exact mutation denials,
connection/transaction boundaries and trust limitations remain unchanged. A caller
cannot forge identity or select an authentication bootstrap by supplying a UUID.

## Required cases

- Group admin with Organization member status and no binding in the selected
  same-Organization Group: no tenant/projection rows; permitted global self row.
  This actor is not an Organization admin, which would be permitted to select
  an existing same-Organization Group.
- Fixed authenticated actor selects Organization A with a Group belonging to B:
  no tenant/projection or shared-scope identity rows; permitted global self row.
- Valid Organization-only and selectable-Group positives retain their existing
  permitted rows/actions. Structurally malformed/zero/partial/illegal contexts
  retain rejection and no protected mutation/data. All forbidden updates remain
  forbidden; context_denied must not be replaced by unauthenticated solely due to
  an invalid tenant selection.

Derive independent expected sets from this clarification, the original Spec,
ADR-004 and canonical AUTHORIZATION contract; do not read policy SQL, implementation
outputs or Plan predicates to manufacture those sets. No unrelated oracle cases
or allowed-operation matrix is revised.
