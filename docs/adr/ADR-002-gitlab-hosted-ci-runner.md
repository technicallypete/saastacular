# ADR-002 — Job-isolated GitLab hosted Docker-in-Docker for foundation CI

> **Editorial note (public export).** This historical decision record is retained for its reusable architecture and security reasoning. Links to private planning, review and evidence records that are not part of this repository are marked "(not included)"; references to "Architecture §N" or ARCHITECTURE.md point to the source project's architecture document, which is not included. Goal IDs, dates, approvals and superseded history are kept as written; product-specific passages are redacted where marked "[Editorial redaction: …]". Current behaviour is documented in the foundation guides under [`docs/foundation/`](../foundation/).

Date: 2026-10-06
Status: ACCEPTED by product owner, 2026-10-06
Decision owner: Product owner
Recorded by: Codex
Scope: bounded G003 CI revision; no production deployment

## Context

G003's existing-runner assumption is unresolved. The owner rejects a $138/month dedicated runner and requests lower-cost hosted CI and shorter routine feedback. Read-only API inspection confirms project hosted runners enabled and online small/medium AMD64 entries. This does not prove quota, execution compatibility or privilege safety.

GitLab documents privileged Docker-in-Docker support for hosted Linux runners. G003 requires explicit review for privilege or runner security-model changes. The product owner accepted the reviewed hosted-runner boundary on 2026-10-06; implementation and hosted actions remain subject to the execution gates below.

## Accepted decision

Use the existing GitLab.com hosted offering with a fresh job-isolated DinD daemon. Start qualification and all fast/integration/full jobs on `saas-linux-medium-amd64`. Small is an optional later optimization, not an additional qualification prerequisite. Pin job/daemon/tool images in Claude's reviewed Plan. Accept runner-configured privilege for the outer GitLab job container and the DinD service within the fresh, isolated, ephemeral job VM. Nested validation, application and test containers remain unprivileged; the nested validation context uses the reviewed no-new-privileges option. This does not authorize privilege on a developer/production host or a persistent/shared runner.

No developer/production Docker socket, persistent/shared daemon, public unauthenticated Docker API, deployment credentials, cloud provisioning or paid runner subscription. Use the reviewed job-private mutual-TLS daemon transport. Verify actual executor/isolation, daemon transport, shared checkout paths, loopback port namespace and bounded cleanup/cancellation during qualification. Public runner documentation and API descriptions are not substitutes for observed compatibility. Failure of the approved transport/isolation model stops execution for review; an unauthenticated fallback is not approved by this decision.

Fast/affected/full gating is defined in the linked G003 revision Spec (not included). Full coverage stays mandatory for foundation changes, exact release candidates and weekly diagnostics. No production runtime platform is selected.

## Consequences and alternatives

Medium has more CPU/memory/disk and consumes compute minutes at factor 2 versus small's factor 1. Measure duration/resource/quota rather than assuming lower elapsed time means lower total cost. A separate local VM runner is an alternative if hosted qualification fails; it requires its own reviewed profile. A dedicated paid runner, new vendor or substantial orchestration is outside this proposal.

Approval of this ADR selects the security/profile direction only. Claude Plan acceptance, exact input binding and named external-action authorization remain required before implementation or pipeline execution. Purchases, project-setting changes and schedules require separate authorization. Historical bound G003 inputs and evidence remain unchanged; the revision must bind this accepted ADR explicitly.

## Decision history

- 2026-10-06 — Initial proposal limited privilege to the DinD service. Plan r2 review identified that this did not describe runner-wide privilege on hosted Linux runners.
- 2026-10-06 — Product owner accepted Codex's recommendation ("Sounds good to me"). This supersedes the proposed sentence "Accept privilege only for the DinD service" with the isolated outer-job/DinD boundary above. Nested container restrictions, medium profile, no production/developer sockets or secrets, qualification and execution gates remain.

Reference: [GitLab hosted Linux runner isolation and privileged DinD support](https://docs.gitlab.com/ci/runners/hosted_runners/linux/).
