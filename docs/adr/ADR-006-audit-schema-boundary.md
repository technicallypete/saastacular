# ADR-006 — Dedicated PostgreSQL audit schema

> **Editorial note (public export).** This historical decision record is retained for its reusable architecture and security reasoning. Links to private planning, review and evidence records that are not part of this repository are marked "(not included)"; references to "Architecture §N" or ARCHITECTURE.md point to the source project's architecture document, which is not included. Goal IDs, dates, approvals and superseded history are kept as written; product-specific passages are redacted where marked "[Editorial redaction: …]". Current behaviour is documented in the foundation guides under [`docs/foundation/`](../foundation/).

Status: ACCEPTED by product owner, 2026-10-14 UTC
Authority: owner approved the recommendation to place G011 durable audit records
in a separate `audit` schema within the existing PostgreSQL database.

## Decision

G011 introduces the `audit` schema for durable audit records and their associated
audit-event tables. Keep the existing database and application schemas; this is
logical separation with explicit privileges, not a new database or service.

- Separate schema/table ownership from ordinary runtime roles. Runtime roles have
  no audit DDL, UPDATE, DELETE or TRUNCATE privileges. Writes use a controlled
  append-only path with trusted actor/scope/resource/outcome attribution.
- Tenant audit reads require current authorization and appropriate RLS. Schema
  placement alone grants no access. Use canonical UUID identifiers and preserve
  Organization/Group scope; no independent profile grants or slug-based identity.
- Retention/deletion uses a separately restricted maintenance path. G011 defines
  that boundary; G013 owns the later retention/recovery behavior. Privileged
  database administrators remain outside any claimed runtime immutability boundary.
- G011's Plan pins ownership/grants, writer/read interfaces and transaction/error
  semantics, including how denied or rolled-back operations remain attributable.
  Do not claim that recording every event in a rolled-back transaction preserves it.

## Verification and transition

G011 must test through actual runtime roles: authorized append/read positives;
refused update/delete/truncate/DDL and cross-tenant reads; trusted attribution,
required-event completeness and secret/payload redaction. Identify any authorized
maintenance exception explicitly; no blanket runtime mutation privilege.

The decision does not put every operational log, delivery queue or business event
in the audit schema. Those stores remain with their owning Goal. It creates no
tables/migration now, changes no G008/G009 implementation or approved inputs, and
authorizes no G011 implementation. G010 dependency, Goal/Spec/Plan review and the
G021 product-feature barrier remain.

## Owner clarification — security-reducing revocation, 2026-10-14

When revocation itself can commit but audit append fails, the owner selects:
“Complete revocation; attempt bounded failure recording and report any audit gap.”
Local logout/current-all/provider-logout revocation and user disablement must not
be prevented solely by an audit fault. Use bounded independent recording and
sanitized gap diagnostics; do not claim a missing event persisted. This is an
explicit exception to ordinary atomic mutation/audit success, not an exception
for sign-in, enablement, permission grants or other mutations. Revocation database
failure/unknown commit remains truthful. Preserve append-only protection and
runtime least privilege; no general trigger-disable/bypass grant. G011's corrected
Plan must pin and test the mechanism before implementation authorization.
