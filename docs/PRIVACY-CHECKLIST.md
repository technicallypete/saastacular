# Privacy checklist

Two lists: what to resolve before an application built on this repository processes real personal data, and what to
review before you make a fork, a screenshot set or a deployment public. Neither list is legal advice, and nothing in
this repository claims GDPR applicability, GDPR or SOC 2 compliance, or production readiness. You must resolve your
actual requirements before real-data use.

## What the repository ships, and what it does not

**Shipped mechanisms** (see [SECURITY.md](SECURITY.md) and the foundation guides): name-only secret resolution and
submitted-source secret scanning; least-privilege database roles with forced row-level security; tenant-scoped,
append-only audit events that store no names, e-mail addresses, tokens or payloads; structured logs that redact
secrets and never log credentials, hostnames in clear, labels or signup secrets ([OPERATIONS.md](foundation/OPERATIONS.md));
tenant-aware object storage with short-lived signed URLs; backup, isolated restore with session invalidation, bounded
retention by a dedicated retention login, backup expiry and quarantine ([RECOVERY.md](foundation/RECOVERY.md)); vendor
and data-handling records ([VENDORS.md](foundation/VENDORS.md)).

**Not shipped:** a consent or subject-rights portal, legal templates, privacy notices, a data map for your
application, retention periods or legal holds chosen for you, production key management, or processor agreements.
Those are adopter work.

## Before processing real personal data

- [ ] **Personal-data inventory and purposes.** List the personal data your application and its providers process
  (including identity-provider profiles, audit rows, logs, object storage and backups) and the purpose of each.
- [ ] **Jurisdiction and role.** Determine which data-protection laws apply and whether you act as controller or
  processor for each processing activity.
- [ ] **Subject-rights procedures.** Define access, correction, deletion, portability and objection procedures for
  your application's data, including what the append-only audit trail and backups mean for deletion requests.
- [ ] **Retention and holds.** Choose lawful retention periods and legal-hold rules; the retention tooling enforces a
  policy you declare and deletes only rows it makes eligible ([RECOVERY.md](foundation/RECOVERY.md)).
- [ ] **Backups and restoration.** Decide backup location, encryption, access control, immutability, cadence, RPO/RTO
  and expiry; a backup holds recoverable PKCE verifiers and CSRF values for unexpired rows, identifiers, names and
  audit rows ([SECURITY.md](SECURITY.md#9-backup-material-and-responsibilities)). A restore must not revive sessions.
- [ ] **Processors and transfers.** Put agreements in place with the identity provider, hosting, DNS, object-storage
  and any AI providers, and review subprocessors and international transfers.
- [ ] **Incident ownership.** Name who triages, contains and notifies; adapt the incident runbook in
  [RECOVERY.md](foundation/RECOVERY.md) and the compromised-client runbook in [OPERATIONS.md](foundation/OPERATIONS.md).
- [ ] **AI providers.** If you enable the intelligence service, review the model provider's data-use, training and
  retention terms before sending any tenant content.
- [ ] **Production prerequisites.** Close the transport and at-rest prerequisites the readiness evaluator reports
  ([DEPLOYMENT.md](DEPLOYMENT.md#current-readiness)).

## Before making a fork or deployment public

Review every item; remove or replace anything that identifies your organization's private systems or people.

- [ ] **Identity values.** `.env.example` and any rendered configuration hold only neutral or intended public values:
  product name, slug, database name, bucket name and domains. The local `.env` (generated secrets, mode 600) is never
  committed; `.gitignore` excludes it and the secret scan fails if it is submitted.
- [ ] **Domains and hostnames.** Example domains in docs, fixtures and Terraform examples are placeholders you intend
  to publish (for example `example.com`, `example.net`), not your private zones, preview hosts or customer domains.
- [ ] **Accounts and project identifiers.** Remove cloud project IDs, regions and zones you consider private, account
  and service-account e-mail addresses, CI project IDs or paths, runner names, registry paths, state-bucket names and
  Cloudflare zone details from committed files. The CI adapters' approved-project pins must name your project only
  where you intend to publish that binding.
- [ ] **Tokens and secrets.** No API keys, tokens, passwords, private keys, certificates or OIDC credentials anywhere
  in the tree or history. Run the secret scan (`bash infra/security/secret-scan.sh --env-file .env secret-scan` in a
  Git repository) and review its exemptions; rotate anything that was ever committed.
- [ ] **Logs and evidence.** Do not publish validation reports, CI evidence directories, diagnostics, database dumps,
  backups or recovery receipts; they can hold identifiers, hostnames and run-specific material. Keep them in the
  ignored `.local/` directory or outside the repository.
- [ ] **Screenshots and recordings.** Use synthetic tenants, users and domains; check browser chrome, address bars,
  e-mail addresses, Organization names, logos and notification content before publishing.
- [ ] **History and metadata.** Inspect Git history, tags, commit authors and e-mail addresses, issue and merge-request
  references, and any generated manifests for private identifiers before pushing to a public remote.
- [ ] **Agent and workflow state.** If you use the bundled development workflow, do not publish agent session
  files, local settings, inboxes or private management records unless you intend to ([DEVELOPMENT-WORKFLOW.md](DEVELOPMENT-WORKFLOW.md)).
- [ ] **Third-party notices.** Keep [LICENSE](../LICENSE), [NOTICE](../NOTICE), [THIRD_PARTY_NOTICES.md](../THIRD_PARTY_NOTICES.md)
  and the license files under `web/` and `distributions/skills/` with the material they cover; do not ship brand assets you are not
  licensed to use.
