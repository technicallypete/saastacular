# CI context image (G003 CI revision, Plan r2 §3.1): the Docker CLI with its compose and buildx plugins on
# a package-manager-free base that provides bash, git, GNU tools, procps and hostname. It runs unprivileged
# beside the job-private DinD daemon; it is built on stdin with no build context.
FROM docker:29.8.2-cli@sha256:b1805116a6a86cc591b5d5f60a910a0715cdcc9d18d866ad68b1457ead25c35c AS dockercli

FROM buildpack-deps:trixie-scm@sha256:ec9476a31fe1d3d74057675bc0bc1672b21ce727315f42ca241643a1c6050f25
COPY --from=dockercli /usr/local/bin/docker /usr/local/bin/docker
COPY --from=dockercli /usr/local/libexec/docker/cli-plugins/ /usr/local/libexec/docker/cli-plugins/
