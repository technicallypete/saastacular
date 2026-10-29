# syntax=docker/dockerfile:1
# G070 staging edge: the official Caddy image rebuilt with xcaddy to add the pinned Cloudflare DNS provider module
# (DNS-01 for B and *.B), as the image documentation shows. Built off the VM in the single build slot; context
# infra/docker, from which nothing is copied. The configuration is infra/edge/Caddyfile, mounted read-only by
# infra/compose/staging.yml; the provenance record is infra/edge/README.md. The final image is the Alpine-based
# official runtime (busybox wget, no curl); its default CMD runs /etc/caddy/Caddyfile.
FROM caddy:2.11.7-builder@sha256:c2fc374fd352439d76e1777b28d97d9f6d9c1dd86116db9227460374d46097d5 AS builder
# The builder's own Go toolchain only (no automatic toolchain download), so the build is attributable to its digest.
ENV GOTOOLCHAIN=local
RUN xcaddy build v2.11.7 --output /usr/bin/caddy --with github.com/caddy-dns/cloudflare@v0.2.4

FROM caddy:2.11.7@sha256:504a44d7d74bd2bc28a66afc4774912fc21223ca8a67734af1c54ab2f999b489
ARG VCS_REF=unknown
COPY --from=builder /usr/bin/caddy /usr/bin/caddy
LABEL org.opencontainers.image.title="staging edge" \
      org.opencontainers.image.version="caddy v2.11.7; github.com/caddy-dns/cloudflare v0.2.4" \
      org.opencontainers.image.revision="${VCS_REF}"
