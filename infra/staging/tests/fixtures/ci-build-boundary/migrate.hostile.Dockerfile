# syntax=docker/dockerfile:1
# G072 EP-1 hostile candidate (infra/staging/tests/ci-build-boundary.sh). The test commits this file as the synthetic
# candidate's infra/docker/migrate.Dockerfile, a path the deployment-infrastructure compatibility contract (Plan S6)
# leaves candidate-controlled, so the real build phase builds it. Each RUN step copies what a candidate build container
# can see into /leak (kept in the image, so `docker export` shows it) and prints it (so the build log shows it): its
# environment, every visible process environment, the GitLab job paths /builds and /certs, its mounts, build secrets,
# the build arguments a careless adapter might pass, and the build context. The test requires these dumps to exist
# and no job canary to appear in them, in the build log or in the artifacts. Never built outside that test.
FROM alpine:3.22 AS runtime
ARG CI_JOB_TOKEN
ARG STAGING_CI_ID_TOKEN
ARG CLOUDFLARE_API_TOKEN
ARG CI_REPOSITORY_URL
RUN mkdir -p /leak && env | tee /leak/env.txt
RUN for f in /proc/[0-9]*/environ; do echo "== $f"; tr '\0' '\n' <"$f"; done 2>&1 | tee /leak/proc-environ.txt
RUN { cat /proc/self/mountinfo; ls -laR /builds /certs /run/secrets; find /builds /certs /run/secrets -type f -exec cat {} +; } 2>&1 | tee /leak/mounts.txt; true
RUN printf 'job=%s id=%s cf=%s repo=%s\n' "${CI_JOB_TOKEN:-unset}" "${STAGING_CI_ID_TOKEN:-unset}" "${CLOUDFLARE_API_TOKEN:-unset}" "${CI_REPOSITORY_URL:-unset}" | tee /leak/build-args.txt
COPY . /leak/context/
RUN find /leak/context -type f | sort | tee /leak/context-files.txt
USER nobody
ENTRYPOINT ["/bin/true"]
