# syntax=docker/dockerfile:1
# Migration runner image. Build context: core/. Target: runtime.
# The binary embeds core/migrations and connects as the migrator role, which
# assumes the owner role. The core server image contains no migration code.
FROM golang:1.27.3-alpine AS build
ENV CGO_ENABLED=0 GOTOOLCHAIN=local GOFLAGS=-trimpath
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -o /out/migrate ./cmd/migrate

FROM gcr.io/distroless/static-debian12:nonroot AS runtime
COPY --from=build /out/migrate /usr/local/bin/migrate
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/migrate"]
