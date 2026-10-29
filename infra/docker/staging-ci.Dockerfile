# Staging CI context image (G072 Plan r0 §2.2, A7): the context of the credential-bearing staging:check, staging:deploy
# and staging:stop jobs (infra/ci/staging-action.sh). google-cloud-cli (gcloud, bash, git, tar, coreutils, OpenSSH
# client) with the Docker CLI and its plugins of the pinned stage that infra/docker/ci-context.Dockerfile uses, and the
# pinned jq 1.7.1 binary (the same image infra/staging/check.sh pins). Built on stdin with no build context; it holds
# no credential, configuration or repository file: the launcher mounts the build directory at run time.
FROM docker:29.8.2-cli@sha256:b1805116a6a86cc591b5d5f60a910a0715cdcc9d18d866ad68b1457ead25c35c AS dockercli

FROM ghcr.io/jqlang/jq:1.7.1@sha256:096b83865ad59b5b02841f103f83f45c51318394331bf1995e187ea3be937432 AS jq

FROM gcr.io/google.com/cloudsdktool/google-cloud-cli:590.0.1-slim@sha256:2a5368c4f3e25f7e202ccbd7b63e4128380d8fd683266f7f015c5d0afcc8f3be
COPY --from=dockercli /usr/local/bin/docker /usr/local/bin/docker
COPY --from=dockercli /usr/local/libexec/docker/cli-plugins/ /usr/local/libexec/docker/cli-plugins/
COPY --from=jq /jq /usr/local/bin/jq
