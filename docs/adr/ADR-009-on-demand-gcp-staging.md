# ADR-009 — Reusable on-demand GCP staging

> **Editorial note (public export).** This historical decision record is retained for its reusable architecture and security reasoning. Links to private planning, review and evidence records that are not part of this repository are marked "(not included)"; references to "Architecture §N" or ARCHITECTURE.md point to the source project's architecture document, which is not included. Goal IDs, dates, approvals and superseded history are kept as written; product-specific passages are redacted where marked "[Editorial redaction: …]". Current behaviour is documented in the foundation guides under [`docs/foundation/`](../foundation/).

Date: 2026-10-26 UTC
Status: ACCEPTED architecture direction; implementation/live deployment not yet released
Decision owner: Product owner; recorded by Codex
Owning delivery: G070 (not included) /
Spec (not included)

## Context

The owner wants staging available for testing sessions without an always-running
managed database or application VM. G057 Part A is merged; its remaining real
WorkOS/domain qualification needs public HTTPS. The owner selects GCP because of
its preferred project/IAM workflow, Terraform provisioning, Cloudflare DNS and
Caddy/Let's Encrypt, with a reusable production extension later.

## Decision

1. Deliver generic SaaStacular deployment configuration: one regular Compute
   Engine VM runs the existing Compose services with local seeded PostgreSQL,
   Redis and MinIO. Terraform owns infrastructure; narrow lifecycle commands
   deploy/start/stop/reset/retire the same environment. Spot is not the initial
   choice. No Kubernetes or new platform scheduler.
2. Domain, project/region, identities, resource names, product identity and secret
   references are inputs. Use a base `B`, `auth.B`, `custom.B` and `*.B`. The
   Beginify adoption supplies its owner-selected domain privately; reusable code
   and examples use generic configuration. No per-tenant DNS/cert allocation.
3. Use Caddy-managed trusted Let's Encrypt certificates for B and *.B through
   DNS-01 with the pinned Cloudflare DNS module. Retain certificate/account
   storage across normal stops. Core's exact-host/tenant/issuance authorization
   remains authoritative; no open on-demand certificate endpoint.
4. Stop preserves the synthetic database/object bytes and certificate state;
   reset/reseed and full retirement are separate explicit operations. Record
   residual disk/IP/state costs, scoped ownership and release conditions. Native
   session run-time expiry and crash/partial-allocation reconciliation apply.
5. G070 owns reusable infrastructure/lifecycle. G057 Part B consumes that exact
   environment for real-provider/custom-domain proof, retaining its privacy and
   operator gates. Defer that execution until readiness; no qualification waiver
   or claim that infrastructure smoke proves real login.
6. Document the production extension: always-on application compute, managed
   PostgreSQL such as Cloud SQL, and private R2. Actual provider compatibility,
   backup/recovery, monitoring and availability qualification remain future work.
   Production services/accounts are not provisioned through G070.

## Consequences and authority

This dated staging choice supersedes the earlier absence of a selected provider
for this environment in Architecture §62. Its managed-container production
comparison remains historical/future context, not a requirement to port the app
to Cloud Run now. Existing role/RLS/storage isolation and foundation/G021 barriers
remain. G070 planning overlaps G069 under one coordinator and disjoint ownership.
Code requires Plan review/release; cloud/DNS/WorkOS writes require an exact later
account/resource/cost release. No purchase or real-provider mutation is authorized
by architecture publication. G057 frozen local evidence and M1 stay protected.

## References

- G057 (not included).
- [Caddy automatic HTTPS](https://caddyserver.com/docs/automatic-https).
- [GCP VM lifecycle](https://docs.cloud.google.com/compute/docs/instances/stop-start-instance).
- [Object-storage production limits](../foundation/OBJECT-STORAGE.md#production-configuration-cloudflare-r2-notes-not-qualified).

## Staging configuration refinement — 2026-10-26 UTC

Source inspection during G070 Plan review (not included)
found the current production loader incompatible with the selected private HTTP
MinIO endpoint. The Spec r1 addendum (not included)
permits a synthetic staging profile using the development loader with real public
HTTPS, explicit secure cookies, private backend ports and the real WorkOS adapter;
test authentication/fixtures remain forbidden. An unconfigured identity pilot is
infrastructure smoke only. Production readiness and G057's provider/privacy proof
remain separate and outstanding. No public MinIO or broader production TLS project
is added. Private Docker runtime secrets may persist on the retained encrypted
disk and must be accounted for; tmpfs delivery alone is not erasure at stop.
This implementation refinement preserves the accepted provider/lifecycle choice
and grants no implementation or live-action release.
