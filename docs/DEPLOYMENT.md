# Deployment

This repository runs locally with Docker Compose and ships optional, provider-specific staging tooling. It does
**not** ship a production deployment, production infrastructure, production credentials or hosted deployment
evidence. Read [SECURITY.md](SECURITY.md) and [PRIVACY-CHECKLIST.md](PRIVACY-CHECKLIST.md) before deploying anything
that handles real data.

## What local Compose is, and is not

`infra/compose/local.yml` is for local development and foundation testing. It binds to loopback, keeps PostgreSQL,
Redis and MinIO in local containers and uses the `loopback-http` cookie profile. It is not evidence of a production
topology, and nothing about it states that a volume or link is encrypted.

## Current readiness

The configuration-only readiness evaluator (`core/cmd/readiness`, [SECURITY.md](SECURITY.md#7-transport-and-at-rest-readiness))
currently reports **blocked**. The named prerequisites are: verified TLS from core and migrate to PostgreSQL and from
core to Redis (or removing Redis), TLS on the internal hops (web to core, core to intelligence) and at the browser edge,
at-rest evidence for PostgreSQL, object storage and backup sets, and an operator entry point for user disablement and
membership removal. Findings F-1 to F-8 in SECURITY.md stay disclosed.

The accepted foundation capabilities (identity and sessions, tenancy, shared authorization, row-level security, audit,
storage authorization, recovery tooling, MCP connections, domains) are implemented and covered by the repository's
checks. What is unqualified is the deployment around them: provider configuration, network and TLS topology, secret
management, backup storage and the live identity-provider and custom-domain behaviour, none of which has been executed
for this repository.

## Choosing a production direction

These are recommendations for adopters, not bundled capabilities or a mandated cloud.

- **Provisioning and orchestration are separate concerns.** Terraform can provision cloud resources, including
  Kubernetes infrastructure, but adopting Terraform does not require Kubernetes.
- **Starting points to evaluate:** Terraform with AWS ECS, comparing Fargate and EC2 after a workload and cost
  comparison; or GCP Cloud Run where its HTTP and streaming, background-worker, networking and scaling requirements
  fit the services.
- **Managed data services:** prefer managed PostgreSQL, Redis and object storage outside the application containers,
  subject to your requirements and to preserving the service contracts: the role separation provisioned by
  `infra/postgres/provision-roles.sql` (managed PostgreSQL runs no init scripts), forced row-level security, a scoped
  object-store credential rather than root credentials, and Redis as ephemeral state only.
- **Kubernetes later, if needed:** adopt it for demonstrated scheduling, operator or environment requirements, not
  merely because the template has several services or tenant subdomains.
- **Cost:** compare supporting-service and operating costs as well as compute. Neither provider is promised to be
  cheaper; measure for your workload.
- **Qualify before production:** close the readiness prerequisites above, qualify your identity-provider configuration
  (WorkOS environments, redirect and wildcard behaviour on your domains), DNS and TLS for platform and custom domains,
  backup location, encryption, immutability, cadence, RPO/RTO and retention periods.

## Included staging tooling (provider-specific)

The repository includes reusable tooling for one on-demand staging environment, documented in
[STAGING.md](foundation/STAGING.md):

- `infra/terraform/`: Terraform modules for a Google Cloud Compute Engine VM with a retained data disk, static
  address and dedicated network (`modules/gcp-staging-vm`), and Cloudflare DNS records (`modules/cloudflare-dns`),
  with an example environment and bootstrap scripts for the deployer and CI federation identities.
- `infra/staging/`: the operator lifecycle command (`staging.sh`), remote scripts, CI adapters, examples and offline
  self-tests; `infra/compose/staging.yml` with `infra/edge/` (a Caddy edge with Let's Encrypt certificates in front
  of the existing services; only the edge publishes host ports); `infra/ci/staging-action.sh` and `infra/docker/staging-ci.Dockerfile`.
- `.gitlab-ci.yml`: manual `staging:prepare`, `staging:check`, `staging:build`, `staging:deploy` and `staging:stop`
  jobs, admitted only for web or API pipelines on the protected integration branch.

It is a bounded synthetic profile, not production: the development loader with private local PostgreSQL, Redis and
MinIO on one VM, real public DNS and HTTPS, plaintext internal hops on the VM's private Docker networks, and no
externally usable signed downloads. With identity secrets empty, a staging deployment is infrastructure smoke only;
authentication is unavailable, never bypassed.

**Truthful status.** This tooling is provider-specific (Google Cloud, Cloudflare, Let's Encrypt and GitLab CI with
OIDC federation). It requires adopter setup: your own cloud project with billing, region and zone, a Cloudflare zone and
zone-scoped tokens, an ACME contact address, a Terraform state bucket, deployer and CI federation identities, protected
CI environments and variables, and re-pinning of the project-specific values in `infra/ci/hosted-context.sh` and the
staging configuration. Every project, region, zone, domain, account and token is a configuration input; the examples use
placeholders such as `example.net` and `PROJECT_ID`. **It has not been exercised against any provider for this
repository**: no provisioning, certificate issuance, lifecycle, sizing or cost result is recorded or claimed here. Its
offline self-tests (`infra/staging/tests/`) and the harness's static checks cover the scripts without contacting any
cloud. Costs in STAGING.md are marked unverified.

Running it creates billable cloud resources and DNS records under your accounts. Review [SECURITY.md](SECURITY.md)
section 7 ("Staging profile constraints") and STAGING.md's secrets, retention and retirement sections first.

## CI provider

Hosted CI is GitLab-specific; see [OPERATIONS.md](foundation/OPERATIONS.md#ci). For GitHub Actions, see
[CI-GITHUB-ACTIONS-PARITY.md](CI-GITHUB-ACTIONS-PARITY.md). Local commands need no GitLab account, token, remote or
hosted runner.
