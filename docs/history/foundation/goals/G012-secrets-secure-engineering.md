> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G012 — Secrets and secure engineering controls

Epic: EPIC-004 — Audit and compliance foundation  
Status: r0 + privacy inheritance amendment — defined for downstream planning; G011 implementation active
Owner: Codex Goal/Spec/review; Claude Code Plan/implementation/validation  
Date: 2026-10-14 UTC

## Objective

Make the existing foundation's credential handling and engineering security
controls assessable: safe secret resolution, least-privilege operational access,
credential withdrawal/rotation, dependency/security scans, and attributable
environment-specific encryption and vendor-readiness evidence. Reuse accepted
interfaces and CI; do not build a security platform or claim SOC 2 compliance.

## In Scope

- Harden the existing core secrets-provider boundary and document every current
  service's secret sources, consumers, failure and reload behavior.
- Exercise one disposable service-credential rotation/withdrawal; reconcile
  existing human/session/role deprovisioning and G011 audit behavior.
- Reuse secret/static checks; add vulnerability scanning for the three service
  dependency graphs and the minimum existing CI/report integration.
- Check encryption configuration/readiness with explicit local exceptions,
  unsupported production paths and externally owned at-rest responsibilities.
- Complete the existing security/vendor documentation and G013 handoff.
- Carry Architecture §56.1's privacy-relevant vendor/data-flow responsibilities
  in those same records, marking unresolved owner/legal decisions explicitly.
  Full subject-rights/product lifecycle implementation belongs to future G056.

## Out of Scope

Managed secrets-manager/KMS adoption, cloud accounts or infrastructure, runtime
secret-management APIs/UI, tenant-managed integrations, new roles or schemas,
mandatory MFA, new identity providers, automatic credential rotation daemons,
whole-stack TLS/mTLS deployment, penetration-testing infrastructure, general CI
redesign, dependency-update bots, bulk dependency modernization, compliance
certification, backup/retention/incident exercises (G013), and product features.

## Dependencies

- Final **G011 ACCEPT and committed integration** before implementation. Its
  approved Plan r2 is available, but its audit schema/API/fault behavior is not
  yet an accepted implementation. Reconcile actual delivered interfaces first.
- Accepted G001–G003 configuration/CI/privacy, G004 identity, G005–G008 roles,
  policy/RLS, G009 storage, G010 attack suite, G049–G053 workflow improvements;
  accepted ADR-002/003/004/006 and canonical architecture.
- [Work Spec](../specs/G012-secrets-secure-engineering-spec.md). No new runtime
  provider/vendor is selected. Production deployment remains unprovisioned.

## Completion Definition

Spec AC1–AC10 have attributable evidence; actual diff/security review, required
staging disposition, compact Result and scoped cleanup are accepted by Codex.
Production-readiness gaps remain explicit blockers to deployment, not falsely
completed encryption claims. G013 receives these contracts and responsibilities.
G021 remains closed.

## Execution Policy

**Architecture/security tier.** Claude writes a concise Plan; Codex accepts it
and the Git-first binding before implementation. Under standing workflow,
provisional planning becomes eligible after G011 integrates; pending G011 final
acceptance permits planning only. This specification request does not dispatch
a parallel planner now or authorize application/test/build work.

After release, implement, diagnose and correct autonomously within approved
scope. Keep scanner integration bounded to existing entry points and changed
paths. Reuse unchanged controls; no speculative harness fixes or qualification
campaign. No real credential/account/project-setting changes are authorized;
rehearsals use synthetic credentials and owned throwaway resources.

## Iteration Budget

**B5: maximum five shared goal-level execution attempts; checkpoint after
attempt 3**, per BUDGET_POLICY (not included). Inner failed-batch
fix/retest loops share the active attempt; review-ready submission or abandoned
strategy ends it. Forecast and feasibility belong in Claude's Plan. Escalate
scope growth instead of silently treating B5 as authority for infrastructure work.

Planning defaults to **xhigh**, implementation to **high**. Remind the operator
at dispatch; record actual setting as verified or unknown. No attempt is used
by specification/planning alone.

## Context / Scope Budget

One finite current-secret inventory, the existing resolver boundary, three
service dependency graphs, one credential rehearsal, and security/readiness
records. No new orchestration or generic policy engine. Bind facts to current
G011 integration before approval; recommend a split if a reliable bounded Plan
requires cloud deployment or whole-stack transport changes.

## Suggested Subagents / Delegation

Zero to two bounded delegates: read-only security review and, if useful,
scanner/configuration analysis. One implementer owns integration and overlapping
configuration/CI files. Delegates share B5; product/security decisions and final
acceptance remain with the owner/Codex respectively.

## Stop / Escalation Conditions

Stop affected work for canonical/ADR contradictions, new provider/vendor or paid
service, new privilege/role/schema, public contract changes, real-secret access
or rotation, destructive changes, unbounded cleanup, or proposed suppression of
security findings. Unresolved blocking vulnerabilities require explicit reviewed
disposition; do not make the job green through automatic fixes/exclusions.

Escalate new query/index/helper/locking costs under the performance policy and
record them in PERFORMANCE.md/RISKS.md as applicable. Stop on native permission
denial without bypass. Missing G011 acceptance, an unresolved Plan decision,
exhausted budget or inability to collect required evidence blocks execution or
completion. No G013 auto-specification is authorized by G012's definition.
