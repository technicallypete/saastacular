# Staging edge (G070)

Caddy terminates public TLS for an on-demand staging environment and passes requests to web and core on the private
`edge` network. Two files, both generic (every deployment value is an environment input; no domain, token or account
literal):

- `infra/docker/edge.Dockerfile` — the official `caddy:2.11.7` runtime with the binary rebuilt by xcaddy from the
  `caddy:2.11.7-builder` image to add `github.com/caddy-dns/cloudflare@v0.2.4`. Built off the VM, one build at a time,
  context `infra/docker`; `--build-arg VCS_REF=<candidate sha>` sets the revision label (default `unknown`).
- `infra/edge/Caddyfile` — mounted read-only at `/etc/caddy/Caddyfile` by `infra/compose/staging.yml`.

The operator runbook is [STAGING.md](../../docs/foundation/STAGING.md); the host model and edge inputs are in
[DOMAINS.md](../../docs/foundation/DOMAINS.md#edge-deployment-inputs-e1-caddy-on-demand-tls).

## Inputs

| Variable | Required | Meaning |
|---|---|---|
| `EDGE_BASE_DOMAIN` | yes | `B` (example `staging.example.net`); the site is `B, *.B` |
| `ACME_EMAIL` | yes | ACME account contact |
| `EDGE_ACME_CA` | no | ACME directory; unset = Let's Encrypt production, `https://acme-staging-v02.api.letsencrypt.org/directory` for rehearsals (use a separate data volume) |
| `EDGE_HSTS_MAX_AGE` | no | `Strict-Transport-Security` max-age in seconds on `B, *.B`; unset = 300; never `includeSubDomains`/`preload` |
| `CF_API_TOKEN` | yes (secret) | Cloudflare token scoped to the one zone, `Zone:Read` + `DNS:Edit`; read at runtime through `{env.CF_API_TOKEN}` |

An empty (rather than unset) `EDGE_ACME_CA` or `EDGE_HSTS_MAX_AGE` is a configuration error: pass a value or leave it out.

## Behaviour and trust notes

- **Certificates.** `B` and `*.B` share one DNS-01 certificate obtained at start (Caddy writes and removes
  `_acme-challenge` TXT records; Terraform never manages that name). Any other name is a custom host: the catch-all
  `https://` site is the only on-demand policy, and Caddy asks `http://core:8180/internal/v1/domains/tls-ask` first;
  only a 200 permits issuance. Setting `EDGE_ACME_CA` (or its default) selects exactly that CA; record the observed
  issuer. ACME account, keys and certificates live in the `/data` volume, which only retirement deletes.
- **Routing.** `Host` passes unchanged; `/mcp/c/*` and `/.well-known/oauth-protected-resource/mcp/c/*` go to
  `core:8180`, everything else to `web:3000` (as `infra/identity/proxy.conf`). Core's `/internal/*` routes are never
  proxied to core: web answers them, and unknown labels, with its fixed 404. `X-Forwarded-For/Proto/Host` are removed
  on the way up (nothing reads them); no `trusted_proxies`. HTTP redirects to HTTPS; HTTP/3 is off (the firewall
  opens only tcp 80/443).
- **Logs.** Access and runtime logs are JSON on stdout with the whole query replaced by `?REDACTED` on every path and
  `Cookie`, `Authorization`, `Referer`, `Set-Cookie` and `Location` dropped, so callback and signup-continuation values
  (`code`, `state`, `error`, `error_description`, `ref`, `n`, `p`) never reach a log; `log_credentials` is unset.
- **Secrets.** The admin API is off. The token is never a literal, so `caddy adapt` output and the autosaved config
  hold only the placeholder; the container's environment still holds it (never capture `docker inspect` without a
  `--format` that excludes `.Config.Env`, nor `docker compose config`). A malformed token can be echoed by the DNS
  module's validation error: check the token before deploying, not by reading edge logs into evidence.
- **Process.** The container runs as the official image's default user (root inside the container) so the rebuilt
  binary binds 80/443 without file capabilities; nothing else is added to the official runtime.

## Provenance record (filled per build; published in the Result with the deploy manifest)

| Field | Value |
|---|---|
| Candidate commit | `<sha>` |
| `infra/docker/edge.Dockerfile` sha256 | `<sha256>` |
| `infra/edge/Caddyfile` sha256 | `<sha256>` |
| Builder image | `caddy:2.11.7-builder@sha256:<digest>` |
| Base image | `caddy:2.11.7@sha256:<digest>` |
| Plugin | `github.com/caddy-dns/cloudflare v0.2.4` |
| `caddy version` | `<output>` |
| Go version (`go version -m /usr/bin/caddy`, first line) | `<go1.x.y>` |
| `caddy list-modules` includes `dns.providers.cloudflare` | `<yes + module count>` |
| Image ID / registry digest | `<sha256:…>` / `<registry>/edge@sha256:<digest>` |
| Build host, date (UTC) | `<host>`, `<yyyy-mm-ddThh:mm:ssZ>` |

Rebuilding the same Dockerfile at the same digests resolves the same module versions (Go checksums), but the binary is
not guaranteed bit-identical across builds; the record above, not a rebuild, identifies what runs.
